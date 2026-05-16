# Design: Spec-Kit Engram-Sync Extension Installation

## Technical Approach

Add step 3d to `sdd.Inject()` — a new `injectSpecKitExtensions` function that detects an existing `.specify/` directory at the project root and writes embedded spec-kit extension files into it. Follows the same `findProjectRoot` + `os.Stat` guard pattern as step 3b (workflowInjector), but without an adapter interface since it runs for ALL adapters. Uses `fs.WalkDir` for recursive asset copy and `filemerge.WriteFileAtomic` for atomic writes with content-comparison idempotency.

## Architecture Decisions

### Decision: Standalone function, not adapter interface

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A. `injectSpecKitExtensions()` standalone function | Simple, no interface overhead, runs for all adapters | **Chosen** — spec-kit extensions are adapter-agnostic |
| B. New `specKitExtensionInjector` interface on adapter | Extensible but forces no-op stubs on all adapters | Rejected — only one consumer now |
| C. Separate component (ComponentSpecKit) | Clean separation but overkill for 2 files | Deferred — refactor if more integrations appear |

### Decision: Error logging via `log` package

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A. Add `"log"` import, use `log.Printf` | Standard, used elsewhere in codebase | **Chosen** |
| B. Silently swallow errors | Risky — hides failures | Rejected |
| C. Return error, let caller decide | Breaks spec REQ-5 (must not abort pipeline) | Rejected |

### Decision: No CLI flag in this change

| Option | Tradeoff | Decision |
|--------|----------|----------|
| A. Add `--no-speckit-extensions` flag now | Exposes control early | Deferred to follow-up |
| B. Skip field defaults to false (enabled) | Simpler, flag can be added later | **Chosen** |

## Data Flow

```
Inject() entry
     │
     ├─ 3b. workflowInjector (per adapter)
     ├─ 3c. sub-agents (per adapter)
     │
     ├─ 3d. injectSpecKitExtensions()  ← NEW
     │       │
     │       ├─ opts.SkipSpecKitExtensions? → skip
     │       ├─ findProjectRoot(opts.WorkspaceDir) → not found? → skip
     │       ├─ os.Stat(projectRoot/.specify) → not found? → skip
     │       └─ fs.WalkDir(assets.FS, "specify")
     │            └─ filemerge.WriteFileAtomic(target, content, 0o644)
     │                 ├─ success → track changed + files
     │                 └─ error → log.Printf, continue
     │
     └─ 4. skill-registry automation
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/components/sdd/inject.go` | Modify | Add `injectSpecKitExtensions` func, call in `Inject()` as step 3d, add `SkipSpecKitExtensions` to `InjectOptions`, add `"log"` import |
| `internal/assets/assets.go` | Modify | Add `all:specify` to embed directive |
| `internal/assets/specify/extensions/engram-sync/extension.yml` | Create | Spec-kit extension manifest for engram-sync |
| `internal/assets/specify/extensions/engram-sync/scripts/bash/engram-sync.sh` | Create | Bash script that syncs specs → engram memory |
| `internal/components/sdd/inject_test.go` | Modify | Add test functions for step 3d |

## Interfaces / Contracts

### InjectOptions new field

```go
type InjectOptions struct {
    // ... existing fields ...

    // SkipSpecKitExtensions disables step 3d (spec-kit extension injection).
    // When false (default), spec-kit extension files are written to
    // <projectRoot>/.specify/ if .specify/ exists at the project root.
    SkipSpecKitExtensions bool
}
```

### injectSpecKitExtensions function

```go
// injectSpecKitExtensions writes embedded spec-kit extension files into
// <projectRoot>/.specify/ when .specify/ exists at the project root.
// Errors are logged but do not abort the Inject() pipeline.
func injectSpecKitExtensions(projectRoot string) (bool, []string) { ... }
```

Called from `Inject()` at step 3d, after sub-agents (step 3c) and before skill-registry automation (step 4).

### Asset structure

```
internal/assets/specify/
└── extensions/
    └── engram-sync/
        ├── extension.yml          # spec-kit extension manifest
        └── scripts/
            └── bash/
                └── engram-sync.sh # sync script
```

### Embed directive change

```go
//go:embed all:claude all:opencode all:generic all:skills all:gga all:gemini all:codex all:antigravity all:windsurf all:cursor all:kimi all:qwen all:kiro all:specify
```

### CLI propagation

No changes to `sync.go` or `run.go`. `SkipSpecKitExtensions` defaults to `false` (enabled). Both `componentSyncStep.Run()` and `componentApplyStep.Run()` construct `InjectOptions` without setting the field, so step 3d runs automatically from all entry points (install, sync, update).

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Happy path — `.specify/` exists, files written | Create temp dir with `.specify/` + `go.mod`, call `Inject()`, assert files exist and content matches assets |
| Unit | Skip — no project root | Empty `WorkspaceDir`, assert no `.specify/` paths in `result.Files` |
| Unit | Skip — no `.specify/` dir | Project root exists but no `.specify/`, assert step 3d is skipped |
| Unit | Skip — `SkipSpecKitExtensions = true` | Set flag, assert no files written even with `.specify/` present |
| Unit | Idempotency — second run unchanged | Inject twice, assert `second.Changed == false` |
| Unit | Overwrite — existing content differs | Pre-write modified file, inject, assert content matches embedded asset |
| Unit | Error resilience — write error logged not fatal | Use permission-denied dir, assert `Inject()` succeeds (no error returned) |

### Test function names

```go
func TestInject_SpecKitExtensionsWrittenToProject(t *testing.T)
func TestInject_SpecKitExtensionsIdempotent(t *testing.T)
func TestInject_SpecKitExtensionsSkippedWithoutWorkspaceDir(t *testing.T)
func TestInject_SpecKitExtensionsSkippedWithoutDotSpecify(t *testing.T)
func TestInject_SpecKitExtensionsSkippedWhenFlagSet(t *testing.T)
func TestInject_SpecKitExtensionsOverwritesExisting(t *testing.T)
```

All tests use `t.TempDir()` for filesystem isolation, matching existing inject_test.go patterns (see `TestInjectWindsurf_WorkflowsCopiedToWorkspace`). No golden files — direct content comparison against `assets.MustRead()`.

## Migration / Rollout

No migration required. Step 3d is purely additive — it only writes files when `.specify/` already exists. Projects without `.specify/` are unaffected.

### Rollback on uninstall

Extension files written to `<projectRoot>/.specify/` are NOT cleaned up on uninstall. Rationale: `.specify/` is owned by the spec-kit tool, not by gentle-ai. Gentle-ai only adds files inside it. Cleanup is the user's responsibility or spec-kit's. This matches the existing pattern where gentle-ai does not remove `.windsurf/` on uninstall.

## Open Questions

None — all decisions resolved by the architecture decision record (Option A).
