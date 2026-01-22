// Package select provides SelectObjectContent helpers.
package s3select

import "encoding/xml"

// CSVFileHeaderInfo controls how CSV headers are handled.
type CSVFileHeaderInfo string

const (
	CSVFileHeaderInfoNone   CSVFileHeaderInfo = "NONE"
	CSVFileHeaderInfoIgnore CSVFileHeaderInfo = "IGNORE"
	CSVFileHeaderInfoUse    CSVFileHeaderInfo = "USE"
)

// SelectCompressionType describes input compression.
type SelectCompressionType string

const (
	SelectCompressionNone   SelectCompressionType = "NONE"
	SelectCompressionGZIP   SelectCompressionType = "GZIP"
	SelectCompressionBZIP   SelectCompressionType = "BZIP2"
	SelectCompressionZSTD   SelectCompressionType = "ZSTD"
	SelectCompressionLZ4    SelectCompressionType = "LZ4"
	SelectCompressionS2     SelectCompressionType = "S2"
	SelectCompressionSNAPPY SelectCompressionType = "SNAPPY"
)

// CSVQuoteFields controls how CSV fields are quoted.
type CSVQuoteFields string

const (
	CSVQuoteFieldsAlways   CSVQuoteFields = "Always"
	CSVQuoteFieldsAsNeeded CSVQuoteFields = "AsNeeded"
)

// QueryExpressionType describes the query language.
type QueryExpressionType string

const (
	QueryExpressionTypeSQL QueryExpressionType = "SQL"
)

// JSONType determines JSON input serialization type.
type JSONType string

const (
	JSONDocumentType JSONType = "DOCUMENT"
	JSONLinesType    JSONType = "LINES"
)

// ParquetInputOptions describes Parquet input options.
type ParquetInputOptions struct{}

// CSVInputOptions describes CSV input options.
type CSVInputOptions struct {
	FileHeaderInfo    CSVFileHeaderInfo
	fileHeaderInfoSet bool

	RecordDelimiter    string
	recordDelimiterSet bool

	FieldDelimiter    string
	fieldDelimiterSet bool

	QuoteCharacter    string
	quoteCharacterSet bool

	QuoteEscapeCharacter    string
	quoteEscapeCharacterSet bool

	Comments    string
	commentsSet bool
}

// SetFileHeaderInfo sets the file header info in the CSV input options.
func (c *CSVInputOptions) SetFileHeaderInfo(val CSVFileHeaderInfo) {
	c.FileHeaderInfo = val
	c.fileHeaderInfoSet = true
}

// SetRecordDelimiter sets the record delimiter in the CSV input options.
func (c *CSVInputOptions) SetRecordDelimiter(val string) {
	c.RecordDelimiter = val
	c.recordDelimiterSet = true
}

// SetFieldDelimiter sets the field delimiter in the CSV input options.
func (c *CSVInputOptions) SetFieldDelimiter(val string) {
	c.FieldDelimiter = val
	c.fieldDelimiterSet = true
}

// SetQuoteCharacter sets the quote character in the CSV input options.
func (c *CSVInputOptions) SetQuoteCharacter(val string) {
	c.QuoteCharacter = val
	c.quoteCharacterSet = true
}

// SetQuoteEscapeCharacter sets the quote escape character in the CSV input options.
func (c *CSVInputOptions) SetQuoteEscapeCharacter(val string) {
	c.QuoteEscapeCharacter = val
	c.quoteEscapeCharacterSet = true
}

// SetComments sets the comments character in the CSV input options.
func (c *CSVInputOptions) SetComments(val string) {
	c.Comments = val
	c.commentsSet = true
}

// MarshalXML renders CSV input options as XML.
func (c CSVInputOptions) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if c.FileHeaderInfo != "" || c.fileHeaderInfoSet {
		if err := e.EncodeElement(c.FileHeaderInfo, xml.StartElement{Name: xml.Name{Local: "FileHeaderInfo"}}); err != nil {
			return err
		}
	}
	if c.RecordDelimiter != "" || c.recordDelimiterSet {
		if err := e.EncodeElement(c.RecordDelimiter, xml.StartElement{Name: xml.Name{Local: "RecordDelimiter"}}); err != nil {
			return err
		}
	}
	if c.FieldDelimiter != "" || c.fieldDelimiterSet {
		if err := e.EncodeElement(c.FieldDelimiter, xml.StartElement{Name: xml.Name{Local: "FieldDelimiter"}}); err != nil {
			return err
		}
	}
	if c.QuoteCharacter != "" || c.quoteCharacterSet {
		if err := e.EncodeElement(c.QuoteCharacter, xml.StartElement{Name: xml.Name{Local: "QuoteCharacter"}}); err != nil {
			return err
		}
	}
	if c.QuoteEscapeCharacter != "" || c.quoteEscapeCharacterSet {
		if err := e.EncodeElement(c.QuoteEscapeCharacter, xml.StartElement{Name: xml.Name{Local: "QuoteEscapeCharacter"}}); err != nil {
			return err
		}
	}
	if c.Comments != "" || c.commentsSet {
		if err := e.EncodeElement(c.Comments, xml.StartElement{Name: xml.Name{Local: "Comments"}}); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// CSVOutputOptions describes CSV output options.
type CSVOutputOptions struct {
	QuoteFields    CSVQuoteFields
	quoteFieldsSet bool

	RecordDelimiter    string
	recordDelimiterSet bool

	FieldDelimiter    string
	fieldDelimiterSet bool

	QuoteCharacter    string
	quoteCharacterSet bool

	QuoteEscapeCharacter    string
	quoteEscapeCharacterSet bool
}

// SetQuoteFields sets the quote field parameter in the CSV output options.
func (c *CSVOutputOptions) SetQuoteFields(val CSVQuoteFields) {
	c.QuoteFields = val
	c.quoteFieldsSet = true
}

// SetRecordDelimiter sets the record delimiter character in the CSV output options.
func (c *CSVOutputOptions) SetRecordDelimiter(val string) {
	c.RecordDelimiter = val
	c.recordDelimiterSet = true
}

// SetFieldDelimiter sets the field delimiter character in the CSV output options.
func (c *CSVOutputOptions) SetFieldDelimiter(val string) {
	c.FieldDelimiter = val
	c.fieldDelimiterSet = true
}

// SetQuoteCharacter sets the quote character in the CSV output options.
func (c *CSVOutputOptions) SetQuoteCharacter(val string) {
	c.QuoteCharacter = val
	c.quoteCharacterSet = true
}

// SetQuoteEscapeCharacter sets the quote escape character in the CSV output options.
func (c *CSVOutputOptions) SetQuoteEscapeCharacter(val string) {
	c.QuoteEscapeCharacter = val
	c.quoteEscapeCharacterSet = true
}

// MarshalXML renders CSV output options as XML.
func (c CSVOutputOptions) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if c.QuoteFields != "" || c.quoteFieldsSet {
		if err := e.EncodeElement(c.QuoteFields, xml.StartElement{Name: xml.Name{Local: "QuoteFields"}}); err != nil {
			return err
		}
	}
	if c.RecordDelimiter != "" || c.recordDelimiterSet {
		if err := e.EncodeElement(c.RecordDelimiter, xml.StartElement{Name: xml.Name{Local: "RecordDelimiter"}}); err != nil {
			return err
		}
	}
	if c.FieldDelimiter != "" || c.fieldDelimiterSet {
		if err := e.EncodeElement(c.FieldDelimiter, xml.StartElement{Name: xml.Name{Local: "FieldDelimiter"}}); err != nil {
			return err
		}
	}
	if c.QuoteCharacter != "" || c.quoteCharacterSet {
		if err := e.EncodeElement(c.QuoteCharacter, xml.StartElement{Name: xml.Name{Local: "QuoteCharacter"}}); err != nil {
			return err
		}
	}
	if c.QuoteEscapeCharacter != "" || c.quoteEscapeCharacterSet {
		if err := e.EncodeElement(c.QuoteEscapeCharacter, xml.StartElement{Name: xml.Name{Local: "QuoteEscapeCharacter"}}); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// JSONInputOptions describes JSON input options.
type JSONInputOptions struct {
	Type    JSONType
	typeSet bool
}

// SetType sets the JSON type in the JSON input options.
func (j *JSONInputOptions) SetType(typ JSONType) {
	j.Type = typ
	j.typeSet = true
}

// MarshalXML renders JSON input options as XML.
func (j JSONInputOptions) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if j.Type != "" || j.typeSet {
		if err := e.EncodeElement(j.Type, xml.StartElement{Name: xml.Name{Local: "Type"}}); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// JSONOutputOptions describes JSON output options.
type JSONOutputOptions struct {
	RecordDelimiter    string
	recordDelimiterSet bool
}

// SetRecordDelimiter sets the record delimiter in the JSON output options.
func (j *JSONOutputOptions) SetRecordDelimiter(val string) {
	j.RecordDelimiter = val
	j.recordDelimiterSet = true
}

// MarshalXML renders JSON output options as XML.
func (j JSONOutputOptions) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if j.RecordDelimiter != "" || j.recordDelimiterSet {
		if err := e.EncodeElement(j.RecordDelimiter, xml.StartElement{Name: xml.Name{Local: "RecordDelimiter"}}); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// InputSerialization describes the input serialization for select.
type InputSerialization struct {
	CompressionType SelectCompressionType `xml:"CompressionType,omitempty"`
	Parquet         *ParquetInputOptions  `xml:"Parquet,omitempty"`
	CSV             *CSVInputOptions      `xml:"CSV,omitempty"`
	JSON            *JSONInputOptions     `xml:"JSON,omitempty"`
}

// OutputSerialization describes the output serialization for select.
type OutputSerialization struct {
	CSV  *CSVOutputOptions  `xml:"CSV,omitempty"`
	JSON *JSONOutputOptions `xml:"JSON,omitempty"`
}
