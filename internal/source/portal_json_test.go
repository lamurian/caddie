package source

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

func TestExtractJSONFixture(t *testing.T) {
	p, err := LoadPortal(filepath.Join(portalDir, "example-json.yml"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(portalDir, "example-json", "listing.json"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := extractJSON(p, body)
	if err != nil {
		t.Fatal(err)
	}
	want := []portalItem{
		{"id": "9007199254740993123", "title": "Senior Economist, Climate Finance",
			"url": "https://jobs.example.net/careers/job/9007199254740993123", "location": "Manila",
			"deadline": "2026-11-30T23:59:59Z", "salary": "USD 120,000 - 150,000 per year"},
		{"id": "4502", "title": "Procurement Officer", "location": "Abidjan, Nairobi", "deadline": "2026-12-15"},
		{"id": "4504", "title": "Data Analyst (Consultant)",
			"url": "https://jobs.example.net/careers/job/4504?lang=en"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestExtractJSONRootArray(t *testing.T) {
	p := Portal{Org: "root", ListingURL: "https://jobs.example.net/api/list",
		Fields: PortalFields{Title: "name", URL: "href", Location: "where", Salary: "paid"}}
	body := `[{"name": "Analyst", "href": "jobs/1", "where": [" Rome ", 2, null, {"x": 1}], "paid": true},
		{"name": null, "href": "jobs/2"}, {"name": "Clerk", "where": {"city": "Oslo"}}]`
	got, err := extractJSON(p, []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	want := []portalItem{
		{"title": "Analyst", "url": "https://jobs.example.net/api/jobs/1", "location": "Rome, 2", "salary": "true"},
		{"title": "Clerk"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestExtractJSONArrayIndex(t *testing.T) {
	p := Portal{Org: "taleo", ListingURL: "https://jobs.example.net/api", JSONPath: "data.1.list",
		Fields: PortalFields{ID: "no", Title: "column.0", Location: "column.1", Deadline: "column.9"}}
	body := `{"data": [{}, {"list": [{"no": "7", "column": ["Analyst", "[\"Rome\"]"]}, {"column": "Clerk"}]}]}`
	got, err := extractJSON(p, []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	want := []portalItem{{"id": "7", "title": "Analyst", "location": "Rome"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestExtractJSONErrors(t *testing.T) {
	cases := []struct{ path, body, want string }{
		{"data.results", `{"data": {"items": []}}`, `key "results" not found in "data"`},
		{"data.results", `{"data": [1]}`, `"data" is an array, not an object`},
		{"data.results", `[]`, `the document root is an array, not an object`},
		{"data.results", `{"data": {"results": {"a": 1}}}`, `is an object, not an array`},
		{"data.results", `{"data": {"results": [1]}}`, `item 0 is a number, not an object`},
		{"data.results", `<html>`, `invalid JSON`},
		{"", `{"results": []}`, `is an object, not an array`},
		{"data.2", `{"data": [[], []]}`, `index 2 out of range in "data" (length 2)`},
		{"data.-1", `{"data": [[]]}`, `"data" is an array, not an object`},
		{"data.0.x", `{"data": ["s"]}`, `"data.0" is a string, not an object`},
	}
	for _, c := range cases {
		p := Portal{Org: "bank", ListingURL: "https://jobs.example.net/api?key=SECRET",
			JSONPath: c.path, Fields: PortalFields{Title: "title"}}
		_, err := extractJSON(p, []byte(c.body))
		if !errors.Is(err, ErrFetch) {
			t.Errorf("%s: got %v, want ErrFetch", c.body, err)
			continue
		}
		prefix := "portal:bank: json_path " + `"` + c.path + `": `
		if msg := err.Error(); !strings.Contains(msg, prefix) || !strings.Contains(msg, c.want) ||
			strings.Contains(msg, "SECRET") {
			t.Errorf("%s: got %q, want %q and %q", c.body, msg, prefix, c.want)
		}
	}
}

func TestJSONText(t *testing.T) {
	cases := map[string]string{
		`" a \n b "`: "a b", `12345678901234567890`: "12345678901234567890", `1.5`: "1.5",
		`false`: "false", `null`: "", `{"a": "b"}`: "", `["x", "", ["y"], "z"]`: "x, z",
		`"[\"Geneva\", \"Nairobi\"]"`: "Geneva, Nairobi", `"[Draft] Analyst"`: "[Draft] Analyst",
		`"[\"a\"] [\"b\"]"`: `["a"] ["b"]`, `"{\"a\": 1}"`: `{"a": 1}`,
	}
	for in, want := range cases {
		dec := json.NewDecoder(strings.NewReader(in))
		dec.UseNumber()
		var v any
		if err := dec.Decode(&v); err != nil {
			t.Fatal(err)
		}
		if got := jsonText(v); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestLoadPortalInvalidJSONPaths(t *testing.T) {
	cases := map[string]string{ // file in invalid/ → field the error names
		"json-path-empty-key": `json_path: "data..results" has an empty key`,
		"json-field-attr":     "fields.url: @attr is HTML syntax",
	}
	for name, want := range cases {
		path := filepath.Join(portalDir, "invalid", name+".yml")
		_, err := LoadPortal(path)
		if !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), path+": "+want) {
			t.Errorf("%s: got %v, want ErrConfig and %q", name, err, want)
		}
	}
}
