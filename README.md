# go-update: Build self-updating Go programs [![godoc reference](https://godoc.org/github.com/inconshreveable/go-update?status.png)](https://godoc.org/github.com/inconshreveable/go-update)

Package update provides functionality to implement secure, self-updating Go programs (or other single-file targets)
A program can update itself by replacing its executable file with a new version.

It provides the flexibility to implement different updating user experiences
like auto-updating, or manual user-initiated updates. It also boasts
advanced features like binary patching and code signing verification.

Example of updating from a URL:

```go
import (
    "fmt"
    "net/http"

    "github.com/inconshreveable/go-update"
)

func doUpdate(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    err := update.Apply(resp.Body, &update.Options{})
    if err != nil {
        // error handling
    }
    return err
}
```

## Features

- Cross platform support (Windows too!)
- Binary patch application
- Checksum verification
- Code signing verification
- Support for updating arbitrary files

## [equinox.io](https://equinox.io)
go-update provides the primitives for building self-updating applications, but there a number of other pieces
involved in a complete updating solution such as hosting, code signing, update channels, gradual rollout,
dynamically computing binary patches, tracking update metrics, etc.

[equinox.io](https://equinox.io) is a provider of that complete solution.

## License
Apache
