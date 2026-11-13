# Coding standards

These rules apply to the whole repo.

## Specialist functions

Each function does one thing, and its name says what. A function that collects, formats and writes is three functions.

## One package per responsibility

Each package owns one job: running external commands (`runner`), listing applications (`apps`), listing Homebrew packages (`brew`), grouping them (`inventory`), formatting the file (`render`), writing it (`store`), reading the config (`config`). A package that needs the outside world (a command, the clock, the home folder) receives it as a parameter or an interface, so its tests can pass a fake. The same goes for what `run` in `cmd/main.go` calls: process exit, the clock and the home folder arrive as parameters, so every path of `run` can be tested.

## Layout

```
cmd/main.go            ← only reads options and wires the packages; no domain logic
internal/<package>/    ← one folder per package, tests next to the code
```
