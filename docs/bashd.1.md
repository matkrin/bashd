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

- **-S**, **--severity** _SEVERITY-LEVEL_
  Minimum severity used for diagnostics. _SEVERITY_LEVEL_ must be one of
  _style_, _info_, _warning_ or _error_. Default: _style_

- **--shellcheck-enable** _OPTIONAL-LINTS_
  Enable **shellcheck** optional lints. See avaible optional lints with
  **shellcheck --list-optional**.

- **--shellcheck-exclude** _RULE-CODES_
  Exclude **shellcheck** lints. _RULES-COES_ is a comma separated list of rules.

- **--shellcheck-include** _RULE-CODES_
  Only include **shellcheck** lints. _RULES-CODES_ is a comma separated list of
  rules. All other rules will be disabled.

- **--fmt-binary-next-line**
  On format, binary operators will appear on the next line when a binary
  command, such as a _|_, _&&_ or _||_, spans multiple lines. A `_\\_` will be
  used.

- **--fmt-case-indent**
  On format, _switch_ cases will be indented. As such, _switch_ case bodies will
  be two levels deeper than the _switch_ itself.

- **--fmt-func-next-line**
  On format, function opening braces are placed on a separate line.

- **--fmt-space-redirects**
  On format, redirect operators such as _>_ will be followed by a space.

- **-h**, **--help**
  Print a help message.

- **-V**, **--version**
  Print the version.

---

# CONFIGURATION
## severity
Minimum severity used for diagnostics. Must be one of
_style_, _info_, _warning_ or _error_. Default: _style_

## shellcheck
---
- **include**
  List of **shellcheck** rule codes. All other rules will be disabled.

- **exclude**
  List of **shellcheck** rules codes.

- **enable**
  List of **shellcheck** optional lints.
---

As the time of writing the following optional lints are available:

| Rule name                    | Description                                                 |
| ---------------------------- | ----------------------------------------------------------- |
| _add-default-case_           | Suggest adding a default case in `case` statements          |
| _avoid-negated-conditions_   | Suggest removing unnecessary comparison negations           |
| _avoid-nullary-conditions_   | Suggest explicitly using -n in `[ $var ]`                   |
| _check-extra-masked-returns_ | Check for additional cases where exit codes are masked      |
| _check-set-e-suppressed_     | Notify when set -e is suppressed during function invocation |
| _check-unassigned-uppercase_ | Warn when uppercase variables are unassigned                |
| _deprecate-which_            | Suggest 'command -v' instead of 'which'                     |
| _quote-safe-variables_       | Suggest quoting variables without metacharacters            |
| _require-double-brackets_    | Require [[ and warn about [ in Bash/Ksh                     |
| _require-variable-braces_    | Suggest putting braces around all variable references       |
| _useless-use-of-cat_         | Check for Useless Use Of Cat (UUOC)                         |

## format
---
- **binary_next_line**
  Binary operators will appear on the next line when a binary
  command, such as a _|_, _&&_ or _||_, spans multiple lines. A _`\\`_ will be
  used between lines.

- **case_indent**
  Switch cases will be indented. As such, _switch_ case bodies will
  be two levels deeper than the _switch_ itself.

- **space_redirects**
  Redirect operators such as _>_ will be followed by a space.

- **func_next_line**
  Function opening braces are placed on a separate line.
---

# EXAMPLES
## Invoke with highest log level and log to a file

```sh
bashd -vvv --logfile <LOG-FILE-NAME>
```

To do the same within Neovim, use
```lua
    vim.lsp.config.bashd = {
      cmd = { "bashd", "-vvv", "--logfile", "<LOG-FILE-NAME>" },
      -- ...
```

or in Helix via

```toml
    [language-server.bashd]
    command = "bashd"
    args = [ "-vvv", "--logfile", "<LOG_FILE-NAME>"]
    # ...
```


## Neovim configuration
```lua
vim.lsp.config.bashd = {
  cmd = { "bashd" },
  filetypes = { "bash", "sh" },
  root_markers = { ".git" },
  settings = {
    bashd = {
      severity = "style",                -- Minimum severity of errors to consider (error, warning, info, style)
      shellcheck = {
        include = {},                    -- Consider only given types of warnings
        exclude = {},                    -- Exclude types of warnings
        enable = {                       -- List of optional checks to enable (or 'all')
          "add-default-case",            -- Suggest adding a default case in `case` statements
          "avoid-negated-conditions",    -- Suggest removing unnecessary comparison negations
          "avoid-nullary-conditions",    -- Suggest explicitly using -n in `[ $var ]`
          "check-extra-masked-returns",  -- Check for additional cases where exit codes are masked
          "check-set-e-suppressed",      -- Notify when set -e is suppressed during function invocation
          "check-unassigned-uppercase",  -- Warn when uppercase variables are unassigned
          "deprecate-which",             -- Suggest 'command -v' instead of 'which'
          "quote-safe-variables",        -- Suggest quoting variables without metacharacters
          "require-double-brackets",     -- Require [[ and warn about [ in Bash/Ksh
          "require-variable-braces",     -- Suggest putting braces around all variable references
          "useless-use-of-cat",          -- Check for Useless Use Of Cat (UUOC)
        },
      },
      format = {
        binary_next_line = true,  -- Binary ops like && and | may start a line
        case_indent = false,      -- Switch cases will be indented
        space_redirects = true,   -- Redirect operators will be followed by a space
        func_next_line = false,   -- Function opening braces are placed on a separate line
      }
    }
  }
}

vim.lsp.enable("bashd")
```

## Helix configuration
```toml
[language-server.bashd]
command = "bashd"

[language-server.bashd.config.bashd]
severity = "style"                 # Minimum severity of errors to consider (error, warning, info, style)
shellcheck.include = []            # Consider only given types of warnings
shellcheck.exclude = []            # Exclude types of warnings
shellcheck.enable = [              # -- List of optional checks to enable (or 'all')
    "add-default-case",            # Suggest adding a default case in `case` statements
    "avoid-negated-conditions",    # Suggest removing unnecessary comparison negations
    "avoid-nullary-conditions",    # Suggest explicitly using -n in `[ $var ]`
    "check-extra-masked-returns",  # Check for additional cases where exit codes are masked
    "check-set-e-suppressed",      # Notify when set -e is suppressed during function invocation
    "check-unassigned-uppercase",  # Warn when uppercase variables are unassigned
    "deprecate-which",             # Suggest 'command -v' instead of 'which'
    "quote-safe-variables",        # Suggest quoting variables without metacharacters
    "require-double-brackets",     # Require [[ and warn about [ in Bash/Ksh
    "require-variable-braces",     # Suggest putting braces around all variable references
    "useless-use-of-cat",          # Check for Useless Use Of Cat (UUOC)
]
format.binary_next_line = true     # Binary ops like && and | may start a line
format.case_indent = false         # Switch cases will be indented
format.space_redirects = true      # Redirect operators will be followed by a space
format.func_next_line = false      # Function opening braces are placed on a separate line

[[language]]
name = "bash"
language-servers = [{ name = "bashd" }]
```

# SEE ALSO

sh(1), bash(1), shellcheck(1)
