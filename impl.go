package update

import (
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/asn1"
	"errors"
	"io"
	"math/big"

	"github.com/kr/binarydist"
)

type patchFn func(io.Reader, io.Writer, io.Reader) error

func (fn patchFn) Patch(old io.Reader, new io.Writer, patch io.Reader) error {
	return fn(old, new, patch)
}

type verifyFn func([]byte, []byte, crypto.Hash, crypto.PublicKey) error

func (fn verifyFn) VerifySignature(checksum []byte, signature []byte, hash crypto.Hash, publicKey crypto.PublicKey) error {
	return fn(checksum, signature, hash, publicKey)
}

var BSDiffPatcher = patchFn(binarydist.Patch)
var RSAVerifier = verifyFn(func(checksum, signature []byte, hash crypto.Hash, publicKey crypto.PublicKey) error {
	key, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("not a valid RSA public key")
	}
	return rsa.VerifyPKCS1v15(key, hash, checksum, signature)
})

type rsDER struct {
	R *big.Int
	S *big.Int
}

var ECDSAVerifier = verifyFn(func(checksum, signature []byte, hash crypto.Hash, publicKey crypto.PublicKey) error {
	key, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("not a valid ECDSA public key")
	}
	var rs rsDER
	if _, err := asn1.Unmarshal(signature, &rs); err != nil {
		return err
	}
	if !ecdsa.Verify(key, checksum, rs.R, rs.S) {
		return errors.New("failed to verify ecsda signature")
	}
	return nil
})

var DSAVerifier = verifyFn(func(checksum, signature []byte, hash crypto.Hash, publicKey crypto.PublicKey) error {
	key, ok := publicKey.(*dsa.PublicKey)
	if !ok {
		return errors.New("not a valid DSA public key")
	}
	var rs rsDER
	if _, err := asn1.Unmarshal(signature, &rs); err != nil {
		return err
	}
	if !dsa.Verify(key, checksum, rs.R, rs.S) {
		return errors.New("failed to verify ecsda signature")
	}
	return nil
})
