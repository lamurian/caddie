package document

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The files Render writes into the temporary render directory for the
// template (ADR 0022).
const (
	BodyFile = "_caddie-body.md"
	DataFile = "_caddie.json"
)

// renderData is the shape of _caddie.json, read by Typst with
// json("_caddie.json"). Empty fields are omitted.
type renderData struct {
	Kind     string       `json:"kind"`
	Name     string       `json:"name,omitempty"`
	Location string       `json:"location,omitempty"`
	Email    string       `json:"email,omitempty"`
	Phone    string       `json:"phone,omitempty"`
	Links    []renderLink `json:"links,omitempty"`
	Position string       `json:"position,omitempty"`
	Company  string       `json:"company,omitempty"`
}

type renderLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// prepareRenderDir copies the template directory into dir and writes the
// body and data files. Contact details are written here and nowhere else.
func prepareRenderDir(dir string, in RenderInput, content []byte) error {
	if err := copyTree(filepath.Dir(in.Template), dir); err != nil {
		return err
	}
	var data bytes.Buffer
	enc := json.NewEncoder(&data)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(newRenderData(in)); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, DataFile), data.Bytes(), 0o600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, BodyFile), escapeBody(stripFrontmatter(content)), 0o600)
}

// escapeBody makes Quarto read the verified body literally, so nothing
// verify never saw reaches the PDF (ADR 0022). Quarto scans the included
// text before Pandoc does, in prose and code alike:
//   - "{{<" becomes "{\{<": no shortcode ({{< env >}}, {{< include >}},
//     {{< meta >}}) or escaped one ({{{< >}}}) is ever parsed, closed or
//     not, and Pandoc prints "{{<" again (in code the backslash shows);
//   - "<!--" becomes "<\!--": Quarto drops HTML comments before it looks
//     for YAML blocks, which could turn "<!-- -->---" into an opener;
//   - a "---" line becomes "----": still a rule or a setext underline, but
//     never a YAML metadata block (header-includes would be raw Typst).
func escapeBody(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte("{{<"), []byte(`{\{<`))
	body = bytes.ReplaceAll(body, []byte("<!--"), []byte(`<\!--`))
	lines := bytes.SplitAfter(body, []byte("\n"))
	for i, l := range lines {
		if string(bytes.TrimRight(l, " \t\r\n")) == "---" {
			lines[i] = append([]byte("-"), l...)
		}
	}
	return bytes.Join(lines, nil)
}

func newRenderData(in RenderInput) renderData {
	d := renderData{
		Kind:     string(in.Kind),
		Name:     strings.TrimSpace(in.Identity.Name),
		Location: strings.TrimSpace(in.Identity.Location),
		Email:    strings.TrimSpace(in.Contact.Email),
		Phone:    strings.TrimSpace(in.Contact.Phone),
		Position: strings.TrimSpace(in.Position),
		Company:  strings.TrimSpace(in.Company),
	}
	for label, url := range in.Identity.Links {
		if strings.TrimSpace(url) != "" {
			d.Links = append(d.Links, renderLink{Label: label, URL: strings.TrimSpace(url)})
		}
	}
	sort.Slice(d.Links, func(i, j int) bool { return d.Links[i].Label < d.Links[j].Label })
	return d
}

// stripFrontmatter removes a leading YAML block fenced by a "---" line and
// the next "---" or "..." line. Without a closing fence it changes nothing.
func stripFrontmatter(content []byte) []byte {
	lines := bytes.SplitAfter(content, []byte("\n"))
	if len(lines) == 0 || string(bytes.TrimSpace(lines[0])) != "---" {
		return content
	}
	for i := 1; i < len(lines); i++ {
		if t := string(bytes.TrimSpace(lines[i])); t == "---" || t == "..." {
			return bytes.Join(lines[i+1:], nil)
		}
	}
	return content
}

// copyTree copies every regular file under src into dst, keeping the
// layout and skipping dot-files and dot-directories.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == "." {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o700)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		return copyFile(path, filepath.Join(dst, rel))
	})
}

// copyFile copies src to dst, replacing dst if it exists.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()
	_, err = io.Copy(out, in)
	return err
}
