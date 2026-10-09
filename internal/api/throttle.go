package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/ratelimit"
)

// Guessing a password is the cheapest attack on this platform. These are the
// two things worth limiting and they protect against different attacks, so
// they are counted separately.
const (
	// failedLoginLimit caps wrong passwords against one address. Only
	// failures count: somebody signing in on a phone and a laptop every
	// morning is not an attacker, and throttling them would be a bug that
	// looked like security.
	failedLoginLimit  = 6
	failedLoginWindow = 15 * time.Minute

	// authAttemptLimit caps everything one caller may try, successful or not.
	// Wider, because it is aimed at a machine working through a list rather
	// than a person getting their own password wrong.
	authAttemptLimit  = 30
	authAttemptWindow = 5 * time.Minute
)

// throttles holds the counters. In memory, so per instance: the API runs at
// most three, and an attacker spread across them gets at most three times
// these numbers. Stated rather than hidden; moving the counters into the
// database is the fix when that stops being tolerable.
type throttles struct {
	byAccount *ratelimit.Limiter
	byCaller  *ratelimit.Limiter
}

func newThrottles() *throttles {
	return &throttles{
		byAccount: ratelimit.New(failedLoginLimit, failedLoginWindow),
		byCaller:  ratelimit.New(authAttemptLimit, authAttemptWindow),
	}
}

// refuse answers a throttled request.
//
// The message never says whether the address exists. The whole point of
// answering "email or password is incorrect" to both cases is undone if the
// throttle then behaves differently for one of them, so an unknown address is
// counted and refused exactly like a known one.
func refuse(w http.ResponseWriter, retry time.Duration) {
	seconds := int(retry.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	writeError(w, http.StatusTooManyRequests,
		"Too many sign-in attempts. Try again in "+humanWait(retry)+".")
}

func humanWait(d time.Duration) string {
	if d < time.Minute {
		return "under a minute"
	}
	m := int(d.Minutes()) + 1
	if m == 1 {
		return "a minute"
	}
	return strconv.Itoa(m) + " minutes"
}

// caller identifies who is asking, for the per-caller limit.
//
// Cloud Run puts the client address first in X-Forwarded-For and appends its
// own hops, so the first entry is the one to count. A client can prepend a
// value it invented, which means this alone would be forgeable — which is
// exactly why the limit that actually protects an account is keyed on the
// address being attacked, not on this.
func caller(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// allowAuthAttempt is called before any credential is checked.
func (s *Server) allowAuthAttempt(w http.ResponseWriter, r *http.Request) bool {
	if s.throttle == nil {
		return true
	}
	if ok, retry := s.throttle.byCaller.Allow("caller:" + caller(r)); !ok {
		refuse(w, retry)
		return false
	}
	return true
}

// allowLoginFor is called before a password is verified, and recordFailedLogin
// after it turns out to be wrong — so only failures are counted.
func (s *Server) allowLoginFor(w http.ResponseWriter, email string) bool {
	if s.throttle == nil {
		return true
	}
	key := "account:" + auth.NormaliseEmail(email)
	if ok, retry := s.throttle.byAccount.Peek(key); !ok {
		refuse(w, retry)
		return false
	}
	return true
}

func (s *Server) recordFailedLogin(email string) {
	if s.throttle == nil {
		return
	}
	s.throttle.byAccount.Allow("account:" + auth.NormaliseEmail(email))
}
