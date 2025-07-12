# bashd

## Features

Diagnostics:

- Check if sourced file exists
- [Parser](https://github.com/mvdan/sh/) errors

Definition:

- Variable assignment in document and sourced files
- Function declaration in document and sourced files
- Sourced file itself

References:

- Function calls in document
- Variable usage in document
- Depending on `ReferenceContext.includeDeclaration` function declarations and
  variable assignments

Completion:

- Variables declared in document (on `$`)
- Functions declared in document
- Environment variables (on `$`)
- Keywords

Document symbols:

- Variable assignment in document
- Function declaration in document

## TODO

- Language features:
  - Better completion: resolve with docs from man pages
  - Go to declaration: `declare` statements
  - Rename
  - Hover: man pages, flags parsing from man pages
  - Formatting, range formatting: sh package?
  - Maybe:
    - Inlay hint
    - Code lens
    - Signature help

  - Find references: Variables, functions in sourced files (does this make
    sense? maybe for other scripts in workspace, would need to check if file is
    sourced...)

- Workspace features:
  - Workspace symbols
  -
