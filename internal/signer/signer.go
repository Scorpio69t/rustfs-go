// Package signer internal/signer/signer.go
package signer

import (
	"net/http"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

// Signer 签名器接口
type Signer interface {
	// Sign 签名请求
	Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request

	// Presign 预签名请求
	Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request
}

// SignerType 签名类型
type SignerType int

const (
	SignerV4 SignerType = iota
	SignerV2
	SignerAnonymous
)

// NewSigner 创建签名器
func NewSigner(signerType SignerType) Signer {
	switch signerType {
	case SignerV2:
		return &V2Signer{}
	case SignerAnonymous:
		return &AnonymousSigner{}
	default:
		return &V4Signer{}
	}
}

// SignRequest 签名请求的便捷函数
func SignRequest(req *http.Request, creds credentials.Value, region string) *http.Request {
	signer := NewSigner(getSignerType(creds.SignerType))
	return signer.Sign(req, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, region)
}

func getSignerType(st credentials.SignatureType) SignerType {
	switch st {
	case credentials.SignatureV2:
		return SignerV2
	case credentials.SignatureAnonymous:
		return SignerAnonymous
	default:
		return SignerV4
	}
}

// AnonymousSigner 匿名签名器
type AnonymousSigner struct{}

// Sign 使用匿名方式签名请求
func (s *AnonymousSigner) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	return req
}

// Presign 使用匿名方式预签名请求
func (s *AnonymousSigner) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	return req
}
