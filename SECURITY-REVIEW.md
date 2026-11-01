# Dependency security review

Assessment date: **2026-11-01** (America/Los_Angeles). Release dates below use UTC. This is a point-in-time dependency and build-supply-chain assessment, not a full application security audit. No dependencies or build settings were changed by this review. Automation below is proposed, not installed or enforced.

## Findings and recommended policy

- The application has **one external Go module and no transitive module dependencies**. The external module is already at its latest stable release and satisfies a seven-day cooldown.
- Keep Go's standard library. Keep Pelletier's TOML implementation as an explicit exception: preserving general TOML compatibility warrants a maintained parser project. Prefer original code for small, bounded features; do not build a general TOML parser inside this application merely to remove a dependency.
- Update the build toolchain: this workstation runs Go 1.26.4, while cooled security patch releases are available. Pin the Docker builder and control automatic toolchain selection.
- Require at least **168 elapsed hours** since public release, no known critical vulnerability affecting the selected version, and review before merge. A cooldown improves detection opportunities; it cannot establish that a release is uncompromised.
- Prefer the standard library, then an appropriate Go-project library, then a demonstrably well-maintained external project when complexity justifies it. A familiar organization, popularity, or a signed commit alone does not establish safety. Apply the same review to test and build tools.

## Inventory and exit assessments

Inventory evidence: `go.mod`, `go.sum`, `go list -m -json all`, cached upstream `go.mod`, imports in application/tests, Dockerfile, Makefile, and build/release scripts. No vendored modules, Go workspace, JavaScript dependencies, or checked-in GitHub Actions workflows were found. Standard-library packages share the Go release lifecycle, so they are assessed together rather than assigned fictitious independent versions.

| Library / module | Current version and use | Primary library from Go project? | Security reputation / evidence | Newer available? | Latest cooled stable version; release date | Latest stable version; release date | Exit assessment | Exit strategy |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Go standard library and runtime | Local build: **Go 1.26.4**, released 2026-06-02. `go.mod` requires Go 1.26; it does not pin the compiler patch. Docker uses `golang:1.26-alpine`, whose actual compiler is unresolved. | **Yes**, shipped by Go. | Preferred first-party baseline with a public security process. Still subject to vulnerabilities; local compiler is behind security patches. | **Yes**: 1.26.10 in the existing series; 1.27.3 overall. | **go1.27.3; 2026-10-19**. Staying on 1.26: **go1.26.10; 2026-10-19**. | **go1.27.3; 2026-10-19**. Latest 1.26 patch: **go1.26.10; 2026-10-19**. | **Keep using; update build compiler.** | Upgrade to 1.26.10 as the smallest patch step; evaluate 1.27.3 separately. Rebuild binaries to incorporate runtime/stdlib fixes. Reimplementing the runtime, crypto, or parsers is not a sensible exit. |
| `github.com/pelletier/go-toml/v2` | **v2.4.3**; configuration decoding, generated config encoding, `ch config`, and tests. Exact version and checksums committed. | **No**. Maintained by the upstream TOML project, independently of Go. | Established specialist project; maintained 2.x line, public security reporting, multiple contributors, and no module dependencies. Favorable but qualified assessment: upstream describes maintenance as a side project; a recent high-severity parsing advisory demonstrates remaining parser risk. No independent security audit was established in this review. | **No**, within the v2 module's stable release line. | **v2.4.3; 2026-07-05**. | **v2.4.3; 2026-07-05**. GitHub publication: **02:35:59Z**. | **Keep using**, as a documented complexity exception while TOML is the configuration contract. | If eliminating this dependency becomes necessary, prefer migration to standard-library `encoding/json` with a deliberate configuration migration and replacement for generated comments. A small original parser is viable only after explicitly narrowing/documenting accepted syntax and migration behavior. Do not silently replace full TOML with an incompatible subset. Fork only for an urgent, bounded fix with an assigned maintainer. |

Go version dates and security patch history come from the [official release history](https://go.dev/doc/devel/release). TOML version/publication comes from the [v2.4.3 release](https://github.com/pelletier/go-toml/releases/tag/v2.4.3), checked against GitHub's public releases API. Its module metadata reports revision time `2026-07-05T02:25:11Z`; use the later publication time for cooldown. Maintenance evidence: [upstream repository](https://github.com/pelletier/go-toml), [security policy](https://github.com/pelletier/go-toml/security), and [versioned module manifest](https://github.com/pelletier/go-toml/blob/v2.4.3/go.mod).

Direct standard-library imports across source and tests are `bytes`, `encoding/json`, `errors`, `fmt`, `io`, `os`, `os/exec`, `path/filepath`, `reflect`, `regexp`, `strings`, `sync`, `syscall`, `testing`, `time`, `unicode/utf8`, and `unsafe`. Their Go-project status, version dates, and keep/upgrade assessment are identical to the first row. Internal runtime dependencies and standard-library packages used by TOML/tests are covered by that row too. The application's own `internal/*` packages are original project code. There are no `golang.org/x/*` dependencies today.

### Vulnerability evidence and limits

On the assessment date, an exact-version query to [OSV's query API](https://google.github.io/osv.dev/post-v1-query/) for ecosystem `Go`, package `github.com/pelletier/go-toml/v2`, version `v2.4.3` returned `{}`: no matching advisories. This is evidence of no known findings in that source, not proof of absence or a successful application scan.

Upstream [GHSA-frc4-6h9q-39fq](https://github.com/pelletier/go-toml/security/advisories/GHSA-frc4-6h9q-39fq), published 2026-07-07, rates unbounded TOML parser recursion **High, CVSS 7.5**; versions through 2.4.2 are affected and **2.4.3 is patched**. The advisory lists no CVE identifier. Therefore check GHSA/Go/OSV advisories as well as CVE identifiers. This application reads local configuration rather than serving a TOML upload endpoint, but still calls the affected parsing API. Do not downgrade to an older vulnerable parser for age alone.

The local Go 1.26.4 predates later security fixes. Reachability and critical-severity status of its individual findings were **not established** here. `govulncheck` is not installed; no source, module-level, or binary vulnerability scan was completed. The current environment is therefore **not certified to meet the proposed vulnerability gate**. Before release, scan with the selected patched compiler for Windows/amd64 and any other supported build configuration. Use [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) for Go-specific analysis and advisory metadata for severity; a successful reachability scan alone does not exclude critical issues elsewhere in a selected module or build toolchain.

## Other supply-chain inputs

| Input | Present exposure | Recommended treatment |
| --- | --- | --- |
| Docker builder | `golang:1.26-alpine` is a mutable tag; actual Go/Alpine versions and image digest are not recorded. | Keep the official Go builder, pin an approved patch, Alpine version, and image digest; review each digest update. Record actual publication/first-seen time and scan the full builder image. Official Docker images are not identical to Go-project library ownership. |
| Alpine packages | Dockerfile runs `apk add --no-cache bash`; repository state can change independently of the image pin. | Prefer a POSIX-shell build script to remove Bash if practical; otherwise record/pin package inputs or use a controlled repository snapshot and scan them. Their latest/cooled versions are **not assessed** because no resolved package inventory exists. |
| Go module distribution | Local `GOPROXY=https://proxy.golang.org,direct`, `GOSUMDB=sum.golang.org`; `go.sum` contains both hashes for TOML. | Retain checksum verification; consider proxy-only resolution for controlled release builds. Hashes detect changed content, not malicious content first published by a compromised maintainer. Run `go mod verify`. |
| Compiler selection | Local `GOTOOLCHAIN=auto`; minimum Go directive does not pin a release build. | Provision the approved exact compiler and set `GOTOOLCHAIN=local` in controlled builds so requirements cannot trigger an unreviewed compiler download. See [Go toolchain selection](https://go.dev/doc/toolchain). |
| Release tools | Bash, Git, `gh`, Docker/buildx, and the host OS are trusted build inputs; scripts run tests/build without age or vulnerability checks. | Keep tools from their original publishers, maintain an inventory of exact versions, and apply publisher-specific update review. They are not application Go modules. No exact installed versions or latest/cooled versions were established for them in this review. |

Go's [module authentication documentation](https://go.dev/ref/mod#authenticating) explains checksum verification. Neither pinning nor vendoring removes the obligation to review the original source and track security fixes.

## Upgrade automation plan

### Choice of implementation

**Recommended: declarative update proposals plus a small repository-owned verification command and human review.** Go's module manifest cannot express a release-age or CVE policy. Avoid adding an updater framework to the application; hosted update services are build infrastructure and need their own trust assessment.

| Approach | What it can do | Remaining work |
| --- | --- | --- |
| Dependabot | Supports Go modules and a declarative `cooldown.default-days: 7`; opens scheduled update PRs. | Its cooldown applies to version updates, **not security updates**. Independent merge checks must enforce age and security policy for every PR. |
| Renovate | Supports Go modules, `minimumReleaseAge: "7 days"`, and `minimumReleaseAgeBehaviour: "timestamp-required"`. | Verify release timestamp provenance and security-update overrides. OSV integration currently covers direct dependencies and is experimental; it is not a complete vulnerability gate. |
| Original Go verifier/updater | Standard-library HTTP/JSON/process support can gather metadata, filter candidates, and invoke Go tooling. | Implement explicit policy and fail-closed handling; do not build a full dependency resolver, advisory database, or general SemVer library from scratch. |
| Periodic assisted manual review | Evaluates reputation, ownership changes, suspicious code, migration costs, and exceptions. | Useful immediately, but easy to miss updates without scheduling and enforced checks. Retain it alongside automation. |

Sources: [Dependabot options](https://docs.github.com/en/code-security/reference/supply-chain-security/dependabot-options-reference#cooldown), [Renovate release age](https://docs.renovatebot.com/key-concepts/minimum-release-age/), and [Renovate OSV integration](https://docs.renovatebot.com/configuration-options/#osvvulnerabilityalerts).

An illustrative Dependabot configuration (not an active repository file):

```yaml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule:
      interval: weekly
    cooldown:
      default-days: 7
```

### Verification and candidate selection

Implement these steps as an independent required check, also called before publishing a release. Weekly proposals are appropriate for routine updates; scan existing pins daily so newly disclosed vulnerabilities are discovered without waiting for an upgrade PR.

1. **Declare policy as JSON**, parsed with `encoding/json`: minimum age `168h`, allowed module paths and major versions, approved Go compiler, blocked advisory IDs/versions, and any time-limited exceptions. Add new modules, replacements, forks, tool dependencies, and major-version migrations only through explicit review. Default to blocking known critical findings; preferably block all known vulnerabilities unless individually reviewed.
2. **Discover candidates without executing them.** Use Go commands such as `go list -m -versions -json` and `go list -m -json module@version` for versions/metadata. Honor retractions, exclude prereleases and pseudo-versions by default, and retain Go's version ordering. Evaluate every stable candidate in the allowed line, not just `@latest`: an older candidate may be the newest cooled, secure release. Enumerate new major module paths separately during manual review.
3. **Verify age from public availability.** Go module `Time` commonly reflects the revision, not first publication. Compare it with publisher release/tag metadata and a trusted first-seen timestamp; use the latest applicable time. An old commit newly tagged today must wait seven days. Missing/conflicting dates, moved tags, or API failures block eligibility pending review. Cache first-seen records with provenance in controlled storage. Restrict metadata fetching to approved publishers/endpoints.
4. **Reject known unsafe candidates.** Query Go's vulnerability database and OSV for the complete selected graph, plus publisher advisories and GitHub advisory severity where needed. Include `stdlib` and `toolchain`. Match affected version ranges, aliases, withdrawals, platform scope, and fix versions; a fix for one advisory may still have another. Block critical findings at module/toolchain level even if a call is currently unreachable. Unknown severity or unavailable/stale advisory data requires review rather than a green result. Store database retrieval times and advisory IDs in the report. Use an approved, pinned Go-project `govulncheck` tool for reachability analysis rather than reimplementing it.
5. **Resolve in isolation and verify the final graph.** Apply an exact candidate with `go get module@version` in a temporary checkout; inspect `go mod tidy` output. Go resolution can raise transitive versions or toolchain requirements. Reapply age, allowlist, retraction, and vulnerability checks to every version actually selected. Do not rely solely on filtering direct dependencies or on a bot's PR description.
6. **Validate and record.** Run `go mod verify`, `go test ./...`, `go vet ./...`, Windows source analysis with `govulncheck -json ./...`, module-level scanning, and a Windows build using the approved compiler. Add race tests where supported and relevant. Record compiler/tool versions, complete module graph, release/publication timestamps, eligibility time, advisories, test results, and source diff links in the PR. Pin the scanner and apply the same age/security review to its own tool dependencies.
7. **Require review and fresh checks.** Review new dependencies, maintainer/ownership changes, source diffs, generated artifacts, network/process behavior, and build changes in an unprivileged environment without release credentials. Keep automatic merging disabled initially. Recheck age and advisory data at merge and release; protect the check/policy files and require the check on all dependency edits, including manual and security-bot PRs.

There may be **no version satisfying both constraints**. Report that explicitly; do not silently select a vulnerable cooled version or bypass the cooldown. For an urgent fix released less than seven days ago, a maintainer may approve a documented, narrowly scoped age exception after reviewing the fix and provenance. Never automatically waive the critical-vulnerability block. Otherwise mitigate/disable affected behavior, use a reviewed original replacement, or delay release. A fork creates continuing patch obligations and is a last resort.

### Rollout and manual review

1. **Before the next release:** choose a patched Go compiler, resolve/pin the builder, complete the missing vulnerability scans, and record results. Reconcile README's old Go 1.22+ instruction with the existing Go 1.26 requirement when updating the toolchain documentation.
2. **First automation change:** add one declarative proposal service and a standard-library-based verifier. Test meaningful boundaries: exactly 168 hours, newly published old commits, missing dates, vulnerable newest releases, retractions, transitive version increases, and no eligible candidate. Enforce checks for dependency changes and release preparation.
3. **Ongoing:** run daily vulnerability checks and weekly update proposals; conduct a monthly assisted reputation/exit review and review immediately on a compromise notice or ownership change. Refresh this document's date, inventory, and evidence with each dependency/toolchain review. Human maintainers own merge and exception decisions.

This policy is implementable with existing Go tooling and declarative services. Manual review remains necessary for trust and exit decisions, not for routine date arithmetic or version discovery.
