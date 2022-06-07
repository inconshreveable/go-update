package selfupdate

import (
	"crypto/ed25519"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPSourceLatestVersion(t *testing.T) {
	client := http.Client{Timeout: time.Duration(60) * time.Second}
	httpSource := NewHTTPSource(&client, "http://geoffrey-test-artefacts.fynelabs.com/nomad-windows-amd64.exe")

	version, err := httpSource.LatestVersion()
	assert.Nil(t, err)
	assert.NotNil(t, version)
}

func TestHTTPSourceCheckSignature(t *testing.T) {
	client := http.Client{Timeout: time.Duration(60) * time.Second}

	publicKey := ed25519.PublicKey{178, 103, 83, 57, 61, 138, 18, 249, 244, 80, 163, 162, 24, 251, 190, 241, 11, 168, 179, 41, 245, 27, 166, 70, 220, 254, 118, 169, 101, 26, 199, 129}
	wrongPublicKey := ed25519.PublicKey{42, 103, 83, 57, 61, 138, 18, 249, 244, 80, 163, 162, 24, 251, 190, 241, 11, 168, 179, 41, 245, 27, 166, 70, 220, 254, 118, 169, 101, 26, 199, 129}

	httpSource := NewHTTPSource(&client, "http://geoffrey-test-artefacts.fynelabs.com/nomad-windows-amd64.exe")
	signature, err := httpSource.GetSignature()
	assert.Nil(t, err)

	file, contentLength, err := httpSource.Get(&Version{Date: time.Unix(100, 0)})
	log.Println(file, " -- ", err)
	assert.Nil(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, int64(32099400), contentLength)

	body, err := ioutil.ReadAll(file)
	assert.Nil(t, err)
	assert.NotNil(t, body)
	file.Close()

	ok := ed25519.Verify(publicKey, body, signature[:])
	assert.True(t, ok)

	ok = ed25519.Verify(wrongPublicKey, body, signature[:])
	assert.False(t, ok)
}

func TestReplaceUrlTemplate(t *testing.T) {
	nochange := "http://localhost/nomad-windows-amd64.exe"
	change := "http://localhost/nomad-{{.OS}}-{{.Arch}}{{.Ext}}"
	expected := ""
	if runtime.GOOS == "windows" {
		expected = "http://localhost/nomad-" + runtime.GOOS + "-" + runtime.GOARCH + ".exe"
	} else {
		expected = "http://localhost/nomad-" + runtime.GOOS + "-" + runtime.GOARCH
	}

	r := replaceUrlTemplate(nochange)
	assert.Equal(t, nochange, r)

	r = replaceUrlTemplate(change)
	assert.NotEqual(t, change, r)
	assert.Equal(t, expected, r)
}
