# Proposal: Spec-Kit Engram-Sync Extension Installation

## Intent

When a project uses spec-kit (`.specify/` exists), Gentle AI's install/sync/update/upgrade must write the `engram-sync` extension files so spec-kit can persist SDD artifacts to Engram. Today this is a manual step — this change automates it inside the existing SDD injection pipeline.

## Scope

### In Scope
- New `specKitExtensionInjector` step inside `sdd.Inject()` (step 3d, after sub-agents, before step 4)
- New `SkipSpecKitExtensions bool` field on `InjectOptions` — CLI flag `--no-speckit-extensions`
- New embedded assets: `internal/assets/specify/extensions/engram-sync/extension.yml` + `internal/assets/specify/scripts/bash/engram-sync.sh`
- Add `all:specify` to `//go:embed` directive in `internal/assets/assets.go`
- Guard: only writes when `.specify/` directory exists at project root (never creates it)
- Installer ALWAYS overwrites — no checksum guard for user modifications
- Works from all entry points: `gentle-ai install`, `gentle-ai sync`, `gentle-ai run` (update/upgrade)
- Tests for the new injector step (idempotency, skip conditions, guard behavior)

### Out of Scope
- **Feature name resolution** (`/sdd-new` and `/sdd-apply` without arguments resolving feature from spec-kit) — separate concern, deferred to a follow-up change
- Creating or modifying `.specify/` directory structure — this change ONLY writes into an existing one
- Any spec-kit CLI dependency at runtime — the Go binary embeds and writes files, never executes spec-kit
- Supporting extensions beyond `engram-sync` — architecture allows future additions but only one is implemented now

## Capabilities

### New Capabilities
- `speckit-extension-injector`: Detects `.specify/` at project root, embeds and writes spec-kit extension files (extension.yml + bash scripts) atomically. Includes skip flag and idempotent re-sync.

### Modified Capabilities
- `gga`: Sync/install pipeline passes `SkipSpecKitExtensions` flag through `InjectOptions` when CLI flag is set

## Approach

Follow the existing `workflowInjector` pattern (step 3b) but **without** a new adapter interface — the injector runs for ALL adapters when `.specify/` exists:

1. **Inside `Inject()`, after step 3c (sub-agents, ~line 628)**, add step 3d:
   - Check `opts.SkipSpecKitExtensions` — skip if true
   - Call `findProjectRoot(opts.WorkspaceDir)` — skip silently if no root found (same as workflowInjector)
   - `os.Stat(projectRoot + "/.specify")` — skip if `.specify/` doesn't exist
   - Walk `assets.FS.ReadDir("specify")` recursively — write each file relative to `projectRoot/.specify/`
   - Use `filemerge.WriteFileAtomic` for atomic writes (already handles content comparison for idempotency)
   - Track `changed` and `files` in the existing result

2. **New asset files**:
   - `internal/assets/specify/extensions/engram-sync/extension.yml` — spec-kit extension manifest
   - `internal/assets/specify/extensions/engram-sync/scripts/bash/engram-sync.sh` — sync script
   - Add `all:specify` to the embed directive

3. **`InjectOptions` extension**: Add `SkipSpecKitExtensions bool`. Wire through from CLI sync/run/install via existing `InjectOptions` variadic pattern.

4. **CLI wiring**: Add `--no-speckit-extensions` flag to sync and install commands. Default is `false` (extensions enabled).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/components/sdd/inject.go` | Modified | Add step 3d injector + `SkipSpecKitExtensions` on `InjectOptions` |
| `internal/assets/assets.go` | Modified | Add `all:specify` to embed directive |
| `internal/assets/specify/` | New | Extension assets (extension.yml, engram-sync.sh) |
| `internal/cli/sync.go` | Modified | Wire `--no-speckit-extensions` flag to `InjectOptions` |
| `internal/cli/run.go` | Modified | Wire `--no-speckit-extensions` flag to `InjectOptions` |
| `internal/components/sdd/inject_test.go` | Modified | New tests for spec-kit extension injection |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `.specify/` directory structure changes across spec-kit versions | Low | Extension files live under `extensions/engram-sync/` — stable spec-kit convention |
| `WriteFileAtomic` overwrites user-customized extension scripts | Med | By design — installer always wins. Documented in decision artifact |
| Binary size increase from embedded assets | Low | Two small text files (~2KB total) — negligible |
| `findProjectRoot` fails in unusual environments (containers, CI) | Med | Silent skip — same behavior as workflowInjector. No error propagated |

## Rollback Plan

1. Remove step 3d code block from `Inject()` — single deletion
2. Remove `SkipSpecKitExtensions` field from `InjectOptions`
3. Remove `all:specify` from embed directive
4. Delete `internal/assets/specify/` directory
5. Remove CLI flag wiring in sync.go and run.go
6. All changes are additive — rollback is clean deletion, no migration needed

## Dependencies

- Existing `filemerge.WriteFileAtomic` — no changes required
- Existing `findProjectRoot` — reused as-is
- Existing `assets.FS` embed filesystem — extended with new directory

## Success Criteria

- [ ] `go build ./...` passes with zero errors
- [ ] `go test ./internal/components/sdd/...` — all existing + new tests pass
- [ ] Injection writes files to `projectRoot/.specify/extensions/engram-sync/` when `.specify/` exists
- [ ] Injection silently skips when `.specify/` does NOT exist
- [ ] Injection silently skips when `SkipSpecKitExtensions` is true
- [ ] Injection is idempotent: second run returns `Changed == false` when content unchanged
- [ ] Works from `install`, `sync`, and `run` entry points
- [ ] Binary size increase < 5KB
