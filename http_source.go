package selfupdate

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type HTTPSource struct {
	client  *http.Client
	baseURL string
}

var _ Source = (*HTTPSource)(nil)

func NewHTTPSource(client *http.Client, base string) Source {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPSource{client: client, baseURL: base}
}

func (h *HTTPSource) Get(v *Version) (io.ReadCloser, int64, error) {
	request, err := http.NewRequest("GET", h.baseURL, nil)
	if err != nil {
		return nil, 0, err
	}

	if v != nil {
		if !v.Date.IsZero() {
			request.Header.Add("If-Modified-Since", v.Date.Format(http.TimeFormat))
		}
	}

	response, err := h.client.Do(request)
	if err != nil {
		return nil, 0, err
	}

	return response.Body, response.ContentLength, nil
}

func (h *HTTPSource) GetSignature() ([64]byte, error) {
	resp, err := h.client.Get(h.baseURL + ".ed25519")
	if err != nil {
		return [64]byte{}, err
	}
	defer resp.Body.Close()

	if resp.ContentLength != 64 {
		return [64]byte{}, fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", resp.ContentLength)
	}

	writer := bytes.NewBuffer(make([]byte, 0, 64))
	n, err := io.Copy(writer, resp.Body)
	if err != nil {
		return [64]byte{}, err
	}

	if n != 64 {
		return [64]byte{}, fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", n)
	}

	r := [64]byte{}
	copy(r[:], writer.Bytes())

	return r, nil
}

func (h *HTTPSource) LatestVersion() (*Version, error) {
	resp, err := h.client.Head(h.baseURL)
	if err != nil {
		return nil, err
	}

	lastModified := resp.Header.Get("Last-Modified")
	if lastModified == "" {
		return nil, fmt.Errorf("no Last-Modified served")
	}

	t, err := http.ParseTime(lastModified)
	if err != nil {
		return nil, err
	}

	return &Version{Date: t}, nil
}
