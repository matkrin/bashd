---
name: bashd
section: 1
date: 2025-09-28
left-footer: bashd Manual
center-footer: User Commands
---

# NAME

**bashd** - Bash language server

# SYNOPSIS

bashd [_OPTIONS_]

# DESCRIPTION

bashd is a Language Server Protocol (LSP) implementation for Bash, built using
the Go sh package and featuring ShellCheck integration for real-time linting.

# OPTIONS

---

- **-j**, **--json**
  Log in JSON format.

- **-l**, **--logfile** _FILE_
  Log to _FILE_ instead of stderr.

- **-v**, **--verbose**
  Increase log message verbosity with repeated usage up to **-vvv**.

- **-h**, **--help**
  Print a help message.

- **-V**, **--version**
  Print the version.

---

# FEATURES

## Diagnostics

- Check if sourced file exists
- Parser errors
- ShellCheck lints
- For document on document change
- For workspace on initialize

## Hover

- Show **help** output as docs for keywords and builtins
- Show man page as docs for executables
- Show location of assignment for variables
- Show location and body for functions

## Definition

- Variable assignment in document and sourced files
- Function declaration in document and sourced files
- Sourced file itself

## References

- Function calls in current document and sourced files
- Variable usage in current document and sourced files
- Function calls in workspace file which source the current file
- Variable usage in workspace file which source the current file
- Depending on `ReferenceContext.includeDeclaration` function declarations and
  variable assignments

## Rename

- Function declarations and calls in document and sourced files
- Variable assignments and usage in document and sourced files
- Function declarations and calls in workspace file which source the current
  file
- Variable assignments and usage in workspace file which source the current file

## Completion

- Variables declared in document (on _`$`_ and _`{`_)
- Functions declared in document
- Environment variables (on _`$`_ and _`{`_)
- Keywords
- Executables in _PATH_
- Resolve with **help** output as docs for keywords and builtins
- Resolve with man page as docs for executables

## Document Symbols

- Variable assignment in document
- Function declaration in document

## Workspace Symbols

- Variable assignment in workspace .sh files and scripts without extension but
  shebang
- Function declaration in workspace .sh files and scripts without extension but
  shebang

## Formatting

- Entire file
- Range formatting (as long as range covers nodes that can be formatted)

## Code Actions

- Fix for shellcheck lints (position dependent)
- Add ignore comment for shellcheck lints (position dependent)
- Fix all auto-fixable lints (only when there are fixable lints)
- Add shebang if not exist
- Minify script

### Inlay Hint

- SGR ANSI escapes

### Document Colors

- 256-color (8-bit) foreground (`\x1b[38;5;<n>m`) and background
  (`\x1b[48;5;<n>m`)
- True color (24-bit) foreground(`\x1b[38;2;<r>;<g>;<b>m`) and background
  (`\x1b[48;2;<r>;<g>;<b>m`)
- 3-bit / 4-bit color foreground (`\x1b[<n>m`; n ∈ [30,37] ∪ [90,97]) and
  background (`\x1b[<n>m`; n ∈ [40,47] ∪ [100,107])
- Also alternative escapes `\e` and `\033`

# SEE ALSO

sh(1), bash(1), shellcheck(1)
