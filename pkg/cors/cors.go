// Package cors provides CORS configuration types and XML helpers.
package cors

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const defaultXMLNS = "http://s3.amazonaws.com/doc/2006-03-01/"

// Config represents a bucket CORS configuration.
type Config struct {
	XMLNS     string   `xml:"xmlns,attr,omitempty"`
	XMLName   xml.Name `xml:"CORSConfiguration"`
	CORSRules []Rule   `xml:"CORSRule"`
}

// Rule represents a single CORS rule.
type Rule struct {
	AllowedHeader []string `xml:"AllowedHeader,omitempty"`
	AllowedMethod []string `xml:"AllowedMethod,omitempty"`
	AllowedOrigin []string `xml:"AllowedOrigin,omitempty"`
	ExposeHeader  []string `xml:"ExposeHeader,omitempty"`
	ID            string   `xml:"ID,omitempty"`
	MaxAgeSeconds int      `xml:"MaxAgeSeconds,omitempty"`
}

// NewConfig creates a new CORS configuration with the given rules.
func NewConfig(rules []Rule) Config {
	return Config{
		XMLNS: defaultXMLNS,
		XMLName: xml.Name{
			Local: "CORSConfiguration",
			Space: defaultXMLNS,
		},
		CORSRules: rules,
	}
}

// ParseBucketCORSConfig parses a CORS configuration in XML from an io.Reader.
func ParseBucketCORSConfig(reader io.Reader) (Config, error) {
	var cfg Config

	if err := xml.NewDecoder(io.LimitReader(reader, 128*1024)).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode cors xml: %w", err)
	}
	if cfg.XMLNS == "" {
		cfg.XMLNS = defaultXMLNS
	}
	for i, rule := range cfg.CORSRules {
		for j, method := range rule.AllowedMethod {
			cfg.CORSRules[i].AllowedMethod[j] = strings.ToUpper(method)
		}
	}
	return cfg, nil
}

// ToXML marshals the CORS configuration to XML.
func (c Config) ToXML() ([]byte, error) {
	if c.XMLNS == "" {
		c.XMLNS = defaultXMLNS
	}
	data, err := xml.Marshal(&c)
	if err != nil {
		return nil, fmt.Errorf("marshal cors xml: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}
