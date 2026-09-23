package works

import (
	"encoding/json"
	"os"
	"testing"
)

const roadsDashboardBody = `{
  "status": 200,
  "message": "Mega CC road project data with images",
  "count": 2,
  "data": [
    {"contractorName":"M/s Example Ltd. W-421","roadName":"Road A","ward":"R/S","status":"Completed",
     "length":400,"width":18.3,"startDate":"2025-11-01","endDate":"2026-01-10",
     "contractorRepName":"A Person","contractorRepMobile":"9999999999"},
    {"contractorName":"M/s Example Ltd. W-421","roadName":"Road B","ward":"R/C","status":"In Progress",
     "length":120,"width":9.15,"startDate":"2026-02-01","endDate":null}
  ]
}`

func TestParseRoadsDashboardReturnsOneRecordPerWork(t *testing.T) {
	recs, err := ParseRoadsDashboard([]byte(roadsDashboardBody), []string{"contractorRepName", "contractorRepMobile"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("got %d records, want 2", len(recs))
	}
	if recs[0].NaturalKey == "" {
		t.Error("record needs a natural key")
	}
	if recs[0].NaturalKey == recs[1].NaturalKey {
		t.Error("different works must not share a natural key")
	}
	if got := recs[0].Fields["status"]; got != "Completed" {
		t.Errorf("status: got %v", got)
	}
}

func TestParseRoadsDashboardDropsBlocklistedFields(t *testing.T) {
	recs, err := ParseRoadsDashboard([]byte(roadsDashboardBody), []string{"contractorRepName", "contractorRepMobile"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, rec := range recs {
		for _, banned := range []string{"contractorRepName", "contractorRepMobile"} {
			if _, present := rec.Fields[banned]; present {
				t.Errorf("record %s kept blocklisted field %q", rec.NaturalKey, banned)
			}
		}
		// And it must not survive anywhere in the serialised record either.
		encoded, _ := json.Marshal(rec.Fields)
		if containsAny(string(encoded), "A Person", "9999999999") {
			t.Errorf("personal data survived in %s: %s", rec.NaturalKey, encoded)
		}
	}
}

func TestParseRoadsDashboardRejectsDuplicateNaturalKeys(t *testing.T) {
	body := `{"count":2,"data":[
	  {"roadName":"Same","ward":"R/S","contractorName":"X"},
	  {"roadName":"Same","ward":"R/S","contractorName":"X"}]}`
	_, err := ParseRoadsDashboard([]byte(body), nil)
	if err == nil {
		t.Fatal("a key collision must be an error, not a silent merge")
	}
}

func TestParseRoadsDashboardRejectsUnexpectedShape(t *testing.T) {
	if _, err := ParseRoadsDashboard([]byte(`{"count":1}`), nil); err == nil {
		t.Error("missing data array should error")
	}
	if _, err := ParseRoadsDashboard([]byte(`not json`), nil); err == nil {
		t.Error("invalid json should error")
	}
}

func TestParseRoadsDashboardAgainstArchivedSample(t *testing.T) {
	raw, err := os.ReadFile("../../docs/01-research/sources/2026-08-24-bmc-roads-sample-25.json")
	if err != nil {
		t.Skipf("archived sample unavailable: %v", err)
	}
	var sample struct {
		Records []map[string]any `json:"records"`
	}
	if err := json.Unmarshal(raw, &sample); err != nil {
		t.Fatalf("read sample: %v", err)
	}
	body, err := json.Marshal(map[string]any{"count": len(sample.Records), "data": sample.Records})
	if err != nil {
		t.Fatal(err)
	}

	recs, err := ParseRoadsDashboard(body, []string{"contractorRepName", "contractorRepMobile"})
	if err != nil {
		t.Fatalf("parse archived sample: %v", err)
	}
	if len(recs) != len(sample.Records) {
		t.Errorf("got %d records from the archived sample, want %d", len(recs), len(sample.Records))
	}
}

func TestParseSWDProgressCardIsASingleRecord(t *testing.T) {
	body := `{"Data":{"TargetQuantity":"833298.17","TotalQuantity":"924335.65","TodayTrips":"7"},
	          "Msg":"Successful","ServiceResponse":1}`
	recs, err := ParseSWDProgressCard([]byte(body), nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("got %d records, want 1", len(recs))
	}
	if recs[0].NaturalKey != "progresscard" {
		t.Errorf("natural key: got %q", recs[0].NaturalKey)
	}
	if recs[0].Fields["TotalQuantity"] != "924335.65" {
		t.Errorf("fields: got %v", recs[0].Fields)
	}
}

func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		for i := 0; i+len(n) <= len(haystack); i++ {
			if haystack[i:i+len(n)] == n {
				return true
			}
		}
	}
	return false
}
