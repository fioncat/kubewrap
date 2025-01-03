function {{name}}() {
  local NEED_SOURCE_CODE=302
  local DEFAULT_EXECUTABLE_PATH="kubewrap"
  declare -a opts

  while test $# -gt 0; do
    opts+=( "$1" )
    shift
  done

  local executable_path
  if [[ -n "$KUBEWRAP_EXECUTABLE_PATH" ]]; then
    executable_path="$KUBEWRAP_EXECUTABLE_PATH"
  else
    executable_path="$DEFAULT_EXECUTABLE_PATH"
  fi

  $executable_path "${opts[@]}"
  local exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    return $exit_code
  fi
  local source_content=$($executable_path source)
  if [[ $? -ne 0 ]]; then
    return 1
  fi
  source <(echo "$source_content")

  return
}
