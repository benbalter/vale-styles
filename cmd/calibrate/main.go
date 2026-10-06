// Command calibrate measures how often rules fire on a corpus, split by era,
// to tell an AI tell from the author's own voice: a good tell is near zero in
// writing from before ChatGPT and rises afterward. The corpus stays outside
// this repo; pass its path.
//
//	go run ./cmd/calibrate -rules AIPatterns.MicDrop,BenBalter.So ~/projects/blog/posts
//
// Files named with a YYYY-MM date prefix are split at -split (default
// 2022-12, ChatGPT's release); other files count toward "undated". Rates are
// hits per 1,000 words, with front matter and fenced code left out of the
// word count.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type alert struct{ Check, Match string }

var (
	frontMatter = regexp.MustCompile(`(?s)\A---\n.*?\n---\n`)
	codeFence   = regexp.MustCompile("(?s)```.*?```")
	datePrefix  = regexp.MustCompile(`^(\d{4}-\d{2})`)
)

func main() {
	rulesFlag := flag.String("rules", "", "comma-separated rule IDs to measure (default: every rule in both styles)")
	split := flag.String("split", "2022-12", "YYYY-MM that starts the \"after\" era")
	top := flag.Int("top", 5, "most common matches to show per rule")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: calibrate [-rules A.B,C.D] [-split YYYY-MM] <corpus-dir>...")
		os.Exit(2)
	}
	repo, err := os.Getwd()
	check(err)

	words := map[string]int{}
	files := map[string][]string{}
	for _, dir := range flag.Args() {
		paths, err := filepath.Glob(filepath.Join(dir, "*.md*"))
		check(err)
		for _, p := range paths {
			era := "undated"
			if m := datePrefix.FindStringSubmatch(filepath.Base(p)); m != nil {
				era = "before"
				if m[1] >= *split {
					era = "after"
				}
			}
			data, err := os.ReadFile(p)
			check(err)
			text := codeFence.ReplaceAllString(frontMatter.ReplaceAllString(string(data), ""), "")
			words[era] += len(strings.Fields(text))
			files[era] = append(files[era], p)
		}
	}

	cfg, err := os.CreateTemp("", "calibrate-*.ini")
	check(err)
	defer os.Remove(cfg.Name())
	fmt.Fprintf(cfg, "StylesPath = %s\nMinAlertLevel = suggestion\n\n[formats]\nmdx = md\n\n[*.{md,mdx}]\n", repo)
	if *rulesFlag == "" {
		fmt.Fprintln(cfg, "BasedOnStyles = BenBalter, AIPatterns")
	} else {
		for _, r := range strings.Split(*rulesFlag, ",") {
			fmt.Fprintf(cfg, "%s = YES\n", strings.TrimSpace(r))
		}
	}
	check(cfg.Close())

	hits := map[string]map[string][]string{} // rule -> era -> matches
	for era, paths := range files {
		cmd := exec.Command("vale", append([]string{"--no-global", "--config=" + cfg.Name(), "--output=JSON", "--no-exit"}, paths...)...)
		cmd.Env = append(os.Environ(), "VALE_STYLES_PATH="+repo)
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "vale: %v\n%s", err, out)
			os.Exit(1)
		}
		var results map[string][]alert
		check(json.Unmarshal(out, &results))
		for _, alerts := range results {
			for _, a := range alerts {
				if hits[a.Check] == nil {
					hits[a.Check] = map[string][]string{}
				}
				hits[a.Check][era] = append(hits[a.Check][era], strings.ToLower(a.Match))
			}
		}
	}

	eras := []string{"before", "after", "undated"}
	fmt.Printf("%-34s", "rule (hits per 1k words)")
	for _, e := range eras {
		if words[e] > 0 {
			fmt.Printf(" %8s", e)
		}
	}
	fmt.Println("  top matches")
	for _, e := range eras {
		if words[e] > 0 {
			fmt.Printf("  %s: %d files, %d words\n", e, len(files[e]), words[e])
		}
	}
	rules := make([]string, 0, len(hits))
	for r := range hits {
		rules = append(rules, r)
	}
	sort.Strings(rules)
	for _, r := range rules {
		fmt.Printf("%-34s", r)
		counts := map[string]int{}
		for _, e := range eras {
			if words[e] == 0 {
				continue
			}
			n := len(hits[r][e])
			fmt.Printf(" %4.2f/%-3d", 1000*float64(n)/float64(words[e]), n)
			for _, m := range hits[r][e] {
				counts[m]++
			}
		}
		matches := make([]string, 0, len(counts))
		for m := range counts {
			matches = append(matches, m)
		}
		sort.Slice(matches, func(i, j int) bool { return counts[matches[i]] > counts[matches[j]] })
		if len(matches) > *top {
			matches = matches[:*top]
		}
		for i, m := range matches {
			matches[i] = fmt.Sprintf("%s×%d", m, counts[m])
		}
		fmt.Println("  " + strings.Join(matches, ", "))
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
