// Package restore provides restore request types for archived objects.
package restore

import (
	"encoding/xml"

	"github.com/Scorpio69t/rustfs-go/pkg/acl"
	s3select "github.com/Scorpio69t/rustfs-go/pkg/select"
)

const defaultXMLNS = "http://s3.amazonaws.com/doc/2006-03-01/"

// RestoreType represents the restore request type.
type RestoreType string

const (
	// RestoreSelect represents a select restore operation.
	RestoreSelect RestoreType = "SELECT"
)

// TierType represents a retrieval tier.
type TierType string

const (
	// TierStandard is the standard retrieval tier.
	TierStandard TierType = "Standard"
	// TierBulk is the bulk retrieval tier.
	TierBulk TierType = "Bulk"
	// TierExpedited is the expedited retrieval tier.
	TierExpedited TierType = "Expedited"
)

// GlacierJobParameters represents the retrieval tier parameter.
type GlacierJobParameters struct {
	Tier TierType `xml:"Tier,omitempty"`
}

// Encryption describes the server-side encryption for the restored object copy.
type Encryption struct {
	EncryptionType string `xml:"EncryptionType,omitempty"`
	KMSContext     string `xml:"KMSContext,omitempty"`
	KMSKeyID       string `xml:"KMSKeyId,omitempty"`
}

// MetadataEntry represents a metadata entry for restore output.
type MetadataEntry struct {
	Name  string `xml:"Name,omitempty"`
	Value string `xml:"Value,omitempty"`
}

// Tag represents a tag key-value pair.
type Tag struct {
	Key   string `xml:"Key,omitempty"`
	Value string `xml:"Value,omitempty"`
}

// Tagging represents a tagging set.
type Tagging struct {
	TagSet []Tag `xml:"TagSet>Tag,omitempty"`
}

// S3 holds properties of the restored object copy.
type S3 struct {
	AccessControlList *acl.ACL       `xml:"AccessControlList,omitempty"`
	BucketName        string         `xml:"BucketName,omitempty"`
	Prefix            string         `xml:"Prefix,omitempty"`
	CannedACL         *string        `xml:"CannedACL,omitempty"`
	Encryption        *Encryption    `xml:"Encryption,omitempty"`
	StorageClass      *string        `xml:"StorageClass,omitempty"`
	Tagging           *Tagging       `xml:"Tagging,omitempty"`
	UserMetadata      *MetadataEntry `xml:"UserMetadata,omitempty"`
}

// SelectParameters holds the select request parameters.
type SelectParameters struct {
	XMLName             xml.Name                     `xml:"SelectParameters"`
	ExpressionType      s3select.QueryExpressionType `xml:"ExpressionType"`
	Expression          string                       `xml:"Expression"`
	InputSerialization  s3select.InputSerialization  `xml:"InputSerialization"`
	OutputSerialization s3select.OutputSerialization `xml:"OutputSerialization"`
}

// OutputLocation holds properties of the copy of the archived object.
type OutputLocation struct {
	XMLName xml.Name `xml:"OutputLocation"`
	S3      S3       `xml:"S3"`
}

// RestoreRequest describes a restore object request.
type RestoreRequest struct {
	XMLName xml.Name `xml:"RestoreRequest"`
	XMLNS   string   `xml:"xmlns,attr,omitempty"`

	Type                 *RestoreType          `xml:"Type,omitempty"`
	Tier                 *TierType             `xml:"Tier,omitempty"`
	Days                 *int                  `xml:"Days,omitempty"`
	GlacierJobParameters *GlacierJobParameters `xml:"GlacierJobParameters,omitempty"`
	Description          *string               `xml:"Description,omitempty"`
	SelectParameters     *SelectParameters     `xml:"SelectParameters,omitempty"`
	OutputLocation       *OutputLocation       `xml:"OutputLocation,omitempty"`
}

// Normalize ensures required XML defaults.
func (r *RestoreRequest) Normalize() {
	if r.XMLNS == "" {
		r.XMLNS = defaultXMLNS
	}
}

// SetDays sets the days parameter of the restore request.
func (r *RestoreRequest) SetDays(v int) {
	r.Days = &v
}

// SetGlacierJobParameters sets the GlacierJobParameters of the restore request.
func (r *RestoreRequest) SetGlacierJobParameters(v GlacierJobParameters) {
	r.GlacierJobParameters = &v
}

// SetType sets the type of the restore request.
func (r *RestoreRequest) SetType(v RestoreType) {
	r.Type = &v
}

// SetTier sets the retrieval tier of the restore request.
func (r *RestoreRequest) SetTier(v TierType) {
	r.Tier = &v
}

// SetDescription sets the description of the restore request.
func (r *RestoreRequest) SetDescription(v string) {
	r.Description = &v
}

// SetSelectParameters sets SelectParameters of the restore select request.
func (r *RestoreRequest) SetSelectParameters(v SelectParameters) {
	r.SelectParameters = &v
}

// SetOutputLocation sets the properties of the copy of the archived object.
func (r *RestoreRequest) SetOutputLocation(v OutputLocation) {
	r.OutputLocation = &v
}
