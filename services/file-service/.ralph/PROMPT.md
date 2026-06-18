# Ralph Loop Instructions

## You are in a Ralph autonomous loop

CLAUDE.md defines your role, rules, and commands. This file adds Ralph loop mechanics.

## How Ralph Works

Ralph runs you repeatedly. Each invocation is one **loop**:
1. Read `.ralph/fix_plan.md` — pick the highest priority unchecked task
2. Execute ONE task (or a small batch of identical repetitive fixes)
3. Verify your work (build, type-check, lint — see CLAUDE.md for commands)
4. Mark the task `[x]` in fix_plan.md
5. Output `RALPH_STATUS` block (see below)
6. Exit — Ralph will start the next loop

## Key Discipline

- **ONE task per loop** — don't try to do everything at once
- **Verify before marking done** — a task isn't complete until build/check passes
- **Minimal changes** — fix the issue, don't refactor unrelated code
- **Commit after meaningful progress** — descriptive conventional commit messages

## Protected Files (DO NOT MODIFY)

These are Ralph's runtime files. Never delete, move, or overwrite:
- `.ralph/` (entire directory and all contents)
- `.ralphrc` (project configuration)

## Status Reporting (REQUIRED)

End EVERY response with this block:

```
---RALPH_STATUS---
STATUS: IN_PROGRESS | COMPLETE | BLOCKED
TASKS_COMPLETED_THIS_LOOP: <number>
FILES_MODIFIED: <number>
TESTS_STATUS: PASSING | FAILING | NOT_RUN
WORK_TYPE: IMPLEMENTATION | REVIEW | TESTING | DOCUMENTATION | REFACTORING
EXIT_SIGNAL: false | true
RECOMMENDATION: <one line — what to do next>
---END_RALPH_STATUS---
```

### EXIT_SIGNAL rules

Set `EXIT_SIGNAL: true` ONLY when ALL items in fix_plan.md are `[x]`.

Set `STATUS: BLOCKED` if:
- You cannot proceed without external input
- The same error persists across multiple attempts
- A required file or dependency is missing

### What Ralph does with this

Ralph parses this block to decide: continue (`IN_PROGRESS`), stop (`COMPLETE`), or halt with circuit breaker (`BLOCKED` / repeated no-progress). Write it exactly as shown — Ralph's parser depends on the format.

## Current Task

Read `.ralph/fix_plan.md` and work on the highest priority unchecked item.
