package binary

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero/wasm"
)

func TestEncodeImport(t *testing.T) {
	tests := []struct {
		name     string
		input    *wasm.Import
		expected []byte
	}{
		{
			name: "func no module, no name, type index 0",
			input: &wasm.Import{ // Ex. (import "" "" (func (type 0)))
				Kind:     wasm.ImportKindFunc,
				Module:   "",
				Name:     "",
				DescFunc: 0,
			},
			expected: []byte{wasm.ImportKindFunc, 0x00, 0x00, 0x00},
		},
		{
			name: "func module, no name, type index 0",
			input: &wasm.Import{ // Ex. (import "$test" "" (func (type 0)))
				Kind:     wasm.ImportKindFunc,
				Module:   "test",
				Name:     "",
				DescFunc: 0,
			},
			expected: []byte{
				0x04, 't', 'e', 's', 't',
				0x00,
				wasm.ImportKindFunc,
				0x00,
			},
		},
		{
			name: "func module, name, type index 0",
			input: &wasm.Import{ // Ex. (import "$math" "$pi" (func (type 0)))
				Kind:     wasm.ImportKindFunc,
				Module:   "math",
				Name:     "pi",
				DescFunc: 0,
			},
			expected: []byte{
				0x04, 'm', 'a', 't', 'h',
				0x02, 'p', 'i',
				wasm.ImportKindFunc,
				0x00,
			},
		},
		{
			name: "func module, name, type index 10",
			input: &wasm.Import{ // Ex. (import "$math" "$pi" (func (type 10)))
				Kind:     wasm.ImportKindFunc,
				Module:   "math",
				Name:     "pi",
				DescFunc: 10,
			},
			expected: []byte{
				0x04, 'm', 'a', 't', 'h',
				0x02, 'p', 'i',
				wasm.ImportKindFunc,
				0x0a,
			},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			bytes := encodeImport(tc.input)
			require.Equal(t, tc.expected, bytes)
		})
	}
}
