// Package replication provides bucket replication configuration types and helpers.
package replication

import (
	"encoding/xml"
	"fmt"
	"io"
)

const defaultXMLNS = "http://s3.amazonaws.com/doc/2006-03-01/"

// ReplicationConfig represents a bucket replication configuration.
type ReplicationConfig struct {
	XMLNS   string   `xml:"xmlns,attr,omitempty"`
	XMLName xml.Name `xml:"ReplicationConfiguration"`
	Role    string   `xml:"Role,omitempty"`
	Rules   []Rule   `xml:"Rule"`
}

// Config is an alias for ReplicationConfig.
type Config = ReplicationConfig

// Rule defines a replication rule.
type Rule struct {
	ID          string      `xml:"ID,omitempty"`
	Status      Status      `xml:"Status"`
	Priority    int         `xml:"Priority,omitempty"`
	Filter      Filter      `xml:"Filter,omitempty"`
	Destination Destination `xml:"Destination"`
}

// Destination defines the replication target.
type Destination struct {
	Bucket       string `xml:"Bucket"`
	StorageClass string `xml:"StorageClass,omitempty"`
}

// Filter restricts which objects are replicated.
type Filter struct {
	Prefix string `xml:"Prefix,omitempty"`
	Tag    Tag    `xml:"Tag,omitempty"`
}

// Tag is a key/value pair for replication filtering.
type Tag struct {
	Key   string `xml:"Key,omitempty"`
	Value string `xml:"Value,omitempty"`
}

// Status represents replication rule status.
type Status string

const (
	Enabled  Status = "Enabled"
	Disabled Status = "Disabled"
)

// Normalize validates and normalizes the replication configuration.
func (c *ReplicationConfig) Normalize() error {
	if c.XMLNS == "" {
		c.XMLNS = defaultXMLNS
	}
	if c.XMLName.Local == "" {
		c.XMLName = xml.Name{Local: "ReplicationConfiguration", Space: defaultXMLNS}
	} else if c.XMLName.Space == "" {
		c.XMLName.Space = defaultXMLNS
	}
	for _, rule := range c.Rules {
		if rule.Status != Enabled && rule.Status != Disabled {
			return fmt.Errorf("invalid replication rule status")
		}
		if rule.Destination.Bucket == "" {
			return fmt.Errorf("replication destination bucket is required")
		}
	}
	return nil
}

// ToXML marshals the replication configuration to XML.
func (c ReplicationConfig) ToXML() ([]byte, error) {
	if err := c.Normalize(); err != nil {
		return nil, err
	}
	data, err := xml.Marshal(&c)
	if err != nil {
		return nil, fmt.Errorf("marshal replication xml: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

// ParseConfig parses replication configuration XML from a reader.
func ParseConfig(reader io.Reader) (ReplicationConfig, error) {
	var cfg ReplicationConfig
	if err := xml.NewDecoder(reader).Decode(&cfg); err != nil {
		return ReplicationConfig{}, fmt.Errorf("decode replication xml: %w", err)
	}
	if cfg.XMLNS == "" {
		cfg.XMLNS = defaultXMLNS
	}
	return cfg, nil
}
