# exclude-merge

Tests that `.dockerignore` and `railpack.json` `exclude` are merged into one
list. `.dockerignore` patterns come first, so a later `exclude` entry can
negate them.

`.dockerignore` drops `secret.txt` and `kept.txt`. `railpack.json` puts
`kept.txt` back with `!kept.txt` and drops `also-secret.txt`. The image keeps
`kept.txt` and does not contain the other two.

Use one ignore list in a real project. This example exists so the merge is
covered when both are present.
