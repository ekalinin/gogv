// SPDX-License-Identifier: MIT
// Provenance: original

package gogv

import (
	"html"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// landingPage returns the HTML of the landing page.
func landingPage(t *testing.T) string {
	t.Helper()
	page, err := os.ReadFile("site/index.html")
	if err != nil {
		t.Fatal(err)
	}
	return string(page)
}

var metaReleasePattern = regexp.MustCompile(`<meta name="gogv-release" content="([^"]*)">`)

// pageRelease returns the release that the page names in its meta tag, or
// "" when it names none or more than one.
func pageRelease(page string) string {
	m := metaReleasePattern.FindAllStringSubmatch(page, -1)
	if len(m) != 1 {
		return ""
	}
	return m[0][1]
}

// pageSection returns the HTML of the element with the given id, from its
// opening tag to the end of the section or footer that it opens.
func pageSection(page, id string) string {
	start := regexp.MustCompile(`<(section|footer)[^>]* id="` + regexp.QuoteMeta(id) + `"`).FindStringSubmatchIndex(page)
	if start == nil {
		return ""
	}
	tag := page[start[2]:start[3]]
	end := strings.Index(page[start[0]:], "</"+tag+">")
	if end < 0 {
		return ""
	}
	return page[start[0] : start[0]+end]
}

var tagPattern = regexp.MustCompile(`<[^>]*>`)

// pageText returns the text of some HTML, without its tags and entities.
func pageText(fragment string) string {
	return html.UnescapeString(tagPattern.ReplaceAllString(fragment, ""))
}

func TestLandingPresentationScenarios(t *testing.T) {
	page := landingPage(t)
	release := pageRelease(page)
	if release == "" {
		t.Fatal(`the page names no release in one <meta name="gogv-release">`)
	}

	t.Run("Visit the page", func(t *testing.T) {
		if !strings.Contains(page, "<title>gogv - ") {
			t.Error("the title does not name gogv")
		}
		lede := pageText(regexp.MustCompile(`(?s)<p class="lede">.*?</p>`).FindString(page))
		if !strings.Contains(lede, "gogv is a Go implementation of the Graphviz dot program") {
			t.Errorf("the lede does not say what gogv is: %q", lede)
		}
		for _, link := range []string{
			`href="https://github.com/ekalinin/gogv"`,
			`href="https://github.com/ekalinin/gogv/releases"`,
			`href="https://pkg.go.dev/github.com/ekalinin/gogv"`,
		} {
			if !strings.Contains(page, link) {
				t.Errorf("no link %s", link)
			}
		}
	})

	t.Run("JavaScript disabled", func(t *testing.T) {
		for _, id := range []string{"ways", "playground", "pprof", "formats", "install", "conformance", "clean-room", "license"} {
			if strings.TrimSpace(pageText(pageSection(page, id))) == "" {
				t.Errorf("no section %q with text in the static HTML", id)
			}
		}
		// The only script is the module of the page; it shows the playground.
		if scripts := regexp.MustCompile(`<script[^>]*>`).FindAllString(page, -1); !slices.Equal(scripts, []string{`<script type="module" src="page.mjs">`}) {
			t.Errorf("scripts %q, want only page.mjs", scripts)
		}
		playground := pageSection(page, "playground")
		if !strings.Contains(playground, `<div class="pg" hidden>`) {
			t.Error("the playground is not hidden until its script shows it")
		}
		if !strings.Contains(pageText(regexp.MustCompile(`(?s)<noscript>.*?</noscript>`).FindString(playground)), "The playground needs JavaScript") {
			t.Error("the playground has no <noscript> message")
		}
		// No third-party assets: everything that the page loads is its own.
		for _, m := range regexp.MustCompile(`<(?:link|script|img)[^>]*(?:src|href)="([^"]*)"`).FindAllStringSubmatch(page, -1) {
			if strings.Contains(m[1], ":") || strings.HasPrefix(m[1], "/") {
				t.Errorf("the page loads %q, which is not one of its files", m[1])
			}
		}
	})

	t.Run("Output formats", func(t *testing.T) {
		var got []string
		list := regexp.MustCompile(`(?s)<ul class="formats">.*?</ul>`).FindString(pageSection(page, "formats"))
		for _, m := range regexp.MustCompile(`<li><code>([^<]*)</code></li>`).FindAllStringSubmatch(list, -1) {
			got = append(got, m[1])
		}
		if want := Formats(); !slices.Equal(got, want) {
			t.Errorf("formats on the page %q, want %q", got, want)
		}
	})

	t.Run("Example graph", func(t *testing.T) {
		src, err := os.ReadFile("docs/example.gv")
		if err != nil {
			t.Fatal(err)
		}
		m := regexp.MustCompile(`(?s)<pre class="sheet-source"><code>(.*?)</code></pre>`).FindStringSubmatch(page)
		if m == nil || html.UnescapeString(m[1]) != strings.TrimRight(string(src), "\n") {
			t.Error("the source of the example is not docs/example.gv")
		}
		svg, err := os.ReadFile("docs/example.svg")
		if err != nil {
			t.Fatal(err)
		}
		size := regexp.MustCompile(`<svg width="(\d+)pt" height="(\d+)pt"`).FindStringSubmatch(string(svg))
		if size == nil || !strings.Contains(page, `<img src="example.svg" width="`+size[1]+`" height="`+size[2]+`"`) {
			t.Error("the picture of the example is not example.svg at the size of docs/example.svg")
		}
	})

	t.Run("Installation", func(t *testing.T) {
		install := pageText(pageSection(page, "install"))
		for _, want := range []string{
			"go install github.com/ekalinin/gogv/cmd/dot@" + release,
			"go get github.com/ekalinin/gogv@" + release,
			"dot -V   # dot - graphviz version " + GraphvizVersion + " (gogv " + strings.TrimPrefix(release, "v") + ")",
		} {
			if !strings.Contains(install, want) {
				t.Errorf("the installation does not show %q", want)
			}
		}
	})
}
