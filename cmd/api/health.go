package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// healthCheck asks a running server whether it is well. It exists because the
// runtime image is distroless: there is no curl and no shell, so the binary has
// to be able to check itself for Docker's HEALTHCHECK.
func healthCheck(ctx context.Context, base string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("health request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("health request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned %d", resp.StatusCode)
	}
	return nil
}
