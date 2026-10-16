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
	"net/url"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/mail"
	"github.com/vinit-churi/tracesarkar/internal/polish"
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
	Reviews  Reviews
	// Labels and Blobs power the labelling surface, which turns captures into
	// the evaluation set system 3 is measured against. Optional: without them
	// the API still serves captures.
	Labels Labels
	// Truth powers the annotation surface the redaction detector is graded
	// against. Optional, like labelling.
	Truth Truth
	// Wards powers the only surface served without a sign-in.
	Wards WardProfiles
	// Issues powers corroboration and the SLA clock.
	Issues Issues
	// PublicIssues powers the issue permalink, the second surface served
	// without a sign-in.
	PublicIssues PublicIssues
	// Handles is the registry of verified social accounts a generated post
	// may tag. Optional: without it the share kit still generates every post,
	// untagged, which is the safe direction.
	Handles Handles
	// Evidence compiles one issue into the record behind the PDF pack. The
	// only share surface that is not public.
	Evidence Evidence
	// Polish rewords the prose lines of a generated post. Optional, and off
	// unless configured: without it every post is the deterministic
	// composition, which is what ADR 0020 shipped and is always correct.
	Polish polish.Polisher
	// Credentials persists one-time links and the token version behind
	// revocation. Optional: without it the reset and verification endpoints
	// refuse cleanly rather than half-working.
	Credentials Credentials
	// Mail delivers a reset or verification link. Optional, and the reset
	// endpoints refuse without it rather than minting a credential nobody
	// will ever receive.
	Mail mail.Sender
	// PublicBaseURL is where the client lives, used to build the link in a
	// reset mail. Without it a link cannot be built and the mail is not sent.
	PublicBaseURL string
	// Operator powers system 8's first surface: collection health, the queues,
	// and the claimed-versus-confirmed counts. Optional, like every other
	// surface that needs a database.
	Operator Operator
	// Confirmations lets a reporter answer the question the classifier could
	// not: which problem this is. Optional, like every other surface that
	// needs a database.
	Confirmations Confirmations
	Blobs         Blobs
	Details       Details
	Issuer        *auth.Issuer
	Google        *auth.GoogleVerifier
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
	reports       Reports
	media         Media
	accounts      Accounts
	reviews       Reviews
	wards         WardProfiles
	issues        Issues
	publicIssues  PublicIssues
	handles       Handles
	evidence      Evidence
	polish        polish.Polisher
	credentials   Credentials
	operator      Operator
	confirmations Confirmations
	mail          mail.Sender
	labels        Labels
	truth         Truth
	blobs         Blobs
	details       Details
	issuer        *auth.Issuer
	google        *auth.GoogleVerifier
	origins       []string
	token         string
	account       string
	publicBaseURL string
	maxSize       int64
	throttle      *throttles
	log           *slog.Logger
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
	// Checked here rather than at the first use. A permalink is burned into a
	// card image and forwarded past any page that could correct it, so a
	// malformed base must stop the process starting rather than appear under
	// every reader's photograph.
	base := strings.TrimSpace(opts.PublicBaseURL)
	if base != "" {
		u, err := url.Parse(base)
		if err != nil {
			return nil, fmt.Errorf("api: PublicBaseURL %q is not a URL: %w", base, err)
		}
		if u.Scheme != "https" && u.Scheme != "http" {
			return nil, fmt.Errorf("api: PublicBaseURL %q is not http or https", base)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("api: PublicBaseURL %q names no host", base)
		}
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
		reports:       opts.Reports,
		media:         opts.Media,
		accounts:      opts.Accounts,
		throttle:      newThrottles(),
		reviews:       opts.Reviews,
		wards:         opts.Wards,
		issues:        opts.Issues,
		publicIssues:  opts.PublicIssues,
		handles:       opts.Handles,
		evidence:      opts.Evidence,
		polish:        opts.Polish,
		credentials:   opts.Credentials,
		operator:      opts.Operator,
		confirmations: opts.Confirmations,
		mail:          opts.Mail,
		publicBaseURL: base,
		labels:        opts.Labels,
		truth:         opts.Truth,
		blobs:         opts.Blobs,
		details:       opts.Details,
		issuer:        opts.Issuer,
		google:        opts.Google,
		origins:       opts.AllowedOrigins,
		token:         opts.Token,
		account:       opts.Account,
		maxSize:       size,
		log:           log,
	}, nil
}

// Handler returns the router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	// Google's frontend answers /healthz itself: on Cloud Run the request never
	// reaches the container and returns Google's own 404, with no entry in the
	// request log. Anything watching the service from outside has to use a path
	// inside our namespace, which nothing upstream reserves.
	mux.HandleFunc("GET /v1/health", s.handleHealth)

	mux.HandleFunc("POST /v1/auth/register", s.handleRegister)
	mux.HandleFunc("POST /v1/auth/login", s.handleLogin)
	mux.HandleFunc("POST /v1/auth/google", s.handleGoogle)
	mux.Handle("GET /v1/auth/me", s.authenticated(http.HandlerFunc(s.handleMe)))
	// Unauthenticated, because somebody who cannot sign in is exactly who
	// needs them. The one-time token is the whole proof.
	mux.HandleFunc("POST /v1/auth/password/forgot", s.handleForgotPassword)
	mux.HandleFunc("POST /v1/auth/password/reset", s.handleResetPassword)
	mux.HandleFunc("POST /v1/auth/email/verify", s.handleVerifyEmail)
	// Authenticated: there is no reason for a stranger to be able to send
	// mail to somebody else's address, or to end their sessions.
	mux.Handle("POST /v1/auth/email/verify/request",
		s.authenticated(http.HandlerFunc(s.handleRequestEmailVerification)))
	mux.Handle("POST /v1/auth/sessions/revoke",
		s.authenticated(http.HandlerFunc(s.handleRevokeSessions)))

	// System 8's first surface. Authenticated: a count of what is stuck is an
	// operational signal, and the published version of the ratio belongs on a
	// surface built for publishing it.
	mux.Handle("GET /v1/ops/health", s.authenticated(http.HandlerFunc(s.handleOperatorHealth)))
	mux.Handle("GET /v1/ops/trend", s.authenticated(http.HandlerFunc(s.handleOperatorTrend)))

	mux.Handle("POST /v1/reports", s.authenticated(http.HandlerFunc(s.handlePostReport)))

	// The attribution review surface. Personal tier: it exists so a human can
	// put a precision number on the join before anything it produces is shown
	// to anyone.
	mux.Handle("GET /v1/reports", s.authenticated(http.HandlerFunc(s.handleReportList)))
	mux.Handle("GET /v1/reports/{id}", s.authenticated(http.HandlerFunc(s.handleReportDetail)))

	mux.Handle("GET /v1/reports/{id}/media", s.authenticated(http.HandlerFunc(s.handleReportMedia)))
	// The question the classifier could not answer, and the reporter's reply.
	// Both scoped to the caller's own captures: a capture carries a position
	// accurate to metres and a photograph that may not be redacted yet.
	mux.Handle("GET /v1/reports/awaiting-confirmation",
		s.authenticated(http.HandlerFunc(s.handleAwaitingConfirmation)))
	mux.Handle("POST /v1/reports/{id}/confirm",
		s.authenticated(http.HandlerFunc(s.handleConfirmClassification)))
	mux.Handle("POST /v1/issues/{id}/lodged", s.authenticated(http.HandlerFunc(s.handleLodged)))
	// The evidence pack. Signed in, unlike the rest of the share kit: it
	// carries the exact position and the complaint reference, because it is
	// the document that gets attached to a complaint.
	mux.Handle("GET /v1/issues/{id}/evidence.pdf",
		s.authenticated(http.HandlerFunc(s.handleEvidencePack)))

	// Public on purpose: published facts about offices and obligations, with
	// nothing a citizen reported on it.
	mux.HandleFunc("GET /v1/public/ward/{authority}/{ward}", s.handleWardProfile)
	mux.HandleFunc("GET /v1/public/issue/{id}", s.handlePublicIssue)
	mux.HandleFunc("GET /v1/public/issue/{id}/image", s.handlePublicIssueImage)
	// The share kit. Public for the same reason the page is: asking a resident
	// to sign in before they can send the post defeats the point of writing it
	// for them.
	mux.HandleFunc("GET /v1/public/issue/{id}/share", s.handleShareKit)
	mux.HandleFunc("GET /v1/public/issue/{id}/card", s.handleShareCard)

	mux.Handle("GET /v1/taxonomy", s.authenticated(http.HandlerFunc(s.handleTaxonomy)))
	mux.Handle("GET /v1/label/queue", s.authenticated(http.HandlerFunc(s.handleLabelQueue)))
	mux.Handle("POST /v1/label/{id}", s.authenticated(http.HandlerFunc(s.handleSaveLabel)))

	// The adversarial redaction set: a photograph marked up by hand is the only
	// thing that can say whether the detector finds a face it has not been
	// shown before (ADR 0008).
	mux.Handle("GET /v1/redaction/truth/queue", s.authenticated(http.HandlerFunc(s.handleTruthQueue)))
	mux.Handle("POST /v1/redaction/truth/{id}", s.authenticated(http.HandlerFunc(s.handleSaveTruth)))
	mux.Handle("GET /v1/media/{id}", s.authenticated(http.HandlerFunc(s.handleMedia)))

	mux.Handle("GET /v1/review/attribution", s.authenticated(http.HandlerFunc(s.handleReviewQueue)))
	mux.Handle("POST /v1/review/attribution/{id}", s.authenticated(http.HandlerFunc(s.handleReviewVerdict)))

	// The field kit is a static page; the token it holds is what authenticates
	// its uploads, so serving the page itself needs no token.
	if kit, err := fieldkitHandler(); err == nil {
		mux.Handle("GET /", kit)

		if reviewPage, err := reviewHandler(); err == nil {
			mux.Handle("GET /review/", reviewPage)
			mux.Handle("GET /review", http.RedirectHandler("/review/", http.StatusFound))
		}
		if page, err := staticPage(wardFS, "ward"); err == nil {
			mux.Handle("GET /ward/", page)
			mux.Handle("GET /ward", http.RedirectHandler("/ward/", http.StatusFound))
		}
		// The issue permalink carries an id in its path, so it cannot be a
		// plain file server. The shell reads the id back and asks the public
		// endpoint for the facts, and the server fills in the metadata an
		// unfurl needs — which means the gate is consulted before a byte is
		// written, rather than afterwards by the page's own script.
		if shell, err := issueShell(); err == nil {
			page := s.issuePageHandler(shell)
			mux.Handle("GET /issue/{id}", page)
			mux.Handle("GET /issue/{id}/", page)
		}
		if page, err := methodologyHandler(); err == nil {
			// Public on purpose, and the only page that is. A platform that
			// publishes facts about named parties has to publish how it got
			// them and what it does not know.
			mux.Handle("GET /methodology/", page)
			mux.Handle("GET /methodology", http.RedirectHandler("/methodology/", http.StatusFound))
		}
		if labelPage, err := labelHandler(); err == nil {
			mux.Handle("GET /label/", labelPage)
			mux.Handle("GET /label", http.RedirectHandler("/label/", http.StatusFound))
		}
		// Where the adversarial redaction set is built. Signed-in only, and it
		// shows un-redacted originals, so it serves them to nobody but the
		// account that captured them.
		if page, err := staticPage(truthFS, "truth"); err == nil {
			mux.Handle("GET /truth/", page)
			mux.Handle("GET /truth", http.RedirectHandler("/truth/", http.StatusFound))
		}
		// The operator console. One page, three blocks, and no moderation:
		// the surfaces that need a human verdict are /review/, /label/ and
		// /truth/, and this is the one that answers whether anything is
		// stuck.
		if page, err := staticPage(opsFS, "ops"); err == nil {
			mux.Handle("GET /ops/", page)
			mux.Handle("GET /ops", http.RedirectHandler("/ops/", http.StatusFound))
		}
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
			if claims, err := s.issuer.Verify(presented); err == nil && s.sessionIsCurrent(r, claims) {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
				return
			}
		}
		writeError(w, http.StatusUnauthorized, "that token is not valid")
	})
}

// sessionIsCurrent is the revocation check: the version in the token against
// the version on the account (migration 0044).
//
// This is the one database read the signed-token design was meant to avoid,
// and it is worth it. Without it "sign out everywhere" and "reset my
// password" cannot end a session before its own expiry, which is thirty days
// — so a stolen token stays good for a month after the theft is noticed. One
// row by primary key is the cheapest form that answer comes in; a deny list
// of revoked tokens would be a read too, and would grow forever.
//
// A read that fails refuses the request. Treating a database blip as "the
// session is fine" would mean a revoked token working again whenever the
// database hiccups.
func (s *Server) sessionIsCurrent(r *http.Request, claims auth.Claims) bool {
	if s.credentials == nil {
		return true
	}
	version, err := s.credentials.TokenVersion(r.Context(), claims.AccountID)
	if err != nil {
		if !errors.Is(err, ErrNoAccount) {
			s.log.Error("could not check whether a session is current",
				"error", err.Error())
		}
		return false
	}
	return claims.TokenVersion == version
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

//go:embed review
var reviewFS embed.FS

//go:embed label
var labelFS embed.FS

//go:embed methodology
var methodologyFS embed.FS

//go:embed ward
var wardFS embed.FS

//go:embed issue
var issueFS embed.FS

//go:embed truth
var truthFS embed.FS

//go:embed ops
var opsFS embed.FS

// reviewHandler serves the attribution review page. Personal tier: it exists
// so a human can put a precision number on the join, and it shows no data of
// its own — everything comes from the authenticated API.
func reviewHandler() (http.Handler, error) {
	sub, err := fs.Sub(reviewFS, "review")
	if err != nil {
		return nil, fmt.Errorf("review assets: %w", err)
	}
	return http.StripPrefix("/review", http.FileServer(http.FS(sub))), nil
}

// staticPage serves one embedded directory at its own prefix.
func staticPage(from embed.FS, name string) (http.Handler, error) {
	sub, err := fs.Sub(from, name)
	if err != nil {
		return nil, fmt.Errorf("%s assets: %w", name, err)
	}
	return http.StripPrefix("/"+name, http.FileServer(http.FS(sub))), nil
}

// issueShell is the permalink page before any issue has been read into it.
func issueShell() ([]byte, error) {
	body, err := issueFS.ReadFile("issue/index.html")
	if err != nil {
		return nil, fmt.Errorf("issue page: %w", err)
	}
	return body, nil
}

// methodologyHandler serves the public methodology page.
func methodologyHandler() (http.Handler, error) {
	sub, err := fs.Sub(methodologyFS, "methodology")
	if err != nil {
		return nil, fmt.Errorf("methodology assets: %w", err)
	}
	return http.StripPrefix("/methodology", http.FileServer(http.FS(sub))), nil
}

// labelHandler serves the labelling page. Like the review page it holds no
// data of its own — the photographs and the taxonomy both come from the
// authenticated API.
func labelHandler() (http.Handler, error) {
	sub, err := fs.Sub(labelFS, "label")
	if err != nil {
		return nil, fmt.Errorf("label assets: %w", err)
	}
	return http.StripPrefix("/label", http.FileServer(http.FS(sub))), nil
}

// fieldkitHandler serves the capture page. It is the one route that returns
// HTML, and it carries no data: everything it shows comes from the device.
func fieldkitHandler() (http.Handler, error) {
	sub, err := fs.Sub(fieldkitFS, "fieldkit")
	if err != nil {
		return nil, fmt.Errorf("field kit assets: %w", err)
	}
	return http.FileServer(http.FS(sub)), nil
}
