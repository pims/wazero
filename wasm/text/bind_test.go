package text

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero/wasm"
)

func TestBindIndices(t *testing.T) {
	i32 := wasm.ValueTypeI32
	paramI32I32ResultI32 := &typeFunc{params: []wasm.ValueType{i32, i32}, result: i32}
	paramI32I32I32I32ResultI32 := &typeFunc{params: []wasm.ValueType{i32, i32, i32, i32}, result: i32}
	indexZero, indexOne := &index{numeric: 0}, &index{numeric: 1}

	tests := []struct {
		name            string
		input, expected *module
	}{
		{
			name: "import function: inlined type to numeric index",
			input: &module{
				types: []*typeFunc{paramI32I32I32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "fd_write",
						typeInlined: &inlinedTypeFunc{paramI32I32I32I32ResultI32, 0, 0}},
				},
				importFuncNames: wasm.NameMap{{Index: wasm.Index(0), Name: "runtime.fd_write"}},
			},
			expected: &module{
				types: []*typeFunc{paramI32I32I32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "fd_write", typeIndex: indexZero},
				},
				importFuncNames: wasm.NameMap{{Index: wasm.Index(0), Name: "runtime.fd_write"}},
			},
		},
		{
			name: "import function: multiple inlined types to numeric indices",
			input: &module{
				types: []*typeFunc{paramI32I32ResultI32, paramI32I32I32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_sizes_get",
						typeInlined: &inlinedTypeFunc{paramI32I32ResultI32, 0, 0}},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "fd_write",
						typeInlined: &inlinedTypeFunc{paramI32I32I32I32ResultI32, 0, 0}},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
			expected: &module{
				types: []*typeFunc{paramI32I32ResultI32, paramI32I32I32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_sizes_get", typeIndex: indexZero},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "fd_write", typeIndex: indexOne},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
		},
		{
			name: "import function: multiple inlined types to same numeric index",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_get",
						typeInlined: &inlinedTypeFunc{paramI32I32ResultI32, 0, 0}},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "args_sizes_get",
						typeInlined: &inlinedTypeFunc{paramI32I32ResultI32, 0, 0}},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
			expected: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_get", typeIndex: indexOne},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "args_sizes_get", typeIndex: indexOne},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
		},
		{
			name: "import function: multiple type names to numeric indices",
			input: &module{
				types: []*typeFunc{
					typeFuncEmpty,
					{name: "i32i32_i32", params: []wasm.ValueType{i32, i32}, result: i32},
					{name: "i32i32i32i32_i32", params: []wasm.ValueType{i32, i32, i32, i32}, result: i32},
				},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_sizes_get",
						typeIndex: &index{ID: "i32i32_i32", line: 5, col: 86}},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "fd_write",
						typeIndex: &index{ID: "i32i32i32i32_i32", line: 6, col: 76}},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
			expected: &module{
				types: []*typeFunc{
					typeFuncEmpty,
					{name: "i32i32_i32", params: []wasm.ValueType{i32, i32}, result: i32},
					{name: "i32i32i32i32_i32", params: []wasm.ValueType{i32, i32, i32, i32}, result: i32},
				},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_sizes_get",
						typeIndex: &index{numeric: 1, line: 5, col: 86}},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "fd_write",
						typeIndex: &index{numeric: 2, line: 6, col: 76}},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
		},
		{
			name: "import function: multiple type numeric indices left alone",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32, paramI32I32I32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_sizes_get",
						typeIndex: &index{numeric: 1, line: 5, col: 86}},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "fd_write",
						typeIndex: &index{numeric: 2, line: 6, col: 76}},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
			expected: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32, paramI32I32I32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_sizes_get",
						typeIndex: &index{numeric: 1, line: 5, col: 86}},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "fd_write",
						typeIndex: &index{numeric: 2, line: 6, col: 76}},
				},
				importFuncNames: wasm.NameMap{
					{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
					{Index: wasm.Index(1), Name: "runtime.fd_write"},
				},
			},
		},
		{
			name: "export imported func",
			input: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{module: "foo", name: "bar", typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"}},
				exportFuncs: []*exportFunc{
					{name: "bar", exportIndex: wasm.Index(0), funcIndex: &index{ID: "bar", line: 3, col: 22}},
				},
			},
			expected: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{module: "foo", name: "bar", typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"}},
				exportFuncs: []*exportFunc{
					{name: "bar", exportIndex: wasm.Index(0), funcIndex: &index{numeric: 0, line: 3, col: 22}},
				},
			},
		},
		{
			name: "export imported func twice",
			input: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{module: "foo", name: "bar", typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"}},
				exportFuncs: []*exportFunc{
					{name: "foo", exportIndex: wasm.Index(0), funcIndex: &index{ID: "bar", line: 3, col: 22}},
					{name: "bar", exportIndex: wasm.Index(1), funcIndex: &index{ID: "bar", line: 4, col: 22}},
				},
			},
			expected: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{module: "foo", name: "bar", typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"}},
				exportFuncs: []*exportFunc{
					{name: "foo", exportIndex: wasm.Index(0), funcIndex: &index{numeric: 0, line: 3, col: 22}},
					{name: "bar", exportIndex: wasm.Index(1), funcIndex: &index{numeric: 0, line: 4, col: 22}},
				},
			},
		},
		{
			name: "export different func",
			input: &module{
				types: []*typeFunc{typeFuncEmpty},
				importFuncs: []*importFunc{
					{module: "foo", name: "bar", importIndex: wasm.Index(0), typeIndex: indexZero},
					{module: "baz", name: "qux", importIndex: wasm.Index(1), typeIndex: indexZero},
				},
				importFuncNames: wasm.NameMap{
					&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"},
					&wasm.NameAssoc{Index: wasm.Index(1), Name: "qux"},
				},
				exportFuncs: []*exportFunc{
					{name: "foo", exportIndex: wasm.Index(0), funcIndex: &index{ID: "bar", line: 4, col: 22}},
					{name: "bar", exportIndex: wasm.Index(1), funcIndex: &index{ID: "qux", line: 5, col: 22}},
				},
			},
			expected: &module{
				types: []*typeFunc{typeFuncEmpty},
				importFuncs: []*importFunc{
					{module: "foo", name: "bar", importIndex: wasm.Index(0), typeIndex: indexZero},
					{module: "baz", name: "qux", importIndex: wasm.Index(1), typeIndex: indexZero},
				},
				importFuncNames: wasm.NameMap{
					&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"},
					&wasm.NameAssoc{Index: wasm.Index(1), Name: "qux"},
				},
				exportFuncs: []*exportFunc{
					{name: "foo", exportIndex: wasm.Index(0), funcIndex: &index{numeric: 0, line: 4, col: 22}},
					{name: "bar", exportIndex: wasm.Index(1), funcIndex: &index{numeric: 1, line: 5, col: 22}},
				},
			},
		},
		{
			name: "start: imported function name to numeric index",
			input: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{typeIndex: indexZero}, {typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{{Index: wasm.Index(0), Name: "one"}, {Index: wasm.Index(1), Name: "two"}},
				startFunction:   &index{ID: "two", line: 3, col: 9},
			},
			expected: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{typeIndex: indexZero}, {typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{{Index: wasm.Index(0), Name: "one"}, {Index: wasm.Index(1), Name: "two"}},
				startFunction:   &index{numeric: 1, line: 3, col: 9},
			},
		},
		{
			name: "start: imported function numeric index left alone",
			input: &module{
				types:         []*typeFunc{typeFuncEmpty},
				importFuncs:   []*importFunc{{name: "hello", importIndex: wasm.Index(0), typeIndex: indexZero}},
				startFunction: &index{numeric: 0, line: 3, col: 9},
			},
			expected: &module{
				types:         []*typeFunc{typeFuncEmpty},
				importFuncs:   []*importFunc{{name: "hello", importIndex: wasm.Index(0), typeIndex: indexZero}},
				startFunction: &index{numeric: 0, line: 3, col: 9},
			},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			err := bindIndices(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func TestBindIndices_Errors(t *testing.T) {
	i32 := wasm.ValueTypeI32
	indexZero := &index{}

	tests := []struct {
		name        string
		input       *module
		expectedErr string
	}{
		{
			name: "function type points out of range",
			input: &module{
				types:       []*typeFunc{typeFuncEmpty},
				importFuncs: []*importFunc{{name: "hello", typeIndex: &index{numeric: 1, line: 3, col: 9}}},
			},
			expectedErr: "3:9: index 1 is out of range [0..0] in module.import[0].func.type",
		},
		{
			name: "function type points nowhere",
			input: &module{
				types:       []*typeFunc{typeFuncEmpty},
				importFuncs: []*importFunc{{name: "hello", typeIndex: &index{ID: "main", line: 3, col: 9}}},
			},
			expectedErr: "3:9: unknown ID $main in module.import[0].func.type",
		},
		{
			name: "import function: inlined type doesn't match indexed",
			input: &module{
				types: []*typeFunc{typeFuncEmpty},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "", name: "hello", typeIndex: indexZero,
						typeInlined: &inlinedTypeFunc{&typeFunc{params: []wasm.ValueType{i32, i32}}, 3, 9}},
				},
			},
			expectedErr: "3:9: inlined type doesn't match type index 0 in module.import[0].func.type",
		},
		{
			name: "export func points out of range",
			input: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{module: "foo", name: "bar", typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"}},
				exportFuncs: []*exportFunc{
					{name: "bar", exportIndex: wasm.Index(0), funcIndex: &index{numeric: 3, line: 3, col: 22}},
				},
			},
			expectedErr: "3:22: index 3 is out of range [0..0] in module.exports[0].func",
		},
		{
			name: "export func points nowhere",
			input: &module{
				types:           []*typeFunc{typeFuncEmpty},
				importFuncs:     []*importFunc{{module: "foo", name: "bar", typeIndex: indexZero}},
				importFuncNames: wasm.NameMap{&wasm.NameAssoc{Index: wasm.Index(0), Name: "bar"}},
				exportFuncs: []*exportFunc{
					{name: "bar", exportIndex: wasm.Index(0), funcIndex: &index{ID: "qux", line: 3, col: 22}},
				},
			},
			expectedErr: "3:22: unknown ID $qux in module.exports[0].func",
		},
		{
			name: "start points out of range",
			input: &module{
				types:         []*typeFunc{typeFuncEmpty},
				importFuncs:   []*importFunc{{name: "hello", importIndex: wasm.Index(0), typeIndex: indexZero}},
				startFunction: &index{numeric: 1, line: 3, col: 9},
			},
			expectedErr: "3:9: index 1 is out of range [0..0] in module.start",
		},
		{
			name: "start points nowhere",
			input: &module{
				startFunction: &index{ID: "main", line: 1, col: 16},
			},
			expectedErr: "1:16: unknown ID $main in module.start",
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			err := bindIndices(tc.input)
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}
