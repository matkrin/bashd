package server

var SNIPPETS = []struct {
	label         string
	documentation string
	insertText    string
}{
	{
		"for",
		"for-in loop",
		"for ${1:ITEM} in ${2:ITEMS}; do\n\t${0::}\ndone",
	},
	{
		"for",
		"C-style foor loop",
		"for (( ${1:i} = 0; ${1:i} < ${2:n}; ${1:i}++ )); do\n\t${0::}\ndone",
	},
	{
		"for in directory",
		"foor loop over files in a directory",
		"for ${1:FILE} in \"${2:DIR}\"/*; do\n\t${0::}\ndone",
	},
	{
		"while",
		"while loop",
		"while ${1:PREDICATE}; do\n\t${0::}\ndone",
	},
	{
		"while read file",
		"while loop over lines in a file",
		"while IFS= read -r ${1:LINE}; do\n\t${0::}\ndone < \"${2:FILE}\"",
	},
	{
		"while getopts",
		"while loop with getopts command",
		`while getopts "${1:ab:c:}" opt; do
	case "\$opt" in
		a) ;;
		b) ${2:VALUE}="\$OPTARG" ;;
		c) ${3:VALUE}="\$OPTARG" ;;
		*) exit 1 ;;
	esac
done`,
	},
	{
		"if",
		"if statement",
		"if [[ ${1:PREDICATE} ]]; then\n\t${0::}\nfi",
	},
	{
		"if else",
		"if else statement",
		"if [[ ${1:PREDICATE} ]]; then\n\t${2::}\nelse\n\t${0::}\nfi",
	},
	{
		"if file exists",
		"if statement that checks existance of a file",
		"if [[ -f \"${1:FILE}\" ]]; then\n\t${0::}\nfi",
	},
	{
		"if directory exists",
		"if statement that checks existance of a directory",
		"if [[ -d \"${1:DIRECTORY}\" ]]; then\n\t${0::}\nfi",
	},
	{
		"if string non-zero",
		"if statement that checks if a string has non-zero length",
		"if [[ -n \"${1:VAR}\" ]]; then\n\t${0::}\nfi",
	},
	{
		"if string zero",
		"if statement that checks if a string has zero length",
		"if [[ -z \"${1:VAR}\" ]]; then\n\t${0::}\nfi",
	},
	{
		"if command exists",
		"if statement that checks if executable exists in PATH",
		"if command -v ${1:EXECUTABLE} >/dev/null 2>&1; then\n\t${0::}\nfi",
	},
	{
		"case",
		"case statement",
		"case \"${1:VALUE}\" in\n${2:PATTERN})\n\t${0::}\n\t;;\n*)\n\t;;\nesac",
	},
	{
		"start",
		"sensible preamble",
		"#!/usr/bin/env bash\n\nset -euo pipefail\n\n${0}",
	},
	{
		"declare array",
		"declare and array with values",
		"declare -a ${1:ARR}=(\n\t\"${2:ITEM}\"\n)\n$0",
	},
	{
		"declare associative array",
		"declare an associative array with keys and values",
		"declare -A ${1:MAP}=(\n\t[\"${2:KEY}\"]=\"${3:VALUE}\"\n)\n$0",
	},
	{
		"read -r",
		"read with -r (no backslash escape)",
		"read -r ${1:VAR}",
	},
	{
		"read -rp",
		"read with -r (no backslash escape) and -p (prompt)",
		"read -rp \"${1:PROMPT: }\" ${2:VAR}",
	},
	{
		"read comma list",
		"read a comma separated input into array",
		"IFS=',' read -ra ${1:PARTS} <<< \"${2:INPUT}\"",
	},
	{
		"printf",
		"",
		"printf '%s\n' \"${1:VALUE}\"",
	},
	{
		"echo stderr",
		"echo to stderr",
		"echo >&2 \"${1:ERROR_MSG}\"",
	},
	{
		"heredoc",
		"",
		"cat <<EOF\n$0\nEOF",
	},
	{
		"heredoc to file",
		"",
		"cat > \"${1:FILE}\" <<EOF\n$0\nEOF",
	},
	{
		"trap EXIT",
		"",
		"trap '${1:CLEANUP}' EXIT",
	},
	{
		"trap INT TERM",
		"",
		"trap '${1:HANDLER}' INT TERM",
	},
	{
		"cleanup",
		"cleanup function",
		"cleanup() {\n\t${0::}\n}\n\ntrap cleanup EXIT",
	},
	{
		"tmpdir",
		"",
		"tmpdir=\"$(mktemp -d)\"\ntrap 'rm -rf \"$tmpdir\"' EXIT\n$0",
	},
	{
		"command-substitution",
		"",
		"\\$( ${0:COMMAND} )",
	},
	{
		"process-substitution",
		"",
		"<( ${0:COMMAND} )",
	},
	{
		"var default",
		"parameter expansion with default value",
		"\"\\${${1:VAR:-${2:DEFAULT}}}\"",
	},
	{
		"var-error",
		"parameter expansion with check for null or unset",
		"\"\\${${1:VAR:?${2:ERROR_MSG}}}\"",
	},
	{
		"find file",
		"find a file with given extension",
		"find \"${1:PATH}\" -type f -name \"${2:*.EXT}\"",
	},
	{
		"mapfile lines",
		"mapfile that creates a array over all lines in a file",
		"mapfile -t ${1:LINES} < \"${2:FILE}\"",
	},
	{
		"scriptdir",
		"gets the directory where scripti is located",
		"script_dir=\"\\$(cd -- \"\\$(dirname -- \"\\${BASH_SOURCE[0]}\")\" && pwd)\"",
	},
	{
		"date",
		"",
		"date +'%Y-%m-%d'",
	},
	{
		"date ISO 8601",
		"",
		"date +'%Y-%m-%dT%H:%M:%S%z'",
	},
	{
		"install",
		"",
		"install -Dm755 \"${1:SRC}\" \"${2:DST}\"",
	},
	{
		"usage",
		"usage function",
		"usage() {\n\tcat <<EOF\nUsage: ${1:script} [options]\n\nEOF\n }",
	},
	{
		"help",
		"help function",
		`print-help() {
	cat <<EOF
${1:PROGRAM_NAME} - ${2:SHORT_DESCRIPTION} (version ${3:VERSION})

Usage: ${1:PROGRAM_NAME} [OPTIONS] <VALUE>

Arguments:
  VALUE  An argument value

Options:
  -o, --optional <VALUE>  Provide an optional argument
  -f, --flag              Set a flag
  -h, --help              Print help
  -V, --version           Print version
EOF
}`,
	},
	{
		"python command",
		"python command with heredoc",
		"python3 - <<'PY'\n${0:PYTHON}\nPY",
	},
	{
		"log",
		"log function",
		`log() { printf '[%s] %s\n' "$(date +'%F %T')" "$*"; }`,
	},
	{
		"die",
		"die function",
		`die() { echo >&2 "$*"; exit 1; }`,
	},
	{
		"require",
		"require function",
		`require() { command -v "$1" >/dev/null 2>&1 || die "missing dependency: $1"; }`,
	},
}
