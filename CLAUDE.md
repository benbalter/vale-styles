# CLAUDE.md

The `BenBalter` [Vale](https://vale.sh) style package: one YAML rule per file in [`BenBalter/`](BenBalter/). Consumers download a release's `BenBalter.zip` through `Packages` in their `.vale.ini`.

## Commands

- `script/lint && script/test` is what CI runs: yamllint on the rules, then Vale against [`testdata/`](testdata/). Both need `yamllint` and `vale` on `PATH`.
- `script/test` asserts that a few rules fire on their fixtures and that [`testdata/clean.md`](testdata/clean.md) trips no rule at all. When you add or tighten a rule, add a fixture that trips it plus an `assert_fires` line, and keep `clean.md` clean.

## Releasing

Pushing a `v*` tag runs [`release.yml`](.github/workflows/release.yml), which zips `BenBalter/` and publishes the release. Consumers only get changes when they bump the version in their `Packages` URL. Tag only after the owner's explicit go-ahead, and bump the version in the README's usage example in the same change. Push branches with `--no-follow-tags`.

## Gotchas

- A rule's file name is its ID (`BenBalter.<Name>`), which consumers reference in their own `.vale.ini`. Don't rename rules, even the misspelled `SentanceSpacing`.
- An invalid rule doesn't just miss matches: Vale refuses to load the config, which breaks linting for every consumer. Vale 3 dropped `scope: link`, for example; [`MeaningfulLinkWords.yml`](BenBalter/MeaningfulLinkWords.yml) matches link syntax with a `raw` scope instead.
- [`CulturalInclusion.yml`](BenBalter/CulturalInclusion.yml) is a `substitution` rule: each `swap` key is a case-insensitive regex, and the value is the suggested replacement.
