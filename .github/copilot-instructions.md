# GitHub Copilot Instructions for assistant-go

- Follow standard Go conventions and idioms.
- Do not prioritize any files, directories, or technologies unless explicitly requested.
- Generate code, tests, documentation, or other content as requested by the user.
- No additional instructions or restrictions apply.

---

## Standard Go CLI App Project Template

When generating a new Go CLI app, use the following conventions:

- Project structure:

  - `cmd/<appname>/main.go` (entry point)
  - `internal/` (private packages)
  - `pkg/` (public packages, if needed)
  - `go.mod` and `go.sum` (module files)
  - `README.md` (project documentation)
  - `.github/` (workflows, instructions)

- Main file (`main.go`) example:

```go
package main

import (
		"flag"
		"fmt"
		"os"
)

func main() {
		var name string
		flag.StringVar(&name, "name", "World", "Name to greet")
		flag.Parse()
		fmt.Printf("Hello, %s!\n", name)
}
```

- Use the `flag` package for CLI arguments.
- Provide clear usage/help output.
- Add tests in `*_test.go` files.
- Document commands and flags in `README.md`.
- Use Go modules.
- Prefer idiomatic error handling.

---
