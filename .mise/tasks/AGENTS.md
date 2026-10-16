## Shell Scripts

When writing shell scripts (including mise tasks which are shell scripts):

- Prefer zsh (assume latest version) over bash when you are certain it exists
- Prefer stricter error handling over silent failure.
- Hoist common environment variables to the top of the script.
- Describe each script once at the top, in 1-4 lines. For mise tasks, use `#MISE description="..."`; do not add a duplicate `Description:` comment.
- Document conditional exits with a one-line explanation if it's not obvious.
- Add line breaks between logical groups of code.

## Mise Tasks

- Prefer native Mise metadata over shell implementations of task configuration,
  argument parsing, input validation, and orchestration.
- Use `#MISE` for task properties such as environment requirements, defaults,
  dependencies, tools, working directories, and timeouts. Use `#USAGE` for
  arguments and flags, including required values, defaults, choices, and
  environment-backed inputs; read the resulting `usage_*` variables in the script.
- Use `env.NAME = { required = true }` when a variable must be set. Add custom
  Tera validation only when stricter checks are explicitly needed. Keep build
  and runtime assertions in the script.
- Share repeated task metadata through `task_templates` and `#MISE extends`.
  Dependency tasks do not inherit the parent task's `env` or `tools` settings.
- After changing task definitions, run `mise tasks validate` for the affected
  tasks and exercise their input validation with the installed Mise version.
