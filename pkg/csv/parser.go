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
	LibURL     = LibUrl
)

// Field represents a single CSV text field.
type Field = string

// Row represents an array of CSV fields.
type Row = []Field

// Configuration options for the parser.
type Configuration struct {
	FieldSeparator any // @defaultValue `,`
	LineSeparator  any // @defaultValue `\n`
	Quote          any // @defaultValue `"`
}

// ParserOptions is kept as the public alias expected by the test suite.
type ParserOptions = Configuration

// DefaultConfiguration represents the fallback options matrix.
var DefaultConfiguration = Configuration{
	FieldSeparator: ",",
	LineSeparator:  "\n",
	Quote:          "\"",
}

func optionString(value any, fallback string) string {
	switch v := value.(type) {
	case string:
		return v
	case rune:
		return string(v)
	case byte:
		return string(v)
	default:
		return fallback
	}
}

func (c Configuration) fieldSeparator() string {
	return optionString(c.FieldSeparator, ",")
}

func (c Configuration) lineSeparator() string {
	return optionString(c.LineSeparator, "\n")
}

func (c Configuration) quote() string {
	return optionString(c.Quote, "\"")
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
	lineOffset: -1,
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
	runes       []rune
}

// NewParser creates a pointer instance initialized with configured separators.
func NewParser(config ...*Configuration) *Parser {
	opts := DefaultConfiguration
	if len(config) > 0 && config[0] != nil {
		c := config[0]
		if c.FieldSeparator != nil {
			opts.FieldSeparator = optionString(c.FieldSeparator, opts.fieldSeparator())
		}
		if c.LineSeparator != nil {
			opts.LineSeparator = optionString(c.LineSeparator, opts.lineSeparator())
		}
		if c.Quote != nil {
			opts.Quote = optionString(c.Quote, opts.quote())
		}
	}
	return &Parser{
		options: opts,
	}
}

// Parse sweeps raw text data matrix layouts array grids.
func (p *Parser) Parse(text string) ([]Row, error) {
	p.reset()

	rows := []Row{}
	row := Row{}
	cell := ""
	inQuotes := false
	quoteStartLine := 0
	quoteStartColumn := 0
	quoteErrorLine := 0
	quoteErrorColumn := 0
	line := 0
	column := 0
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		current := string(runes[i])

		if inQuotes {
			if current == p.options.quote() {
				if i+1 < len(runes) && string(runes[i+1]) == p.options.quote() {
					cell += current
					quoteErrorLine = line
					quoteErrorColumn = column + 1
					i++
					column += 2
					continue
				}
				inQuotes = false
				column++
				continue
			}
			cell += current
			column++
			continue
		}

		switch current {
		case p.options.quote():
			if cell == "" {
				inQuotes = true
				quoteStartLine = line
				quoteStartColumn = column
			} else {
				return nil, &ParseError{Line: line, Column: column}
			}
		case p.options.fieldSeparator():
			row = append(row, cell)
			cell = ""
		case p.options.lineSeparator():
			row = append(row, cell)
			rows = append(rows, row)
			row = Row{}
			cell = ""
			line++
			column = 0
			continue
		default:
			cell += current
		}
		column++
	}

	if inQuotes {
		if quoteErrorColumn > 0 {
			return nil, &ParseError{Line: quoteErrorLine, Column: quoteErrorColumn}
		}
		return nil, &ParseError{Line: quoteStartLine, Column: quoteStartColumn}
	}

	if cell != "" || len(row) > 0 {
		row = append(row, cell)
		rows = append(rows, row)
	}

	p.rows = cloneRows(rows)
	p.makeImmutable()
	return cloneRows(rows), nil
}

// Rows exposes the internal row grid database matching the rows getter method.
func (p *Parser) Rows() []Row {
	return cloneRows(p.rows)
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

	return cloneJSON(result)
}

func (p *Parser) handleNext() {
	_ = p.handleQuote() || p.handleFieldSeparator() || p.handleLineSeparator()
	p.processState()
}

func (p *Parser) handleQuote() bool {
	if p.current == p.options.quote() {
		p.quoteState = p.state
		if p.state.quoted && p.index+1 < len(p.runes) && string(p.runes[p.index+1]) == p.options.quote() {
			p.cell += p.current
			p.state.appendCell = true
			p.index++
			return true
		}
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
	if p.current == p.options.fieldSeparator() {
		if !p.state.quoted {
			p.state.appendCell = false
			p.state.appendField = true
		}
		return true
	}
	return false
}

func (p *Parser) handleLineSeparator() bool {
	if p.current == p.options.lineSeparator() {
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

func cloneRows(rows []Row) []Row {
	copied := make([]Row, len(rows))
	for i, row := range rows {
		copied[i] = append(Row(nil), row...)
	}
	return copied
}

func cloneJSON(rows []map[string]string) []map[string]string {
	copied := make([]map[string]string, len(rows))
	for i, row := range rows {
		entry := make(map[string]string, len(row))
		for key, value := range row {
			entry[key] = value
		}
		copied[i] = entry
	}
	return copied
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
	p.runes = nil
}
