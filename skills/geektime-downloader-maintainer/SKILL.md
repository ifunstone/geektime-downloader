---
name: geektime-downloader-maintainer
description: Maintain and extend the Geektime Downloader repository. Use when Codex needs to modify this Go codebase, trace download behavior, add support for new Geektime product types, adjust the Fyne desktop UI, fix download bugs across PDF/Markdown/Audio/Video flows, or choose the correct package to edit based on the current repository architecture.
---

# Geektime Downloader Maintainer

## Overview

Work from the current repository structure instead of guessing. Route each change to the layer that owns it, preserve the existing separation between UI, service orchestration, API client code, and content downloaders, and verify changes with focused Go tests when possible.

## Follow the Execution Path

Start by locating which entrypoint the user is exercising.

- For the current desktop app flow, begin at `main.go`, then `internal/uiapp/app.go`, then `internal/app/service.go`.
- For legacy CLI behavior, inspect `cmd/root.go` and `internal/fsm/runner.go`.
- For actual download work, continue into `internal/course/downloader.go` and then the media-specific package.

Read [references/repo-map.md](references/repo-map.md) before making non-trivial changes. Use it to choose the correct package and avoid duplicating logic that already exists in another layer.

## Route Changes to the Right Layer

- Change `internal/uiapp` when the request is about desktop fields, buttons, persistence, status display, or log refresh behavior.
- Change `internal/app` when the request is about product-type selection, course resolution, direct-download rules, or UI-facing orchestration.
- Change `internal/fsm` and `internal/ui` only when preserving or fixing the legacy terminal workflow.
- Change `internal/geektime` when API paths, request/response handling, auth behavior, or product parsing must change.
- Change `internal/course` when the issue is cross-media download orchestration, skip logic, directory layout, or retry/wait behavior.
- Change `internal/pdf`, `internal/markdown`, `internal/audio`, or `internal/video` when the bug is format-specific.
- Change `internal/config` when validation, defaults, or cookie-loading behavior must change.
- Change `internal/pkg` only for genuinely shared helpers.

## Preserve Existing Patterns

- Reuse `app.Service` for UI-triggered workflows instead of duplicating API and downloader wiring in `internal/uiapp`.
- Keep product-type rules centralized in `internal/app/service.go` unless the change is strictly legacy-CLI-specific.
- Keep Geektime HTTP concerns inside `internal/geektime`; do not spread raw Resty calls into UI or downloader packages.
- Keep file naming and filesystem behavior aligned with `internal/pkg/filenamify` and `internal/pkg/files`.
- Respect the current output bitmask for text courses: PDF `1`, Markdown `2`, Audio `4`.

## Validate Pragmatically

Use the smallest validation that proves the change.

- Run `go test ./...` for broad verification when the change touches shared behavior.
- Run targeted tests first if the change is isolated, for example `go test ./internal/markdown ./internal/pkg/...`.
- If the change affects the desktop app and GUI execution is impractical, at least verify buildability with `go test ./...` and review compile paths through `internal/uiapp` and `internal/app`.
- If a bug involves runtime dependencies like Chrome or FFmpeg, state clearly whether you verified only compile/test coverage or also exercised the end-to-end flow.

## Common Tasks

- Add a new downloadable product type: update product option definitions in `internal/app/service.go`, then ensure the matching Geektime client method and download path exist.
- Fix course loading or access checks: inspect `internal/app/service.go` and `internal/geektime`.
- Fix article/video skipping, overwrite, or directory behavior: inspect `internal/course/downloader.go`.
- Fix empty PDF, timeout, or page-render issues: inspect `internal/pdf` plus config fields consumed by the UI or CLI.
- Fix UI config persistence or defaults: inspect `internal/uiapp/settings.go`, `internal/uiapp/app.go`, and `internal/app/service.go`.
