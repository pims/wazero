package text

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero/wasm"
)

func TestBeginBody(t *testing.T) {
	i32Const1End := []byte{wasm.OpcodeI32Const, 0x01, wasm.OpcodeEnd}

	tests := []struct {
		name, input  string
		expectedCode []byte
	}{
		{
			name: "empty",
		},
		{
			name:         "i32.const",
			input:        "(i32.const 1)",
			expectedCode: i32Const1End,
		},
		{
			name:         "i32.const twice",
			input:        "(i32.const 1) (i32.const 2)",
			expectedCode: []byte{wasm.OpcodeI32Const, 0x01, wasm.OpcodeI32Const, 0x02, wasm.OpcodeEnd},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			lp := &funcParser{m: &moduleParser{}, onBodyEnd: noopTokenParser}
			_, _, err := lex(lp.beginBody(), []byte(tc.input))
			require.NoError(t, err)
			if tc.expectedCode == nil {
				require.Equal(t, end, lp.getCode())
			} else {
				require.Equal(t, tc.expectedCode, lp.getCode())
			}
		})
	}
}

func TestBeginBody_Errors(t *testing.T) {
	tests := []struct {
		name, input, expectedErr string
	}{
		{
			name:        "i32.const wrong value",
			input:       "(i32.const a)",
			expectedErr: "unexpected keyword: a",
		},
		{
			name:        "i32.const overflow",
			input:       "(i32.const 4294967296)",
			expectedErr: "malformed i32 4294967296: value out of range",
		},
		{
			name:        "not yet supported",
			input:       "(f32.const 1.1)",
			expectedErr: "unsupported instruction: f32.const",
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			lp := &funcParser{m: &moduleParser{}, onBodyEnd: noopTokenParser}
			_, _, err := lex(lp.beginBody(), []byte(tc.input))
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}
