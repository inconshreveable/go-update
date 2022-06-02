package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"
)

type Application struct {
	Executable string
	PrivateKey string
	PublicKey  string
}

func Sign() *cli.Command {
	a := &Application{}

	return &cli.Command{
		Name:        "sign",
		Usage:       "Generate a signature for a Fyne binary and store it in FyneApp.toml",
		Description: "You may specify a filename for the Private Key, the executable and the FyneApp.toml you want to use",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "private-key",
				Aliases:     []string{"priv"},
				Usage:       "The private key file to use to sign this executable.",
				Destination: &a.PrivateKey,
				Value:       "ed25519.key",
			},
			&cli.StringFlag{
				Name:        "executable",
				Aliases:     []string{"exe"},
				Usage:       "The executable to generate a signature for.",
				Destination: &a.Executable,
			},
		},
		Action: func(ctx *cli.Context) error {
			if a.Executable != "" {
				err := a.Sign(a.Executable)
				if err != nil {
					return err
				}
			}

			for _, exe := range ctx.Args().Slice() {
				err := a.Sign(exe)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func (a *Application) Sign(executable string) error {
	signer, err := privateKeySigner(a.PrivateKey)
	if err != nil {
		return err
	}

	content, err := executableContent(executable)
	if err != nil {
		return err
	}

	signature := ed25519.Sign(signer, content)

	if len(signature) != 64 {
		return fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", len(signature))
	}

	err = ioutil.WriteFile(executable+".ed25519", signature, 0644)
	return err
}

func privateKeySigner(privateKey string) (ed25519.PrivateKey, error) {
	privateKeyFile, err := os.Open(privateKey)
	if err != nil {
		return []byte{}, err
	}
	defer privateKeyFile.Close()

	b, err := ioutil.ReadAll(privateKeyFile)
	if err != nil {
		return []byte{}, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return []byte{}, fmt.Errorf("unable to decode Private Key PEM")
	}

	signer, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return []byte{}, nil
	}

	ed25519signer, ok := signer.(ed25519.PrivateKey)
	if !ok {
		return []byte{}, fmt.Errorf("private Key is not an ED25519")
	}

	return ed25519signer, nil
}

func executableContent(executable string) ([]byte, error) {
	executableFile, err := os.Open(executable)
	if err != nil {
		return []byte{}, err
	}
	defer executableFile.Close()

	return ioutil.ReadAll(executableFile)
}
