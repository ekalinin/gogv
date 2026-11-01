// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/attrs.html
// Provenance: doc https://graphviz.org/doc/info/shapes.html
// Provenance: doc https://graphviz.org/doc/info/arrows.html

package attrs

import (
	"regexp"
	"slices"

	"github.com/ekalinin/gogv/internal/graph"
)

// spec is what gogv supports of an attribute: the values it accepts (all
// when values is nil) and the documented default it uses in place of a
// value it does not support.
type spec struct {
	values []string
	color  bool
	def    string
}

func oneOf(def string, values ...string) spec { return spec{values: values, def: def} }

var anyValue = spec{}

var colorValue = spec{color: true, def: "black"}

// Shapes are the node shapes gogv draws, with their documented synonyms.
var Shapes = []string{"rect", "box", "rectangle", "plaintext", "point", "ellipse", "oval", "circle", "diamond", "hexagon", "octagon"}

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
		"style":     oneOf("solid", "solid", "rounded"),
	},
	graph.EdgeKind: {
		"label":      anyValue,
		"xlabel":     anyValue,
		"headlabel":  anyValue,
		"taillabel":  anyValue,
		"fontname":   anyValue,
		"fontsize":   anyValue,
		"fontcolor":  colorValue,
		"sametail":   anyValue,
		"tailport":   anyValue,
		"headport":   anyValue,
		"minlen":     anyValue,
		"color":      colorValue,
		"style":      oneOf("solid", "solid", "invis"),
		"constraint": anyValue,
		"arrowhead":  oneOf("normal", arrowTypes...),
		"arrowtail":  oneOf("normal", arrowTypes...),
		"dir":        oneOf("forward", "forward", "back", "both", "none"),
		"arrowsize":  anyValue,
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
	case s.values != nil:
		return slices.Contains(s.values, value)
	}
	return true
}

// defaultOf returns the documented default that replaces an unsupported
// value of a supported attribute, or "" when there is none to name.
func defaultOf(k graph.Kind, name string) string {
	return supported[k][name].def
}
