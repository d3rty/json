package dirtytests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReadSampleFile(t *testing.T, fileName string) []byte {
	contents, err := os.ReadFile(fmt.Sprintf("./testdata/%s.json", fileName))
	require.NoError(t, err)

	return contents
}
