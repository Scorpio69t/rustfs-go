package rustfs

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3utils"
)

// Tag - object tag
type Tag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// Tagging - object tagging structure
type Tagging struct {
	XMLName xml.Name `xml:"Tagging"`
	TagSet  TagSet   `xml:"TagSet"`
}

// TagSet - tag set
type TagSet struct {
	Tag []Tag `xml:"Tag"`
}

// SetObjectTagging - set object tags
func (c *Client) SetObjectTagging(ctx context.Context, bucketName, objectName string, tags map[string]string) error {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return err
	}

	// Build query values
	queryValues := make(url.Values)
	queryValues.Set("tagging", "")

	// Build tagging XML
	tagSet := TagSet{}
	for k, v := range tags {
		tagSet.Tag = append(tagSet.Tag, Tag{
			Key:   k,
			Value: v,
		})
	}
	tagging := Tagging{
		TagSet: tagSet,
	}

	body, err := xml.Marshal(tagging)
	if err != nil {
		return err
	}

	// Build metadata
	metadata := requestMetadata{
		bucketName:    bucketName,
		objectName:    objectName,
		contentBody:   bytes.NewReader(body),
		contentLength: int64(len(body)),
		queryValues:   queryValues,
		customHeader:  make(http.Header),
	}
	metadata.customHeader.Set("Content-Type", "application/xml")

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodPut, metadata)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	return nil
}

// GetObjectTagging - get object tags
func (c *Client) GetObjectTagging(ctx context.Context, bucketName, objectName string) (map[string]string, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return nil, err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return nil, err
	}

	// Build query values
	queryValues := make(url.Values)
	queryValues.Set("tagging", "")

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		objectName:   objectName,
		queryValues:  queryValues,
		customHeader: make(http.Header),
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodGet, metadata)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	// Parse response
	var tagging Tagging
	if err := parseResponse(resp.Body, &tagging); err != nil {
		return nil, err
	}

	// Convert to map
	tags := make(map[string]string)
	for _, tag := range tagging.TagSet.Tag {
		tags[tag.Key] = tag.Value
	}

	return tags, nil
}

// RemoveObjectTagging - remove object tags
func (c *Client) RemoveObjectTagging(ctx context.Context, bucketName, objectName string) error {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return err
	}

	// Build query values
	queryValues := make(url.Values)
	queryValues.Set("tagging", "")

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		objectName:   objectName,
		queryValues:  queryValues,
		customHeader: make(http.Header),
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodDelete, metadata)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	return nil
}
