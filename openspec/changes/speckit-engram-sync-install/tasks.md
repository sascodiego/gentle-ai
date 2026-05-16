# Tasks: Spec-Kit Engram-Sync Extension Installation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 250–310 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | auto-chain |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Assets + embed directive | PR 1 | Foundation — new files + 1-line assets.go change |
| 2 | Injection logic + 6 tests | PR 1 | Core feature, depends on Unit 1 |

## Phase 1: Asset Foundation

- [x] 1.1 Create `internal/assets/specify/extensions/engram-sync/extension.yml` — spec-kit extension manifest
- [x] 1.2 Create `internal/assets/specify/extensions/engram-sync/scripts/bash/engram-sync.sh` — sync script
- [x] 1.3 Add `all:specify` to embed directive in `internal/assets/assets.go`
- [x] 1.4 Verify: `go build ./internal/assets/` compiles, `assets.FS.ReadDir("specify")` returns entries

## Phase 2: Core Implementation (TDD)

- [x] 2.1 RED: Write `TestInject_SpecKitExtensionsWrittenToProject` — temp dir with `.specify/` + `go.mod`, assert files written and content matches `assets.MustRead()` (REQ-2 happy path)
- [x] 2.2 RED: Write `TestInject_SpecKitExtensionsSkippedWithoutWorkspaceDir` — empty WorkspaceDir, assert no `.specify/` paths in result.Files (REQ-2 skip: no root)
- [x] 2.3 RED: Write `TestInject_SpecKitExtensionsSkippedWithoutDotSpecify` — project root found but no `.specify/`, assert step skipped (REQ-2 skip: no .specify)
- [x] 2.4 RED: Write `TestInject_SpecKitExtensionsSkippedWhenFlagSet` — `SkipSpecKitExtensions: true`, assert no files written (REQ-3)
- [x] 2.5 GREEN: Add `SkipSpecKitExtensions bool` to `InjectOptions` in `internal/components/sdd/inject.go`
- [x] 2.6 GREEN: Write `injectSpecKitExtensions(projectRoot string) (bool, []string)` in `inject.go` — guard chain: findProjectRoot → os.Stat(.specify) → fs.WalkDir → WriteFileAtomic, errors logged via `log.Printf`
- [x] 2.7 GREEN: Wire step 3d call in `Inject()` — after step 3c, before step 4, aggregate `(changed, files)` into result. Skip if `opts.SkipSpecKitExtensions`
- [x] 2.8 Run tests 2.1–2.4 — all pass: `go test ./internal/components/sdd/... -run TestInject_SpecKitExtensions`

## Phase 3: Additional Test Scenarios (TDD)

- [x] 3.1 RED: Write `TestInject_SpecKitExtensionsIdempotent` — inject twice, assert `second.Changed == false` (REQ-2 idempotency)
- [x] 3.2 RED: Write `TestInject_SpecKitExtensionsOverwritesExisting` — pre-write modified file, inject, assert content matches embedded asset (REQ-2 overwrite)
- [x] 3.3 Run all 6 tests: `go test ./internal/components/sdd/... -run TestInject_SpecKitExtensions -v`

## Phase 4: Verification

- [x] 4.1 Run full test suite: `go test ./internal/components/sdd/...` — zero regressions
- [x] 4.2 Run build: `go build ./...` — compiles clean
