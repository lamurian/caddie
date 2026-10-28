package source

import (
	"strconv"
	"strings"
)

// ldDetail maps a schema.org JobPosting object to a Detail.
func ldDetail(o map[string]any) Detail {
	desc := str(o["description"])
	j := Job{
		Title:    str(o["title"]),
		Company:  orgName(o["hiringOrganization"]),
		Location: ldLocation(o["jobLocation"]),
		Deadline: Date(str(o["validThrough"])),
		Posted:   Date(str(o["datePosted"])),
	}
	if strings.EqualFold(str(o["jobLocationType"]), "TELECOMMUTE") {
		j.WorkMode = "wfh"
	} else {
		j.WorkMode = WorkMode(j.Location, j.Title)
	}
	ldSalary(obj(o["baseSalary"])).apply(&j)
	return Detail{Job: j, HTML: desc}
}

func orgName(v any) string {
	if m := obj(v); m != nil {
		return str(m["name"])
	}
	return str(v)
}

func ldLocation(v any) string {
	list, ok := v.([]any)
	if !ok {
		list = []any{v}
	}
	var places []string
	for _, x := range list {
		if p := ldPlace(x); p != "" {
			places = append(places, p)
		}
	}
	return strings.Join(places, "; ")
}

func ldPlace(v any) string {
	m := obj(v)
	if m == nil {
		return str(v)
	}
	addr := obj(m["address"])
	if addr == nil {
		return firstNonEmpty(str(m["address"]), str(m["name"]))
	}
	var parts []string
	for _, k := range []string{"addressLocality", "addressRegion", "addressCountry"} {
		p := str(addr[k])
		if c := obj(addr[k]); c != nil {
			p = str(c["name"])
		}
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, ", ")
}

var ldPeriods = map[string]string{"YEAR": "year", "MONTH": "month", "HOUR": "hour"}

func ldSalary(m map[string]any) Salary {
	if m == nil {
		return Salary{}
	}
	s := Salary{Currency: str(m["currency"])}
	val := obj(m["value"])
	if val == nil {
		s.Min = int(num(m["value"]))
		s.Max = s.Min
	} else {
		s.Min, s.Max = int(num(val["minValue"])), int(num(val["maxValue"]))
		if v := int(num(val["value"])); s.Min == 0 && v > 0 {
			s.Min, s.Max = v, v
		}
		s.Period = ldPeriods[strings.ToUpper(str(val["unitText"]))]
	}
	if s.Min == 0 {
		return Salary{}
	}
	if s.Max == 0 {
		s.Max = s.Min
	}
	return s
}

func obj(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func str(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	}
	return ""
}

func num(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	}
	return 0
}
