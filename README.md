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
- **TODO**: function preview, variable preview?, flags parsing from man pages?

### Definition

- Variable assignment in document and sourced files
- Function declaration in document and sourced files
- Sourced file itself
- **TODO**: Workspace

### References

- Function calls in document
- Variable usage in document
- Depending on `ReferenceContext.includeDeclaration` function declarations and
  variable assignments
- **TODO**: Workspace

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
- **TODO**: Workspace

### Document symbols

- Variable assignment in document
- Function declaration in document

### Workspace symbols

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

### Document colors

- [256-color (8-bit)][8b-color] foreground (`\x1b[38;5;<n>m`) and background
  (`\x1b[48;5;<n>m`)
- [True color (24-bit)][24b-color] foreground(`\x1b[38;2;<r>;<g>;<b>m`) and
  background (`\x1b[48;2;<r>;<g>;<b>m`)
- [3-bit / 4-bit color][3-4b-color] foreground (`\x1b[<n>`; n ∈ [30,37] ∪
  [90,97]) and background (`\x1b[<n>`; n ∈ [40,47] ∪ [100,107])
- Also alternative escapes `\e` and `\033`

## TODO

- Refactoring!!!
- Inlay hint for ansi escapes
- Go to declaration: `declare` statements
- Maybe:
  - Code lens
  - Signature help
  - Find references: Variables, functions in sourced files (does this make
    sense? maybe for other scripts in workspace, would need to check if file is
    sourced...)

[8b-color]: https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
[24b-color]: (https://en.wikipedia.org/wiki/ANSI_escape_code#24-bit)
[3-4b-color]: (https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit)
