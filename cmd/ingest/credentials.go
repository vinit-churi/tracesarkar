package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

// runCredentials removes one-time credentials that can no longer be used.
//
// Not a privacy sweep — the rows hold a hash, an account id and two
// timestamps, and no address. It is housekeeping with one deliberate delay:
// a spent or expired row is kept for a week, so that "my reset link never
// arrived" can still be answered with when it was issued and whether it was
// used. After that it answers nothing anybody will ask.
func runCredentials(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("credentials", flag.ExitOnError)
	keep := fs.Duration("keep", 7*24*time.Hour,
		"how long a spent or expired credential is kept before deletion")
	if err := fs.Parse(args); err != nil {
		return err
	}

	_, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	n, err := db.DeleteSpentCredentials(ctx, *keep)
	if err != nil {
		return err
	}
	if n == 0 {
		fmt.Fprintln(os.Stdout, "no spent credentials to remove")
		return nil
	}
	fmt.Fprintf(os.Stdout, "removed %d spent credential(s) older than %s\n", n, *keep)
	return nil
}
