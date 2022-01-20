package text

import (
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero/wasm"
	"github.com/tetratelabs/wazero/wasm/leb128"
)

// funcParser parses any instructions and dispatches to onBodyEnd.
//
// Ex.  `(module (func (nop)))`
//       starts here --^    ^
// onBodyEnd resumes here --+
//
// Note: funcParser is reusable. The caller resets when reaching the appropriate tokenRParen via beginBody.
type funcParser struct {
	// m is used as a function pointer to moduleParser.tokenParser. This updates based on state changes.
	m *moduleParser

	// onBodyEnd is called when complete parsing the body. Unless testing, this should be moduleParser.parseFuncEnd
	onBodyEnd tokenParser

	currentInstruction wasm.Opcode
	// currentParameters are the current parameters to currentInstruction in WebAssembly 1.0 (MVP) binary format
	currentParameters []byte

	// currentCode is the current function body encoded in WebAssembly 1.0 (MVP) binary format
	currentCode []byte
}

var end = []byte{wasm.OpcodeEnd}

func (p *funcParser) getCode() []byte {
	if p.currentCode == nil {
		return end
	}
	return append(p.currentCode, wasm.OpcodeEnd)
}

// beginLocalsOrBody returns a parser that consumes a function body
//
// The onBodyEnd field is invoked once any instructions are written into currentCode.
//
// Ex. Given the source `(module (func (nop)))`
//             beginBody starts here --^    ^
//                 onBodyEnd resumes here --+
//
//
// NOTE: An empty function is valid and will not reach a tokenLParen! Ex. `(module (func))`
func (p *funcParser) beginBody() tokenParser {
	p.currentCode = nil
	p.m.tokenParser = p.parseBody
	return p.m.parse
}

// beginBodyField returns a parser that starts inside the first field of a function that isn't a type use.
//
// The onBodyEnd field is invoked once any instructions are written into currentCode.
//
// Ex. Given the source `(module (func $main (param i32) (nop)))`
//                          beginBodyField starts here --^    ^
//                                   onBodyEnd resumes here --+
//
//
// NOTE: An empty function is valid and will not reach a tokenLParen! Ex. `(module (func))`
func (p *funcParser) beginBodyField() tokenParser {
	p.currentCode = nil
	p.m.tokenParser = p.beginInstruction
	return p.m.parse
}

func (p *funcParser) parseBody(tok tokenType, tokenBytes []byte, line, col uint32) error {
	if tok == tokenLParen {
		p.beginBodyField()
		return nil
	}
	// If we reached this point, we have one or more locals, but no result. Ex. (func (local i32)) or (func)
	return p.onBodyEnd(tok, tokenBytes, line, col)
}

func (p *funcParser) parseMoreBody(tok tokenType, tokenBytes []byte, line, col uint32) error {
	if tok == tokenLParen {
		p.m.tokenParser = p.beginInstruction
		return nil
	}
	return p.onBodyEnd(tok, tokenBytes, line, col)
}

// beginInstruction is a tokenParser called after a tokenLParen and accepts an instruction field.
func (p *funcParser) beginInstruction(tok tokenType, tokenBytes []byte, line, col uint32) error {
	if tok == tokenKeyword {
		switch string(tokenBytes) {
		case "i32.const":
			p.currentInstruction = wasm.OpcodeI32Const
			p.m.tokenParser = p.parseI32Const
			return nil
		}
		return fmt.Errorf("unsupported instruction: %s", tokenBytes)
	}
	return p.onBodyEnd(tok, tokenBytes, line, col)
}

func (p *funcParser) parseI32Const(tok tokenType, tokenBytes []byte, _, _ uint32) error {
	if tok == tokenUN { // Ex. 1
		i, err := decodeUint32(tokenBytes)
		if err != nil {
			return fmt.Errorf("malformed i32 %s: %w", tokenBytes, errors.Unwrap(err))
		}
		p.currentParameters = leb128.EncodeUint32(i)
		p.m.tokenParser = p.parseEndInstruction
		return nil
	}
	return unexpectedToken(tok, tokenBytes)
}

func (p *funcParser) parseEndInstruction(tok tokenType, tokenBytes []byte, _, _ uint32) error {
	if tok == tokenRParen { // end of this field
		p.currentCode = append(p.currentCode, p.currentInstruction)
		p.currentCode = append(p.currentCode, p.currentParameters...)
		p.currentParameters = nil
		p.m.tokenParser = p.parseMoreBody
		return nil
	}
	return unexpectedToken(tok, tokenBytes)
}

func (p *funcParser) errorContext() string {
	return "" // TODO: add locals etc
}
