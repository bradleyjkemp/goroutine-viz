# goroutine-viz

A simple tool to convert the goroutine ancestry information from pprof into a more readable tree diagram.

This tool is experimental and likely to change.

## Usage
Run `goroutine-viz <your pprof url>` and pipe the output into Graphviz (SVG output is recommended).

## Requirements
* Use `GODEBUG="tracebackancestors=N"` when building/running your code to track goroutine ancestors up to N levels.
* Have pprof running.
* Have Graphviz installed to be able to render the output.

## TODOs (PRs welcome)
### Features
- [ ] Support reading from standard input to render saved profiles.
- [ ] Better handling of terminated goroutines.
- [ ] More readable output diagrams: colours? text formatting?

### Code quality
- [ ] Snapshot tests of Graphviz outputs.
- [ ] Linting/unit tests in CI.
