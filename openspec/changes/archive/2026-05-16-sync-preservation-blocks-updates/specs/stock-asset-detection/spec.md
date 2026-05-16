# Stock Asset Detection Specification

## Purpose

Deterministic detection of stock vs user-customized orchestrator prompt content via hash comparison, enabling safe asset updates while preserving genuine user edits during sync.

## ADDED Requirements

### Requirement: Hash Computation and Storage (REQ-1)

The system MUST compute a SHA-256 hex digest of the orchestrator prompt string that was last written to `opencode.json` and store it as a `_gentle-ai-asset-hash` string field on the `agent.gentle-orchestrator` object. The hash MUST be computed against the exact prompt value before any migration transforms (i.e., against the raw string stored in JSON).

#### Scenario: Fresh sync writes hash alongside prompt

- GIVEN no `opencode.json` exists
- WHEN sync runs with `PreserveOpenCodeOrchestratorPrompt=false`
- THEN the merged `opencode.json` contains `agent.gentle-orchestrator.prompt` with the embedded asset
- AND `agent.gentle-orchestrator._gentle-ai-asset-hash` equals SHA-256 of that prompt string

#### Scenario: Hash updated after stock replacement

- GIVEN `opencode.json` has a matching hash (stock prompt)
- WHEN sync replaces the prompt with a new embedded asset
- THEN `_gentle-ai-asset-hash` is updated to SHA-256 of the new prompt

### Requirement: Stock Detection — Replace on Hash Match (REQ-2)

When `PreserveOpenCodeOrchestratorPrompt=true` and the stored `_gentle-ai-asset-hash` matches SHA-256 of the existing prompt on disk, the system MUST treat the content as stock and replace it with the current embedded asset. The hash MUST then be updated to match the new asset.

#### Scenario: Stock prompt replaced with new version

- GIVEN `opencode.json` has `agent.gentle-orchestrator.prompt = "old stock"`
- AND `_gentle-ai-asset-hash` equals SHA-256 of `"old stock"`
- WHEN sync runs
- THEN the prompt is replaced with the current embedded asset
- AND `_gentle-ai-asset-hash` is updated to SHA-256 of the new prompt

#### Scenario: Stock detection is byte-exact after migration

- GIVEN `opencode.json` has a prompt with `sdd-orchestrator` references
- AND `_gentle-ai-asset-hash` equals SHA-256 of the pre-migration prompt
- WHEN sync runs with `PreserveOpenCodeOrchestratorPrompt=true`
- THEN the system compares the hash against the raw prompt (pre-migration)
- AND on match, replaces with the new embedded asset (already containing `gentle-orchestrator`)

### Requirement: Customization Detection — Preserve on Hash Mismatch (REQ-3)

When `PreserveOpenCodeOrchestratorPrompt=true` and the stored `_gentle-ai-asset-hash` does NOT match SHA-256 of the existing prompt on disk, the system MUST treat the content as user-customized and preserve it unchanged. The stored hash MUST NOT be modified.

#### Scenario: User-edited prompt preserved

- GIVEN `opencode.json` has `agent.gentle-orchestrator.prompt = "my custom prompt"`
- AND `_gentle-ai-asset-hash` equals SHA-256 of `"old stock"` (mismatch)
- WHEN sync runs
- THEN the prompt remains `"my custom prompt"` unchanged
- AND `_gentle-ai-asset-hash` remains unchanged

#### Scenario: External tool modified prompt between syncs

- GIVEN a previous sync wrote a stock prompt and its hash
- AND an external profile manager changed the prompt text without updating the hash
- WHEN sync runs
- THEN the hash mismatch is detected and the modified prompt is preserved

### Requirement: First-Run Migration — No Hash Field (REQ-4)

When `PreserveOpenCodeOrchestratorPrompt=true` and the `gentle-orchestrator` agent definition exists on disk but contains no `_gentle-ai-asset-hash` field, the system MUST assume the existing prompt is user-customized (safe default), preserve it, and set `_gentle-ai-asset-hash` to SHA-256 of the preserved prompt as a baseline for future comparisons.

#### Scenario: Upgrade from pre-hash version preserves content

- GIVEN `opencode.json` has `agent.gentle-orchestrator.prompt = "existing prompt"` with no hash field
- WHEN sync runs
- THEN the existing prompt is preserved unchanged
- AND `_gentle-ai-asset-hash` is set to SHA-256 of `"existing prompt"`
- AND subsequent syncs correctly detect stock vs customized

#### Scenario: No existing prompt — fallback to embedded asset

- GIVEN `opencode.json` has no `gentle-orchestrator` agent
- AND no legacy `sdd-orchestrator` or `gentleman` agent
- WHEN sync runs with `PreserveOpenCodeOrchestratorPrompt=true`
- THEN the embedded asset is used as prompt
- AND `_gentle-ai-asset-hash` is set to SHA-256 of the embedded asset

### Requirement: Hash Passed Through Deep Merge (REQ-5)

The `_gentle-ai-asset-hash` field MUST be included in the overlay JSON passed to `mergeJSONFile()` so that the existing deep merge writes it atomically alongside the prompt. The deep merge logic in `mergeObjects()` MUST treat `_gentle-ai-asset-hash` as a standard scalar string field — no special handling required.

#### Scenario: Hash survives deep merge

- GIVEN the overlay contains `agent.gentle-orchestrator._gentle-ai-asset-hash = "abc123"`
- WHEN `mergeJSONFile()` deep-merges the overlay into the existing `opencode.json`
- THEN the resulting file contains `_gentle-ai-asset-hash = "abc123"` on the orchestrator agent

### Requirement: Adapter Consistency (REQ-6)

The hash-based stock detection logic MUST apply identically for OpenCode and Kilocode adapters. Both adapters use the same `inlineOpenCodeSDDPrompts()` code path and the same `opencode.json` settings file.

#### Scenario: Kilocode adapter behaves identically to OpenCode

- GIVEN `inlineOpenCodeSDDPrompts()` is called for a Kilocode adapter
- WHEN the existing prompt hash matches stock
- THEN the prompt is replaced with the current embedded asset
- AND the behavior is identical to the OpenCode adapter case
