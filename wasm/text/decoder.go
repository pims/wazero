package text

import (
	"github.com/tetratelabs/wazero/wasm"
)

// DecodeModule implements wasm.DecodeModule for the WebAssembly 1.0 (MVP) Text Format
// See https://www.w3.org/TR/wasm-core-1/#text-format%E2%91%A0
func DecodeModule(source []byte) (result *wasm.Module, err error) {
	// First, attempt to parse the module into a basic structure representing the text format. If this errs, return
	// immediately without wrapping parseModule returns FormatError, which is pre-formatted.
	var m *module
	if m, err = parseModule(source); err != nil {
		return nil, err
	}

	result = &wasm.Module{}

	// Next, we need to convert the types from the text format into the binary one. This is easy because the only
	// difference is that the text format has type names and the binary format does not.
	typeCount := len(m.types)
	if typeCount > 0 {
		result.TypeSection = make([]*wasm.FunctionType, typeCount)
		for i, t := range m.types {
			var results []wasm.ValueType
			if t.result != 0 {
				results = []wasm.ValueType{t.result}
			}
			result.TypeSection[i] = &wasm.FunctionType{Params: t.params, Results: results}
		}
	}

	// Now, handle any imported functions. Notably, we retain the same insertion order as defined in the text format in
	// case a numeric index is used for the start function (or another reason such as the call instruction).
	importFuncCount := len(m.importFuncs)
	if importFuncCount > 0 {
		result.ImportSection = make([]*wasm.Import, importFuncCount)
		for i, f := range m.importFuncs {
			result.ImportSection[i] = &wasm.Import{
				Module: f.module, Name: f.name,
				Kind:     wasm.ImportKindFunc,
				DescFunc: f.typeIndex.numeric,
			}
		}
	}

	// Now, handle any exported functions. Notably, we retain the same insertion order as defined in the text format.
	exportFuncCount := len(m.exportFuncs)
	if exportFuncCount > 0 {
		result.ExportSection = make(map[string]*wasm.Export, exportFuncCount)
		for _, f := range m.exportFuncs {
			e := &wasm.Export{
				Name:  f.name,
				Kind:  wasm.ExportKindFunc,
				Index: f.funcIndex.numeric,
			}
			result.ExportSection[e.Name] = e
		}
	}

	// The start function is called on Module.Instantiate.
	if m.startFunction != nil {
		result.StartSection = &m.startFunction.numeric
	}

	// Don't set the name section unless we found a name!
	if localNames := mergeLocalNames(m); localNames != nil || m.name != "" || m.importFuncNames != nil {
		result.NameSection = &wasm.NameSection{
			ModuleName:    m.name,
			FunctionNames: m.importFuncNames,
			LocalNames:    localNames,
		}
	}
	return
}

// mergeLocalNames produces wasm.NameSection LocalNames. This has to be done post-parse as types can be defined after
// functions that use them.
func mergeLocalNames(m *module) wasm.IndirectNameMap {
	j, jLen := 0, len(m.importFuncParamNames)
	if m.typeParamNames == nil && jLen == 0 {
		return nil
	}

	var result wasm.IndirectNameMap
	for i, f := range m.importFuncs {
		funcIdx := wasm.Index(i)

		// seek to see if we have any function-defined parameter names
		var inlinedNames wasm.NameMap
		for ; j < jLen; j++ {
			next := m.importFuncParamNames[j]
			if next.Index > funcIdx {
				break // we have parameter names, but starting at a later index
			} else if next.Index == funcIdx {
				inlinedNames = next.NameMap
				break
			}
		}
		// TODO: module function and local names

		typeNames, hasType := m.typeParamNames[f.typeIndex.numeric]
		var localNames wasm.NameMap
		if inlinedNames == nil && !hasType {
			continue
		} else if inlinedNames == nil {
			localNames = typeNames
		} else {
			// On conflict, choose the function names, as merge rules aren't defined in the specification. If there are
			// names on the function, the user added them. They may not intend to inherit names they didn't define!
			localNames = inlinedNames
		}
		result = append(result, &wasm.NameMapAssoc{Index: funcIdx, NameMap: localNames})
	}
	return result
}
