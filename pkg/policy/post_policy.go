// Package policy provides S3 POST policy helpers.
package policy

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

const expirationDateFormat = "2006-01-02T15:04:05.000Z"

// Condition represents a single policy condition.
type Condition struct {
	MatchType string
	Condition string
	Value     string
}

// ContentLengthRange represents the allowed content length range.
type ContentLengthRange struct {
	Min int64
	Max int64
}

// PostPolicy describes the policy used for browser uploads.
type PostPolicy struct {
	expiration         time.Time
	conditions         []Condition
	contentLengthRange *ContentLengthRange
	formData           map[string]string
}

// NewPostPolicy creates a new PostPolicy instance.
func NewPostPolicy() *PostPolicy {
	return &PostPolicy{
		conditions: make([]Condition, 0),
		formData:   make(map[string]string),
	}
}

// Expiration returns the policy expiration time.
func (p *PostPolicy) Expiration() time.Time {
	if p == nil {
		return time.Time{}
	}
	return p.expiration
}

// FormData returns a copy of the form data map.
func (p *PostPolicy) FormData() map[string]string {
	if p == nil {
		return nil
	}
	out := make(map[string]string, len(p.formData))
	for k, v := range p.formData {
		out[k] = v
	}
	return out
}

// SetExpires sets the policy expiration time.
func (p *PostPolicy) SetExpires(t time.Time) error {
	if p == nil {
		return errors.New("post policy is nil")
	}
	if t.IsZero() {
		return errors.New("expiration time must be specified")
	}
	p.expiration = t
	return nil
}

// SetBucket sets the target bucket condition.
func (p *PostPolicy) SetBucket(bucketName string) error {
	if strings.TrimSpace(bucketName) == "" {
		return errors.New("bucket name is empty")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "bucket",
		Value:     bucketName,
	}); err != nil {
		return err
	}
	p.formData["bucket"] = bucketName
	return nil
}

// SetKey sets the exact object key condition.
func (p *PostPolicy) SetKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("object key is empty")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "key",
		Value:     key,
	}); err != nil {
		return err
	}
	p.formData["key"] = key
	return nil
}

// SetKeyStartsWith sets the object key prefix condition.
func (p *PostPolicy) SetKeyStartsWith(prefix string) error {
	if err := p.AddCondition(Condition{
		MatchType: "starts-with",
		Condition: "key",
		Value:     prefix,
	}); err != nil {
		return err
	}
	p.formData["key"] = prefix
	return nil
}

// SetContentType sets the content-type condition.
func (p *PostPolicy) SetContentType(contentType string) error {
	if strings.TrimSpace(contentType) == "" {
		return errors.New("content type is empty")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "Content-Type",
		Value:     contentType,
	}); err != nil {
		return err
	}
	p.formData["Content-Type"] = contentType
	return nil
}

// SetContentTypeStartsWith sets the content-type prefix condition.
func (p *PostPolicy) SetContentTypeStartsWith(prefix string) error {
	if err := p.AddCondition(Condition{
		MatchType: "starts-with",
		Condition: "Content-Type",
		Value:     prefix,
	}); err != nil {
		return err
	}
	p.formData["Content-Type"] = prefix
	return nil
}

// SetContentDisposition sets the content-disposition condition.
func (p *PostPolicy) SetContentDisposition(disposition string) error {
	if strings.TrimSpace(disposition) == "" {
		return errors.New("content disposition is empty")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "Content-Disposition",
		Value:     disposition,
	}); err != nil {
		return err
	}
	p.formData["Content-Disposition"] = disposition
	return nil
}

// SetContentEncoding sets the content-encoding condition.
func (p *PostPolicy) SetContentEncoding(encoding string) error {
	if strings.TrimSpace(encoding) == "" {
		return errors.New("content encoding is empty")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "Content-Encoding",
		Value:     encoding,
	}); err != nil {
		return err
	}
	p.formData["Content-Encoding"] = encoding
	return nil
}

// SetContentLengthRange sets the allowed content length range.
func (p *PostPolicy) SetContentLengthRange(minLen, maxLen int64) error {
	if minLen > maxLen {
		return errors.New("minimum limit is larger than maximum limit")
	}
	if minLen < 0 {
		return errors.New("minimum limit cannot be negative")
	}
	if maxLen <= 0 {
		return errors.New("maximum limit must be positive")
	}
	p.contentLengthRange = &ContentLengthRange{Min: minLen, Max: maxLen}
	return nil
}

// SetSuccessActionRedirect sets the success redirect URL condition.
func (p *PostPolicy) SetSuccessActionRedirect(redirect string) error {
	if strings.TrimSpace(redirect) == "" {
		return errors.New("redirect is empty")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "success_action_redirect",
		Value:     redirect,
	}); err != nil {
		return err
	}
	p.formData["success_action_redirect"] = redirect
	return nil
}

// SetSuccessStatusAction sets the success status condition.
func (p *PostPolicy) SetSuccessStatusAction(status string) error {
	if strings.TrimSpace(status) == "" {
		return errors.New("success action status is empty")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "success_action_status",
		Value:     status,
	}); err != nil {
		return err
	}
	p.formData["success_action_status"] = status
	return nil
}

// SetUserMetadata sets user metadata condition.
func (p *PostPolicy) SetUserMetadata(key, value string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("metadata key is empty")
	}
	if strings.TrimSpace(value) == "" {
		return errors.New("metadata value is empty")
	}
	headerName := "x-amz-meta-" + key
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: headerName,
		Value:     value,
	}); err != nil {
		return err
	}
	p.formData[headerName] = value
	return nil
}

// SetUserMetadataStartsWith sets user metadata prefix condition.
func (p *PostPolicy) SetUserMetadataStartsWith(key, value string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("metadata key is empty")
	}
	headerName := "x-amz-meta-" + key
	if err := p.AddCondition(Condition{
		MatchType: "starts-with",
		Condition: headerName,
		Value:     value,
	}); err != nil {
		return err
	}
	p.formData[headerName] = value
	return nil
}

// SetUserData sets custom x-amz- prefixed data condition.
func (p *PostPolicy) SetUserData(key, value string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("user data key is empty")
	}
	if strings.TrimSpace(value) == "" {
		return errors.New("user data value is empty")
	}
	headerName := "x-amz-" + key
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: headerName,
		Value:     value,
	}); err != nil {
		return err
	}
	p.formData[headerName] = value
	return nil
}

// SetTagging sets the tagging condition for the post policy.
func (p *PostPolicy) SetTagging(taggingXML string) error {
	if strings.TrimSpace(taggingXML) == "" {
		return errors.New("tagging XML is empty")
	}
	var parsed taggingConfig
	if err := xml.Unmarshal([]byte(taggingXML), &parsed); err != nil {
		return errors.New("malformed tagging XML")
	}
	if err := p.AddCondition(Condition{
		MatchType: "eq",
		Condition: "tagging",
		Value:     taggingXML,
	}); err != nil {
		return err
	}
	p.formData["tagging"] = taggingXML
	return nil
}

// SetEncryption adds SSE headers to the form data.
func (p *PostPolicy) SetEncryption(enc sse.Encrypter) {
	if enc == nil || p == nil {
		return
	}
	headers := http.Header{}
	enc.ApplyHeaders(headers)
	for k, v := range headers {
		if len(v) > 0 {
			p.formData[k] = v[0]
		}
	}
}

// SetCondition sets a condition for x-amz-credential/date/algorithm.
func (p *PostPolicy) SetCondition(matchType, condition, value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("condition value is empty")
	}
	normalized := strings.TrimSpace(condition)
	lower := strings.ToLower(normalized)
	switch lower {
	case "x-amz-credential", "x-amz-date", "x-amz-algorithm":
		if err := p.AddCondition(Condition{
			MatchType: matchType,
			Condition: lower,
			Value:     value,
		}); err != nil {
			return err
		}
		p.formData[lower] = value
		return nil
	default:
		return errors.New("invalid condition in policy")
	}
}

// AddCondition adds a policy condition.
func (p *PostPolicy) AddCondition(condition Condition) error {
	if p == nil {
		return errors.New("post policy is nil")
	}
	if strings.TrimSpace(condition.MatchType) == "" || strings.TrimSpace(condition.Condition) == "" {
		return errors.New("policy fields are empty")
	}
	if condition.MatchType != "starts-with" && strings.TrimSpace(condition.Value) == "" {
		return errors.New("policy value is empty")
	}
	p.conditions = append(p.conditions, condition)
	return nil
}

// Base64 returns the Base64-encoded policy document.
func (p *PostPolicy) Base64() (string, error) {
	policyJSON, err := p.marshalJSON()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(policyJSON), nil
}

// String returns the JSON policy as a string when available.
func (p *PostPolicy) String() string {
	if p == nil {
		return ""
	}
	raw, err := p.marshalJSON()
	if err != nil {
		return ""
	}
	return string(raw)
}

func (p *PostPolicy) marshalJSON() ([]byte, error) {
	if p == nil {
		return nil, errors.New("post policy is nil")
	}
	if p.expiration.IsZero() {
		return nil, errors.New("expiration time must be specified")
	}

	doc := policyDocument{
		Expiration: p.expiration.UTC().Format(expirationDateFormat),
	}
	if len(p.conditions) > 0 || p.contentLengthRange != nil {
		doc.Conditions = make([]interface{}, 0, len(p.conditions)+1)
		for _, cond := range p.conditions {
			doc.Conditions = append(doc.Conditions, []interface{}{
				cond.MatchType,
				normalizeCondition(cond.Condition),
				cond.Value,
			})
		}
		if p.contentLengthRange != nil {
			doc.Conditions = append(doc.Conditions, []interface{}{
				"content-length-range",
				p.contentLengthRange.Min,
				p.contentLengthRange.Max,
			})
		}
	}

	return json.Marshal(doc)
}

func normalizeCondition(condition string) string {
	if strings.HasPrefix(condition, "$") {
		return condition
	}
	return "$" + condition
}

type policyDocument struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions,omitempty"`
}

type taggingConfig struct {
	XMLName xml.Name     `xml:"Tagging"`
	TagSet  []taggingTag `xml:"TagSet>Tag"`
}

type taggingTag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}
