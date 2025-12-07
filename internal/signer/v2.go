package signer

import (
	"net/http"
	"time"
)

// V2Signer AWS Signature Version 2 签名器
type V2Signer struct{}

// Sign 使用 V2 算法签名请求
func (s *V2Signer) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	// TODO: 实现 V2 签名逻辑
	return req
}

// Presign 使用 V2 算法预签名请求
func (s *V2Signer) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	// TODO: 实现 V2 预签名逻辑
	return req
}
