// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/attrs.html
// Provenance: doc https://graphviz.org/docs/attr-types/bool/

package layout

import (
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/attrs"
	"github.com/ekalinin/gogv/internal/graph"
)

// pointsPerInch converts the inches of the size attributes to points.
const pointsPerInch = 72

// parseNumber reads the longest prefix of s that is a number, as the
// reference reads numeric attributes, and reports whether there was one.
func parseNumber(s string) (float64, bool) {
	s = strings.TrimLeft(s, " \t\n")
	end := 0
	seenDigit, seenDot, seenExp := false, false, false
	for end < len(s) {
		c := s[end]
		switch {
		case c >= '0' && c <= '9':
			seenDigit = true
		case (c == '+' || c == '-') && (end == 0 || s[end-1] == 'e' || s[end-1] == 'E'):
		case c == '.' && !seenDot && !seenExp:
			seenDot = true
		case (c == 'e' || c == 'E') && seenDigit && !seenExp:
			seenExp = true
		default:
			goto done
		}
		end++
	}
done:
	for end > 0 {
		if v, err := strconv.ParseFloat(s[:end], 64); err == nil {
			return v, true
		}
		end--
	}
	return 0, false
}

// number returns a numeric attribute value, def when it is empty or not a
// number, and min when it is smaller.
func number(v graph.Value, def, min float64) float64 {
	f, ok := parseNumber(v.S)
	if !ok {
		return def
	}
	return max(f, min)
}

// boolean reads a boolean attribute: "true", "yes" or a non-zero integer is
// true, "false", "no" or zero false, anything else def.
func boolean(v graph.Value, def bool) bool {
	s := strings.ToLower(strings.TrimSpace(v.S))
	switch s {
	case "true", "yes":
		return true
	case "false", "no":
		return false
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n != 0
	}
	return def
}

// supported returns the value when gogv supports it for the attribute, or
// def otherwise.
func supported(k graph.Kind, name string, v graph.Value, def string) string {
	if v.S == "" || v.HTML || !attrs.IsSupported(k, name, v.S) {
		return def
	}
	return v.S
}

func nodeValue(g *graph.Graph, n *graph.Node, name string) graph.Value {
	v, _ := n.Attr(g, name)
	return v
}

func edgeValue(g *graph.Graph, e *graph.Edge, name string) graph.Value {
	v, _ := e.Attr(g, name)
	return v
}

func graphValue(g *graph.Graph, name string) graph.Value {
	v, _ := g.Attr(name)
	return v
}

// isSet reports whether an attribute has a value.
func isSet(v graph.Value) bool { return v.S != "" }
