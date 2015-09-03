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
    err := update.Apply(resp.Body, update.Options{})
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
[equinox.io](https://equinox.io) is a complete ready-to-go updating solution built on top of go-update that provides:

- Hosted updates
- Update channels (stable, beta, nightly, ...)
- Dynamically computed binary diffs
- Automatic key generation and code
- Release tooling with proper code signing
- Update/download metrics

## Breaking API Changes
- Sept 3, 2015: The `Options` struct passed to `Apply` was changed to be passed by value instead of passed by pointer.

## Older API Versions
Did your build just break because the go-update API changed? You have two options:

1. Update your import to `gopkg.in/inconshreveable/go-update.v0`
1. Vendor your dependency on it with a tool like [gb](http://getgb.io/) or [govendor](https://github.com/kardianos/govendor)

## License
Apache
