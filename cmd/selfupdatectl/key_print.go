package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func keyPrint() *cli.Command {
	a := &application{}

	return &cli.Command{
		Name:        "print-key",
		Usage:       "Display public key in Go format that can be used directly in your project",
		Description: "You may specify a public ed25519 PEM public key to display.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "public-key",
				Aliases:     []string{"pub"},
				Usage:       "The public key file to use to verify the signature for this executable.",
				Destination: &a.publicKey,
				Value:       "ed25519.pem",
			},
		},
		Action: func(ctx *cli.Context) error {
			return a.keyPrint()
		},
	}
}

func (a *application) keyPrint() error {
	verifier, err := publicKeyVerifier(a.publicKey)
	if err != nil {
		return err
	}

	output := "publicKey := ed25519.PublicKey{"

	for i, b := range verifier {
		if i == 0 {
			output += fmt.Sprintf("%v", b)
		} else {
			output += fmt.Sprintf(", %v", b)
		}
	}
	output += "}"

	fmt.Println(output)
	return nil
}
