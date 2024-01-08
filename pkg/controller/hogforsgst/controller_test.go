package hogforsgst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvgCOP(t *testing.T) {
	cont := New(nil, nil)
	assert.Equal(t, 3.5, cont.avgCOP())

	cont.addCOP(3.0)
	cont.addCOP(4.0)
	assert.Equal(t, 3.5, cont.avgCOP())

	cont.addCOP(10.0)
	assert.Equal(t, 5.125, cont.avgCOP())
}
