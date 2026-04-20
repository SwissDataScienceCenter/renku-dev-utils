package clipboard

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClipboardInit(t *testing.T) {
	supported, err := Init()
	assert.NoError(t, err)
	fmt.Fprintf(os.Stderr, "NOTE: clipboard supported: %t\n", supported)
}
