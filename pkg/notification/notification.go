// Package notification provides bucket notification configuration types.
package notification

import (
	"encoding/xml"
	"fmt"
	"io"
)

// EventType represents a bucket notification event.
type EventType string

const (
	ObjectCreatedAll                     EventType = "s3:ObjectCreated:*"
	ObjectCreatedPut                     EventType = "s3:ObjectCreated:Put"
	ObjectCreatedPost                    EventType = "s3:ObjectCreated:Post"
	ObjectCreatedCopy                    EventType = "s3:ObjectCreated:Copy"
	ObjectCreatedCompleteMultipartUpload EventType = "s3:ObjectCreated:CompleteMultipartUpload"
	ObjectRemovedAll                     EventType = "s3:ObjectRemoved:*"
	ObjectRemovedDelete                  EventType = "s3:ObjectRemoved:Delete"
	ObjectRemovedDeleteMarkerCreated     EventType = "s3:ObjectRemoved:DeleteMarkerCreated"
	ObjectAccessedAll                    EventType = "s3:ObjectAccessed:*"
	ObjectAccessedGet                    EventType = "s3:ObjectAccessed:Get"
	ObjectAccessedHead                   EventType = "s3:ObjectAccessed:Head"
)

// FilterRule defines a prefix/suffix filter rule.
type FilterRule struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

// S3Key carries prefix/suffix rules for notification filters.
type S3Key struct {
	FilterRules []FilterRule `xml:"FilterRule,omitempty"`
}

// Filter defines notification filters.
type Filter struct {
	S3Key S3Key `xml:"S3Key,omitempty"`
}

// Config represents a notification configuration target.
type Config struct {
	ID     string      `xml:"Id,omitempty"`
	Events []EventType `xml:"Event"`
	Filter *Filter     `xml:"Filter,omitempty"`
}

// QueueConfig carries one queue notification configuration.
type QueueConfig struct {
	Config
	Queue string `xml:"Queue"`
}

// TopicConfig carries one topic notification configuration.
type TopicConfig struct {
	Config
	Topic string `xml:"Topic"`
}

// LambdaConfig carries one lambda notification configuration.
type LambdaConfig struct {
	Config
	Lambda string `xml:"CloudFunction"`
}

// Configuration represents the full notification configuration.
type Configuration struct {
	XMLName       xml.Name       `xml:"NotificationConfiguration"`
	LambdaConfigs []LambdaConfig `xml:"CloudFunctionConfiguration,omitempty"`
	TopicConfigs  []TopicConfig  `xml:"TopicConfiguration,omitempty"`
	QueueConfigs  []QueueConfig  `xml:"QueueConfiguration,omitempty"`
}

// ToXML marshals the notification configuration to XML.
func (c Configuration) ToXML() ([]byte, error) {
	data, err := xml.Marshal(&c)
	if err != nil {
		return nil, fmt.Errorf("marshal notification xml: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

// ParseConfig parses a notification configuration from an XML reader.
func ParseConfig(reader io.Reader) (Configuration, error) {
	var cfg Configuration
	if err := xml.NewDecoder(reader).Decode(&cfg); err != nil {
		return Configuration{}, fmt.Errorf("decode notification xml: %w", err)
	}
	return cfg, nil
}
