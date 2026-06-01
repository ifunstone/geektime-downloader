# Repository Map

## Primary entrypoints

- `main.go`: current application entrypoint; launches the Fyne desktop app in `internal/uiapp`.
- `cmd/root.go`: legacy Cobra CLI entrypoint; validates config, creates the Geektime client, and runs the FSM flow.

## Application layers

### `internal/uiapp`

Own the desktop GUI.

- `app.go`: four-tab GUI shell, form state, button handlers, async load/download actions, progress display, pause/resume/cancel behavior, log refresh, and settings save/restore.
- `conversion.go`: TS scan, pending-file list, conversion progress updates, and TS-to-MP4 batch processing triggered from the GUI.
- `settings.go`: persisted desktop settings.
- `logs.go`: log file reading, truncation-to-tail, and log deletion support for the GUI.

Use this package for UI behavior changes, not for new downloader rules.

### `internal/app`

Own UI-facing orchestration and product resolution.

- `service.go`: default config, product-type definitions, course resolution, direct-video handling, and download dispatch.
- `legacy.go`: adapters between the newer service model and legacy option types.

Use this package when the user asks for new product types, different selection rules, direct-product handling, or different orchestration from the desktop app.

### `internal/fsm` and `internal/ui`

Own the legacy terminal flow.

- `internal/fsm/runner.go`: state machine for product type selection, course loading, action selection, and article selection.
- `internal/ui/*`: prompt-based terminal interactions.

Touch these packages only if the terminal workflow still matters for the requested behavior.

## Domain/API layer

### `internal/geektime`

Own all Geektime API integration and product-specific fetch logic.

- `client.go`: shared Resty client setup, request execution, auth/rate-limit error mapping.
- `account.go`, `enterprise.go`, `geektime.go`, `university.go`: API-specific fetch methods.
- `response/`: typed response models.

If the bug is "wrong endpoint", "bad parsing", "auth expired", "enterprise/university fetch mismatch", or "need a new API call", start here.

## Download orchestration

### `internal/course`

Own cross-format download flow.

- `downloader.go`: output bitmask handling, skip logic, directory creation, article loops, random waits, direct-video helpers, and dispatch into specific media packages.

This is the right place for:

- download-all vs single-article behavior
- overwrite/skip decisions
- section subdirectory behavior
- choosing which media downloader to call

## Media-specific implementations

### `internal/pdf`

Own Chrome-based PDF rendering and PDF stream writing.

Check this package when the issue mentions blank PDFs, timeouts, rendering waits, or stream output.

### `internal/markdown`

Own article-to-Markdown conversion and related tests.

### `internal/audio`

Own article audio downloads.

### `internal/video`

Own video retrieval, m3u8/ts handling, and mp4 conversion.

- `vod/`: video-on-demand API structs and helpers.

Use this area for download-time quality selection, ts/mp4 generation, FFmpeg integration, or university/enterprise video specifics.

## Shared support packages

### `internal/config`

Own config schema, validation, and cookie input behavior.

### `internal/pkg`

Own reusable helpers.

- `ffmpeg`: FFmpeg integration and tests.
- `files`: file existence helpers.
- `filenamify`: safe filename normalization and tests.
- `logger`: logging setup and discard logger.
- `m3u8`, `downloader`, `crypto`: lower-level shared utilities.

Avoid putting business rules here unless they are truly reusable across multiple higher-level packages.

## Change routing cheat sheet

- "Add a new UI field/button" -> `internal/uiapp`
- "Adjust the four-page GUI or settings persistence" -> `internal/uiapp/app.go` + `internal/uiapp/settings.go`
- "Fix log tab rendering, truncation, refresh, or delete behavior" -> `internal/uiapp/logs.go` + `internal/uiapp/app.go`
- "Fix TS directory scan or batch conversion in the GUI" -> `internal/uiapp/conversion.go` + `internal/pkg/ffmpeg`
- "Support another Geektime product type" -> `internal/app` + `internal/geektime` + maybe `internal/course` or `internal/video`
- "Course ID accepted but wrong product check" -> `internal/app/service.go` or `internal/fsm/runner.go`
- "Download should skip or overwrite differently" -> `internal/course/downloader.go`
- "PDF generation too slow or blank" -> `internal/pdf` + config consumers
- "Markdown output wrong" -> `internal/markdown`
- "Video conversion or quality issue" -> `internal/video` + `internal/pkg/ffmpeg`
- "Cookie/auth problem" -> `internal/config` + `internal/geektime/client.go`
