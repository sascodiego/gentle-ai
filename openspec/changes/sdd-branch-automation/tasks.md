# Tasks: SDD Branch Automation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~850-950 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | auto-chain |
| Chain strategy | feature-branch-chain |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Pure functions: slug, lifecycle, active-change + tests | PR 1 | Base: feat/sdd-branch-automation; ~300 lines |
| 2 | Git + context + commands (current, slug, feature) + tests | PR 2 | Base: PR 1 branch; ~350 lines |
| 3 | Wiring: app.go dispatch + .gitignore + integration tests | PR 3 | Base: PR 2 branch; ~200 lines |

## Phase 1: Foundation — Pure Logic (PR 1)

- [x] 1.1 Create `internal/sdd/slug.go` — `Generate(desc, SlugOptions)` with stop-word list, sanitization, 244-byte truncation
- [x] 1.2 Create `internal/sdd/slug_test.go` — table-driven: basic, stop-words, truncation, numeric, all-stop-words fallback, 244-byte boundary, with-type prefix
- [x] 1.3 Create `internal/sdd/lifecycle.go` — `Phase` type with 9 constants, `DetectStatus(changeDir)` checking artifact presence in order
- [x] 1.4 Create `internal/sdd/lifecycle_test.go` — table-driven: artifact file sets mapped to expected phase, empty dir → exploring, all artifacts → completed
- [x] 1.5 Create `internal/sdd/active.go` — `ActiveChange` struct, `ReadActiveChange(path)`, `WriteActiveChange(path, ac)` using direct string format
- [x] 1.6 Create `internal/sdd/active_test.go` — round-trip write/read, malformed yaml error, missing file error

## Phase 2: Commands & Git Integration (PR 2)

- [x] 2.1 Create `internal/sdd/git.go` — `currentBranch()`, `createBranch(name)` using `os/exec`
- [x] 2.2 Create `internal/sdd/context.go` — `ResolveContext()` calling `currentBranch()` → parse type/slug → `active-change.yaml` fallback → error
- [x] 2.3 Create `internal/sdd/context_test.go` — branch parse, yaml fallback, error case; mock git via test-only `CurrentBranch` variable seam
- [x] 2.4 Create `internal/sdd/cmd_current.go` — `currentCmd(args, stdout)` calling ResolveContext + DetectStatus; handle `--json`, `--branch`, `--status`
- [x] 2.5 Create `internal/sdd/cmd_slug.go` — `slugCmd(args, stdout)` parsing `--type`, calling Generate, outputting result
- [x] 2.6 Create `internal/sdd/cmd_feature.go` — `featureCmd` dispatching init/list/complete; init creates branch+dir+yaml; list enumerates changes; complete validates status=verifying
- [x] 2.7 Create `internal/sdd/cmd_current_test.go` — integration: Run(["current"]) in temp git repo, verify output text/JSON/branch/status flags
- [x] 2.8 Create `internal/sdd/sdd.go` — `Run(args, stdout) error` dispatching current/slug/feature subcommands, usage on empty/unknown

## Phase 3: Wiring & Integration (PR 3)

- [ ] 3.1 Modify `internal/app/app.go` — add `case "sdd":` in first switch block (line ~63), import `internal/sdd`, call `sdd.Run(args[1:], stdout)`
- [ ] 3.2 Modify `.gitignore` — add `openspec/active-change.yaml`
- [ ] 3.3 Verify `go build ./...` compiles, `go test ./...` passes, `go vet ./...` is clean
- [ ] 3.4 Manual smoke test: `gentle-ai sdd slug "Add branch automation"` → `add-branch-automation`; `gentle-ai sdd` → usage error exit 1
