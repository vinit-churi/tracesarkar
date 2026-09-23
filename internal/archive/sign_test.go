package archive

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// The "get-vanilla" case from the AWS Signature Version 4 test suite.
func TestSignMatchesAWSTestVector(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.amazonaws.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "example.amazonaws.com"

	when := time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)
	creds := credentials{accessKey: "AKIDEXAMPLE", secret: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY"}

	sign(req, creds, "us-east-1", "service", emptyPayloadHash, when)

	const want = "AWS4-HMAC-SHA256 " +
		"Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, " +
		"SignedHeaders=host;x-amz-date, " +
		"Signature=5fa00fa31553b73ebf1942676e86291e8372ff2a2260956d9b8aae1d763fbf31"

	if got := req.Header.Get("Authorization"); got != want {
		t.Errorf("Authorization:\n got %s\nwant %s", got, want)
	}
	if got := req.Header.Get("X-Amz-Date"); got != "20150830T123600Z" {
		t.Errorf("X-Amz-Date: got %q", got)
	}
}

func TestSignAddsPayloadHashHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodPut, "https://acc.r2.cloudflarestorage.com/bucket/key", strings.NewReader("body"))
	if err != nil {
		t.Fatal(err)
	}
	payload := sha256Hex([]byte("body"))

	sign(req, credentials{accessKey: "ak", secret: "sk"}, "auto", "s3", payload, time.Now().UTC())

	if got := req.Header.Get("X-Amz-Content-Sha256"); got != payload {
		t.Errorf("X-Amz-Content-Sha256: got %q, want %q", got, payload)
	}
	if !strings.Contains(req.Header.Get("Authorization"), "x-amz-content-sha256") {
		t.Errorf("payload hash header must be signed: %q", req.Header.Get("Authorization"))
	}
}

func TestSha256HexOfEmptyPayload(t *testing.T) {
	if got := sha256Hex(nil); got != emptyPayloadHash {
		t.Errorf("got %q, want %q", got, emptyPayloadHash)
	}
}
