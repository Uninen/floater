# Floater

A simple native tool for running a Web page in a window that can be toggled to float always on top.

Pre-Alpha stage.

## Goals

- [ ] Ability to set an URL to be displayesd in a single chromeless window.
- [ ] Ability to toggle a window staying on top.
- [ ] Ability to run multiple windows.

## North Star

This app should look and feel as native as possible. The target is macOS 27 with dark theme. Every single UI/UX-decision you do must keep this North Star in mind.

## Agent Workflow

Design, preview, screenshot, test, and visually verify only dark-mode UI. Never spend time on light-mode UI.

## Tooling

- MyGo framework: <https://github.com/egoist/mygo>
- Local checkout: `~/Code/studyrepos/mygo` -- study this when you need to study the tool or docs
  - Docs: `~/Code/studyrepos/mygo/docs`
- `learnings.md` - your self-documented learnings based on past work and failures

## Commands

- `go tool mygo dev` - develop with live reload
- `go test` - run functional tests without a window
- `go tool mygo build` - package the app

When user asks to build the app it means build it and copy result to `/Applications/`, do NOTHING else.

## Do Not Interrupt the User's Desktop

- Edit files normally with a user-started development runner active. Its automatic reloads are authorized; do not ask for confirmation or stop that runner.
- Do not start `go tool mygo dev`, launch/relaunch GUI apps, activate/raise windows, or send desktop keyboard/mouse events without explicit permission to interrupt the user's work.
- Stop agent-owned live-reload runners after validation; edits must not trigger unexpected app launches.

## Code Style

- ZERO unneeded complexity or unnecessary defensive coding; always strive for simplest best possible architecture.
- Every added comment or docstring MUST be justified: **why**-comments for decisions the code can't show. Trivial comments that restate the code are FORBIDDEN. If you can't justify it, don't write it.
- One fact lives in exactly ONE place. Never duplicate the same explanation across files — refer to the other module, component, or function by name, never by line number.
- Every test is a permanent cost: add one ONLY if breaking the code makes it fail AND no existing test already would. Test whole user-facing flows, never MyGo or the stdlib; fake only system boundaries; regression tests must fail before the fix.
- Touching tests means pruning them: delete low-value ones, merge duplicates, report the delta ("added 2, removed 3").
- Write or rewrite all text (comments, docstrings, learnings, your responses, everything) in plain, layman-readable English in the spirit of ASD-STE100 Simplified Technical English: short sentences, active voice, simple tenses, one word one meaning, condition before command, every technical term defined at first use, no AI slop.

## Document Your Learnings

`learnings.md` exists so that a problem which took you many failed attempts to solve gets solved on the first try next time. Record a lesson only when you hit that kind of struggle and found the solution: state the symptom you'll recognize andthe fix that worked.

Never record task-specific trivia or anything reading the docs would already tell you. If it wouldn't have saved you real time on this task, it doesn't belong.

Keep it lean, prune stale entries, and never duplicate this file.

## Post-Task Checklist

Check when the full task is finished, not after every user message.

- [ ] Tests pass AND are properly gardened. You MUST provide a VERY compelling justification for EVERY added new test.
- [ ] Learnings recorded per [Document Your Learnings](#document-your-learnings), or none.

Always report that you have followed this checklist after you complete any meaningfully large task that requires things like testing.
