# bashd

## Features

### Diagnostics:

- Check if sourced file exists
- [Parser](https://github.com/mvdan/sh/) errors
- [shellcheck](https://github.com/koalaman/shellcheck) integration
- For document on document change
- For workspace on initialize

### Hover

- Show `help` output as docs for keywords and builtins
- Show man page as docs for executables
- **TODO**: function preview, variable preview?, flags parsing from man pages?

### Definition:

- Variable assignment in document and sourced files
- Function declaration in document and sourced files
- Sourced file itself
- **TODO**: Workspace

### References:

- Function calls in document
- Variable usage in document
- Depending on `ReferenceContext.includeDeclaration` function declarations and
  variable assignments
- **TODO**: Workspace

### Completion:

- Variables declared in document (on `$` and `{`)
- Functions declared in document
- Environment variables (on `$` and `{`)
- Keywords
- Executables in PATH
- Resolve with `help` output as docs for keywords and builtins
- Resolve with man page as docs for executables

### Rename:

- Function declarations and calls in document and sourced files
- Variable assignments and usage in document and sourced files
- **TODO**: Workspace

### Document symbols:

- Variable assignment in document
- Function declaration in document

### Workspace symbols:

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

## TODO

- Go to declaration: `declare` statements
- documentsColors for ansi escapes
- Inlay hint for ansi escapes
- Maybe:
  - Code lens
  - Signature help
  - Find references: Variables, functions in sourced files (does this make
    sense? maybe for other scripts in workspace, would need to check if file is
    sourced...)
