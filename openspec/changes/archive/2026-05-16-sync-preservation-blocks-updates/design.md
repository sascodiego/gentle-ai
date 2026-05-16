# Design: Sync Preservation Logic Fix

## Technical Approach

Replace the binary preserve/replace logic in `inlineOpenCodeSDDPrompts()` with hash-based stock detection. When `PreserveOpenCodeOrchestratorPrompt=true`, compute SHA-256 of the existing orchestrator prompt from disk and compare it against a stored hash in `opencode.json`. If they match, the content is stock — replace with the latest embedded asset. If they differ, the user customized it — preserve unchanged. On first run (no hash field), assume customized (safe default) and set the baseline hash.

## Architecture Decisions

### Decision: Hash Algorithm

| Option | Tradeoff | Decision |
|--------|----------|----------|
| SHA-256 | Go stdlib (`crypto/sha256`), no dependencies, 64-char hex string | **Chosen** |
| xxhash | Faster, but requires external dependency | Rejected |
| CRC32 | Fast but collision-prone for prompt-length strings | Rejected |

**Rationale**: Prompts are ~1-4 KB. SHA-256 performance is negligible at this size. No new dependency.

### Decision: What Gets Hashed

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Raw prompt string bytes (UTF-8) | Deterministic, matches what's on disk | **Chosen** |
| Trimmed prompt | Removes trailing-newline ambiguity | Rejected — could mask real user edits |
| Post-migration prompt | After `migratePreserved` applied | Rejected — hash must match disk state |

**Rationale**: Hash what's stored on disk. After migration, the content changes, so the hash would no longer match the stored value. Hash BEFORE migration; compare against stored hash; then decide.

### Decision: Hash Storage Location

| Option | Tradeoff | Decision |
|--------|----------|----------|
| `_gentle-ai-asset-hash` field in agent def | Co-located with prompt, deep-merge passes strings through | **Chosen** |
| Sidecar file | Separate state to manage, cleanup burden | Rejected |
| Top-level `_gentle-ai` metadata section | Requires custom merge logic | Rejected |

**Rationale**: The agent map (`agent.gentle-orchestrator`) is a flat string→any map. Deep merge treats unknown string fields as scalars — they pass through unchanged. OpenCode ignores unknown fields. No merge logic changes needed.

### Decision: First-Run Behavior

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Assume customized (preserve + set hash) | Safe; first sync after upgrade is no-op | **Chosen** |
| Assume stock (replace + set hash) | Risky; overwrites genuine customizations | Rejected |
| Compare against current embedded asset | Can't distinguish old-stock from customized | Rejected |

**Rationale**: Safe default. Users who never customized won't notice (prompt already works). Their next asset update WILL be applied. Document the one-time manual step.

## Data Flow

```
inlineOpenCodeSDDPrompts(overlay, homeDir, settingsPath, preserve=true)
  │
  ├─ Read existing prompt from opencode.json (lines 780-795)
  │
  ├─ Read stored hash: agent.gentle-orchestrator._gentle-ai-asset-hash
  │
  ├─ Compute SHA-256 of existing prompt (raw bytes)
  │
  ├─ stored hash == "" (absent)?
  │   YES → FIRST RUN
  │         Store computed hash, preserve existing prompt
  │
  ├─ computed == stored?
  │   YES → STOCK CONTENT
  │         Use new embedded asset as prompt
  │         Compute new hash from embedded asset
  │
  │   NO  → USER CUSTOMIZED
  │         Preserve existing prompt
  │         Keep stored hash unchanged
  │
  ├─ Apply migratePreservedOpenCodeOrchestratorPrompt()
  │
  ├─ Set orchestratorMap["prompt"] = resolved prompt
  │
  └─ Set orchestratorMap["_gentle-ai-asset-hash"] = resolved hash
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/components/sdd/inject.go` | Modify | `inlineOpenCodeSDDPrompts()` — add hash read/compute/compare/write logic |
| `internal/components/sdd/inject.go` | Create | New `computeAssetHash(content string) string` helper |
| `internal/components/sdd/inject.go` | Create | New `readOpenCodeAgentField(settingsPath, agentKey, field string) string` helper |
| `internal/components/sdd/inject_test.go` | Modify | Add unit tests for 3 detection branches + edge cases |

## Interfaces / Contracts

```go
// computeAssetHash returns the SHA-256 hex digest of the input string.
func computeAssetHash(content string) string

// readOpenCodeAgentField reads a single string field from an agent definition
// in opencode.json. Returns "" if the file/agent/field doesn't exist.
func readOpenCodeAgentField(settingsPath, agentKey, field string) (string, error)
```

The hash is stored in `opencode.json` as:

```json
{
  "agent": {
    "gentle-orchestrator": {
      "prompt": "...",
      "_gentle-ai-asset-hash": "sha256:a1b2c3..."
    }
  }
}
```

Prefix `sha256:` identifies the algorithm for future-proofing.

## Function Changes Detail

### `inlineOpenCodeSDDPrompts()` (inject.go:754-848)

The `preserveExistingOrchestratorPrompt=true` branch (lines 779-803) changes from:

```
preserve=true → read existing → migratePreserved → set as overlay prompt
```

To:

```
preserve=true → read existing prompt
             → read stored hash from agent._gentle-ai-asset-hash
             → compute hash of existing prompt
             → decide: stock | customized | first-run
             → set prompt and hash on orchestratorMap
```

When `preserveExistingOrchestratorPrompt=false`, the hash is still written (computed from the embedded asset) so future switches to `preserve=true` work correctly.

### `readOpenCodeAgentPrompt()` (inject.go:864-900)

No changes. The new `readOpenCodeAgentField()` generalizes the same pattern but extracts any string field.

### `mergeJSONFile()` (inject.go:1353-1382)

No changes. Deep merge already handles unknown string fields. The `_gentle-ai-asset-hash` field in the overlay's `gentle-orchestrator` map is a scalar string — `mergeObjects` writes it directly (line 237: `result[key] = overlayValue`).

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `computeAssetHash()` determinism | Same input → same hash; different input → different hash |
| Unit | First-run (no hash field) | Existing prompt preserved, hash set in overlay |
| Unit | Stock content (hash matches) | Prompt replaced with embedded asset, new hash set |
| Unit | User customized (hash differs) | Prompt preserved, stored hash kept |
| Unit | `preserve=false` path | Hash still computed and written from embedded asset |
| Unit | Corrupted hash field (non-sha256 prefix) | Treated as mismatch → preserve (safe degradation) |
| Integration | Full `Inject()` with preserve=true | End-to-end: opencode.json has correct prompt + hash |

## Migration / Rollout

No data migration required. The hash field is absent from existing configs — the first-run path handles this as "assume customized, set baseline hash." The `sha256:` prefix allows algorithm changes without breaking existing hashes.

**Rollback**: Remove hash comparison from `inlineOpenCodeSDDPrompts()`, revert to binary preserve/replace. The `_gentle-ai-asset-hash` field in `opencode.json` is inert metadata — no removal needed.

## Open Questions

None. The proposal and exploration resolved all design questions.
