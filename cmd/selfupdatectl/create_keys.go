package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

func createKeys() *cli.Command {
	a := &application{}

	return &cli.Command{
		Name:        "create-keys",
		Usage:       "Create public and private keys to be use to certify.",
		Description: "You may specify a filename for the Private and the Public Keys",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "private-key",
				Aliases:     []string{"priv"},
				Usage:       "The private key file to store the new key in.",
				Destination: &a.privateKey,
				Value:       "ed25519.key",
			},
			&cli.StringFlag{
				Name:        "public-key",
				Aliases:     []string{"pub"},
				Usage:       "The public key file to store the new key in.",
				Destination: &a.publicKey,
				Value:       "ed25519.pem",
			},
		},
		Action: func(_ *cli.Context) error {
			return a.createKeys()
		},
	}
}

func (a *application) createKeys() error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}

	err = ioutil.WriteFile(a.privateKey, pem.EncodeToMemory(block), 0600)
	if err != nil {
		return err
	}

	b, err = x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return err
	}

	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}

	err = ioutil.WriteFile(a.publicKey, pem.EncodeToMemory(block), 0644)
	return err
}
