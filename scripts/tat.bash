# Source this in ~/.bashrc: source ~/.tat/tat.bash

if [[ -n "$BASH_VERSION" ]]; then
  tat_preexec() {
    [[ -n "$TAT_DISABLE" ]] && return
    printf '{"event":"preexec","cmd":%q,"cwd":%q,"ts":%q}\n' \
      "$BASH_COMMAND" "$PWD" "$(date -Iseconds)" | tat record >/dev/null 2>&1
    export TAT_LAST_START_TS=$(date +%s%3N)
  }
  tat_postexec() {
    [[ -n "$TAT_DISABLE" ]] && return
    local ec=$?
    local now=$(date +%s%3N)
    local dur=$(( now - ${TAT_LAST_START_TS:-now} ))
    printf '{"event":"postexec","cmd":%q,"cwd":%q,"ts":%q,"exit":%d,"duration_ms":%d}\n' \
      "$BASH_COMMAND" "$PWD" "$(date -Iseconds)" "$ec" "$dur" | tat record >/dev/null 2>&1
  }
  trap 'tat_preexec' DEBUG
  PROMPT_COMMAND='tat_postexec; '"$PROMPT_COMMAND"
fi
