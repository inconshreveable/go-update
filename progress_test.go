package selfupdate

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ProgressReaderWithContentLength(t *testing.T) {
	data := []byte("this is some test data that should arrive on the other end of the pipe")

	reader := bytes.NewReader(data)

	var lastPercentage float64
	pr := progressReader{
		Reader:        reader,
		contentLength: int64(len(data)),
		progressCallback: func(f float64, err error) {
			lastPercentage = f
		},
	}

	r, err := ioutil.ReadAll(&pr)
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), lastPercentage)
	assert.Equal(t, data, r)
}
func Test_ProgressReaderWithoutContentLength(t *testing.T) {
	data := []byte("this is some test data that should arrive on the other end of the pipe")

	reader := bytes.NewReader(data)

	var lastPercentage float64
	pr := progressReader{
		Reader:        reader,
		contentLength: 0,
		progressCallback: func(f float64, err error) {
			lastPercentage = f
		},
	}

	r, err := ioutil.ReadAll(&pr)
	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), lastPercentage)
	assert.Equal(t, data, r)
}
