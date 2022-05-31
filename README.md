# self-update: Build self-updating Fyne programs [![godoc reference](https://godoc.org/github.com/fynelabs/self-update?status.png)](https://godoc.org/github.com/fynelabs/self-update)

Package update provides functionality to implement secure, self-updating Fyne programs (or other single-file targets)
A program can update itself by replacing its executable file with a new version.

It provides the flexibility to implement different updating user experiences
like auto-updating, or manual user-initiated updates. It also boasts
advanced features like binary patching and code signing verification.

Example of updating from a URL:

```go
import (
    "fmt"
    "net/http"

    "github.com/fynelabs/self-update"
)

func doUpdate(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    err = update.Apply(resp.Body, update.Options{})
    if err != nil {
        // error handling
    }
    return err
}
```

## Features

- Cross platform support
- Binary patch application
- Checksum verification
- Code signing verification
- Support for updating arbitrary files

## API Compatibility Promises
The master branch of `self-update` is *not* guaranteed to have a stable API over time. For any production application, you should vendor
your dependency on `self-update` with `go vendor`.

The `self-update` package makes the following promises about API compatibility:
1. A list of all API-breaking changes will be documented in this README.
1. `self-update` will strive for as few API-breaking changes as possible.

## API Breaking Changes
- **May 30, 2022**: Many changes moving to a new API that will be supported going forward.

## License
Apache
