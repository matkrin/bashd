# Upcoming release


# v0.4.0 (2026-10-30)
## New
- Workspace diagnostics perf (#12)
  - Workspace diagnostics now run entirely in a separate go-routine
  - For workspace sh-files shellcheck now runs concurrently with 8 workers
  - Now uses 8 workers and only reads 256 bytes of files that could be
    shell script files
  - Handling of `shutdown` and `exit` are handled before the client message queue
  - The final `os.Exit` was lifted to `main`

## Fixes
- Make sure null is returned by response to `shutdown`

# v0.3.1 (2026-10-23)
## Fixes
- textDocument/documentColor: Return empty slice instead of nil


# v0.3.0 (2026-10-18)
## New
- Configurable workspace diagnostics by @matkrin in #10
  - This adds the functionality to suppress workspace diagnostics when the
    configuration updates (workspace/didChangeConfiguration) by the boolean
    settings workspace_diagnostics_on_update with the default value of true.

  - Further, there is a boolean flag --no-workspace-diagnostics-on-update,
    serving the same purpose.

   - The workspace diagnostics checking was removed from initialize as it was
     redundant, as on startup, the initial configuration is also transferred via
     workspace/didChangeConfiguration, i.e. diagnostics where checked twice in the
     past.

## Fixes
- Omit empty documentSymbol.children
- Double checking of workspace diagnostics on startup


# v0.2.3 (2026-07-28)
## Fixes
- Dereference of nil pointer of clientInfo in initialize


# v0.2.2 (2026-06-23)
## Fixes
- Omit empty optional fields of completionItem in textdocument/completion


# v0.2.1 (2026-06-23)
## Fixes
- Diagnostics in workspace/didChangeConfiguration (fixes #4) by @matkrin in #5


# v0.2.0 (2026-06-08)
## New
- Add snippets support by @matkrin in #1


# v0.1.1 (2026-02-28)
## Fixes
- Find references and rename for arithmetic expressions and C-style for-loops


# v0.1.0 (2025-11-17)
## New
- VSCode support (198059e)

## Fixes
- Removal of empty code actions (0a3fbf8)
- Not allow - prefix for definitions in read statements (f3186c4)
- Empty array for document symbol children instead of null (700d626)
