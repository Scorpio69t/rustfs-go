/*
 * RustFS Go SDK
 * Copyright 2025 RustFS Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package credentials

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AssumeRoleResponse contains the result of successful AssumeRole request.
type AssumeRoleResponse struct {
	XMLName xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleResponse" json:"-"`

	Result           AssumeRoleResult `xml:"AssumeRoleResult"`
	ResponseMetadata struct {
		RequestID string `xml:"RequestId,omitempty"`
	} `xml:"ResponseMetadata,omitempty"`
}

// AssumeRoleResult - Contains the response to a successful AssumeRole
// request, including temporary credentials that can be used to make
// RustFS API requests.
type AssumeRoleResult struct {
	// The identifiers for the temporary security credentials that the operation
	// returns.
	AssumedRoleUser AssumedRoleUser `xml:",omitempty"`

	// The temporary security credentials, which include an access key ID, a secret
	// access key, and a security (or session) token.
	//
	// Note: The size of the security token that STS APIs return is not fixed. We
	// strongly recommend that you make no assumptions about the maximum size. As
	// of this writing, the typical size is less than 4096 bytes, but that can vary.
	// Also, future updates to AWS might require larger sizes.
	Credentials struct {
		AccessKey    string    `xml:"AccessKeyId" json:"accessKey,omitempty"`
		SecretKey    string    `xml:"SecretAccessKey" json:"secretKey,omitempty"`
		Expiration   time.Time `xml:"Expiration" json:"expiration,omitempty"`
		SessionToken string    `xml:"SessionToken" json:"sessionToken,omitempty"`
	} `xml:",omitempty"`

	// A percentage value that indicates the size of the policy in packed form.
	// The service rejects any policy with a packed size greater than 100 percent,
	// which means the policy exceeded the allowed space.
	PackedPolicySize int `xml:",omitempty"`
}

// A STSAssumeRole retrieves credentials from RustFS service, and keeps track if
// those credentials are expired.
type STSAssumeRole struct {
	Expiry

	// Optional http Client to use when connecting to RustFS STS service
	// (overrides default client in CredContext)
	Client *http.Client

	// STS endpoint to fetch STS credentials.
	STSEndpoint string

	// various options for this request.
	Options STSAssumeRoleOptions
}

// STSAssumeRoleOptions collection of various input options
// to obtain AssumeRole credentials.
type STSAssumeRoleOptions struct {
	// Mandatory inputs.
	AccessKey string
	SecretKey string

	SessionToken string // Optional if the first request is made with temporary credentials.
	Policy       string // Optional to assign a policy to the assumed role

	Location        string // Optional commonly needed with AWS STS.
	DurationSeconds int    // Optional defaults to 1 hour.

	// Optional only valid if using with AWS STS
	RoleARN         string
	RoleSessionName string
	ExternalID      string

	TokenRevokeType string // Optional, used for token revokation (RustFS extension)
}

// NewSTSAssumeRole returns a pointer to a new
// Credentials object wrapping the STSAssumeRole.
func NewSTSAssumeRole(stsEndpoint string, opts STSAssumeRoleOptions) (*Credentials, error) {
	if opts.AccessKey == "" || opts.SecretKey == "" {
		return nil, errors.New("AssumeRole credentials access/secretkey is mandatory")
	}
	return New(&STSAssumeRole{
		STSEndpoint: stsEndpoint,
		Options:     opts,
	}), nil
}

const defaultDurationSeconds = 3600

// closeResponse close non nil response with any response Body.
// convenient wrapper to drain any remaining data on response body.
//
// Subsequently this allows golang http RoundTripper
// to re-use the same connection for future requests.
func closeResponse(resp *http.Response) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if resp != nil && resp.Body != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		//  - http://stackoverflow.com/a/17961593/4465767
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func getAssumeRoleCredentials(clnt *http.Client, endpoint string, opts STSAssumeRoleOptions) (AssumeRoleResponse, error) {
	v := url.Values{}
	v.Set("Action", "AssumeRole")
	v.Set("Version", STSVersion)
	if opts.RoleARN != "" {
		v.Set("RoleArn", opts.RoleARN)
	}
	if opts.RoleSessionName != "" {
		v.Set("RoleSessionName", opts.RoleSessionName)
	}
	if opts.DurationSeconds > defaultDurationSeconds {
		v.Set("DurationSeconds", strconv.Itoa(opts.DurationSeconds))
	} else {
		v.Set("DurationSeconds", strconv.Itoa(defaultDurationSeconds))
	}
	if opts.Policy != "" {
		v.Set("Policy", opts.Policy)
	}
	if opts.ExternalID != "" {
		v.Set("ExternalId", opts.ExternalID)
	}
	if opts.TokenRevokeType != "" {
		v.Set("TokenRevokeType", opts.TokenRevokeType)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return AssumeRoleResponse{}, err
	}
	u.Path = "/"

	postBody := strings.NewReader(v.Encode())
	hash := sha256.New()
	if _, err = io.Copy(hash, postBody); err != nil {
		return AssumeRoleResponse{}, err
	}
	postBody.Seek(0, 0)

	req, err := http.NewRequest(http.MethodPost, u.String(), postBody)
	if err != nil {
		return AssumeRoleResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(hash.Sum(nil)))
	if opts.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", opts.SessionToken)
	}
	req = signV4STS(*req, opts.AccessKey, opts.SecretKey, opts.Location)

	resp, err := clnt.Do(req)
	if err != nil {
		return AssumeRoleResponse{}, err
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return AssumeRoleResponse{}, err
		}
		_, err = xmlDecodeAndBody(bytes.NewReader(buf), &errResp)
		if err != nil {
			var s3Err Error
			if _, err = xmlDecodeAndBody(bytes.NewReader(buf), &s3Err); err != nil {
				return AssumeRoleResponse{}, err
			}
			errResp.RequestID = s3Err.RequestID
			errResp.STSError.Code = s3Err.Code
			errResp.STSError.Message = s3Err.Message
		}
		return AssumeRoleResponse{}, errResp
	}

	a := AssumeRoleResponse{}
	if _, err = xmlDecodeAndBody(resp.Body, &a); err != nil {
		return AssumeRoleResponse{}, err
	}
	return a, nil
}

// RetrieveWithCredContext retrieves credentials from the RustFS service.
// Error will be returned if the request fails, optional cred context.
func (m *STSAssumeRole) RetrieveWithCredContext(cc *CredContext) (Value, error) {
	if cc == nil {
		cc = defaultCredContext
	}

	client := m.Client
	if client == nil {
		client = cc.Client
	}
	if client == nil {
		client = defaultCredContext.Client
	}

	stsEndpoint := m.STSEndpoint
	if stsEndpoint == "" {
		stsEndpoint = cc.Endpoint
	}
	if stsEndpoint == "" {
		return Value{}, errors.New("STS endpoint unknown")
	}

	a, err := getAssumeRoleCredentials(client, stsEndpoint, m.Options)
	if err != nil {
		return Value{}, err
	}

	// Expiry window is set to 10secs.
	m.SetExpiration(a.Result.Credentials.Expiration, DefaultExpiryWindow)

	return Value{
		AccessKeyID:     a.Result.Credentials.AccessKey,
		SecretAccessKey: a.Result.Credentials.SecretKey,
		SessionToken:    a.Result.Credentials.SessionToken,
		Expiration:      a.Result.Credentials.Expiration,
		SignerType:      SignatureV4,
	}, nil
}

// Retrieve retrieves credentials from the RustFS service.
// Error will be returned if the request fails.
func (m *STSAssumeRole) Retrieve() (Value, error) {
	return m.RetrieveWithCredContext(nil)
}

// 以下是 STS V4 签名的本地实现，避免循环依赖

const (
	signV4Algorithm   = "AWS4-HMAC-SHA256"
	iso8601DateFormat = "20060102T150405Z"
	yyyymmdd          = "20060102"
)

// signV4STS 为 STS 请求签名
func signV4STS(req http.Request, accessKeyID, secretAccessKey, location string) *http.Request {
	if accessKeyID == "" || secretAccessKey == "" {
		return &req
	}

	t := time.Now().UTC()
	req.Header.Set("X-Amz-Date", t.Format(iso8601DateFormat))

	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}

	region := location
	if region == "" {
		region = "us-east-1"
	}

	scope := buildCredentialScope(t, region)
	canonicalRequest := buildCanonicalRequest(&req)
	stringToSign := buildStringToSign(canonicalRequest, t, scope)
	signingKey := deriveSigningKey(secretAccessKey, t, region)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	signedHeaders := getSignedHeaders(req.Header)
	authorization := signV4Algorithm + " " +
		"Credential=" + accessKeyID + "/" + scope + ", " +
		"SignedHeaders=" + signedHeaders + ", " +
		"Signature=" + signature

	req.Header.Set("Authorization", authorization)

	return &req
}

func buildCredentialScope(t time.Time, region string) string {
	return strings.Join([]string{
		t.Format(yyyymmdd),
		region,
		"sts",
		"aws4_request",
	}, "/")
}

func buildCanonicalRequest(req *http.Request) string {
	var canonicalHeaders, signedHeaders strings.Builder

	headers := make([]string, 0, len(req.Header))
	for k := range req.Header {
		headers = append(headers, strings.ToLower(k))
	}
	sort.Strings(headers)

	for i, k := range headers {
		if i > 0 {
			signedHeaders.WriteString(";")
		}
		signedHeaders.WriteString(k)

		canonicalHeaders.WriteString(k)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(k)))
		canonicalHeaders.WriteString("\n")
	}

	encodedPath := req.URL.EscapedPath()
	if encodedPath == "" {
		encodedPath = "/"
	}

	canonicalQuery := req.URL.Query().Encode()

	return strings.Join([]string{
		req.Method,
		encodedPath,
		canonicalQuery,
		canonicalHeaders.String(),
		signedHeaders.String(),
		req.Header.Get("X-Amz-Content-Sha256"),
	}, "\n")
}

func buildStringToSign(canonicalRequest string, t time.Time, scope string) string {
	hash := sha256.Sum256([]byte(canonicalRequest))
	return strings.Join([]string{
		signV4Algorithm,
		t.Format(iso8601DateFormat),
		scope,
		hex.EncodeToString(hash[:]),
	}, "\n")
}

func deriveSigningKey(secretAccessKey string, t time.Time, region string) []byte {
	kSecret := []byte("AWS4" + secretAccessKey)
	kDate := hmacSHA256(kSecret, []byte(t.Format(yyyymmdd)))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte("sts"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignedHeaders(header http.Header) string {
	headers := make([]string, 0, len(header))
	for k := range header {
		headers = append(headers, strings.ToLower(k))
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}
