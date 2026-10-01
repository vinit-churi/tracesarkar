package classify

import (
	"context"
	"errors"
	"strconv"
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

type fakeBlobs struct {
	data map[string][]byte
	err  error
}

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

// A pass in which nothing at all succeeded is a different thing from a pass in
// which one photograph was unreadable, and the difference matters most on a day
// of fieldwork: a provider that is down, a key that has expired or a model name
// that no longer exists fails every capture identically, and reports nothing.
func TestRunReportsATotalFailure(t *testing.T) {
	tests := []struct {
		name    string
		results []error // one per pending capture; nil means it classified
		wantErr bool
	}{
		{name: "all succeeded", results: []error{nil, nil}},
		{name: "one bad photograph among good ones", results: []error{errors.New("unsupported image"), nil}},
		{name: "nothing succeeded", results: []error{errors.New("401 unauthorized"), errors.New("401 unauthorized")}, wantErr: true},
		{name: "nothing pending is not a failure", results: nil},
		// One capture failing alone is indistinguishable from one bad
		// photograph, so it is not reported as a run failure.
		{name: "a lone failure is ambiguous", results: []error{errors.New("unsupported image")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{}
			blobs := &fakeBlobs{data: map[string][]byte{}}
			byLabel := map[string]Result{}
			failFor := map[string]error{}
			for i, want := range tt.results {
				key := "k" + strconv.Itoa(i)
				label := "img" + strconv.Itoa(i)
				store.pending = append(store.pending, Pending{
					ReportID: "r" + strconv.Itoa(i), ArchiveKey: key,
				})
				blobs.data[key] = []byte(label)
				if want != nil {
					failFor[label] = want
				} else {
					byLabel[label] = goodResult()
				}
			}

			_, err := Run(context.Background(), RunnerOptions{
				Store:      store,
				Blobs:      blobs,
				Classifier: &fakeClassifier{byLabel: byLabel, errByLabel: failFor},
				Coverage:   CoverageFor("road_defect"),
				Model:      "test", Prompt: "v1",
			})

			if tt.wantErr && err == nil {
				t.Error("a pass where every capture failed must be reported, not logged and forgotten")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// Either way every attempt is recorded — that is what makes the
			// next pass a retry rather than a repeat.
			if len(store.saved) != len(tt.results) {
				t.Errorf("recorded %d attempts, want %d", len(store.saved), len(tt.results))
			}
		})
	}
}
