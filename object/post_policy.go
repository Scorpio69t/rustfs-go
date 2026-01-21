// Package object object/post_policy.go
package object

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/policy"
	"github.com/Scorpio69t/rustfs-go/pkg/signer"
)

const (
	postPolicyAlgorithm = "AWS4-HMAC-SHA256"
	postPolicyDateFmt   = "20060102T150405Z"
)

// PresignedPostPolicy generates a POST policy and form data for browser uploads.
func (s *objectService) PresignedPostPolicy(ctx context.Context, p *policy.PostPolicy) (*url.URL, map[string]string, error) {
	if p == nil {
		return nil, nil, errors.New("post policy is nil")
	}
	if p.Expiration().IsZero() {
		return nil, nil, errors.New("expiration time must be specified")
	}

	formData := p.FormData()
	bucketName, ok := formData["bucket"]
	if !ok || strings.TrimSpace(bucketName) == "" {
		return nil, nil, errors.New("bucket name must be specified")
	}
	if key, ok := formData["key"]; !ok || strings.TrimSpace(key) == "" {
		return nil, nil, errors.New("object key must be specified")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	location, err := s.executor.ResolveBucketLocation(ctx, bucketName)
	if err != nil {
		return nil, nil, err
	}

	targetURL, err := s.executor.TargetURL(ctx, bucketName, "", nil)
	if err != nil {
		return nil, nil, err
	}

	credValues, err := s.executor.GetCredentials(ctx)
	if err != nil {
		return nil, nil, err
	}
	if credValues.SignerType.IsAnonymous() || credValues.AccessKeyID == "" || credValues.SecretAccessKey == "" {
		return nil, nil, errors.New("presigned operations are not supported for anonymous credentials")
	}

	t := time.Now().UTC()

	if credValues.SignerType.IsV2() {
		policyBase64, err := p.Base64()
		if err != nil {
			return nil, nil, err
		}
		formData["policy"] = policyBase64
		if strings.Contains(targetURL.Host, ".storage.googleapis.com") {
			formData["GoogleAccessId"] = credValues.AccessKeyID
		} else {
			formData["AWSAccessKeyId"] = credValues.AccessKeyID
		}
		formData["signature"] = signer.PostPresignSignatureV2(policyBase64, credValues.SecretAccessKey)
		return targetURL, formData, nil
	}

	date := t.Format(postPolicyDateFmt)
	credential := signer.GetCredential(credValues.AccessKeyID, location, t)

	if err := p.AddCondition(policy.Condition{
		MatchType: "eq",
		Condition: "x-amz-date",
		Value:     date,
	}); err != nil {
		return nil, nil, err
	}
	if err := p.AddCondition(policy.Condition{
		MatchType: "eq",
		Condition: "x-amz-algorithm",
		Value:     postPolicyAlgorithm,
	}); err != nil {
		return nil, nil, err
	}
	if err := p.AddCondition(policy.Condition{
		MatchType: "eq",
		Condition: "x-amz-credential",
		Value:     credential,
	}); err != nil {
		return nil, nil, err
	}
	if credValues.SessionToken != "" {
		if err := p.AddCondition(policy.Condition{
			MatchType: "eq",
			Condition: "x-amz-security-token",
			Value:     credValues.SessionToken,
		}); err != nil {
			return nil, nil, err
		}
	}

	policyBase64, err := p.Base64()
	if err != nil {
		return nil, nil, err
	}

	formData["policy"] = policyBase64
	formData["x-amz-algorithm"] = postPolicyAlgorithm
	formData["x-amz-credential"] = credential
	formData["x-amz-date"] = date
	if credValues.SessionToken != "" {
		formData["x-amz-security-token"] = credValues.SessionToken
	}
	formData["x-amz-signature"] = signer.PostPresignSignatureV4(policyBase64, t, credValues.SecretAccessKey, location)

	return targetURL, formData, nil
}
