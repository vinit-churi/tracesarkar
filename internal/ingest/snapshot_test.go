package ingest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/works"
)

type fakeArchive struct {
	put    map[string][]byte
	exists map[string]bool
}

func newFakeArchive() *fakeArchive {
	return &fakeArchive{put: map[string][]byte{}, exists: map[string]bool{}}
}

func (f *fakeArchive) Put(_ context.Context, key string, body []byte, _ string) error {
	f.put[key] = body
	f.exists[key] = true
	return nil
}

func (f *fakeArchive) Exists(_ context.Context, key string) (bool, error) { return f.exists[key], nil }

type fakeStore struct {
	lastSHA     map[string]string
	fetches     []FetchRecord
	documents   []RawDocument
	previous    map[string]map[string]works.Record
	appliedKeys []string
	changes     []works.Change
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		lastSHA:  map[string]string{},
		previous: map[string]map[string]works.Record{},
	}
}

func (f *fakeStore) LastDocumentSHA(_ context.Context, sourceID, endpoint string) (string, error) {
	return f.lastSHA[sourceID+"/"+endpoint], nil
}

func (f *fakeStore) RecordFetch(_ context.Context, rec FetchRecord) error {
	f.fetches = append(f.fetches, rec)
	return nil
}

func (f *fakeStore) SaveRawDocument(_ context.Context, doc RawDocument) (string, error) {
	f.documents = append(f.documents, doc)
	f.lastSHA[doc.SourceID+"/"+doc.Endpoint] = doc.SHA256
	return "doc-" + doc.SHA256[:8], nil
}

func (f *fakeStore) LoadWorks(_ context.Context, sourceID, endpoint string) (map[string]works.Record, error) {
	return f.previous[sourceID+"/"+endpoint], nil
}

func (f *fakeStore) ApplySnapshot(_ context.Context, in SnapshotWrite) error {
	for _, rec := range in.Records {
		f.appliedKeys = append(f.appliedKeys, rec.NaturalKey)
	}
	f.changes = append(f.changes, in.Changes...)
	return nil
}

func roadsServer(t *testing.T, bodies ...string) (*httptest.Server, *int) {
	t.Helper()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := bodies[min(calls, len(bodies)-1)]
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

const snapshotBodyV1 = `{"count":1,"data":[{"roadName":"Road A","ward":"R/S","contractorName":"X",
 "status":"In Progress","contractorRepMobile":"9999999999"}]}`

const snapshotBodyV2 = `{"count":1,"data":[{"roadName":"Road A","ward":"R/S","contractorName":"X",
 "status":"Completed","contractorRepMobile":"9999999999"}]}`

func newTestRunner(srvClient *http.Client, arch Archive, st Store) *Runner {
	f := NewFetcher(srvClient, "test-agent")
	f.Retries = 1
	f.Backoff = time.Millisecond
	return &Runner{Fetcher: f, Archive: arch, Store: st, Now: func() time.Time {
		return time.Date(2026, 9, 23, 6, 0, 0, 0, time.UTC)
	}}
}

func TestSnapshotArchivesAndRecordsFirstRun(t *testing.T) {
	srv, _ := roadsServer(t, snapshotBodyV1)
	arch, st := newFakeArchive(), newFakeStore()
	r := newTestRunner(srv.Client(), arch, st)

	job := Job{SourceID: "bmc_roads_api", Blocklist: []string{"contractorRepMobile"}, Endpoints: []Endpoint{
		{Name: "publicdashboard", URL: srv.URL, Parse: works.ParseRoadsDashboard, Ext: ".json"},
	}}

	summary, err := r.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(arch.put) != 1 {
		t.Errorf("expected one archived object, got %d", len(arch.put))
	}
	if len(st.documents) != 1 {
		t.Fatalf("expected one raw document, got %d", len(st.documents))
	}
	if len(st.fetches) != 1 || st.fetches[0].Unchanged {
		t.Errorf("the attempt must be logged as changed: %+v", st.fetches)
	}
	if summary.Changed != 1 || summary.Unchanged != 0 {
		t.Errorf("summary: %+v", summary)
	}
	if len(st.changes) != 1 || st.changes[0].Kind != works.KindAdded {
		t.Errorf("first run should record an addition: %+v", st.changes)
	}
}

func TestSnapshotSkipsArchiveWhenBytesAreUnchanged(t *testing.T) {
	srv, calls := roadsServer(t, snapshotBodyV1, snapshotBodyV1)
	arch, st := newFakeArchive(), newFakeStore()
	r := newTestRunner(srv.Client(), arch, st)
	job := Job{SourceID: "bmc_roads_api", Endpoints: []Endpoint{
		{Name: "publicdashboard", URL: srv.URL, Parse: works.ParseRoadsDashboard, Ext: ".json"},
	}}

	if _, err := r.Run(context.Background(), job); err != nil {
		t.Fatalf("first run: %v", err)
	}
	archivedAfterFirst := len(arch.put)

	summary, err := r.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	if *calls != 2 {
		t.Errorf("expected two fetches, got %d", *calls)
	}
	if len(arch.put) != archivedAfterFirst {
		t.Errorf("unchanged bytes must not be archived again: %d objects", len(arch.put))
	}
	if len(st.documents) != 1 {
		t.Errorf("unchanged bytes must not create a second raw document: %d", len(st.documents))
	}
	if summary.Unchanged != 1 {
		t.Errorf("summary should report the unchanged endpoint: %+v", summary)
	}
	if len(st.fetches) != 2 || !st.fetches[1].Unchanged {
		t.Errorf("every attempt is logged, marked unchanged: %+v", st.fetches)
	}
}

func TestSnapshotRecordsFieldLevelChangeBetweenRuns(t *testing.T) {
	srv, _ := roadsServer(t, snapshotBodyV1, snapshotBodyV2)
	arch, st := newFakeArchive(), newFakeStore()
	r := newTestRunner(srv.Client(), arch, st)
	job := Job{SourceID: "bmc_roads_api", Blocklist: []string{"contractorRepMobile"}, Endpoints: []Endpoint{
		{Name: "publicdashboard", URL: srv.URL, Parse: works.ParseRoadsDashboard, Ext: ".json"},
	}}

	if _, err := r.Run(context.Background(), job); err != nil {
		t.Fatalf("first run: %v", err)
	}
	// The store now holds the first snapshot.
	st.previous["bmc_roads_api/publicdashboard"] = map[string]works.Record{
		"Road A|R/S|X": {NaturalKey: "Road A|R/S|X", Fields: map[string]any{
			"roadName": "Road A", "ward": "R/S", "contractorName": "X", "status": "In Progress",
		}},
	}
	st.changes = nil

	if _, err := r.Run(context.Background(), job); err != nil {
		t.Fatalf("second run: %v", err)
	}

	if len(st.changes) != 1 {
		t.Fatalf("expected one change, got %+v", st.changes)
	}
	c := st.changes[0]
	if c.Kind != works.KindChanged || c.Field != "status" || c.Old != "In Progress" || c.New != "Completed" {
		t.Errorf("got %+v", c)
	}
}

func TestSnapshotNeverPassesBlocklistedFieldsToTheStore(t *testing.T) {
	srv, _ := roadsServer(t, snapshotBodyV1)
	arch, st := newFakeArchive(), newFakeStore()
	r := newTestRunner(srv.Client(), arch, st)
	job := Job{SourceID: "bmc_roads_api", Blocklist: []string{"contractorRepMobile"}, Endpoints: []Endpoint{
		{Name: "publicdashboard", URL: srv.URL, Parse: works.ParseRoadsDashboard, Ext: ".json"},
	}}

	var captured []SnapshotWrite
	r.Store = &capturingStore{fakeStore: st, captured: &captured}

	if _, err := r.Run(context.Background(), job); err != nil {
		t.Fatalf("run: %v", err)
	}

	for _, w := range captured {
		for _, rec := range w.Records {
			if _, present := rec.Fields["contractorRepMobile"]; present {
				t.Fatalf("blocklisted field reached the store: %+v", rec.Fields)
			}
		}
	}
	// The archived bytes are the untouched original, by design.
	for _, body := range arch.put {
		if !strings.Contains(string(body), "contractorRepMobile") {
			t.Error("the archive must keep the artefact exactly as retrieved")
		}
	}
}

func TestSnapshotLogsTheAttemptWhenFetchFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	arch, st := newFakeArchive(), newFakeStore()
	r := newTestRunner(srv.Client(), arch, st)
	job := Job{SourceID: "bmc_roads_api", Endpoints: []Endpoint{
		{Name: "publicdashboard", URL: srv.URL, Parse: works.ParseRoadsDashboard, Ext: ".json"},
	}}

	summary, err := r.Run(context.Background(), job)

	if err == nil {
		t.Fatal("a failing endpoint must surface an error")
	}
	if len(st.fetches) != 1 {
		t.Fatalf("the failed attempt must still be logged: %+v", st.fetches)
	}
	if st.fetches[0].Error == "" || st.fetches[0].StatusCode != 404 {
		t.Errorf("failure detail missing: %+v", st.fetches[0])
	}
	if summary.Failed != 1 {
		t.Errorf("summary should count the failure: %+v", summary)
	}
}

func TestSnapshotContinuesAfterOneEndpointFails(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()
	good, _ := roadsServer(t, snapshotBodyV1)

	arch, st := newFakeArchive(), newFakeStore()
	r := newTestRunner(good.Client(), arch, st)
	job := Job{SourceID: "bmc_roads_api", Endpoints: []Endpoint{
		{Name: "broken", URL: bad.URL, Parse: works.ParseRoadsDashboard, Ext: ".json"},
		{Name: "publicdashboard", URL: good.URL, Parse: works.ParseRoadsDashboard, Ext: ".json"},
	}}

	summary, _ := r.Run(context.Background(), job)

	if summary.Failed != 1 || summary.Changed != 1 {
		t.Errorf("one failure must not stop the other endpoint: %+v", summary)
	}
}

type capturingStore struct {
	*fakeStore
	captured *[]SnapshotWrite
}

func (c *capturingStore) ApplySnapshot(ctx context.Context, in SnapshotWrite) error {
	*c.captured = append(*c.captured, in)
	return c.fakeStore.ApplySnapshot(ctx, in)
}
