// SPDX-License-Identifier: MIT
// Provenance: original

package gogv

import (
	"fmt"
	"html"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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

var (
	releaseNotesPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)\.md$`)
	resultsPattern      = regexp.MustCompile(`(\d+) cases: (\d+) passed, (\d+) known differences, (\d+) not supported yet, (\d+) failed`)
	// versionPattern matches the mentions of a release of gogv.
	versionPattern = regexp.MustCompile(`\bv\d+\.\d+\.\d+\b|gogv \d+\.\d+\.\d+`)
)

// newestRelease returns the newest release among the names of release notes,
// as "v0.3.0.md", by semantic version.
func newestRelease(names []string) string {
	var newest string
	var best [3]int
	for _, name := range names {
		m := releaseNotesPattern.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		var v [3]int
		for i := range v {
			v[i], _ = strconv.Atoi(m[i+1])
		}
		if newest == "" || slices.Compare(v[:], best[:]) > 0 {
			newest, best = strings.TrimSuffix(name, ".md"), v
		}
	}
	return newest
}

// landingReleaseProblems checks the page against the release notes, given
// by file name: the page must name the newest release, name no other one,
// and give the conformance results of its notes.
func landingReleaseProblems(page string, notes map[string]string) []string {
	var problems []string
	release := pageRelease(page)
	newest := newestRelease(slices.Collect(maps.Keys(notes)))
	if release != newest {
		problems = append(problems, fmt.Sprintf("the page names %q, but the newest release notes are those of %q", release, newest))
	}
	for _, mention := range versionPattern.FindAllString(page, -1) {
		if !strings.HasSuffix(mention, release) && !strings.HasSuffix(mention, " "+strings.TrimPrefix(release, "v")) {
			problems = append(problems, fmt.Sprintf("the page mentions %q, not the release %q", mention, release))
		}
	}
	want := resultsPattern.FindStringSubmatch(strings.Join(strings.Fields(notes[release+".md"]), " "))
	if want == nil {
		return append(problems, fmt.Sprintf("the notes of %q state no conformance results", release))
	}
	section := pageSection(page, "conformance")
	var got []string
	for _, p := range []string{
		`<strong class="cases">(\d+)</strong> cases`,
		`<strong>(\d+)</strong> cases pass`,
		`<strong>(\d+)</strong> cases are known differences`,
		`<strong>(\d+)</strong> cases use features that are not supported yet`,
		`<strong>(\d+)</strong> cases fail`,
	} {
		m := regexp.MustCompile(p).FindStringSubmatch(section)
		if m == nil {
			got = append(got, "")
			continue
		}
		got = append(got, m[1])
	}
	if !slices.Equal(got, want[1:]) {
		problems = append(problems, fmt.Sprintf("the conformance results of the page (cases, passed, known differences, not supported yet, failed) are %q, but those of the notes of %s are %q", got, release, want[1:]))
	}
	return problems
}

func TestLandingReleaseScenarios(t *testing.T) {
	page := landingPage(t)
	entries, err := os.ReadDir("docs/releases")
	if err != nil {
		t.Fatal(err)
	}
	notes := map[string]string{}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join("docs/releases", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		notes[e.Name()] = string(data)
	}
	if problems := landingReleaseProblems(page, notes); len(problems) > 0 {
		t.Fatalf("the page and the release notes disagree:\n%s", strings.Join(problems, "\n"))
	}
	release := pageRelease(page)

	t.Run("Page behind the notes", func(t *testing.T) {
		ahead := maps.Clone(notes)
		ahead["v9.0.0.md"] = notes[release+".md"]
		if len(landingReleaseProblems(page, ahead)) == 0 {
			t.Error("newer release notes than the release of the page give no problem")
		}
		// A mention of another release in the text of the page.
		other := strings.Replace(page, "cmd/dot@"+release, "cmd/dot@v0.0.1", 1)
		if len(landingReleaseProblems(other, notes)) == 0 {
			t.Error("an installation command at another release gives no problem")
		}
	})

	t.Run("Conformance results out of date", func(t *testing.T) {
		section := pageSection(page, "conformance")
		for _, m := range regexp.MustCompile(`<strong[^>]*>(\d+)</strong>`).FindAllStringSubmatch(section, -1) {
			changed := strings.Replace(section, m[0], strings.Replace(m[0], ">"+m[1]+"<", ">"+m[1]+"1<", 1), 1)
			if len(landingReleaseProblems(strings.Replace(page, section, changed, 1), notes)) == 0 {
				t.Errorf("changing %s on the page gives no problem", m[0])
			}
		}
	})
}

func TestPlaygroundFirstExample(t *testing.T) {
	// The first example of the playground, its initial source, is the
	// example of the README.
	js, err := os.ReadFile("site/playground.mjs")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile("(?s)export const examples = \\[\\s*\\{[^`]*source: `((?:[^`\\\\]|\\\\.)*)`").FindSubmatch(js)
	if m == nil {
		t.Fatal("no first example in site/playground.mjs")
	}
	source := strings.NewReplacer("\\\\", "\\", "\\`", "`", "\\$", "$").Replace(string(m[1]))
	want, err := os.ReadFile("docs/example.gv")
	if err != nil {
		t.Fatal(err)
	}
	if source != string(want) {
		t.Errorf("the first example of the playground is not docs/example.gv:\n%s", source)
	}
}

// pagesWorkflowProblems checks the text of the pages workflow: it has no
// YAML parser to use, as the module has no dependencies.
func pagesWorkflowProblems(workflow string) []string {
	var problems []string
	for _, want := range []string{
		"\n  push:\n    branches: [main]\n",
		"\n    tags: [\"v*\"]\n",
		"\n  workflow_dispatch:\n",
		"\n  pages: write\n",
		"\n  id-token: write\n",
		"\n          ref: main\n",
		"\n          fetch-depth: 0\n",
		"\n        run: go run ./tools/site -release -out _site\n",
	} {
		if !strings.Contains(workflow, want) {
			problems = append(problems, fmt.Sprintf("no %q", strings.TrimSpace(want)))
		}
	}
	for _, path := range []string{"site/**", "tools/site/**", ".github/workflows/pages.yml", "docs/example.svg", "LICENSE", "NOTICE"} {
		if !strings.Contains(workflow, "\n      - \""+path+"\"\n") {
			problems = append(problems, fmt.Sprintf("no path %q", path))
		}
	}
	if strings.Contains(workflow, "pull_request") {
		problems = append(problems, "a pull_request trigger, which would deploy unmerged pages")
	}
	steps := strings.Split(workflow, "\n      - ")
	for _, action := range []string{"actions/configure-pages@", "actions/upload-pages-artifact@", "actions/deploy-pages@"} {
		found := false
		for _, step := range steps {
			if strings.Contains(step, action) {
				found = true
				if !strings.Contains(step, "\n        if: steps.site.outputs.deploy == 'true'\n") {
					problems = append(problems, fmt.Sprintf("the step of %s does not depend on deploy=true", strings.TrimSuffix(action, "@")))
				}
			}
		}
		if !found {
			problems = append(problems, fmt.Sprintf("no step of %s", strings.TrimSuffix(action, "@")))
		}
	}
	return problems
}

func TestPagesWorkflow(t *testing.T) {
	// The triggers of the spec scenarios "Release tag pushed after the
	// notes" and "Change of the notices" of landing-page, and the
	// conditions of the deployment.
	data, err := os.ReadFile(".github/workflows/pages.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	if problems := pagesWorkflowProblems(workflow); len(problems) > 0 {
		t.Fatalf(".github/workflows/pages.yml:\n%s", strings.Join(problems, "\n"))
	}
	for _, m := range []struct{ name, old, new string }{
		{"no trigger on tags", "    tags: [\"v*\"]\n", ""},
		{"a pull_request trigger", "  workflow_dispatch:\n", "  workflow_dispatch:\n  pull_request:\n"},
		{"no id-token: write", "  id-token: write\n", ""},
		{"a deployment not conditional on deploy=true", "        id: deployment\n        if: steps.site.outputs.deploy == 'true'\n", "        id: deployment\n"},
		{"a checkout without ref: main", "          ref: main\n", ""},
		{"NOTICE missing from the paths", "      - \"NOTICE\"\n", ""},
	} {
		mutated := strings.Replace(workflow, m.old, m.new, 1)
		if mutated == workflow {
			t.Errorf("%s: the mutation does not apply", m.name)
			continue
		}
		if len(pagesWorkflowProblems(mutated)) == 0 {
			t.Errorf("%s: no problem found", m.name)
		}
	}
}
