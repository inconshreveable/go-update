# self-update: Build self-updating Fyne programs

[![godoc reference](https://godoc.org/github.com/fynelabs/selfupdate?status.png)](https://godoc.org/github.com/fynelabs/selfupdate)
[![Coverage Status](https://coveralls.io/repos/github/fynelabs/selfupdate/badge.svg?branch=main)](https://coveralls.io/github/fynelabs/selfupdate?branch=main)

Package update provides functionality to implement secure, self-updating Fyne programs (or other single-file targets)

A program can update itself by replacing its executable file with a new version.

It provides the flexibility to implement different updating user experiences
like auto-updating, or manual user-initiated updates. It also boasts
advanced features like binary patching and code signing verification.

## Unmanaged update

Example of updating from a URL:

```go
import (
    "fmt"
    "net/http"

    "github.com/fynelabs/selfupdate"
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

## Managed update

To help make self updating Fyne application a new API and a tool, `selfupdatectl` have been introduced. The new API allow to provide a source where to get an update, configure a schedule for the update and a ed25519 public key to ensure that the update is only applied if they come from the proper source.

Example with the new API:

```go
import (
	"crypto/ed25519"
	"log"
	"time"

	"github.com/fynelabs/selfupdate"
)

func main() {
	done := make(chan struct{}, 2)

	// Used `selfupdatectl create-keys` followed by `selfupdatectl print-key`
	publicKey := ed25519.PublicKey{178, 103, 83, 57, 61, 138, 18, 249, 244, 80, 163, 162, 24, 251, 190, 241, 11, 168, 179, 41, 245, 27, 166, 70, 220, 254, 118, 169, 101, 26, 199, 129}

	// The public key above match the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "http://localhost/{{.Executable}}-{{.OS}}-{{.Arch}}{{.Ext}}")
	config := &selfupdate.Config{
		Source:    httpSource,
		Schedule:  selfupdate.Schedule{FetchOnStart: true, Interval: time.Minute * time.Duration(60)},
		PublicKey: publicKey,

		ProgressCallback: func(f float64, err error) { log.Println("Download", f, "%") },
		RestartConfirmCallback: func() bool { return true},
		UpgradeConfirmCallback: func(_ string) bool { return true },
		ExitCallback: func(_ error) { os.Exit(1) }
	}

	_, err := selfupdate.Manage(config)
	if err != nil {
		log.Println("Error while setting up update manager: ", err)
		return
	}

	<-done
}
```

If you desire a GUI element and visual integration with Fyne, you should check [fyneselfupdate](https://github.com/fynelabs/fyneselfupdate).

To help you manage your key, sign binary and upload them to an online S3 bucket the `selfupdatectl` tool is provided. You can check its documentation [here](https://github.com/fynelabs/selfupdate/tree/main/cmd/selfupdatectl).

## Logging

We provide three package wide variables: `LogError`, `LogInfo` and `LogDebug` that follow `log.Printf` API to provide an easy way to hook any logger in. To use it with go logger, you can just do

```go
selfupdate.LogError = log.Printf
```

If you are using logrus for example, you could do the following:

```go
selfupdate.LogError = logrus.Errorf
selfupdate.LogInfo = logrus.Infof
selfupdate.LogDebug = logrus.Debugf
```

Most logger module in the go ecosystem do provide an API that match the `log.Printf` and it should be straight forward to use in the same way as with logrus.

## Features

- Cross platform support
- Binary patch application
- Checksum verification
- Code signing verification
- Support for updating arbitrary files

## API Compatibility Promises
The main branch of `selfupdate` is *not* guaranteed to have a stable API over time. Still we will try hard to not break its API unecessarily and will follow a proper versioning of our release when necessary.

The `selfupdate` package makes the following promises about API compatibility:
1. A list of all API-breaking changes will be documented in this README.
1. `selfupdate` will strive for as few API-breaking changes as possible.

## API Breaking Changes
- **May 30, 2022**: Many changes moving to a new API that will be supported going forward.
- **June 22, 2022**: First tagged release, v0.1.0.

## License
Apache

## Sponsors

This project is kindly sponsored by the following companies:

<a href="https://dentagraphics.com/" style="text-decoration: none">
<img width="190" src="img/sponsor/dentagraphics.jpg" />
</a>