# Tasks: Sync Preservation Logic Fix

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 120–180 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Hash helper + read helper + 3 detection branches + tests | PR 1 | ~120–180 lines; tests first (RED→GREEN) |

## Phase 1: Foundation — Pure Functions (Strict TDD)

- [ ] 1.1 RED: Add `TestComputeAssetHash` in `inject_test.go` — determinism, empty string, collision resistance between different inputs
- [ ] 1.2 GREEN: Create `computeAssetHash(content string) string` in `inject.go` using `crypto/sha256`; prefix `sha256:`, hex-encode
- [ ] 1.3 RED: Add `TestReadOpenCodeAgentField` — returns field value when present, empty string when absent, empty string on file-not-found
- [ ] 1.4 GREEN: Create `readOpenCodeAgentField(settingsPath, agentKey, field string) (string, error)` in `inject.go` — reads JSON, walks `root.agent.{agentKey}.{field}`
- [ ] 1.5 RED: Add `TestResolveOrchestratorPrompt_FirstRun` — no stored hash → preserve existing prompt, set hash from current content
- [ ] 1.6 RED: Add `TestResolveOrchestratorPrompt_StockDetected` — hash matches → return embedded asset + new hash
- [ ] 1.7 RED: Add `TestResolveOrchestratorPrompt_Customized` — hash differs → preserve existing prompt, keep stored hash
- [ ] 1.8 GREEN: Create `resolveOrchestratorPrompt(settingsPath, agentKey string, preserve bool) (prompt, hash string, err error)` — implements 3-branch detection logic per design data flow

## Phase 2: Wiring — Integrate into `inlineOpenCodeSDDPrompts` (Strict TDD)

- [ ] 2.1 RED: Add `TestInlineOpenCodeSDDPrompts_PreserveStockPrompt_ReplacesWithAsset` — existing prompt = embedded asset hash, verify overlay gets NEW asset + updated hash
- [ ] 2.2 RED: Add `TestInlineOpenCodeSDDPrompts_PreserveCustomizedPrompt_KeepsExisting` — existing prompt differs from hash, verify overlay keeps existing prompt + stored hash
- [ ] 2.3 RED: Add `TestInlineOpenCodeSDDPrompts_PreserveFirstRun_SetsBaselineHash` — no hash field, verify overlay preserves current prompt + sets baseline hash
- [ ] 2.4 GREEN: Replace `preserveExistingOrchestratorPrompt` branch (lines 779–803) with call to `resolveOrchestratorPrompt`; set `orchestratorMap["_gentle-ai-asset-hash"]` alongside `orchestratorMap["prompt"]`
- [ ] 2.5 RED: Add `TestInlineOpenCodeSDDPrompts_PreserveFalse_WritesAssetHash` — when preserve=false, hash is still computed from embedded asset and written
- [ ] 2.6 GREEN: In non-preserve branch, compute and set `_gentle-ai-asset-hash` from embedded asset

## Phase 3: Integration Verification (Strict TDD)

- [ ] 3.1 RED: Add `TestInject_PreserveStockOrchestrator_UpdatesOnSecondSync` — first sync sets baseline hash; second sync with unchanged prompt detects stock → replaces with new asset
- [ ] 3.2 RED: Add `TestInject_PreserveCustomizedOrchestrator_NeverOverwrites` — customize prompt after first sync; second sync preserves custom content
- [ ] 3.3 Run `go test ./internal/components/sdd/...` — all existing + new tests pass
