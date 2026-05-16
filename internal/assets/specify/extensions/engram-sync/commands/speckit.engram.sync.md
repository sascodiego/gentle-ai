---
description: "Sync spec-kit artifacts to engram persistent memory"
---

# Sync Artifacts to Engram

Automatically sync spec-kit artifacts to engram persistent memory so they survive across sessions.

## Behavior

This command is invoked as a hook after core spec-kit commands complete. It:

1. Resolves the current feature directory using spec-kit's `common.sh` helpers
2. Syncs all available artifacts (spec, exploration, plan, research, data-model, quickstart, contracts) to engram
3. Uses topic key prefix `spec-kit/{feature-name}/{artifact}` for organized retrieval
4. Skips artifacts that don't exist yet (graceful degradation)

## Execution

Run the engram-sync script:

```bash
bash .specify/extensions/engram-sync/scripts/bash/engram-sync.sh
```

The script auto-detects the current feature from `.specify/feature.json` and syncs all existing artifacts to engram using the CLI (`engram save`).

## Graceful Degradation

- If engram CLI is not found: skips with a warning
- If no feature directory is set: syncs at project level
- If an artifact file doesn't exist: skips that artifact silently
- If `engram save` fails for a specific artifact: logs warning, continues with remaining artifacts
