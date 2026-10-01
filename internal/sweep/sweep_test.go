package sweep

import (
	"context"
	"errors"
	"testing"
)

func TestRun(t *testing.T) {
	boom := errors.New("provider unavailable")

	tests := []struct {
		name      string
		cancelled bool
		stages    []string // names; a name ending in "!" fails
		wantRan   []string
		wantErrs  int
	}{
		{
			name:    "every stage runs in order",
			stages:  []string{"enrich", "classify"},
			wantRan: []string{"enrich", "classify"},
		},
		{
			// The stages are independent. A classification provider being
			// down must not stop jurisdiction from resolving, and a report
			// that failed one pass is picked up by the next.
			name:     "a failing stage does not stop the ones after it",
			stages:   []string{"enrich!", "classify"},
			wantRan:  []string{"enrich", "classify"},
			wantErrs: 1,
		},
		{
			name:     "every failure is reported, none swallowed",
			stages:   []string{"enrich!", "classify!"},
			wantRan:  []string{"enrich", "classify"},
			wantErrs: 2,
		},
		{
			// Cloud Run sends SIGTERM before it stops a job. Starting a paid
			// vision call into a shutdown wastes money and leaves a half-done
			// pass behind.
			name:      "a cancelled context stops before the first stage",
			cancelled: true,
			stages:    []string{"enrich", "classify"},
			wantRan:   nil,
			wantErrs:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancelled {
				cancel()
			}

			var ran []string
			var stages []Stage
			for _, s := range tt.stages {
				name, fails := s, false
				if name[len(name)-1] == '!' {
					name, fails = name[:len(name)-1], true
				}
				stages = append(stages, Stage{Name: name, Run: func(context.Context) error {
					ran = append(ran, name)
					if fails {
						return boom
					}
					return nil
				}})
			}

			results := Run(ctx, stages)

			if len(results) != len(tt.stages) {
				t.Fatalf("got %d results, want %d — every stage must be accounted for",
					len(results), len(tt.stages))
			}
			if got := len(ran); got != len(tt.wantRan) {
				t.Errorf("ran %v, want %v", ran, tt.wantRan)
			}
			for i := range tt.wantRan {
				if i < len(ran) && ran[i] != tt.wantRan[i] {
					t.Errorf("stage %d ran %q, want %q", i, ran[i], tt.wantRan[i])
				}
			}

			var errs int
			for _, r := range results {
				if r.Err != nil {
					errs++
				}
			}
			if errs != tt.wantErrs {
				t.Errorf("got %d errors, want %d", errs, tt.wantErrs)
			}
		})
	}
}

func TestFailed(t *testing.T) {
	// The exit code is what the scheduler alerts on, so "did anything fail"
	// has to be answerable without re-walking the results.
	if Failed(nil) {
		t.Error("no results is not a failure")
	}
	if Failed([]Result{{Name: "enrich"}}) {
		t.Error("a clean pass reported as failed")
	}
	if !Failed([]Result{{Name: "enrich"}, {Name: "classify", Err: errors.New("x")}}) {
		t.Error("a failed stage was not reported")
	}
}
