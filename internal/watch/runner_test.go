package watch

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/ingest"
)

type fakeArchive struct{ put map[string][]byte }

func (f *fakeArchive) Put(_ context.Context, key string, body []byte, _ string) error {
	if f.put == nil {
		f.put = map[string][]byte{}
	}
	f.put[key] = body
	return nil
}

func (f *fakeArchive) Exists(_ context.Context, key string) (bool, error) {
	_, ok := f.put[key]
	return ok, nil
}

type fakeGRStore struct {
	known map[string]bool
	saved []GRRecord
	docs  []ingest.RawDocument
}

func (f *fakeGRStore) KnownGRs(_ context.Context, sanketanks []string) (map[string]bool, error) {
	out := map[string]bool{}
	for _, s := range sanketanks {
		if f.known[s] {
			out[s] = true
		}
	}
	return out, nil
}

func (f *fakeGRStore) SaveGR(_ context.Context, rec GRRecord) error {
	f.saved = append(f.saved, rec)
	return nil
}

func (f *fakeGRStore) SaveRawDocument(_ context.Context, doc ingest.RawDocument) (string, error) {
	f.docs = append(f.docs, doc)
	return "doc-" + doc.SHA256[:6], nil
}

// mirror serves the three endpoints the watcher uses.
func mirror(t *testing.T, grText string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "advancedsearch"):
			fmt.Fprint(w, `{"response":{"numFound":2,"docs":[
			 {"identifier":"in.gov.maharashtra.gr.111","addeddate":"2026-09-23T01:55:41Z","title":"Maharashtra GR: #111"},
			 {"identifier":"in.gov.maharashtra.gr.222","addeddate":"2026-09-23T01:45:18Z","title":"Maharashtra GR: #222"}]}}`)
		case strings.Contains(r.URL.Path, "/metadata/"):
			id := strings.TrimPrefix(r.URL.Path, "/metadata/")
			num := strings.TrimPrefix(id, GRIdentifierPrefix)
			fmt.Fprintf(w, `{"metadata":{"title":"Maharashtra GR: #%s","date":"22-09-2026"},
			 "files":[{"name":"%s.pdf","format":"Text PDF"},{"name":"%s_djvu.txt","format":"DjVuTXT"}]}`, num, num, num)
		case strings.HasSuffix(r.URL.Path, ".pdf"):
			fmt.Fprintf(w, "%%PDF-1.4 fake pdf for %s", r.URL.Path)
		case strings.HasSuffix(r.URL.Path, "_djvu.txt"):
			fmt.Fprint(w, grText)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestWatcher(srv *httptest.Server, arch Archive, st GRStore) *GRWatcher {
	f := ingest.NewFetcher(srv.Client(), "test-agent")
	f.Retries = 1
	f.Backoff = time.Millisecond
	w := NewGRWatcher(f, arch, st)
	w.SearchURL = srv.URL + "/advancedsearch.php"
	w.MetadataBase = srv.URL + "/metadata/"
	w.DownloadBase = srv.URL + "/download/"
	w.Limit = 10
	return w
}

func TestGRWatcherArchivesNewResolutions(t *testing.T) {
	srv := mirror(t, "ordinary resolution text")
	arch, st := &fakeArchive{}, &fakeGRStore{known: map[string]bool{}}
	w := newTestWatcher(srv, arch, st)

	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if result.New != 2 {
		t.Errorf("expected 2 new resolutions, got %d", result.New)
	}
	if len(arch.put) != 2 {
		t.Errorf("each new GR's PDF should be archived: %d objects", len(arch.put))
	}
	if len(st.saved) != 2 {
		t.Fatalf("each new GR should be recorded: %d", len(st.saved))
	}
	if st.saved[0].Sanketank != "111" {
		t.Errorf("sanketank: got %q", st.saved[0].Sanketank)
	}
	if st.saved[0].IssuedOn == nil {
		t.Error("the GR's own date should be recorded")
	}
	if st.saved[0].DocumentID == "" {
		t.Error("the record must point at the archived document")
	}
}

func TestGRWatcherSkipsResolutionsAlreadySeen(t *testing.T) {
	srv := mirror(t, "text")
	arch := &fakeArchive{}
	st := &fakeGRStore{known: map[string]bool{"111": true, "222": true}}
	w := newTestWatcher(srv, arch, st)

	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if result.New != 0 {
		t.Errorf("nothing is new: %+v", result)
	}
	if len(arch.put) != 0 {
		t.Errorf("known resolutions must not be downloaded again: %v", arch.put)
	}
}

func TestGRWatcherFlagsWatchedTerms(t *testing.T) {
	srv := mirror(t, "हा शासन निर्णय माहिती अधिकार शुल्क सुधारतो")
	arch, st := &fakeArchive{}, &fakeGRStore{known: map[string]bool{}}
	w := newTestWatcher(srv, arch, st)
	w.Keywords = []string{"माहिती अधिकार"}

	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(result.Hits) != 2 {
		t.Fatalf("both resolutions mention the term: %+v", result.Hits)
	}
	if len(result.Hits[0].Keywords) == 0 {
		t.Error("the matched term should be reported")
	}
	if len(st.saved) == 0 || len(st.saved[0].Keywords) == 0 {
		t.Error("matches should be stored with the record, for a human to read later")
	}
}

func TestGRWatcherRespectsItsLimit(t *testing.T) {
	srv := mirror(t, "text")
	arch, st := &fakeArchive{}, &fakeGRStore{known: map[string]bool{}}
	w := newTestWatcher(srv, arch, st)
	w.Limit = 1

	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.New != 1 {
		t.Errorf("the limit caps how many are fetched per run: %+v", result)
	}
}

func TestGRWatcherContinuesWhenOneItemFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "advancedsearch"):
			fmt.Fprint(w, `{"response":{"docs":[
			 {"identifier":"in.gov.maharashtra.gr.111","addeddate":"2026-09-23T01:55:41Z","title":"a"},
			 {"identifier":"in.gov.maharashtra.gr.222","addeddate":"2026-09-23T01:45:18Z","title":"b"}]}}`)
		case strings.Contains(r.URL.Path, "/metadata/in.gov.maharashtra.gr.111"):
			w.WriteHeader(http.StatusNotFound)
		case strings.Contains(r.URL.Path, "/metadata/"):
			fmt.Fprint(w, `{"metadata":{"title":"b","date":"22-09-2026"},"files":[{"name":"222.pdf","format":"Text PDF"}]}`)
		case strings.HasSuffix(r.URL.Path, ".pdf"):
			fmt.Fprint(w, "%PDF-1.4")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	arch, st := &fakeArchive{}, &fakeGRStore{known: map[string]bool{}}
	w := newTestWatcher(srv, arch, st)

	result, err := w.Run(context.Background())

	if err == nil {
		t.Error("the failure should be reported")
	}
	if result.New != 1 {
		t.Errorf("the healthy item should still be archived: %+v", result)
	}
}
