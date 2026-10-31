// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import "math"

// The arithmetic below writes every product that feeds an addition as
// float64(a*b), so that the compiler cannot fuse it into a multiply-add and
// results stay the same on every architecture.

func add(p, q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }
func sub(p, q Point) Point { return Point{p.X - q.X, p.Y - q.Y} }

func scale(p Point, k float64) Point { return Point{float64(p.X * k), float64(p.Y * k)} }

// half is x/2 rounded on its own: the compiler turns a division by 2 into
// a multiplication, which it could otherwise fuse with an addition.
func half(x float64) float64 { return float64(x / 2) }

// hypot is sqrt(x^2 + y^2) with exact operations only.
func hypot(x, y float64) float64 { return math.Sqrt(float64(x*x) + float64(y*y)) }

func dist(p, q Point) float64 { return hypot(p.X-q.X, p.Y-q.Y) }

func lerp(p, q Point, t float64) Point {
	return Point{p.X + float64((q.X-p.X)*t), p.Y + float64((q.Y-p.Y)*t)}
}

// cross is the z component of (b - a) x (c - a).
func cross(a, b, c Point) float64 {
	return float64((b.X-a.X)*(c.Y-a.Y)) - float64((b.Y-a.Y)*(c.X-a.X))
}

func normalize(p Point) Point {
	l := hypot(p.X, p.Y)
	if l == 0 {
		return Point{}
	}
	return Point{p.X / l, p.Y / l}
}

// cubic evaluates a cubic Bezier curve at t by de Casteljau's construction.
func cubic(p [4]Point, t float64) Point {
	a, b, c := lerp(p[0], p[1], t), lerp(p[1], p[2], t), lerp(p[2], p[3], t)
	d, e := lerp(a, b, t), lerp(b, c, t)
	return lerp(d, e, t)
}

// splitCubic splits a cubic Bezier curve at t.
func splitCubic(p [4]Point, t float64) (left, right [4]Point) {
	a, b, c := lerp(p[0], p[1], t), lerp(p[1], p[2], t), lerp(p[2], p[3], t)
	d, e := lerp(a, b, t), lerp(b, c, t)
	m := lerp(d, e, t)
	return [4]Point{p[0], a, d, m}, [4]Point{m, e, c, p[3]}
}

// cubicExtent returns the bounding box of a cubic Bezier curve, from the
// extrema of its coordinates.
func cubicExtent(p [4]Point) (minP, maxP Point) {
	minP = Point{math.Min(p[0].X, p[3].X), math.Min(p[0].Y, p[3].Y)}
	maxP = Point{math.Max(p[0].X, p[3].X), math.Max(p[0].Y, p[3].Y)}
	for _, axis := range []int{0, 1} {
		c := func(q Point) float64 {
			if axis == 0 {
				return q.X
			}
			return q.Y
		}
		p0, p1, p2, p3 := c(p[0]), c(p[1]), c(p[2]), c(p[3])
		// The derivative is a quadratic a t^2 + b t + k.
		a := -p0 + float64(3*p1) - float64(3*p2) + p3
		b := float64(2*p0) - float64(4*p1) + float64(2*p2)
		k := -p0 + p1
		var roots []float64
		if math.Abs(a) < 1e-12 {
			if b != 0 {
				roots = append(roots, -k/b)
			}
		} else if disc := float64(b*b) - float64(4*a*k); disc >= 0 {
			s := math.Sqrt(disc)
			roots = append(roots, (-b+s)/(2*a), (-b-s)/(2*a))
		}
		for _, t := range roots {
			if t <= 0 || t >= 1 {
				continue
			}
			q := cubic(p, t)
			minP = Point{math.Min(minP.X, q.X), math.Min(minP.Y, q.Y)}
			maxP = Point{math.Max(maxP.X, q.X), math.Max(maxP.Y, q.Y)}
		}
	}
	return minP, maxP
}
