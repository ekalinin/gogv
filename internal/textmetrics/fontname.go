// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package textmetrics

import "strings"

// familyNames maps the font names that the reference recognizes for its
// Times, Helvetica and Courier metrics, in lower case, to their family.
var familyNames = map[string]Family{
	"times":            Serif,
	"times-roman":      Serif,
	"times new roman":  Serif,
	"liberation serif": Serif,
	"helvetica":        Sans,
	"arial":            Sans,
	"liberation sans":  Sans,
	"arimo":            Sans,
	"nimbus sans":      Sans,
	"freesans":         Sans,
	"courier":          Mono,
	"courier new":      Mono,
	"liberation mono":  Mono,
	"cousine":          Mono,
	"nimbus mono":      Mono,
	"freemono":         Mono,
}

// notSupportedNames are, in lower case, the font names for which the
// reference has metrics that gogv does not have yet.
var notSupportedNames = map[string]bool{
	"dejavu sans":  true,
	"verdana":      true,
	"georgia":      true,
	"consolas":     true,
	"calibri":      true,
	"trebuchet ms": true,
	"open sans":    true,
}

// Lookup returns the face that measures text set in fontname, as the
// reference chooses it: a whole family name, compared without regard to
// case, selects that family; any other name, including one with a weight or
// a slant such as "Helvetica-Bold", selects Times. supported is false for the
// names whose own metrics the reference has and gogv does not have yet; their
// text is measured with Times.
func Lookup(fontname string) (face Face, supported bool) {
	name := strings.ToLower(fontname)
	if family, ok := familyNames[name]; ok {
		return Face{Family: family}, true
	}
	return Face{Family: Serif}, !notSupportedNames[name]
}
