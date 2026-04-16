# Bubble Spice

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]
[![codecov][codecov-badge]][codecov]

## Development

**Requirements:** Go 1.24 or later

For detailed development setup, build commands, and AI agent guidance:

* [AGENTS.md](./AGENTS.md) - Development guidelines, build system, and testing
  patterns.

### Quick Start

```bash
make all    # Full build cycle (get deps, generate, tidy, build)
make test   # Run tests
make tidy   # Format and tidy (run before committing)
```

## See also

* [charmbracelet/bubbletea][bubbletea] — the Bubble Tea framework.
* [charmbracelet/bubbles][bubbles] — official widgets.
* [charmbracelet/lipgloss][lipgloss] — styling.
* [Awesome Apptly][awesome-apptly] — directory.

[godoc]: https://pkg.go.dev/github.com/apptly-oss/bubblespice
[godoc-badge]: https://pkg.go.dev/badge/github.com/apptly-oss/bubblespice.svg
[goreport]: https://goreportcard.com/report/github.com/apptly-oss/bubblespice
[goreport-badge]: https://goreportcard.com/badge/github.com/apptly-oss/bubblespice
[codecov]: https://codecov.io/gh/apptly-oss/bubblespice
[codecov-badge]: https://codecov.io/gh/apptly-oss/bubblespice/graph/badge.svg

[bubbletea]: https://pkg.go.dev/github.com/charmbracelet/bubbletea
[bubbles]: https://pkg.go.dev/github.com/charmbracelet/bubbles
[lipgloss]: https://pkg.go.dev/github.com/charmbracelet/lipgloss
[awesome-apptly]: https://awesome-apptly.com/
