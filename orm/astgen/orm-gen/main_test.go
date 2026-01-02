package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_gen(t *testing.T) {
	buffer := &bytes.Buffer{}
	err := gen(buffer, "testdata/user.go")
	require.NoError(t, err)
	assert.Equal(t, `package testdata`, buffer.String())
}
