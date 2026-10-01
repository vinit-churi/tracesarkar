package classify

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	pending []Pending
	saved   []Attempt
	saveErr error
}

func (f *fakeStore) Pending(context.Context, int) ([]Pending, error) { return f.pending, nil }
func (f *fakeStore) Save(_ context.Context, a Attempt) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, a)
	return nil
}

type fakeBlobs struct{ data map[string][]byte; err error }

func (f *fakeBlobs) Get(_ context.Context, key string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	b, ok := f.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}

func goodResult() Result {
	return Result{Category: "road_defect", Subcategory: "pothole", Severity: "high",
		IsCivicIssue: true, ImageQuality: "good", Confidence: 0.9}
}

func TestTheRunnerClassifiesEveryPendingCapture(t *testing.T) {
	store := &fakeStore{pending: []Pending{
		{ReportID: "r1", ArchiveKey: "k1"},
		{ReportID: "r2", ArchiveKey: "k2"},
	}}
	blobs := &fakeBlobs{data: map[string][]byte{"k1": []byte("a"), "k2": []byte("b")}}
	c := &fakeClassifier{byLabel: map[string]Result{"a": goodResult(), "b": goodResult()}}

	n, err := Run(context.Background(), RunnerOptions{
		Store: store, Blobs: blobs, Classifier: c,
		Coverage: CoverageFor("road_defect"), Model: "m", Prompt: "v1",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 2 || len(store.saved) != 2 {
		t.Fatalf("classified %d, saved %d", n, len(store.saved))
	}
	if store.saved[0].Decision.Outcome != Accepted {
		t.Errorf("outcome: %+v", store.saved[0].Decision)
	}
}

func TestOneFailureDoesNotStopTheRest(t *testing.T) {
	// A capture that cannot be classified must not block the queue behind it.
	// Hard rule 7: nothing captured is ever lost, and a stuck queue loses
	// everything after the stuck item.
	store := &fakeStore{pending: []Pending{
		{ReportID: "bad", ArchiveKey: "missing"},
		{ReportID: "good", ArchiveKey: "k"},
	}}
	blobs := &fakeBlobs{data: map[string][]byte{"k": []byte("a")}}
	c := &fakeClassifier{byLabel: map[string]Result{"a": goodResult()}}

	n, err := Run(context.Background(), RunnerOptions{
		Store: store, Blobs: blobs, Classifier: c,
		Coverage: CoverageFor("road_defect"), Model: "m", Prompt: "v1",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 1 {
		t.Errorf("one should have succeeded, got %d", n)
	}
	if len(store.saved) != 2 {
		t.Fatalf("both attempts should be recorded, got %d", len(store.saved))
	}
}

func TestAFailedAttemptIsRecordedRatherThanForgotten(t *testing.T) {
	// A classification that silently never ran looks exactly like one that
	// ran and found nothing. The two need different responses.
	store := &fakeStore{pending: []Pending{{ReportID: "r1", ArchiveKey: "k"}}}
	blobs := &fakeBlobs{data: map[string][]byte{"k": []byte("a")}}
	c := &fakeClassifier{err: errors.New("provider exploded")}

	if _, err := Run(context.Background(), RunnerOptions{
		Store: store, Blobs: blobs, Classifier: c,
		Coverage: CoverageFor("road_defect"), Model: "m", Prompt: "v1",
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(store.saved) != 1 {
		t.Fatalf("the failure should be recorded: %d", len(store.saved))
	}
	if store.saved[0].Err == "" {
		t.Error("the attempt should carry the reason it failed")
	}
}

func TestEveryAttemptRecordsWhatProducedIt(t *testing.T) {
	// A classification whose model and prompt version are unknown cannot go
	// into an eval, because there is no telling what it is evidence of.
	store := &fakeStore{pending: []Pending{{ReportID: "r1", ArchiveKey: "k"}}}
	blobs := &fakeBlobs{data: map[string][]byte{"k": []byte("a")}}
	c := &fakeClassifier{byLabel: map[string]Result{"a": goodResult()}}

	if _, err := Run(context.Background(), RunnerOptions{
		Store: store, Blobs: blobs, Classifier: c,
		Coverage: CoverageFor("road_defect"), Model: "deepseek-flash", Prompt: "v1",
	}); err != nil {
		t.Fatal(err)
	}

	got := store.saved[0]
	if got.Model != "deepseek-flash" || got.Prompt != "v1" {
		t.Errorf("provenance lost: %+v", got)
	}
	if got.Latency <= 0 {
		t.Error("latency should be measured; it is the cost signal")
	}
}

func TestNothingPendingIsNotAnError(t *testing.T) {
	n, err := Run(context.Background(), RunnerOptions{
		Store: &fakeStore{}, Blobs: &fakeBlobs{}, Classifier: &fakeClassifier{},
		Coverage: CoverageFor("road_defect"), Model: "m", Prompt: "v1",
	})
	if err != nil {
		t.Fatalf("an empty queue is the normal case: %v", err)
	}
	if n != 0 {
		t.Errorf("classified %d from an empty queue", n)
	}
}

var _ = time.Now
