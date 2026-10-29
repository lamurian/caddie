package source

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

func TestExtractHTMLExample(t *testing.T) {
	p, err := LoadPortal(filepath.Join(portalDir, "example-html.yml"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(portalDir, "example-html", "listing.html"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := extractHTML(p, body)
	if err != nil {
		t.Fatal(err)
	}
	want := []portalItem{
		{"id": "EHO-2041", "title": "Technical Officer, Health Data Systems",
			"url": "https://careers.example.org/job?id=EHO-2041", "location": "Geneva, Switzerland",
			"deadline": "Closing: 30 November 2026", "salary": "P-3"},
		{"id": "EHO-2057", "title": "Epidemiologist",
			"url": "https://careers.example.org/job?id=EHO-2057", "location": "Remote",
			"deadline": "15 December 2026", "salary": ""},
		{"id": "EHO-2063", "title": "Programme Assistant (Fixed-term)",
			"url": "https://careers.example.org/job?id=EHO-2063", "location": "",
			"deadline": "8 January 2027", "salary": "G-5"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %v\nwant %v", got, want)
	}
}

func TestExtractHTMLFields(t *testing.T) {
	body := []byte(`<table><tr class="job" data-ref=" R1 "><td><a href="mailto:jobs@example.org">Analyst</a></td></tr>
		<tr class="job" data-ref="R2"><td>Unlinked</td></tr><tr class="job"><td>No ref</td></tr></table>`)
	p := Portal{File: "t.yml", ListingURL: "https://careers.example.org/list", ItemSelector: "tr.job",
		Fields: PortalFields{ID: "@DATA-REF", Title: "td", URL: `a[href*="@"]@href`, Location: "td@title"}}
	got, err := extractHTML(p, body)
	if err != nil {
		t.Fatal(err)
	}
	want := []portalItem{
		{"id": "R1", "title": "Analyst", "url": "mailto:jobs@example.org", "location": ""},
		{"id": "R2", "title": "Unlinked", "url": "", "location": ""},
		{"id": "", "title": "No ref", "url": "", "location": ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %v\nwant %v", got, want)
	}
}

func TestExtractHTMLBadSelector(t *testing.T) {
	base := Portal{File: "portals/x.yml", ListingURL: "https://careers.example.org/",
		ItemSelector: "li", Fields: PortalFields{Title: "a", URL: "a@href"}}
	badItem, badField, badAttr := base, base, base
	badItem.ItemSelector = "ul >> li"
	badField.Fields.Location = "span:nope"
	badAttr.Fields.Salary = "[@grade"
	cases := map[string]Portal{"item_selector": badItem, "fields.location": badField, "fields.salary": badAttr}
	for field, p := range cases {
		_, err := extractHTML(p, []byte("<li><a>x</a></li>"))
		if !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), "portals/x.yml: "+field+": ") {
			t.Errorf("%s: got %v", field, err)
		}
	}
}

func TestLoadPortalBadSelector(t *testing.T) {
	cases := map[string]string{"bad-item-selector": "item_selector: ", "bad-field-selector": "fields.title: "}
	for name, want := range cases {
		path := filepath.Join(portalDir, "invalid", name+".yml")
		_, err := LoadPortal(path)
		if !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), path+": "+want) {
			t.Errorf("%s: got %v, want the file and %q", name, err, want)
		}
	}
}
