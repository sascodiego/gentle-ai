# Spec-Kit Extension Injector Specification

## Purpose

Detects an existing `.specify/` directory at the project root and writes embedded spec-kit extension files (extension.yml + bash scripts) into it. This automates the distribution of spec-kit extensions — particularly `engram-sync` — through Gentle AI's existing install/sync/update/upgrade pipeline.

## Requirements

### REQ-1: Asset Embedding

The system SHALL embed all files under `internal/assets/specify/` via `//go:embed all:specify` in `assets.go`, making them accessible through `assets.FS`.

Extension files for `engram-sync` MUST include at minimum:
- `specify/extensions/engram-sync/extension.yml`
- `specify/extensions/engram-sync/scripts/bash/engram-sync.sh`

#### Scenario: Embedded assets are accessible at runtime

- GIVEN the binary is built with the `all:specify` embed directive
- WHEN `assets.FS.ReadDir("specify/extensions/engram-sync")` is called
- THEN the result contains entries for `extension.yml` and `scripts/`

### REQ-2: Injection Step (3d)

The system SHALL add step 3d to the `Inject()` pipeline, after step 3c (sub-agents) and before step 4 (skill-registry automation).

Step 3d MUST:
1. Check `opts.SkipSpecKitExtensions` — skip entirely if true
2. Call `findProjectRoot(opts.WorkspaceDir)` — skip silently if no root found
3. `os.Stat(projectRoot + "/.specify")` — skip silently if `.specify/` does not exist
4. NEVER create `.specify/` or any parent directory
5. Walk `assets.FS` under `specify/` recursively and write each file to `<projectRoot>/.specify/<relPath>` using `filemerge.WriteFileAtomic`
6. Track `changed` and `files` in the existing `InjectionResult`

The installer MUST always overwrite existing files — no checksum comparison, no "do not edit" header.

#### Scenario: Happy path — project has `.specify/`

- GIVEN `opts.WorkspaceDir` points to a directory inside a project with `.specify/` at its root
- AND `opts.SkipSpecKitExtensions` is false
- WHEN `Inject()` is called
- THEN embedded files under `specify/` are written to `<projectRoot>/.specify/` preserving relative paths
- AND `InjectionResult.Changed` is true
- AND `InjectionResult.Files` contains the written paths

#### Scenario: Skip — no project root found

- GIVEN `opts.WorkspaceDir` is empty or does not contain any project root markers
- WHEN `Inject()` is called
- THEN step 3d is silently skipped
- AND `InjectionResult` is unaffected by step 3d

#### Scenario: Skip — project root found but no `.specify/`

- GIVEN `findProjectRoot` returns a valid project root
- AND `<projectRoot>/.specify` does not exist
- WHEN `Inject()` is called
- THEN step 3d is silently skipped
- AND no directories are created

#### Scenario: Overwrite — existing files differ

- GIVEN `<projectRoot>/.specify/extensions/engram-sync/extension.yml` exists with user-modified content
- WHEN `Inject()` is called
- THEN the file is overwritten with the embedded content
- AND `InjectionResult.Changed` is true

#### Scenario: Idempotency — content unchanged

- GIVEN `<projectRoot>/.specify/` contains files that match the embedded assets exactly
- WHEN `Inject()` is called
- THEN `InjectionResult.Changed` is false (from step 3d's contribution)
- AND no files are re-written

### REQ-3: Skip Flag

`InjectOptions` SHALL include a `SkipSpecKitExtensions bool` field. When true, step 3d is entirely skipped regardless of other conditions.

#### Scenario: Skip flag disables injection

- GIVEN `opts.SkipSpecKitExtensions` is true
- AND project root exists with `.specify/` present
- WHEN `Inject()` is called
- THEN step 3d is skipped
- AND no spec-kit extension files are written

### REQ-4: Entry Point Coverage

Step 3d MUST execute during all paths that call `Inject()`: `install`, `sync`, `update`, and `upgrade`. No separate call site is needed — the step lives inside `Inject()`.

#### Scenario: Injection runs from all entry points

- GIVEN any CLI command (`install`, `sync`, `run`) that calls `sdd.Inject()`
- AND the project has `.specify/` at its root
- WHEN the command is executed
- THEN spec-kit extension files are written to `<projectRoot>/.specify/`

### REQ-5: Error Handling

Errors during individual file writes in step 3d MUST be logged but SHALL NOT fail the entire `Inject()` pipeline. Missing project root and missing `.specify/` are not errors — they are silent skips.

#### Scenario: File write error does not abort pipeline

- GIVEN step 3d attempts to write a file and `WriteFileAtomic` returns an error
- WHEN the error is encountered
- THEN the error is logged
- AND `Inject()` continues to step 4 without returning an error
- AND previously written files in step 3d remain on disk
