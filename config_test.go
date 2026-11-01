package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// zipDir builds a package zip the way release.yml does: the directory itself
// is the archive's one top-level folder.
func zipDir(t *testing.T, dir string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		f, err := w.Create(filepath.ToSlash(p))
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// lintAsConsumer syncs a consumer config that lists BenBalterConfig.zip
// (served locally, so the test covers the zip layout and `vale sync`, not just
// the .vale.ini) and lints doc with it. A config Vale can't load fails here,
// which matters because it would break linting for every consumer.
func lintAsConsumer(t *testing.T, packages, basedOn, extra, doc string) map[string]bool {
	t.Helper()
	pkg := zipDir(t, "BenBalterConfig")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(pkg)
	}))
	defer srv.Close()

	dir := t.TempDir()
	if packages != "" {
		packages += ", "
	}
	cfg := "StylesPath = styles\nPackages = " + packages + srv.URL + "/BenBalterConfig.zip\n\n" +
		"[*.md]\nBasedOnStyles = " + basedOn + "\n" + extra
	for name, body := range map[string]string{".vale.ini": cfg, "doc.md": doc} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "styles"), 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) []byte {
		cmd := exec.Command("vale", append([]string{"--no-global"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("vale %v: %v\n%s", args, err, out)
		}
		return out
	}
	run("sync")
	var results map[string][]valeAlert
	if err := json.Unmarshal(run("--output=JSON", "--no-exit", "doc.md"), &results); err != nil {
		t.Fatal(err)
	}
	checks := map[string]bool{}
	for _, a := range results["doc.md"] {
		checks[a.Check] = true
	}
	return checks
}

// TestConfigPackage checks that BenBalterConfig turns its rules off for a
// consumer that loads write-good and proselint, leaves their other rules on,
// and still lets the consumer's own .vale.ini turn a rule back on.
func TestConfigPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("downloads write-good and proselint")
	}
	doc := "The plan was approved by the team, and it is very unique.\n"
	baseline := []string{"write-good.E-Prime", "write-good.Passive", "write-good.TooWordy"}

	got := lintAsConsumer(t, "write-good, proselint", "write-good, proselint", "", doc)
	for _, c := range baseline {
		if got[c] {
			t.Errorf("%s still fires with BenBalterConfig loaded", c)
		}
	}
	if !got["proselint.Uncomparables"] {
		t.Errorf("proselint.Uncomparables didn't fire, so the package switched off more than its list (got %v)", got)
	}

	got = lintAsConsumer(t, "write-good, proselint", "write-good, proselint", "write-good.Passive = YES\n", doc)
	if !got["write-good.Passive"] {
		t.Errorf("a consumer's write-good.Passive = YES didn't override the package (got %v)", got)
	}
}

// TestConfigPackageWithoutItsStyles covers a consumer that loads none of the
// styles the package configures: Vale has to load the config anyway.
func TestConfigPackageWithoutItsStyles(t *testing.T) {
	lintAsConsumer(t, "", "Vale", "", "Plain prose.\n")
}
