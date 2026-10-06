# CLAUDE.md

Two [Vale](https://vale.sh) style packages, one YAML rule per file: the house rules in [`BenBalter/`](BenBalter/) and the AI-tell rules in [`AIPatterns/`](AIPatterns/) (moved here from ben.balter.com so the blog, the book, and other repos share one copy). Consumers download a release's `BenBalter.zip` and/or `AIPatterns.zip` through `Packages` in their `.vale.ini`.

## Commands

- `script/lint && script/test` is what CI runs: yamllint and gofmt, then [go-cmdtest](https://github.com/google/go-cmdtest) golden-file tests ([`main_test.go`](main_test.go)). They need `yamllint`, `go`, and `vale` on `PATH`. Keep the local Vale version in step with `VALE_VERSION` in [`ci.yml`](.github/workflows/ci.yml), since output can change between releases.
- Every rule needs a [`fixtures/<Style>.<Rule>/`](fixtures/) (a `.vale.ini` enabling only that rule, plus a `test.md` with `## Flag`, `## Leave alone`, and any `## Known false positives` or `## Known misses`) and a matching [`testdata/<Style>.<Rule>.ct`](testdata/) case; `TestEveryRuleHasACase` fails otherwise. After changing a rule or fixture, run `script/test -update` and review the `.ct` diff line by line before committing: it's the record of exactly what each rule flags. [`fixtures/clean`](fixtures/clean/) runs every rule at once and must stay empty.

## Releasing

Pushing a `v*` tag runs [`release.yml`](.github/workflows/release.yml), which zips `BenBalter/` and `AIPatterns/` and publishes the release. Consumers only get changes when they bump the version in their `Packages` URL. Tag only after the owner's explicit go-ahead, and bump the version in the README's usage example in the same change. Push branches with `--no-follow-tags`.

## Gotchas

- A rule's file name is its ID (`BenBalter.<Name>` or `AIPatterns.<Name>`), which consumers reference in their own `.vale.ini`. Don't rename rules, even the misspelled `SentanceSpacing`.
- An invalid rule doesn't just miss matches: Vale refuses to load the config, which breaks linting for every consumer. Vale 3 dropped `scope: link`, for example; [`MeaningfulLinkWords.yml`](BenBalter/MeaningfulLinkWords.yml) matches link syntax with a `raw` scope instead.
- [`CulturalInclusion.yml`](BenBalter/CulturalInclusion.yml) is a `substitution` rule: each `swap` key is a case-insensitive regex, and the value is the suggested replacement.
