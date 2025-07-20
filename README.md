# bashd

## Features

### Diagnostics:

- Check if sourced file exists
- [Parser](https://github.com/mvdan/sh/) errors

### Definition:

- Variable assignment in document and sourced files
- Function declaration in document and sourced files
- Sourced file itself
- TODO: Workspace

### References:

- Function calls in document
- Variable usage in document
- Depending on `ReferenceContext.includeDeclaration` function declarations and
  variable assignments
- TODO: Workspace

### Completion:

- Variables declared in document (on `$`)
- Functions declared in document
- Environment variables (on `$`)
- Keywords

### Rename:

- Function declarations and calls in document and sourced files
- Variable assignments and usage in document and sourced files
- TODO: Workspace

### Document symbols:

- Variable assignment in document
- Function declaration in document

### Workspace symbols:

- Variable assignment in workspace .sh files
- Function declaration in workspace .sh files
- TODO: Files without extension and check shebang??

## TODO

- Language features:
  - Better completion: resolve with docs from man pages
  - Go to declaration: `declare` statements
  - Hover: man pages, flags parsing from man pages
  - Formatting, range formatting: sh package?
  - Maybe:
    - Inlay hint
    - Code lens
    - Signature help

  - Find references: Variables, functions in sourced files (does this make
    sense? maybe for other scripts in workspace, would need to check if file is
    sourced...)
