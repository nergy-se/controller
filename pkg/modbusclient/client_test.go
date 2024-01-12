package modbusclient

import (
	"testing"
)

func TestDecode(t *testing.T) {

	var tests = []struct {
		name     string
		expected int
		given    []byte
	}{
		{
			name:     "8bit negative",
			expected: -28,
			given:    []byte{0xe4},
		},
		{
			name:     "16bit negative",
			expected: -28,
			given:    []byte{0xff, 0xe4},
		},
		{
			name:     "16bit postive",
			expected: 31,
			given:    []byte{0x00, 0x1f},
		},
		{
			name:     "large 32bit positive",
			expected: 514773,
			given:    []byte{0x00, 0x07, 0xda, 0xd5},
		},
		{
			name:     "32bit postive",
			expected: 31,
			given:    []byte{0x00, 0x00, 0x00, 0x1f},
		},
		{
			name:     "32bit negative",
			expected: -29,
			given:    []byte{0xff, 0xff, 0xff, 0xe3},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := Decode(tt.given)
			if actual != tt.expected {
				t.Errorf("given(%#v): expected %d, actual %d", tt.given, tt.expected, actual)
			}
		})
	}

}
