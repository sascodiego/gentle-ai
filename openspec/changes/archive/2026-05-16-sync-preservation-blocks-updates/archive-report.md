# Archive Report: sync-preservation-blocks-updates

**Change**: sync-preservation-blocks-updates
**Archived**: 2026-05-16
**Verdict**: PASS
**Artifact Store**: hybrid (engram + openspec)

## Change Summary

Fixed the sync system's binary preserve/replace logic in `inlineOpenCodeSDDPrompts()` that treated ALL existing orchestrator prompts as user-customized when `PreserveOpenCodeOrchestratorPrompt=true`, blocking asset updates for stock content that was never edited. Replaced with hash-based stock detection using SHA-256 via a `_gentle-ai-asset-hash` field in opencode.json.

## Artifact Lineage (Engram Observation IDs)

| Artifact | Obs ID | Topic Key |
|----------|--------|-----------|
| Exploration | #814 | sdd/sync-preservation-blocks-updates/explore |
| Proposal | #815 | sdd/sync-preservation-blocks-updates/proposal |
| Spec | #816 | sdd/sync-preservation-blocks-updates/spec |
| Design | #817 | sdd/sync-preservation-blocks-updates/design |
| Tasks | #819 | sdd/sync-preservation-blocks-updates/tasks |
| Apply Progress | #820 | sdd/sync-preservation-blocks-updates/apply-progress |
| Verify Report | #821 | sdd/sync-preservation-blocks-updates/verify-report |

## Spec Sync

| Domain | Action | Details |
|--------|--------|---------|
| stock-asset-detection | Created | New main spec created with 6 requirements (REQ-1 through REQ-6), 12 scenarios |

No existing main spec existed for `stock-asset-detection` — the delta spec was a full spec and was copied directly to `openspec/specs/stock-asset-detection/spec.md`.

## Archive Location

`openspec/changes/archive/2026-05-16-sync-preservation-blocks-updates/`

### Archive Contents
- proposal.md ✅
- specs/stock-asset-detection/spec.md ✅
- design.md ✅
- tasks.md ✅ (18/18 tasks complete)
- verify-report.md ✅

## Verification Summary

| Metric | Value |
|--------|-------|
| Tasks total | 18 |
| Tasks complete | 18 |
| Build | ✅ Passed |
| Tests | ✅ 240 passed / 0 failed |
| Spec compliance | 12/12 scenarios COMPLIANT |
| TDD compliance | 6/6 checks passed |
| Critical issues | None |
| Verdict | PASS |

## Files Changed During Implementation

| File | Action | What Was Done |
|------|--------|---------------|
| internal/components/sdd/inject.go | Modified | Added crypto/sha256 import, computeAssetHash(), readOpenCodeAgentField(), resolveOrchestratorPrompt(); replaced preserve branch in inlineOpenCodeSDDPrompts() |
| internal/components/sdd/inject_test.go | Modified | Added 17 new test functions covering all 3 branches + integration |

## SDD Cycle Complete

The change has been fully explored, proposed, specified, designed, tasked, implemented (18/18 tasks), verified (PASS), and archived.

## Source of Truth Updated

The following spec now reflects the new behavior:
- `openspec/specs/stock-asset-detection/spec.md`
