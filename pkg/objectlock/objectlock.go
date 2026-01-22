// Package objectlock provides object lock, legal hold, and retention types.
package objectlock

import (
	"encoding/xml"
	"time"
)

// ObjectLockEnabledValue is the only valid object lock state for S3.
const ObjectLockEnabledValue = "Enabled"

// Config represents bucket-level object lock configuration.
type Config struct {
	XMLName           xml.Name `xml:"ObjectLockConfiguration"`
	ObjectLockEnabled string   `xml:"ObjectLockEnabled"`
	Rule              *Rule    `xml:"Rule,omitempty"`
}

// Rule represents object lock default retention rules.
type Rule struct {
	DefaultRetention DefaultRetention `xml:"DefaultRetention"`
}

// DefaultRetention defines default retention settings.
type DefaultRetention struct {
	Mode  RetentionMode `xml:"Mode"`
	Days  int           `xml:"Days,omitempty"`
	Years int           `xml:"Years,omitempty"`
}

// RetentionMode is the retention mode for object locks.
type RetentionMode string

const (
	RetentionGovernance RetentionMode = "GOVERNANCE"
	RetentionCompliance RetentionMode = "COMPLIANCE"
)

// IsValid reports whether the retention mode is supported.
func (r RetentionMode) IsValid() bool {
	return r == RetentionGovernance || r == RetentionCompliance
}

// LegalHoldStatus indicates whether legal hold is enabled.
type LegalHoldStatus string

const (
	LegalHoldOn  LegalHoldStatus = "ON"
	LegalHoldOff LegalHoldStatus = "OFF"
)

// IsValid reports whether the legal hold status is supported.
func (l LegalHoldStatus) IsValid() bool {
	return l == LegalHoldOn || l == LegalHoldOff
}

// Retention represents object-level retention configuration.
type Retention struct {
	XMLName         xml.Name      `xml:"Retention"`
	Mode            RetentionMode `xml:"Mode,omitempty"`
	RetainUntilDate time.Time     `xml:"RetainUntilDate,omitempty"`
}

// LegalHold represents object-level legal hold configuration.
type LegalHold struct {
	XMLName xml.Name        `xml:"LegalHold"`
	Status  LegalHoldStatus `xml:"Status,omitempty"`
}

// Normalize validates and normalizes the configuration in-place.
func (c *Config) Normalize() error {
	if c.ObjectLockEnabled == "" {
		c.ObjectLockEnabled = ObjectLockEnabledValue
	}
	if c.ObjectLockEnabled != ObjectLockEnabledValue {
		return ErrInvalidObjectLockState
	}
	if c.Rule == nil {
		return nil
	}
	if !c.Rule.DefaultRetention.Mode.IsValid() {
		return ErrInvalidRetentionMode
	}
	days := c.Rule.DefaultRetention.Days
	years := c.Rule.DefaultRetention.Years
	if (days > 0 && years > 0) || (days <= 0 && years <= 0) {
		return ErrInvalidRetentionPeriod
	}
	return nil
}
