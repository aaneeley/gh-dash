# Source this from Nushell's config.nu. Each Rex pane has its own checkpoint.
# No checkpoint (or an already-running Dash) exits silently to the shell.
if $nu.is-interactive and (($env.REX_BLOCK? | default '') != '') {
    try { ^gh dash --resume }
}
