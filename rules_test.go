package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// rule is the subset of a Vale rule file these tests inspect.
type rule struct {
	ID         string
	Extends    string            `yaml:"extends"`
	Message    string            `yaml:"message"`
	Level      string            `yaml:"level"`
	Link       string            `yaml:"link"`
	IgnoreCase bool              `yaml:"ignorecase"`
	Tokens     []string          `yaml:"tokens"`
	Raw        []string          `yaml:"raw"`
	Swap       map[string]string `yaml:"swap"`
}

func loadRules(t *testing.T) []rule {
	t.Helper()
	var rules []rule
	for _, style := range styles {
		paths, err := filepath.Glob(filepath.Join(style, "*.yml"))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			var r rule
			if err := yaml.Unmarshal(data, &r); err != nil {
				t.Fatalf("%s: %v", p, err)
			}
			r.ID = style + "." + strings.TrimSuffix(filepath.Base(p), ".yml")
			rules = append(rules, r)
		}
	}
	return rules
}

// patterns returns a rule's match patterns: tokens, raw, or swap keys.
func (r rule) patterns() []string {
	ps := append(append([]string{}, r.Tokens...), r.Raw...)
	for k := range r.Swap {
		ps = append(ps, k)
	}
	sort.Strings(ps)
	return ps
}

// TestRuleMetadata keeps every rule's metadata consistent: a known level, a
// link to its source, and a message that says what matched whenever a rule has
// more than one pattern (a single-pattern rule's message can name it outright).
func TestRuleMetadata(t *testing.T) {
	levels := map[string]bool{"suggestion": true, "warning": true, "error": true}
	for _, r := range loadRules(t) {
		if !levels[r.Level] {
			t.Errorf("%s: level %q isn't suggestion, warning, or error", r.ID, r.Level)
		}
		style, name, _ := strings.Cut(r.ID, ".")
		want := "https://github.com/benbalter/vale-styles/blob/master/" + style + "/" + name + ".yml"
		if r.Link != want {
			t.Errorf("%s: link is %q, want %q", r.ID, r.Link, want)
		}
		if len(r.patterns()) > 1 && !strings.Contains(r.Message, "%") {
			t.Errorf("%s: message doesn't say what matched (no %%s)", r.ID)
		}
	}
}

// TestReadmeListsEveryRule guards the README's rule tables, a hand-kept copy
// of what's in BenBalter/ and AIPatterns/: every rule needs a row, every row
// needs a rule, and the level column has to match the rule file.
func TestReadmeListsEveryRule(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	row := regexp.MustCompile("(?m)^\\| `(\\w+)` \\| (\\w+) \\|")
	sections := map[string]string{"BenBalter": "## Rules\n", "AIPatterns": "## AIPatterns rules\n"}
	readme := string(data)

	rules := map[string]rule{}
	for _, r := range loadRules(t) {
		rules[r.ID] = r
	}

	seen := map[string]bool{}
	for style, heading := range sections {
		start := strings.Index(readme, heading)
		if start < 0 {
			t.Fatalf("README has no %q section", strings.TrimSpace(heading))
		}
		section := readme[start+len(heading):]
		if end := strings.Index(section, "\n## "); end >= 0 {
			section = section[:end]
		}
		for _, m := range row.FindAllStringSubmatch(section, -1) {
			id := style + "." + m[1]
			seen[id] = true
			r, ok := rules[id]
			if !ok {
				t.Errorf("README lists %s, which doesn't exist", id)
				continue
			}
			if m[2] != r.Level {
				t.Errorf("README says %s is %s; the rule says %s", id, m[2], r.Level)
			}
		}
	}
	for id := range rules {
		if !seen[id] {
			t.Errorf("README doesn't list %s", id)
		}
	}
}

// literal reports whether a pattern is plain text once the curly-apostrophe
// class is folded back to an apostrophe, and returns that text.
func literal(p string) (string, bool) {
	p = strings.ReplaceAll(p, "['’]", "'")
	if strings.ContainsAny(p, `\[](){}?*+|^$.`) {
		return "", false
	}
	return p, true
}

// allowedOverlap lists rules that may flag words another rule owns. So and
// But check a sentence-opening conjunction, which is grammar, not the AI tell
// ("So how do we ...") RhetoricalQuestions flags, and consumers like the book
// turn them off; they'd collide with any token that starts with "So" or "But".
var allowedOverlap = map[string]bool{"BenBalter.So": true, "BenBalter.But": true}

type valeAlert struct {
	Check string
	Line  int
	Span  [2]int
	Match string
}

// TestEveryLiteralTokenFires covers the long lists the fixtures only sample
// (Cliches alone has ~700 entries). It writes every plain-text token of every
// rule into one document, one paragraph each, runs all rules at once, and
// checks two things per paragraph: the token's own rule flags it, and no other
// rule flags the same words (a double alert for one phrase).
func TestEveryLiteralTokenFires(t *testing.T) {
	type entry struct{ rule, token string }
	var entries []entry
	for _, r := range loadRules(t) {
		if r.Extends != "existence" && r.Extends != "substitution" {
			continue
		}
		for _, p := range r.patterns() {
			if tok, ok := literal(p); ok {
				entries = append(entries, entry{r.ID, tok})
			}
		}
	}

	dir := t.TempDir()
	var doc strings.Builder
	lineOf := map[int]entry{}
	line := 1
	for _, e := range entries {
		lineOf[line] = e
		fmt.Fprintf(&doc, "%s placeholder text\n\n", e.token)
		line += 2
	}
	cfg := fmt.Sprintf("StylesPath = %s\nMinAlertLevel = suggestion\n\n[*.md]\nBasedOnStyles = %s\n",
		mustAbs(t, "."), strings.Join(styles, ", "))
	for name, body := range map[string]string{".vale.ini": cfg, "tokens.md": doc.String()} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cmd := exec.Command("vale", "--no-global", "--output=JSON", "--no-exit", "tokens.md")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "VALE_STYLES_PATH="+mustAbs(t, "."))
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("vale: %v\n%s", err, out)
	}
	var results map[string][]valeAlert
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatalf("parse vale output: %v\n%s", err, out)
	}

	byLine := map[int][]valeAlert{}
	for _, alerts := range results {
		for _, a := range alerts {
			byLine[a.Line] = append(byLine[a.Line], a)
		}
	}
	for ln, e := range lineOf {
		end := len([]rune(e.token))
		fired := false
		var others []string
		for _, a := range byLine[ln] {
			if a.Check == e.rule {
				fired = true
			} else if a.Span[0] <= end && !allowedOverlap[a.Check] {
				others = append(others, a.Check)
			}
		}
		if !fired {
			t.Errorf("%s: %q doesn't fire", e.rule, e.token)
		}
		if len(others) > 0 {
			sort.Strings(others)
			t.Errorf("%s: %q is also flagged by %s", e.rule, e.token, strings.Join(others, ", "))
		}
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
