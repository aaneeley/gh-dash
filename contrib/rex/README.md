# Rex recovery

Dash automatically checkpoints navigation when `REX_BLOCK` is set. Each pane
gets its own file, so two dashboards in the same repository stay independent.

Rex restores ordinary terminal programs as shells. To reopen an interrupted
Dash from that shell, source `resume.nu` from Nushell's `config.nu`:

```nu
source "/absolute/path/to/gh-dash/contrib/rex/resume.nu"
```

Find the actual config path with `$nu.config-path`. On macOS it is normally
`~/Library/Application Support/nushell/config.nu`, not `~/.config/nushell/config.nu`.

For another interactive shell, the equivalent startup hook is:

```sh
if [ -n "${REX_BLOCK:-}" ]; then
  gh dash --resume
fi
```

`--resume` is a no-op if there is no interrupted Dash, if Dash was intentionally
quit with `q`/Ctrl-C, or if another Dash already owns this pane's checkpoint.
It does not change OpenCode's own recovery. Restart an existing Dash once after
installing this build so that it starts checkpointing.

## What is restored

- Working directory and explicit/environment-selected config file
- Active view and section, including scratch sections
- Committed search filters and smart filtering
- Selected PR/issue by URL, or notification by ID, even if results reordered
- Previously loaded pages and the list scroll position
- PR preview tab, expanded summary, and scroll position
- Preview visibility, position override, and expanded help

The data is fetched afresh from GitHub. If the selected item no longer matches
the section, selection falls back to the saved row (clamped to the new list).
PR recovery can page through GitHub's 1000-result search window to find it.

Text currently being edited, confirmation dialogs, background commands, and
GitHub mutations are **not replayed**. Notifications are selected but not opened
automatically, to avoid marking them read. Local branches in the experimental
repository view are not checkpointed.

## Storage and manual use

Files live under `$XDG_STATE_HOME/gh-dash/rex` (default
`~/.local/state/gh-dash/rex`). `GH_DASH_STATE_DIR` can override that directory.
Only navigation metadata is stored: no credentials or GitHub response bodies.
Files are private, updated atomically, and only written when navigation changes.
A process-lifetime lock prevents nested shells from opening a duplicate Dash.

Outside Rex, opt in with an explicit checkpoint file:

```sh
gh dash --state-file /absolute/path/to/dashboard.json
# After an interruption (not an intentional quit):
gh dash --resume --state-file /absolute/path/to/dashboard.json
```

If a checkpoint is malformed or its directory/config is unavailable, recovery
reports an error and leaves the shell usable without overwriting the checkpoint.
