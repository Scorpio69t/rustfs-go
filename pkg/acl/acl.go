// Package acl provides ACL types and XML helpers.
package acl

import (
	"encoding/xml"
	"fmt"
	"io"
)

const defaultXMLNS = "http://s3.amazonaws.com/doc/2006-03-01/"

// ACL represents an access control policy.
type ACL struct {
	XMLNS   string   `xml:"xmlns,attr,omitempty"`
	XMLName xml.Name `xml:"AccessControlPolicy"`
	Owner   Owner    `xml:"Owner,omitempty"`
	Grants  []Grant  `xml:"AccessControlList>Grant,omitempty"`
	Canned  CannedACL
}

// Owner identifies the owner of the resource.
type Owner struct {
	ID          string `xml:"ID,omitempty"`
	DisplayName string `xml:"DisplayName,omitempty"`
}

// Grant defines a permission granted to a grantee.
type Grant struct {
	Grantee    Grantee    `xml:"Grantee"`
	Permission Permission `xml:"Permission"`
}

// Grantee identifies the entity being granted access.
type Grantee struct {
	XMLName      xml.Name `xml:"Grantee"`
	Type         string   `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr,omitempty"`
	ID           string   `xml:"ID,omitempty"`
	DisplayName  string   `xml:"DisplayName,omitempty"`
	EmailAddress string   `xml:"EmailAddress,omitempty"`
	URI          string   `xml:"URI,omitempty"`
}

// Permission represents the access level.
type Permission string

const (
	PermissionFullControl Permission = "FULL_CONTROL"
	PermissionWrite       Permission = "WRITE"
	PermissionWriteACP    Permission = "WRITE_ACP"
	PermissionRead        Permission = "READ"
	PermissionReadACP     Permission = "READ_ACP"
)

// CannedACL is a predefined ACL.
type CannedACL string

const (
	ACLPrivate                CannedACL = "private"
	ACLPublicRead             CannedACL = "public-read"
	ACLPublicReadWrite        CannedACL = "public-read-write"
	ACLAuthenticatedRead      CannedACL = "authenticated-read"
	ACLBucketOwnerRead        CannedACL = "bucket-owner-read"
	ACLBucketOwnerFullControl CannedACL = "bucket-owner-full-control"
)

// IsValid reports whether the canned ACL is supported.
func (c CannedACL) IsValid() bool {
	switch c {
	case ACLPrivate,
		ACLPublicRead,
		ACLPublicReadWrite,
		ACLAuthenticatedRead,
		ACLBucketOwnerRead,
		ACLBucketOwnerFullControl:
		return true
	default:
		return false
	}
}

// Normalize validates and normalizes the ACL.
func (a *ACL) Normalize() error {
	if a.Canned != "" {
		if !a.Canned.IsValid() {
			return ErrInvalidCannedACL
		}
		if len(a.Grants) > 0 || a.Owner != (Owner{}) {
			return ErrMixedACL
		}
		return nil
	}
	if a.XMLNS == "" {
		a.XMLNS = defaultXMLNS
	}
	if a.XMLName.Local == "" {
		a.XMLName = xml.Name{Local: "AccessControlPolicy", Space: defaultXMLNS}
	} else if a.XMLName.Space == "" {
		a.XMLName.Space = defaultXMLNS
	}
	return nil
}

// ToXML marshals the ACL policy to XML.
func (a ACL) ToXML() ([]byte, error) {
	if err := a.Normalize(); err != nil {
		return nil, err
	}
	data, err := xml.Marshal(&a)
	if err != nil {
		return nil, fmt.Errorf("marshal acl xml: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}

// ParseACL parses ACL policy XML from a reader.
func ParseACL(reader io.Reader) (ACL, error) {
	var a ACL
	if err := xml.NewDecoder(reader).Decode(&a); err != nil {
		return ACL{}, fmt.Errorf("decode acl xml: %w", err)
	}
	if a.XMLNS == "" {
		a.XMLNS = defaultXMLNS
	}
	return a, nil
}
