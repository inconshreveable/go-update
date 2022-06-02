package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

type Keys struct {
	PrivateKey string
	PublicKey  string
}

func CreateKeys() *cli.Command {
	k := &Keys{}

	return &cli.Command{
		Name:        "create-keys",
		Usage:       "Create public and private keys to be use to certify.",
		Description: "You may specify a filename for the Private and the Public Keys",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "private-key",
				Aliases:     []string{"priv"},
				Usage:       "The private key file to store the new key in.",
				Destination: &k.PrivateKey,
				Value:       "ed25519.key",
			},
			&cli.StringFlag{
				Name:        "public-key",
				Aliases:     []string{"pub"},
				Usage:       "The public key file to store the new key in.",
				Destination: &k.PublicKey,
				Value:       "ed25519.pem",
			},
		},
		Action: func(_ *cli.Context) error {
			return k.CreateKeys()
		},
	}
}

func (k *Keys) CreateKeys() error {
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

	err = ioutil.WriteFile(k.PrivateKey, pem.EncodeToMemory(block), 0600)
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

	err = ioutil.WriteFile(k.PublicKey, pem.EncodeToMemory(block), 0644)
	return err
}
