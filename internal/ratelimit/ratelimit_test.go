package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestAllowsUpToTheLimitThenRefuses(t *testing.T) {
	now := time.Now()
	l := New(3, time.Minute)
	l.now = func() time.Time { return now }

	for i := range 3 {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("attempt %d was refused inside the limit", i+1)
		}
	}
	ok, retry := l.Allow("a")
	if ok {
		t.Fatal("the fourth attempt was allowed")
	}
	// A refusal has to say when to come back, or the caller cannot tell the
	// person anything useful.
	if retry <= 0 || retry > time.Minute {
		t.Errorf("retry after %v, want something inside the window", retry)
	}
}

func TestTheWindowSlides(t *testing.T) {
	now := time.Now()
	l := New(2, time.Minute)
	l.now = func() time.Time { return now }

	l.Allow("a")
	l.Allow("a")
	if ok, _ := l.Allow("a"); ok {
		t.Fatal("allowed past the limit")
	}

	// Just inside the window: still refused.
	now = now.Add(59 * time.Second)
	if ok, _ := l.Allow("a"); ok {
		t.Error("allowed before the window had passed")
	}

	// Past it: the earliest attempts have aged out.
	now = now.Add(2 * time.Second)
	if ok, _ := l.Allow("a"); !ok {
		t.Error("still refused after the window passed")
	}
}

func TestKeysAreIndependent(t *testing.T) {
	l := New(1, time.Minute)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("first key refused")
	}
	// One person hitting their limit must not lock anybody else out.
	if ok, _ := l.Allow("b"); !ok {
		t.Error("a second key was refused because of the first")
	}
}

// Keys arrive from the internet — an email in a login body is attacker
// controlled. Without eviction, a few million distinct keys is the memory of
// the process.
func TestSpentKeysAreForgotten(t *testing.T) {
	now := time.Now()
	l := New(5, time.Minute)
	l.now = func() time.Time { return now }

	for i := range 1000 {
		l.Allow(string(rune(i)))
	}
	if n := l.size(); n != 1000 {
		t.Fatalf("holding %d keys, expected 1000", n)
	}

	now = now.Add(2 * time.Minute)
	l.Allow("anything")
	if n := l.size(); n > 2 {
		t.Errorf("holding %d keys after the window passed; they must be dropped", n)
	}
}

func TestConcurrentUseIsSafe(t *testing.T) {
	l := New(100, time.Minute)
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				l.Allow("shared")
			}
		}()
	}
	wg.Wait()
	// 1000 attempts against a limit of 100: the count must not exceed it.
	if ok, _ := l.Allow("shared"); ok {
		t.Error("the limit was exceeded under concurrency")
	}
}
