// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/attrs.html
// Provenance: doc https://graphviz.org/doc/info/shapes.html
// Provenance: doc https://graphviz.org/doc/info/arrows.html

package attrs

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/graph"
)

// spec is what gogv supports of an attribute: the values it accepts (all
// when values is nil), as a comma-separated list of them when list is set,
// or those that valid accepts, and the documented default it uses in place
// of a value it does not support.
type spec struct {
	values []string
	list   bool
	color  bool
	valid  func(string) bool
	def    string
}

func oneOf(def string, values ...string) spec { return spec{values: values, def: def} }

// listOf accepts a comma-separated list of values, as style takes them.
func listOf(def string, values ...string) spec { return spec{values: values, list: true, def: def} }

// atLeastZero accepts a number of 0 or more, as penwidth takes it.
var atLeastZero = spec{valid: func(v string) bool {
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	return err == nil && f >= 0
}}

// wholeNumber accepts a whole number of 0 or more, as weight takes it.
var wholeNumber = spec{valid: func(v string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return err == nil && n >= 0
}}

// Styles returns the values of a style list.
func Styles(value string) []string {
	var out []string
	for v := range strings.SplitSeq(value, ",") {
		out = append(out, strings.TrimSpace(v))
	}
	return out
}

var anyValue = spec{}

var colorValue = spec{color: true, def: "black"}

// Shapes are the node shapes gogv draws, with their documented synonyms.
var Shapes = []string{"rect", "box", "rectangle", "box3d", "plaintext", "point", "ellipse", "oval", "circle", "diamond", "hexagon", "octagon"}

var arrowTypes = []string{"none", "empty", "normal"}

// supported lists, per kind of object, the attributes and values that the
// layout acts on (the dot-language spec, "Supported attributes").
var supported = [3]map[string]spec{
	graph.GraphKind: {
		"label":      anyValue,
		"fontname":   anyValue,
		"fontsize":   anyValue,
		"fontcolor":  colorValue,
		"labeljust":  oneOf("c", "c", "l", "r"),
		"style":      oneOf("solid", "solid"),
		"color":      colorValue,
		"rank":       oneOf("", "same", "min", "max", "source", "sink"),
		"nodesep":    anyValue,
		"ranksep":    anyValue,
		"rankdir":    oneOf("TB", "TB", "LR"),
		"splines":    oneOf("spline", "spline", "true", "polyline"),
		"remincross": anyValue,
		"searchsize": anyValue,
	},
	graph.NodeKind: {
		"label":     anyValue,
		"xlabel":    anyValue,
		"fontname":  anyValue,
		"fontsize":  anyValue,
		"fontcolor": colorValue,
		"shape":     oneOf("ellipse", Shapes...),
		"width":     anyValue,
		"height":    anyValue,
		"color":     colorValue,
		"fillcolor": spec{color: true, def: "lightgrey"},
		"style":     listOf("solid", "solid", "filled", "rounded"),
		"tooltip":   anyValue,
		"id":        anyValue,
	},
	graph.EdgeKind: {
		"label":        anyValue,
		"xlabel":       anyValue,
		"headlabel":    anyValue,
		"taillabel":    anyValue,
		"fontname":     anyValue,
		"fontsize":     anyValue,
		"fontcolor":    colorValue,
		"sametail":     anyValue,
		"tailport":     anyValue,
		"headport":     anyValue,
		"minlen":       anyValue,
		"color":        colorValue,
		"style":        listOf("solid", "solid", "invis", "dotted"),
		"penwidth":     atLeastZero,
		"weight":       wholeNumber,
		"constraint":   anyValue,
		"arrowhead":    oneOf("normal", arrowTypes...),
		"arrowtail":    oneOf("normal", arrowTypes...),
		"dir":          oneOf("forward", "forward", "back", "both", "none"),
		"arrowsize":    anyValue,
		"tooltip":      anyValue,
		"labeltooltip": anyValue,
	},
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// IsDocumented reports whether the Graphviz documentation describes the
// attribute for objects of kind k.
func IsDocumented(k graph.Kind, name string) bool {
	used, ok := documented[name]
	if !ok {
		return false
	}
	switch k {
	case graph.GraphKind:
		return used&(UsedByGraphs|UsedBySubgraphs|UsedByClusters) != 0
	case graph.NodeKind:
		return used&UsedByNodes != 0
	}
	return used&UsedByEdges != 0
}

// IsSupported reports whether gogv acts on the attribute with that value for
// objects of kind k.
func IsSupported(k graph.Kind, name, value string) bool {
	s, ok := supported[k][name]
	switch {
	case !ok:
		return false
	case s.color:
		return hexColor.MatchString(value)
	case s.list:
		for _, v := range Styles(value) {
			if !slices.Contains(s.values, v) {
				return false
			}
		}
		return true
	case s.values != nil:
		return slices.Contains(s.values, value)
	case s.valid != nil:
		return s.valid(value)
	}
	return true
}

// defaultOf returns the documented default that replaces an unsupported
// value of a supported attribute, or "" when there is none to name.
func defaultOf(k graph.Kind, name string) string {
	return supported[k][name].def
}
