#!/usr/bin/env bash
set -euo pipefail

# Resolve common helpers from the spec-kit installation.
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMMON_SH="$SCRIPT_DIR/../../../../scripts/bash/common.sh"

if [[ -f "$COMMON_SH" ]]; then
  # shellcheck disable=SC1090
  source "$COMMON_SH"
fi

# Resolve the current feature paths.
# get_feature_paths prints shell variable assignments; we must eval them.
FEATURE_NAME=""
FEATURE_DIR=""
if declare -f get_feature_paths > /dev/null 2>&1; then
  _paths_output=$(get_feature_paths) || true
  if [[ -n "$_paths_output" ]]; then
    eval "$_paths_output"
  fi
  unset _paths_output
  FEATURE_NAME="${FEATURE_NAME:-}"
fi

# Extract the feature directory basename for the topic prefix.
FEATURE_BASENAME=""
if [[ -n "${FEATURE_DIR:-}" ]]; then
  FEATURE_BASENAME="$(basename "$FEATURE_DIR")"
fi

# Determine the topic key prefix. If we have a feature directory, scope to it;
# otherwise sync at the project level.
TOPIC_PREFIX="spec-kit"
if [[ -n "$FEATURE_BASENAME" ]]; then
  TOPIC_PREFIX="spec-kit/${FEATURE_BASENAME}"
fi

# Sync each known artifact type to engram.
ARTIFACTS=(spec exploration plan research data-model quickstart contracts)

for artifact in "${ARTIFACTS[@]}"; do
  # Derive the artifact file path from the feature directory.
  ARTIFACT_FILE=""
  if [[ -n "${FEATURE_DIR:-}" ]] && [[ -f "${FEATURE_DIR}/${artifact}.md" ]]; then
    ARTIFACT_FILE="${FEATURE_DIR}/${artifact}.md"
  fi

  if [[ -z "$ARTIFACT_FILE" ]]; then
    continue
  fi

  TOPIC_KEY="${TOPIC_PREFIX}/${artifact}"
  echo "[engram-sync] syncing ${artifact} -> ${TOPIC_KEY}"

  if command -v engram > /dev/null 2>&1; then
    ARTIFACT_CONTENT="$(cat "$ARTIFACT_FILE")"
    engram save \
      "$TOPIC_KEY" \
      "$ARTIFACT_CONTENT" \
      --topic "$TOPIC_KEY" \
      --type architecture \
      || echo "[engram-sync] warning: engram save failed for ${artifact}" >&2
  else
    echo "[engram-sync] skipping ${artifact} — engram CLI not found" >&2
  fi
done

echo "[engram-sync] done"
