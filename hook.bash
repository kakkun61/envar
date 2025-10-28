

_gh_auth_switch() {
  local previous_exit_status=$?
  # Control-C を一旦無効に
  trap -- '' SIGINT
  # GH_TOKEN について
  # 参照 https://cli.github.com/manual/gh_help_environment
  local previous_GH_TOKEN="$GH_TOKEN"
  # ディレクトリーにもとづいたユーザーを返す
  local user="$(gh-auth-switch)"
  GH_TOKEN="$(gh auth token -u "$user")"
  # ユーザーが切り替わるときだけ表示する
  if [[ "$GH_TOKEN" != "$previous_GH_TOKEN" ]]
  then
    echo "gh-auth-switch: $user"
  fi
  export GH_TOKEN
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
