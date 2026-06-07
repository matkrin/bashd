package server

var SNIPPETS = []struct {
	label         string
	documentation string
	insertText    string
}{
	{"for", "for-in loop", "for ${1:ITEM} in ${2:ITEMS}; do\n\t${0::}\ndone"},
	{"for", "C-style foor loop", "for (( ${1:i} = 0; ${1:i} < ${2:n}; ${1:i}++ )); do\n\t${0::}\ndone"},
	{"while", "while loop", "while ${1:PREDICATE}; do\n\t${0::}\ndone"},
	{"if", "if statement", "if [[ ${1:PREDICATE} ]]; then\n\t${0::}\nfi"},
	{"case", "case statement", "case \"${1:VALUE}\" in\n${2:PATTERN})\n\t$0\n\t;;\n*)\n\t;;\nesac"},
}
