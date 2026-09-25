// Package api serves the platform's HTTP surface.
//
// At Phase 0 it is personal-tier only: a bearer token, a health check, and the
// capture endpoint the field kit posts to. v0.1 opens the same endpoint to the
// public behind phone verification.
package api

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Re-exported so handlers and tests speak the store's vocabulary without the
// package depending on a live database.
type (
	NewReport   = store.NewReport
	NewMedia    = store.NewMedia
	ReportLabel = store.ReportLabel
)

// Reports is the persistence the capture endpoint needs.
type Reports interface {
	SaveReport(ctx context.Context, in NewReport) (id string, created bool, err error)
	AddReportMedia(ctx context.Context, in NewMedia) error
	SaveReportLabel(ctx context.Context, in ReportLabel) error
}

// Media stores the photographs themselves.
type Media interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
}

// Options configures a Server.
type Options struct {
	Reports Reports
	Media   Media
	// Accounts, Issuer and Google are optional: without them the server still
	// serves captures against the static token, which is what Phase 0 needs.
	Accounts Accounts
	Issuer   *auth.Issuer
	Google   *auth.GoogleVerifier
	// AllowedOrigins are the browser origins permitted to call this API. The
	// Flutter web client runs on a different origin, so without this it cannot.
	AllowedOrigins []string
	// Token authenticates every route but /healthz while the tier is personal.
	Token string
	// Account owns captures at the personal tier; v0.1 replaces this with the
	// verified account behind the request.
	Account        string
	MaxUploadBytes int64
	Log            *slog.Logger
}

// Server is the HTTP surface.
type Server struct {
	reports  Reports
	media    Media
	accounts Accounts
	issuer   *auth.Issuer
	google   *auth.GoogleVerifier
	origins  []string
	token    string
	account  string
	maxSize  int64
	log      *slog.Logger
}

// contextKey is unexported so nothing outside this package can collide with it.
type contextKey struct{ name string }

var claimsKey = contextKey{"claims"}

func claimsFrom(ctx context.Context) (auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(auth.Claims)
	return claims, ok
}

// DefaultMaxUploadBytes is generous enough for several phone photographs and
// small enough that a runaway upload cannot exhaust the process.
const DefaultMaxUploadBytes = 32 << 20 // 32 MiB

// New validates options and returns a Server.
func New(opts Options) (*Server, error) {
	if opts.Reports == nil {
		return nil, errors.New("api: a reports store is required")
	}
	if opts.Media == nil {
		return nil, errors.New("api: a media store is required")
	}
	if opts.Token == "" {
		return nil, errors.New("api: a token is required; refusing to serve unauthenticated")
	}
	size := opts.MaxUploadBytes
	if size <= 0 {
		size = DefaultMaxUploadBytes
	}
	log := opts.Log
	if log == nil {
		log = slog.Default()
	}
	return &Server{
		reports:  opts.Reports,
		media:    opts.Media,
		accounts: opts.Accounts,
		issuer:   opts.Issuer,
		google:   opts.Google,
		origins:  opts.AllowedOrigins,
		token:    opts.Token,
		account:  opts.Account,
		maxSize:  size,
		log:      log,
	}, nil
}

// Handler returns the router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)

	mux.HandleFunc("POST /v1/auth/register", s.handleRegister)
	mux.HandleFunc("POST /v1/auth/login", s.handleLogin)
	mux.HandleFunc("POST /v1/auth/google", s.handleGoogle)
	mux.Handle("GET /v1/auth/me", s.authenticated(http.HandlerFunc(s.handleMe)))

	mux.Handle("POST /v1/reports", s.authenticated(http.HandlerFunc(s.handlePostReport)))

	// The field kit is a static page; the token it holds is what authenticates
	// its uploads, so serving the page itself needs no token.
	if kit, err := fieldkitHandler(); err == nil {
		mux.Handle("GET /", kit)
	} else {
		s.log.Warn("field kit unavailable", "error", err.Error())
	}
	return s.withCORS(mux)
}

// withCORS answers browser preflights and echoes an allowed origin. A browser
// client on another origin — the Flutter web build — cannot call the API
// without this, and an unlisted origin is simply not answered.
func (s *Server) withCORS(next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range s.origins {
		allowed[strings.TrimSuffix(strings.TrimSpace(o), "/")] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSuffix(r.Header.Get("Origin"), "/")
		if origin != "" && (allowed[origin] || allowed["*"]) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}

// authenticated is a bearer-token check. It is deliberately the only thing
// standing in front of the capture endpoint at the personal tier, and it is
// replaced by phone verification when the endpoint opens to the public.
func (s *Server) authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			writeError(w, http.StatusUnauthorized, "a bearer token is required")
			return
		}
		presented := header[len(prefix):]

		// Two kinds of caller: the field kit with the server's own token, and a
		// signed-in person with a session token.
		if s.token != "" && subtleCompare(presented, s.token) {
			next.ServeHTTP(w, r)
			return
		}
		if s.issuer != nil {
			if claims, err := s.issuer.Verify(presented); err == nil {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
				return
			}
		}
		writeError(w, http.StatusUnauthorized, "that token is not valid")
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"status":  status,
		"message": message,
	}})
}

// subtleCompare is a constant-time comparison, so a token cannot be guessed by
// timing the response.
func subtleCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

//go:embed fieldkit
var fieldkitFS embed.FS

// fieldkitHandler serves the capture page. It is the one route that returns
// HTML, and it carries no data: everything it shows comes from the device.
func fieldkitHandler() (http.Handler, error) {
	sub, err := fs.Sub(fieldkitFS, "fieldkit")
	if err != nil {
		return nil, fmt.Errorf("field kit assets: %w", err)
	}
	return http.FileServer(http.FS(sub)), nil
}
