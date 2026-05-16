# Gentle AI — Agent Skills Index

When working on this project, load the relevant skill(s) BEFORE writing any code.

Naming convention: `gentle-ai-*` skills are repo-specific workflow skills. Unprefixed skills are portable writing or work-unit skills and intentionally keep their canonical names.

## How to Use

1. Check the trigger column to find skills that match your current task
2. Load the skill by reading the SKILL.md file at the listed path
3. Follow ALL patterns and rules from the loaded skill
4. Multiple skills can apply simultaneously

## Skills

| Skill | Trigger | Path |
|-------|---------|------|
| `gentle-ai-issue-creation` | When creating a GitHub issue, reporting a bug, or requesting a feature. | [`skills/issue-creation/SKILL.md`](skills/issue-creation/SKILL.md) |
| `gentle-ai-branch-pr` | When creating a pull request, opening a PR, or preparing changes for review. | [`skills/branch-pr/SKILL.md`](skills/branch-pr/SKILL.md) |
| `gentle-ai-chained-pr` | When a change is too large for one review, or when creating chained/stacked pull requests. | [`skills/chained-pr/SKILL.md`](skills/chained-pr/SKILL.md) |
| `cognitive-doc-design` | When writing docs that must reduce cognitive load for readers or reviewers. | [`skills/cognitive-doc-design/SKILL.md`](skills/cognitive-doc-design/SKILL.md) |
| `comment-writer` | When drafting human comments, PR feedback, issue replies, or async updates. | [`skills/comment-writer/SKILL.md`](skills/comment-writer/SKILL.md) |
| `work-unit-commits` | When splitting implementation work into deliverable commits or chained PRs. | [`skills/work-unit-commits/SKILL.md`](skills/work-unit-commits/SKILL.md) |

## Spec-Kit Delegation Rules (MANDATORY)

Spec-kit commands that generate multiple artifact files MUST be delegated to their corresponding sub-agents. Do NOT execute inline.

| Command | Sub-agent | Reason |
|---------|-----------|--------|
| `/speckit.specify` | `speckit-specify` | Creates spec.md + checklists + runs hooks |
| `/speckit.plan` | `speckit-plan` | Generates 5+ files: plan, research, data-model, quickstart, contracts |
| `/speckit.tasks` | `speckit-tasks` | Generates tasks.md with dependency analysis |
| `/speckit.implement` | `speckit-implement` | Executes tasks across multiple source files |

**Inline is acceptable ONLY for**: reading 1-3 files to decide/verify, writing a single atomic file, running git/gh state commands.

**Violation signal**: if you're about to write 2+ non-trivial files for a speckit phase — STOP — delegate to the sub-agent instead.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
at `specs/003-installer-pm-choice/plan.md`
<!-- SPECKIT END -->
