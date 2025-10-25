

_gh_auth_switch() {
  local previous_exit_status=$?
  # Control-C を一旦無効に
  trap -- '' SIGINT
  # 今のユーザーをキャッシュから取得
  local cache_dir="${XDG_CACHE_HOME:-}"
  if [[ "$cache_dir" == "" ]]
  then
    if [[ "$(uname)" == "Darwin" ]]
    then
      cache_dir="$HOME/Library/Caches"
    else
      cache_dir="$HOME/.cache"
    fi
  fi
  local cached_current_user_file="$cache_dir/gh-auth-switch/current_user"
  if [[ ! -f "$cached_current_user_file" ]]
  then
    mkdir -p "$(dirname "$cached_current_user_file")"
    touch "$cached_current_user_file"
  fi
  local cached_current_user="$(cat "$cached_current_user_file")"
  # ディレクトリーにもとづいてユーザーを切り替え、そのユーザーを返す
  local current_user="$(gh-auth-switch)"
  if [[ "$current_user" != "$cached_current_user" ]]
  then
    echo "gh-auth-switch: " "$current_user"
  fi
  echo "$current_user" > "$cached_current_user_file"
  # Control-C を元に戻す
  trap - SIGINT
  return $previous_exit_status
}

if [[ ";${PROMPT_COMMAND[*]:-};" != *";_gh_auth_switch;"* ]]
then
  if [[ "$(declare -p PROMPT_COMMAND 2>&1)" == "declare -a"* ]]
  then
    PROMPT_COMMAND=(_gh_auth_switch "${PROMPT_COMMAND[@]}")
  else
    PROMPT_COMMAND="_gh_auth_switch${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
  fi
fi
