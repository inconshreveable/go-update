package selfupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlwaysFindExecutableTime(t *testing.T) {
	_, err := lastModifiedExecutable()
	assert.Nil(t, err)
}

func TestAlwaysFindExecutable(t *testing.T) {
	exe, err := getExecutableRealPath()
	assert.Nil(t, err)
	assert.NotEmpty(t, exe)
}
