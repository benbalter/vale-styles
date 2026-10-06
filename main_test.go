package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmdtest"
)

var update = flag.Bool("update", false, "replace test file contents with output")

var styles = []string{"BenBalter", "AIPatterns"}

// TestRules runs every testdata/*.ct case: each one cds into its fixture under
// fixtures/ and runs vale there, and the output has to match the .ct file line
// for line (rule, line, column, and message). Run `go test -update` to rewrite
// the .ct files after an intended change, then review the diff.
func TestRules(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// REPO lets each case cd into fixtures/<name> (go-cmdtest reserves ROOTDIR
	// for its own scratch directory). VALE_STYLES_PATH points
	// Vale's default styles directory at this checkout too; otherwise a style
	// installed globally (e.g. by `vale sync` elsewhere) loads first and the
	// tests silently run against that copy instead.
	t.Setenv("REPO", root)
	t.Setenv("VALE_STYLES_PATH", root)

	ts, err := cmdtest.Read("testdata")
	if err != nil {
		t.Fatal(err)
	}

	vale, err := exec.LookPath("vale")
	if err != nil {
		t.Fatal(err)
	}
	ts.Commands["vale"] = cmdtest.Program(vale)
	ts.Commands["cdf"] = cmdtest.InProcessProgram("cdf", cdf)

	ts.Run(t, *update)
}

// TestEveryRuleHasACase keeps the suite complete: a new rule without a fixture
// and .ct case fails here, as does a case left behind after a rule is removed.
func TestEveryRuleHasACase(t *testing.T) {
	rules := map[string]bool{}
	for _, style := range styles {
		paths, err := filepath.Glob(filepath.Join(style, "*.yml"))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range paths {
			rules[style+"."+strings.TrimSuffix(filepath.Base(p), ".yml")] = true
		}
	}
	if len(rules) == 0 {
		t.Fatal("no rules found")
	}

	for rule := range rules {
		for _, want := range []string{
			filepath.Join("testdata", rule+".ct"),
			filepath.Join("fixtures", rule, ".vale.ini"),
			filepath.Join("fixtures", rule, "test.md"),
		} {
			if _, err := os.Stat(want); err != nil {
				t.Errorf("%s has no %s", rule, want)
			}
		}
	}

	cases, err := filepath.Glob(filepath.Join("testdata", "*.ct"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		name := strings.TrimSuffix(filepath.Base(c), ".ct")
		if name == "clean" {
			continue
		}
		if !rules[name] {
			t.Errorf("%s doesn't match any rule", c)
			continue
		}

		// A rule whose regex silently stopped matching would regenerate to
		// an empty case with -update, so require at least one alert.
		data, err := os.ReadFile(c)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), ":"+name+":") {
			t.Errorf("%s expects no alerts; its fixture should trip %s at least once", c, name)
		}
	}
}
