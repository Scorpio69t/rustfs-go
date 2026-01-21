package signer

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"time"
)

// GetCredential builds the credential string for SigV4.
func GetCredential(accessKeyID, region string, t time.Time) string {
	signer := &V4Signer{}
	return signer.getCredential(accessKeyID, region, t)
}

// PostPresignSignatureV4 generates a SigV4 signature for POST policies.
func PostPresignSignatureV4(policyBase64 string, t time.Time, secretAccessKey, region string) string {
	signer := &V4Signer{}
	signingKey := signer.deriveSigningKey(secretAccessKey, region, t)
	signature := hmacSHA256(signingKey, []byte(policyBase64))
	return hex.EncodeToString(signature)
}

// PostPresignSignatureV2 generates a SigV2 signature for POST policies.
func PostPresignSignatureV2(policyBase64, secretAccessKey string) string {
	hm := hmac.New(sha1.New, []byte(secretAccessKey))
	hm.Write([]byte(policyBase64))
	return base64.StdEncoding.EncodeToString(hm.Sum(nil))
}
