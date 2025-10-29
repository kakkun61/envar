_envar() {
  local previous_exit_status=$?
  # Control-C を一旦無効に
  trap -- '' SIGINT
  # envar コマンドの出力を評価して環境変数を設定・解除する
  local script="$(envar $$)"
  if [[ -n "$script" ]]
  then
    echo "$script" | sed -E 's/^export ([^=]+).*/export \1/'| sed 's/^/envar: /'
    eval "$script"
  fi
  # Control-C を元に戻す
  trap - SIGINT
  return $previous_exit_status
}

if [[ ";${PROMPT_COMMAND[*]:-};" != *";_envar;"* ]]
then
  if [[ "$(declare -p PROMPT_COMMAND 2>&1)" == "declare -a"* ]]
  then
    PROMPT_COMMAND=(_envar "${PROMPT_COMMAND[@]}")
  else
    PROMPT_COMMAND="_envar${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
  fi
fi
