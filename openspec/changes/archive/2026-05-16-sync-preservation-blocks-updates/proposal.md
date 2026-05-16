# Proposal: Sync Preservation Logic Fix

## Intent

The sync system's `inlineOpenCodeSDDPrompts()` preserves ALL existing orchestrator prompts when `PreserveOpenCodeOrchestratorPrompt=true`, with no way to distinguish stock content from genuine user edits. This blocks asset updates — users must manually delete their orchestrator prompt from `opencode.json` to receive improvements shipped in new releases.

## Scope

### In Scope

- Hash-based stock detection in `inlineOpenCodeSDDPrompts()` (`inject.go:779-803`)
- Adding `_gentle-ai-asset-hash` field to orchestrator agent in `opencode.json`
- Hash computation and comparison logic for the orchestrator prompt
- Safe first-run migration: assume customized when hash field is absent

### Out of Scope

- Sub-agent preservation logic (always overwrites via `WriteFileAtomic` — no bug)
- Model assignment preservation in `injectModelAssignments()` (intentional behavior)
- `WriteFileAtomic()` changes (content-comparison idempotency works correctly)
- Changes to markdown section markers (`InjectMarkdownSection`)

## Capabilities

### New Capabilities

- `stock-asset-detection`: Deterministic detection of stock vs user-customized orchestrator prompt content via hash comparison, enabling safe asset updates while preserving genuine user edits.

### Modified Capabilities

None.

## Approach

Add a `_gentle-ai-asset-hash` field (SHA-256) to the orchestrator agent definition in `opencode.json`. On each sync:

1. Read existing prompt and stored hash from disk
2. Compute hash of existing prompt
3. **Hash matches stored** → stock content → replace with new embedded asset and new hash
4. **Hash differs** → user customized → preserve existing prompt, keep stored hash unchanged
5. **Hash absent** (first run after upgrade) → assume customized (safe default), set hash from current content

After determining the prompt, inject both `prompt` and `_gentle-ai-asset-hash` into the overlay. Deep merge writes both atomically.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/opencode/inject.go:754-848` | Modified | `inlineOpenCodeSDDPrompts()` — add hash read/compare/write |
| `internal/opencode/inject.go:1353-1382` | Modified | `mergeJSONFile()` — ensure `_gentle-ai-asset-hash` passes through deep merge |
| `internal/opencode/json_merge.go` | Unchanged | Deep merge already handles unknown string fields correctly |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| First sync after upgrade doesn't update stock content | High (by design) | Document one-time manual step; set baseline hash so future syncs work |
| OpenCode rejects unknown `_gentle-ai-asset-hash` field | Low | Underscore-prefixed convention is standard for private metadata; OpenCode uses loose JSON parsing |
| Concurrent external tool modifies prompt between hash read and write | Low | Acceptable race for a sync tool; hash is point-in-time |

## Rollback Plan

Remove hash comparison logic from `inlineOpenCodeSDDPrompts()`, revert to binary preserve/replace. The `_gentle-ai-asset-hash` field in `opencode.json` is harmless metadata — no removal needed.

## Dependencies

- None beyond the existing codebase.

## Success Criteria

- [ ] When orchestrator prompt is stock (hash matches), sync updates it to the latest embedded asset
- [ ] When orchestrator prompt is user-customized (hash differs), sync preserves it unchanged
- [ ] When hash field is absent (first run), sync preserves existing content and sets the baseline hash
- [ ] Hash is written to `opencode.json` alongside the prompt after every sync
- [ ] Existing tests pass; new unit tests cover all three detection branches
