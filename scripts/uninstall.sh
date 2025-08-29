#!/usr/bin/env bash
set -euo pipefail

BIN="${HOME}/.local/bin/tat"
BRC="${HOME}/.bashrc"
ZRC="${HOME}/.zshrc"
HOOK_BASH='source ~/.tat/tat.bash'
HOOK_ZSH='source ~/.tat/tat.zsh'

remove_line() { file="$1"; line="$2"; [ -f "$file" ] || return 0; tmp="$(mktemp)"; grep -Fv "$line" "$file" > "$tmp" || true; mv "$tmp" "$file"; }

echo "Removing tat binary (if exists) ..."
rm -f "$BIN"

echo "Removing shell hook source lines ..."
remove_line "$BRC" "$HOOK_BASH"
remove_line "$ZRC" "$HOOK_ZSH"

echo "Keeping your data at ~/.tat (DB/config). To remove it permanently, run:"
echo "  rm -rf ~/.tat"
echo "Uninstall complete. Restart your terminal."