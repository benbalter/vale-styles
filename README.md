# Vale styles

Personal prose style checks for [Vale](https://vale.sh) — the house rules behind
[ben.balter.com](https://ben.balter.com) and the book *Open and Async*. They
catch buzzwords, clichés, corporate-speak, and other tells so the writing stays
plain and direct.

## Usage

Add the package to your `.vale.ini` and enable the style:

```ini
StylesPath = .github/styles
Packages = https://github.com/benbalter/vale-styles/releases/download/v0.0.4/BenBalter.zip

[*.md]
BasedOnStyles = BenBalter
```

Then `vale sync` to download it and `vale <files>` to lint. Override any rule's
severity (or turn it off) in your own `.vale.ini` — e.g. `BenBalter.Buzzwords = error`.

## Rules

| Rule | Level | Flags |
| --- | --- | --- |
| `Assumptive` | warning | Reader-condescending throat-clearing ("simply", "of course", "needless to say"). |
| `AvoidJargon` | warning | Jargon with a plainer substitute. |
| `Buzzwords` | warning | Corporate buzzwords ("leverage", "synergy", "robust", "holistic") with concrete replacements. |
| `But` | error | Paragraphs that open with "but". |
| `Cliches` | warning | Clichés ("back to the drawing board", "think outside the box"). |
| `CorporateSpeak` | error | Corporate-speak ("synergy", "circle back"). |
| `CulturalInclusion` | error | Non-inclusive language, with inclusive alternatives. |
| `FillerWords` | error | Filler that adds no meaning. |
| `HyphenatedAdverbs` | warning | Unneeded hyphens after `-ly` adverbs. |
| `MeaningfulLinkWords` | warning | Non-descriptive link text ("click here", "this"). |
| `OxfordComma` | warning | Missing serial comma in lists of four or more. |
| `SentanceSpacing` | error | Double spaces between sentences. |
| `So` | error | Sentences that open with "so". |

## Development

```sh
script/lint   # yamllint the rule files
script/test   # assert rules fire on fixtures in testdata/ and leave clean prose alone
```

CI runs both on every push and pull request.

## Releasing

Push a tag — the [release workflow](.github/workflows/release.yml) builds
`BenBalter.zip` and attaches it to the release:

```sh
git tag v0.0.4
git push origin v0.0.4
```

Consumers pick up the new rules on their next `vale sync` after bumping the
version in their `Packages` URL.
