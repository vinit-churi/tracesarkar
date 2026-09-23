package archive

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// emptyPayloadHash is the SHA-256 of an empty body, which SigV4 requires for
// requests without one.
const emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

type credentials struct {
	accessKey string
	secret    string
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// sign adds the SigV4 headers to req in place.
func sign(req *http.Request, creds credentials, region, service, payloadHash string, when time.Time) {
	when = when.UTC()
	amzDate := when.Format("20060102T150405Z")
	dateStamp := when.Format("20060102")

	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	req.Header.Set("X-Amz-Date", amzDate)
	// S3 requires the payload hash as a signed header; other services do not
	// carry it, and signing it there would break their canonical request.
	if service == "s3" {
		req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	}

	signed, canonicalHeaders := canonicalHeaders(req, host)
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI(req),
		canonicalQuery(req),
		canonicalHeaders,
		signed,
		payloadHash,
	}, "\n")

	scope := strings.Join([]string{dateStamp, region, service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	key := hmacSHA256([]byte("AWS4"+creds.secret), dateStamp)
	key = hmacSHA256(key, region)
	key = hmacSHA256(key, service)
	key = hmacSHA256(key, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(key, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		creds.accessKey, scope, signed, signature))
}

// canonicalHeaders returns the signed-header list and the canonical block. Only
// host and the x-amz-* headers are signed, which is all R2 requires.
func canonicalHeaders(req *http.Request, host string) (signed, block string) {
	values := map[string]string{"host": host}
	for name, vs := range req.Header {
		lower := strings.ToLower(name)
		if strings.HasPrefix(lower, "x-amz-") {
			values[lower] = strings.TrimSpace(strings.Join(vs, ","))
		}
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		b.WriteString(name)
		b.WriteByte(':')
		b.WriteString(values[name])
		b.WriteByte('\n')
	}
	return strings.Join(names, ";"), b.String()
}

func canonicalURI(req *http.Request) string {
	path := req.URL.EscapedPath()
	if path == "" {
		return "/"
	}
	return path
}

func canonicalQuery(req *http.Request) string {
	q := req.URL.Query()
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		values := append([]string(nil), q[k]...)
		sort.Strings(values)
		for _, v := range values {
			parts = append(parts, uriEscape(k)+"="+uriEscape(v))
		}
	}
	return strings.Join(parts, "&")
}

// uriEscape implements RFC 3986 escaping, which differs from url.QueryEscape in
// its treatment of spaces and tildes.
func uriEscape(s string) string {
	const unreserved = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if strings.IndexByte(unreserved, c) >= 0 {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "%%%02X", c)
	}
	return b.String()
}
