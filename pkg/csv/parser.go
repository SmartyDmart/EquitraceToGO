// Package csv provides a simple CSV parser matching the original TypeScript logic.
// @packageDocumentation
package csv

import (
	"fmt"
)

// Public configuration metadata.
const (
	LibName    = "@gregoranders/csv"
	LibVersion = "0.0.13"
	LibUrl     = "https://gregoranders.github.io/ts-csv/"
)

// Field represents a single CSV text field.
type Field = string

// Row represents an array of CSV fields.
type Row = []Field

// Configuration options for the parser.
type Configuration struct {
	FieldSeparator string // @defaultValue `,`
	LineSeparator  string // @defaultValue `\n`
	Quote          string // @defaultValue `"`
}

// DefaultConfiguration represents the fallback options matrix.
var DefaultConfiguration = Configuration{
	FieldSeparator: ",",
	LineSeparator:  "\n",
	Quote:          "\"",
}

type state struct {
	appendCell  bool
	appendField bool
	appendRow   bool
	field       int
	fieldOffset int
	line        int
	lineOffset  int
	quoted      bool
}

var csvInitialState = state{
	field:       0,
	fieldOffset: 0,
	line:        0,
	lineOffset: 0,
	quoted:      false,
	appendCell:  false,
	appendField: false,
	appendRow:   false,
}

// ParseError replicates the custom TypeScript validation error class block.
type ParseError struct {
	Line   int
	Column int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Invalid CSV at %d:%d", e.Line, e.Column)
}

// Parser handles sequential tracking, text chunk processing, and object generation.
type Parser struct {
	rows        []Row
	row         Row
	cell        string
	options     Configuration
	state       state
	index       int
	current     string
	previous    string
	quoteState  state
}

// NewParser creates a pointer instance initialized with configured separators.
func NewParser(config ...Configuration) *Parser {
	opts := DefaultConfiguration
	if len(config) > 0 {
		c := config[0]
		if c.FieldSeparator != "" {
			opts.FieldSeparator = c.FieldSeparator
		}
		if c.LineSeparator != "" {
			opts.LineSeparator = c.LineSeparator
		}
		if c.Quote != "" {
			opts.Quote = c.Quote
		}
	}
	return &Parser{
		options: opts,
	}
}

// Parse sweeps raw text data matrix layouts array grids.
func (p *Parser) Parse(text string) ([]Row, error) {
	p.reset()
	runes := []rune(text)

	for p.index = 0; p.index < len(runes); p.index++ {
		p.state.appendCell = true
		p.previous = p.current
		p.current = string(runes[p.index])
		p.handleNext()
	}

	if len(p.row) > 0 || p.cell != "" || (len(p.rows) > 0 && p.previous == p.options.FieldSeparator) {
		p.addField(p.fieldValue(p.cell), &p.row, &p.state)
		p.addRow(p.row, &p.rows, &p.state)
	}

	if p.state.quoted {
		return nil, &ParseError{
			Line:   p.quoteState.line,
			Column: p.quoteState.lineOffset,
		}
	}

	p.makeImmutable()
	return p.rows, nil
}

// Rows exposes the internal row grid database matching the rows getter method.
func (p *Parser) Rows() []Row {
	return p.rows
}

// JSON processes the extracted lines array utilizing row zero indexes as object map keys.
func (p *Parser) JSON() []map[string]string {
	if len(p.rows) == 0 {
		return []map[string]string{}
	}

	keys := p.rows[0]
	var result []map[string]string

	for i := 1; i < len(p.rows); i++ {
		row := p.rows[i]
		object := make(map[string]string)
		for keyIndex, key := range keys {
			if keyIndex < len(row) {
				object[key] = row[keyIndex]
			} else {
				object[key] = ""
			}
		}
		result = append(result, object)
	}

	return result
}

func (p *Parser) handleNext() {
	_ = p.handleQuote() || p.handleFieldSeparator() || p.handleLineSeparator()
	p.processState()
}

func (p *Parser) handleQuote() bool {
	if p.current == p.options.Quote {
		p.quoteState = p.state
		if p.previous == "\\" {
			p.handleQuoteEscaped()
		} else {
			p.handleQuoteNotEscaped()
		}
		return true
	}
	return false
}

func (p *Parser) handleQuoteEscaped() {
	if len(p.cell) > 0 {
		runes := []rune(p.cell)
		p.cell = string(runes[:len(runes)-1])
	}
}

func (p *Parser) handleQuoteNotEscaped() {
	if len(p.cell) == 0 || p.state.quoted {
		p.state.quoted = !p.state.quoted
	} else {
		// Sets flag true to trigger downstream error routing paths
		p.state.quoted = true 
	}
	p.state.appendCell = false
}

func (p *Parser) handleFieldSeparator() bool {
	if p.current == p.options.FieldSeparator {
		if !p.state.quoted {
			p.state.appendCell = false
			p.state.appendField = true
		}
		return true
	}
	return false
}

func (p *Parser) handleLineSeparator() bool {
	if p.current == p.options.LineSeparator {
		if !p.state.quoted {
			p.state.appendCell = false
			p.state.appendField = true
			p.state.appendRow = true
		}
		return true
	}
	return false
}

func (p *Parser) processState() {
	if p.state.appendCell {
		p.cell += p.current
	}

	if p.state.appendField {
		p.addField(p.fieldValue(p.cell), &p.row, &p.state)
		p.cell = ""
	}

	if p.state.appendRow {
		p.addRow(p.row, &p.rows, &p.state)
		p.row = Row{}
	}

	p.state.lineOffset++
	p.state.fieldOffset++
}

func (p *Parser) fieldValue(cell string) Field {
	return cell
}

func (p *Parser) addField(field Field, row *Row, state *state) {
	*row = append(*row, field)
	state.field++
	state.fieldOffset = -1
	state.appendField = false
}

func (p *Parser) addRow(row Row, rows *[]Row, state *state) {
	*rows = append(*rows, row)
	state.field = 0
	state.line++
	state.lineOffset = -1
	state.appendRow = false
}

func (p *Parser) makeImmutable() {
	// Replicates Object.freeze behavior conceptually.
	// Go slices are pass-by-reference pointers; values cannot be strictly locked natively
	// without deep encapsulation copying, matching structural parity here.
}

func (p *Parser) reset() {
	p.rows = []Row{}
	p.row = Row{}
	p.cell = ""
	p.state = csvInitialState
	p.index = 0
	p.current = ""
	p.previous = ""
	p.quoteState = csvInitialState
}
