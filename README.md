# bashd

## Features

### Diagnostics

- Check if sourced file exists
- [Parser](https://github.com/mvdan/sh/) errors
- [shellcheck](https://github.com/koalaman/shellcheck) integration
- For document on document change
- For workspace on initialize

### Hover

- Show `help` output as docs for keywords and builtins
- Show man page as docs for executables
- Show location of assignment for variables
- Show location and body and functions

### Definition

- Variable assignment in document and sourced files
- Function declaration in document and sourced files
- Sourced file itself

### References

- Function calls in current document and sourced files
- Variable usage in current document and sourced files
- Function calls in workspace file which source the current file
- Variable usage in workspace file which source the current file
- Depending on `ReferenceContext.includeDeclaration` function declarations and
  variable assignments

### Completion

- Variables declared in document (on `$` and `{`)
- Functions declared in document
- Environment variables (on `$` and `{`)
- Keywords
- Executables in PATH
- Resolve with `help` output as docs for keywords and builtins
- Resolve with man page as docs for executables

### Rename

- Function declarations and calls in document and sourced files
- Variable assignments and usage in document and sourced files
- Function declarations and calls in workspace file which source the current
  file
- Variable assignments and usage in workspace file which source the current file

### Document Symbols

- Variable assignment in document
- Function declaration in document

### Workspace Symbols

- Variable assignment in workspace .sh files and scripts without extension but
  shebang
- Function declaration in workspace .sh files and scripts without extension but
  shebang

### Formatting

- Entire file
- Range formatting (as long as range covers nodes that can be formatted)

### Code Actions

- Fix for shellcheck lints (position depenent)
- Fix all auto-fixable lints (only when there are fixable lints)
- Add shebang if not exist
- Put document on single line

### Inlay Hint

- [SGR][sgr] ANSI escapes

### Document Colors

- [256-color (8-bit)][8b-color] foreground (`\x1b[38;5;<n>m`) and background
  (`\x1b[48;5;<n>m`)
- [True color (24-bit)][24b-color] foreground(`\x1b[38;2;<r>;<g>;<b>m`) and
  background (`\x1b[48;2;<r>;<g>;<b>m`)
- [3-bit / 4-bit color][3-4b-color] foreground (`\x1b[<n>m`; n ∈ [30,37] ∪
  [90,97]) and background (`\x1b[<n>m`; n ∈ [40,47] ∪ [100,107])
- Also alternative escapes `\e` and `\033`

## TODO

- Refactoring!!!
- Go to declaration: `declare` statements
- Maybe:
  - Hover: flags parsing from man pages?
  - Code lens
  - Signature help

[sgr]: https://en.wikipedia.org/wiki/ANSI_escape_code#Select_Graphic_Rendition_parameters
[8b-color]: https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
[24b-color]: https://en.wikipedia.org/wiki/ANSI_escape_code#24-bit
[3-4b-color]: https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
