// Package notification provides bucket notification event types.
package notification

// Identity represents the user identity in notification records.
type Identity struct {
	PrincipalID string `json:"principalId"`
}

// BucketMeta holds bucket metadata from an event.
type BucketMeta struct {
	Name          string   `json:"name"`
	OwnerIdentity Identity `json:"ownerIdentity"`
	ARN           string   `json:"arn"`
}

// ObjectMeta holds object metadata from an event.
type ObjectMeta struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size,omitempty"`
	ETag         string            `json:"eTag,omitempty"`
	ContentType  string            `json:"contentType,omitempty"`
	UserMetadata map[string]string `json:"userMetadata,omitempty"`
	VersionID    string            `json:"versionId,omitempty"`
	Sequencer    string            `json:"sequencer"`
}

// EventMeta holds S3 metadata from an event.
type EventMeta struct {
	SchemaVersion   string     `json:"s3SchemaVersion"`
	ConfigurationID string     `json:"configurationId"`
	Bucket          BucketMeta `json:"bucket"`
	Object          ObjectMeta `json:"object"`
}

// SourceInfo represents information about the client that triggered the event.
type SourceInfo struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	UserAgent string `json:"userAgent"`
}

// Event represents an S3 bucket notification event.
type Event struct {
	EventVersion      string            `json:"eventVersion"`
	EventSource       string            `json:"eventSource"`
	AwsRegion         string            `json:"awsRegion"`
	EventTime         string            `json:"eventTime"`
	EventName         string            `json:"eventName"`
	UserIdentity      Identity          `json:"userIdentity"`
	RequestParameters map[string]string `json:"requestParameters"`
	ResponseElements  map[string]string `json:"responseElements"`
	S3                EventMeta         `json:"s3"`
	Source            SourceInfo        `json:"source"`
}

// Info represents notification events and any errors encountered.
type Info struct {
	Records []Event `json:"Records"`
	Err     error   `json:"-"`
}
