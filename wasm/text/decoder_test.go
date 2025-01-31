package text

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero/wasm"
)

func TestDecodeModule(t *testing.T) {
	zero := uint32(0)
	i32, i64 := wasm.ValueTypeI32, wasm.ValueTypeI64
	tests := []struct {
		name     string
		input    string
		expected *wasm.Module
	}{
		{
			name:     "empty",
			input:    "(module)",
			expected: &wasm.Module{},
		},
		{
			name:  "import func empty",
			input: "(module (import \"foo\" \"bar\" (func)))", // ok empty sig
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{{}},
				ImportSection: []*wasm.Import{{
					Module: "foo", Name: "bar",
					Kind:     wasm.ImportKindFunc,
					DescFunc: 0,
				}},
			},
		},
		{
			name: "multiple import func with different inlined type",
			input: `(module
	(type (func) (; ensures no false match on index 0 ;))
	(import "wasi_snapshot_preview1" "path_open" (func $runtime.path_open (param i32 i32 i32 i32 i32 i64 i64 i32 i32) (result i32)))
	(import "wasi_snapshot_preview1" "fd_write" (func $runtime.fd_write (param i32 i32 i32 i32) (result i32)))
)`,
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{
					{},
					{Params: []wasm.ValueType{i32, i32, i32, i32, i32, i64, i64, i32, i32}, Results: []wasm.ValueType{i32}},
					{Params: []wasm.ValueType{i32, i32, i32, i32}, Results: []wasm.ValueType{i32}},
				},
				ImportSection: []*wasm.Import{
					{
						Module: "wasi_snapshot_preview1", Name: "path_open",
						Kind:     wasm.ImportKindFunc,
						DescFunc: 1,
					}, {
						Module: "wasi_snapshot_preview1", Name: "fd_write",
						Kind:     wasm.ImportKindFunc,
						DescFunc: 2,
					},
				},
				NameSection: &wasm.NameSection{
					FunctionNames: wasm.NameMap{
						{Index: wasm.Index(0), Name: "runtime.path_open"},
						{Index: wasm.Index(1), Name: "runtime.fd_write"},
					},
				},
			},
		},
		{
			name: "multiple import func different type - name index",
			input: `(module
	(type (func) (; ensures no false match on index 0 ;))
	(type $i32i32_i32 (func (param i32 i32) (result i32)))
	(type $i32i32i32i32_i32 (func (param i32 i32 i32 i32) (result i32)))
	(import "wasi_snapshot_preview1" "args_sizes_get" (func $runtime.args_sizes_get (type $i32i32_i32)))
	(import "wasi_snapshot_preview1" "fd_write" (func $runtime.fd_write (type $i32i32i32i32_i32)))
)`,
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{
					{},
					{Params: []wasm.ValueType{i32, i32}, Results: []wasm.ValueType{i32}},
					{Params: []wasm.ValueType{i32, i32, i32, i32}, Results: []wasm.ValueType{i32}},
				},
				ImportSection: []*wasm.Import{
					{
						Module: "wasi_snapshot_preview1", Name: "args_sizes_get",
						Kind:     wasm.ImportKindFunc,
						DescFunc: 1,
					}, {
						Module: "wasi_snapshot_preview1", Name: "fd_write",
						Kind:     wasm.ImportKindFunc,
						DescFunc: 2,
					},
				},
				NameSection: &wasm.NameSection{
					FunctionNames: wasm.NameMap{
						{Index: wasm.Index(0), Name: "runtime.args_sizes_get"},
						{Index: wasm.Index(1), Name: "runtime.fd_write"},
					},
				},
			},
		},
		{
			name: "export imported func",
			input: `(module
	(import "foo" "bar" (func $bar))
	(export "bar" (func $bar))
)`,
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{{}},
				ImportSection: []*wasm.Import{
					{Module: "foo", Name: "bar", Kind: wasm.ImportKindFunc, DescFunc: 0},
				},
				ExportSection: map[string]*wasm.Export{
					"bar": {Name: "bar", Kind: wasm.ExportKindFunc, Index: wasm.Index(0)},
				},
				NameSection: &wasm.NameSection{FunctionNames: wasm.NameMap{{Index: wasm.Index(0), Name: "bar"}}},
			},
		},
		{
			name: "export different func",
			input: `(module
	(import "foo" "bar" (func $bar))
	(import "baz" "qux" (func $qux))
	(export "foo" (func $bar))
	(export "bar" (func $qux))
)`,
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{{}},
				ImportSection: []*wasm.Import{
					{Module: "foo", Name: "bar", Kind: wasm.ImportKindFunc, DescFunc: 0},
					{Module: "baz", Name: "qux", Kind: wasm.ImportKindFunc, DescFunc: 0},
				},
				ExportSection: map[string]*wasm.Export{
					"foo": {Name: "foo", Kind: wasm.ExportKindFunc, Index: wasm.Index(0)},
					"bar": {Name: "bar", Kind: wasm.ExportKindFunc, Index: wasm.Index(1)},
				},
				NameSection: &wasm.NameSection{
					FunctionNames: wasm.NameMap{
						{Index: wasm.Index(0), Name: "bar"},
						{Index: wasm.Index(1), Name: "qux"},
					},
				},
			},
		},
		{
			name: "export different func - numeric",
			input: `(module
	(import "foo" "bar" (func))
	(import "baz" "qux" (func))
	(export "foo" (func 0))
	(export "bar" (func 1))
)`,
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{{}},
				ImportSection: []*wasm.Import{
					{Module: "foo", Name: "bar", Kind: wasm.ImportKindFunc, DescFunc: 0},
					{Module: "baz", Name: "qux", Kind: wasm.ImportKindFunc, DescFunc: 0},
				},
				ExportSection: map[string]*wasm.Export{
					"foo": {Name: "foo", Kind: wasm.ExportKindFunc, Index: wasm.Index(0)},
					"bar": {Name: "bar", Kind: wasm.ExportKindFunc, Index: wasm.Index(1)},
				},
			},
		},
		{
			name: "start imported function by name",
			input: `(module
	(import "" "hello" (func $hello))
	(start $hello)
)`,
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{{}},
				ImportSection: []*wasm.Import{{
					Module: "", Name: "hello",
					Kind:     wasm.ImportKindFunc,
					DescFunc: 0,
				}},
				StartSection: &zero,
				NameSection:  &wasm.NameSection{FunctionNames: wasm.NameMap{{Index: wasm.Index(0), Name: "hello"}}},
			},
		},
		{
			name: "start imported function by index",
			input: `(module
	(import "" "hello" (func))
	(start 0)
)`,
			expected: &wasm.Module{
				TypeSection: []*wasm.FunctionType{{}},
				ImportSection: []*wasm.Import{{
					Module: "", Name: "hello",
					Kind:     wasm.ImportKindFunc,
					DescFunc: 0,
				}},
				StartSection: &zero,
			},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			m, err := DecodeModule([]byte(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.expected, m)
		})
	}
}

func TestMergeLocalNames(t *testing.T) {
	i32 := wasm.ValueTypeI32
	paramI32I32ResultI32 := &typeFunc{params: []wasm.ValueType{i32, i32}, result: i32}
	indexZero, indexOne := &index{numeric: 0}, &index{numeric: 1}

	tests := []struct {
		name     string
		input    *module
		expected wasm.IndirectNameMap
	}{
		{
			name: "no parameter names",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_get", typeIndex: indexOne},
				},
			},
		},
		{
			name: "type parameter names, but no import function parameter names",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_get", typeIndex: indexOne},
				},
				typeParamNames: map[wasm.Index]wasm.NameMap{
					wasm.Index(1): {{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}},
				},
			},
			expected: wasm.IndirectNameMap{
				{Index: wasm.Index(0), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
			},
		},
		{
			name: "import function parameter names, but no type parameter names",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_get", typeIndex: indexOne},
				},
				importFuncParamNames: wasm.IndirectNameMap{
					{Index: wasm.Index(0), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
				},
			},
			expected: wasm.IndirectNameMap{
				{Index: wasm.Index(0), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
			},
		},
		{
			name: "type parameter names, but no import function parameter names - function 2",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "", name: "", typeIndex: indexZero},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "args_get", typeIndex: indexOne},
				},
				typeParamNames: map[wasm.Index]wasm.NameMap{
					wasm.Index(1): {{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}},
				},
			},
			expected: wasm.IndirectNameMap{
				{Index: wasm.Index(1), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
			},
		},
		{
			name: "import function parameter names, but no type parameter names - function 2",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "", name: "", typeIndex: indexZero},
					{importIndex: wasm.Index(1), module: "wasi_snapshot_preview1", name: "args_get", typeIndex: indexOne},
				},
				importFuncParamNames: wasm.IndirectNameMap{
					{Index: wasm.Index(1), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
				},
			},
			expected: wasm.IndirectNameMap{
				{Index: wasm.Index(1), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
			},
		},
		{
			name: "conflict on import function parameter names and type parameter names",
			input: &module{
				types: []*typeFunc{typeFuncEmpty, paramI32I32ResultI32},
				importFuncs: []*importFunc{
					{importIndex: wasm.Index(0), module: "wasi_snapshot_preview1", name: "args_get", typeIndex: indexOne},
				},
				typeParamNames: map[wasm.Index]wasm.NameMap{
					wasm.Index(1): {{Index: wasm.Index(0), Name: "x"}, {Index: wasm.Index(0), Name: "y"}},
				},
				importFuncParamNames: wasm.IndirectNameMap{
					{Index: wasm.Index(0), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
				},
			},
			expected: wasm.IndirectNameMap{
				{Index: wasm.Index(0), NameMap: wasm.NameMap{{Index: wasm.Index(0), Name: "argv"}, {Index: wasm.Index(0), Name: "argv_buf"}}},
			},
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, mergeLocalNames(tc.input))
		})
	}
}

func TestDecodeModule_Errors(t *testing.T) {
	tests := []struct{ name, input, expectedErr string }{
		{
			name:        "invalid",
			input:       "module",
			expectedErr: "1:1: expected '(', but found keyword: module",
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeModule([]byte(tc.input))
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}
