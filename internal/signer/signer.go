// Package signer internal/signer/signer.go
// 提供内部签名器接口和实现
package signer

import (
	"net/http"
	"time"
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
