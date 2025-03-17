package thermiagenesis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeHeatCurve(t *testing.T) {
	data := []byte{0x07, 0x6c, 0x0a, 0x28, 0x0c, 0x1c, 0x0d, 0xac, 0x0e, 0xd8, 0x11, 0x94, 0x14, 0x50}
	assert.Equal(t, []float64{19, 26, 31, 35, 38, 45, 52}, decodeHeatCurve(data, 0))
}
