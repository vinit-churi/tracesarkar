package archive_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/config"
)

// Live tests run against the real bucket. They are skipped unless
// TRACESARKAR_LIVE=1, so `go test ./...` stays offline by default.
func liveClient(t *testing.T) *archive.Client {
	t.Helper()
	if os.Getenv("TRACESARKAR_LIVE") != "1" {
		t.Skip("set TRACESARKAR_LIVE=1 to run live object-storage tests")
	}
	cfg, err := config.Load(".env")
	if err != nil {
		cfg, err = config.Load("../../.env")
	}
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	c, err := archive.New(archive.Options{
		Endpoint:  cfg.R2.Endpoint,
		Bucket:    cfg.R2.Bucket,
		Region:    cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey,
		Secret:    cfg.R2.Secret,
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	return c
}

func TestLiveRoundTrip(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	body := []byte(fmt.Sprintf(`{"probe":"phase0","at":%q}`, time.Now().UTC().Format(time.RFC3339)))
	key := archive.KeyFor("_selftest", time.Now().UTC(), archive.Sum(body), ".json")

	if err := c.Put(ctx, key, body, "application/json"); err != nil {
		t.Fatalf("put: %v", err)
	}

	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if !exists {
		t.Fatal("object should exist after put")
	}

	got, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("round trip mismatch:\n got %s\nwant %s", got, body)
	}

	missing, err := c.Exists(ctx, "archive/_selftest/definitely/not/here.json")
	if err != nil {
		t.Fatalf("exists(absent): %v", err)
	}
	if missing {
		t.Error("absent key reported as present")
	}
}
