// Package select provides SelectObjectContent helpers.
package s3select

import (
	"encoding/xml"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

// Options represents the select request payload.
type Options struct {
	XMLName xml.Name `xml:"SelectObjectContentRequest"`

	Expression           string              `xml:"Expression"`
	ExpressionType       QueryExpressionType `xml:"ExpressionType"`
	InputSerialization   InputSerialization  `xml:"InputSerialization"`
	OutputSerialization  OutputSerialization `xml:"OutputSerialization"`
	RequestProgress      RequestProgress     `xml:"RequestProgress"`
	ServerSideEncryption sse.Encrypter       `xml:"-"`
}

// RequestProgress controls select progress messages.
type RequestProgress struct {
	Enabled bool `xml:"Enabled"`
}

// Header returns HTTP headers for a select request.
func (o Options) Header() http.Header {
	headers := make(http.Header)
	if o.ServerSideEncryption != nil {
		o.ServerSideEncryption.ApplyHeaders(headers)
	}
	return headers
}
