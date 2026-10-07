# Vale styles

Personal prose style checks for [Vale](https://vale.sh) — the house rules behind
[ben.balter.com](https://ben.balter.com) and the book *Open and Async*. They
catch buzzwords, clichés, corporate-speak, and other tells so the writing stays
plain and direct.

## Usage

Add the package to your `.vale.ini` and enable the style:

```ini
StylesPath = .github/styles
Packages = https://github.com/benbalter/vale-styles/releases/download/v0.0.6/BenBalter.zip, https://github.com/benbalter/vale-styles/releases/download/v0.0.6/AIPatterns.zip

[*.md]
BasedOnStyles = BenBalter, AIPatterns
```

Each release ships two styles. `BenBalter` holds the house rules below;
`AIPatterns` flags the cadence and vocabulary of AI-drafted prose. Use either or
both.

Then `vale sync` to download it and `vale <files>` to lint. Override any rule's
severity (or turn it off) in your own `.vale.ini` — e.g. `BenBalter.Buzzwords = error`.

## Rules

| Rule | Level | Flags |
| --- | --- | --- |
| `Assumptive` | warning | Reader-condescending throat-clearing ("simply", "of course", "needless to say"). |
| `AvoidJargon` | warning | Jargon with a plainer substitute. |
| `Buzzwords` | warning | Corporate buzzwords ("leverage", "synergy", "robust", "holistic") with concrete replacements. |
| `But` | warning | Paragraphs that open with "but". |
| `Cliches` | warning | Clichés ("back to the drawing board", "think outside the box"). |
| `CorporateSpeak` | warning | Corporate-speak and idioms that don't translate ("circle back", "drop the ball", "move the needle"). |
| `CulturalInclusion` | warning | Non-inclusive language, with inclusive alternatives. |
| `FillerWords` | warning | Filler that adds no meaning. |
| `HyphenatedAdverbs` | warning | Unneeded hyphens after `-ly` adverbs. |
| `MeaningfulLinkWords` | warning | Non-descriptive link text ("click here", "this"). |
| `NumericDates` | suggestion | All-numeric dates ("6/5/2026") that read differently across regions. |
| `OxfordComma` | warning | Missing serial comma in lists of three or more. |
| `Redundancy` | suggestion | Wordy phrases with a shorter form ("in order to", "the vast majority of"). |
| `SentanceSpacing` | warning | Double spaces between sentences. |
| `Terms` | warning | Product names spelled wrong ("Github", "Javascript", "Wordpress"). |
| `So` | warning | Sentences that open with "so". |

## AIPatterns rules

Calibrated against pre-AI posts on ben.balter.com, so they fire on AI tells
rather than on Ben's deliberate voice. `DisguiseMetaphors`, `EmphaticItalics`,
`FalseExclusivity`, `LabelAndExplain`, and `MicDrop` are curated subsets of
[vale-ai-tells](https://github.com/tbhb/vale-ai-tells) rules (MIT), keeping only
the patterns that are near zero in Ben's pre-2023 posts.

| Rule | Level | Flags |
| --- | --- | --- |
| `AITics` | error | Throat-clearing ("it's worth noting that", "as previously mentioned"). |
| `Antithesis` | suggestion | "It's not X, it's Y" constructions. |
| `Aphorisms` | suggestion | Formulaic aphorisms in place of a concrete claim. |
| `ChatbotArtifacts` | error | Chatbot leftovers ("I hope this helps", "Certainly!"). |
| `DisguiseMetaphors` | warning | One thing "dressed up as" or "in a trench coat" as another. |
| `EmDash` | warning | Em dashes. En dashes for ranges are fine. |
| `EmDashDensity` | suggestion | Three or more em dashes in one paragraph, for consumers that turn `EmDash` off. |
| `EmphaticItalics` | suggestion | Italicized common words (*is*, *and*, *actually*) doing the sentence's work. |
| `FalseExclusivity` | warning | Insider-knowledge claims ("what nobody tells you", "the hidden cost"). |
| `Fingerprints` | error | Generation fingerprints ("as an AI language model", leaked citation markup). |
| `Flourishes` | warning | Dramatic flourishes ("Buckle up", "Let that sink in"). |
| `LabelAndExplain` | warning | "The catch: ..." label-and-colon reveals. |
| `MetaCommentary` | error | "Let's dive into", "Let's explore". |
| `MicDrop` | suggestion | Stock closers ("Full stop.", "And it shows."). |
| `PerformativeEnthusiasm` | error | "Great question!", "Excellent point!". |
| `PlainWords` | warning | Inflated words with plain substitutes. |
| `RhetoricalQuestions` | suggestion | Self-answered setups ("So why does this matter?"). |
| `SetupPhrases` | warning | "When it comes to", "In today's world". |
| `Signposts` | error | Signposts that announce a twist instead of delivering one. |
| `ThisConstructions` | error | "This ensures", "This allows" constructions. |
| `Transitions` | warning | "Furthermore", "Moreover", "Additionally". |
| `VagueIntensifiers` | suggestion | "highly effective", "incredibly valuable". |
| `Vocabulary` | warning | Words AI overuses ("delve", "tapestry", "robust"). |
| `WeakOpenings` | suggestion | Weak sentence openers. |

## Development

```sh
script/lint          # yamllint the rule files, gofmt the test harness
script/test          # golden-file tests for every rule (needs Go and Vale)
script/test -update  # regenerate the expected output after an intended change
```

Each rule has a fixture in `fixtures/<Style>.<Rule>/` that enables only that
rule, with prose it should flag, prose it should leave alone, and any known
false positives. `testdata/<Style>.<Rule>.ct` holds the exact Vale output
expected for it, so a change to what a rule matches, where, or what it says
shows up as a diff. `fixtures/clean` runs every rule against plain prose and
must trip nothing. The harness is
[go-cmdtest](https://github.com/google/go-cmdtest), the same one
[errata-ai/Google](https://github.com/errata-ai/Google) uses.

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
