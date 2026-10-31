// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package output

import (
	"strconv"
	"strings"
)

// svgNumber writes a coordinate of the SVG output: at most two decimals,
// without trailing zeros, as the reference writes them.
func svgNumber(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "-0" {
		return "0"
	}
	return s
}

// plainNumber writes a number of the plain output: five significant digits,
// as C's %.5g does.
func plainNumber(v float64) string {
	s := strconv.FormatFloat(v, 'g', 5, 64)
	if s == "-0" {
		return "0"
	}
	return s
}

// xmlText escapes text for the SVG output as the reference does: the XML
// special characters, '-' (which may not appear twice in a comment), and
// every space that follows a space; an '&' that starts an entity is kept.
func xmlText(s string) string {
	var b strings.Builder
	prevSpace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '&' && startsEntity(s[i:]):
			b.WriteByte('&')
		case c == '&':
			b.WriteString("&amp;")
		case c == '<':
			b.WriteString("&lt;")
		case c == '>':
			b.WriteString("&gt;")
		case c == '"':
			b.WriteString("&quot;")
		case c == '\'':
			b.WriteString("&#39;")
		case c == '-':
			b.WriteString("&#45;")
		case c == ' ' && prevSpace:
			b.WriteString("&#160;")
		default:
			b.WriteByte(c)
		}
		prevSpace = c == ' '
	}
	return b.String()
}

// startsEntity reports whether s starts with an XML entity or character
// reference: &name; &#123; or &#x1F;.
func startsEntity(s string) bool {
	end := strings.IndexByte(s, ';')
	if end < 2 {
		return false
	}
	body := s[1:end]
	if strings.HasPrefix(body, "#x") || strings.HasPrefix(body, "#X") {
		body = body[2:]
		if body == "" {
			return false
		}
		for _, c := range body {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(body, "#") {
		body = body[1:]
		if body == "" {
			return false
		}
		for _, c := range body {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	for _, c := range body {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// svgFont is how the SVG output names a font.
type svgFont struct {
	family, weight, style, stretch string
}

// postScriptFonts maps the PostScript names of the Times, Helvetica and
// Courier faces to the SVG font attributes the reference writes for them.
var postScriptFonts = map[string]svgFont{
	"Times-Roman":                  {family: "Times,serif"},
	"Times-Bold":                   {family: "Times,serif", weight: "bold"},
	"Times-Italic":                 {family: "Times,serif", style: "italic"},
	"Times-BoldItalic":             {family: "Times,serif", weight: "bold", style: "italic"},
	"Helvetica":                    {family: "Helvetica,sans-Serif"},
	"Helvetica-Bold":               {family: "Helvetica,sans-Serif", weight: "bold"},
	"Helvetica-Oblique":            {family: "Helvetica,sans-Serif", style: "oblique"},
	"Helvetica-BoldOblique":        {family: "Helvetica,sans-Serif", weight: "bold", style: "oblique"},
	"Helvetica-Narrow":             {family: "Helvetica,sans-Serif", stretch: "condensed"},
	"Helvetica-Narrow-Bold":        {family: "Helvetica,sans-Serif", weight: "bold", stretch: "condensed"},
	"Helvetica-Narrow-Oblique":     {family: "Helvetica,sans-Serif", style: "oblique", stretch: "condensed"},
	"Helvetica-Narrow-BoldOblique": {family: "Helvetica,sans-Serif", weight: "bold", style: "oblique", stretch: "condensed"},
	"Courier":                      {family: "Courier,monospace"},
	"Courier-Bold":                 {family: "Courier,monospace", weight: "bold"},
	"Courier-Oblique":              {family: "Courier,monospace", style: "oblique"},
	"Courier-BoldOblique":          {family: "Courier,monospace", weight: "bold", style: "oblique"},
}

// fontAttrs returns the font attributes of a text element.
func fontAttrs(fontname string) string {
	f, ok := postScriptFonts[fontname]
	if !ok {
		f = svgFont{family: fontname}
	}
	s := ` font-family="` + xmlText(f.family) + `"`
	if f.weight != "" {
		s += ` font-weight="` + f.weight + `"`
	}
	if f.style != "" {
		s += ` font-style="` + f.style + `"`
	}
	if f.stretch != "" {
		s += ` font-stretch="` + f.stretch + `"`
	}
	return s
}
