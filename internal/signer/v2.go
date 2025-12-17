package signer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Signature and API related constants for V2
const (
	signV2Algorithm = "AWS"
)

// v2ResourceList V2 签名中需要包含的查询参数白名单
// 按字母顺序排序
var v2ResourceList = []string{
	"acl",
	"cors",
	"delete",
	"encryption",
	"lifecycle",
	"location",
	"logging",
	"notification",
	"partNumber",
	"policy",
	"replication",
	"requestPayment",
	"response-cache-control",
	"response-content-disposition",
	"response-content-encoding",
	"response-content-language",
	"response-content-type",
	"response-expires",
	"tagging",
	"torrent",
	"uploadId",
	"uploads",
	"versionId",
	"versioning",
	"versions",
	"website",
}

// V2Signer AWS Signature Version 2 签名器
type V2Signer struct {
	virtualHost bool
}

// Sign 使用 V2 算法签名请求
// Authorization = "AWS" + " " + AWSAccessKeyId + ":" + Signature
// Signature = Base64( HMAC-SHA1( YourSecretAccessKeyID, UTF-8-Encoding-Of( StringToSign ) ) )
func (s *V2Signer) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	// 匿名凭证不需要签名
	if accessKey == "" || secretKey == "" {
		return req
	}

	// 初始化时间
	d := time.Now().UTC()

	// 添加 Date 头（如果不存在）
	if date := req.Header.Get("Date"); date == "" {
		req.Header.Set("Date", d.Format(http.TimeFormat))
	}

	// 计算 HMAC
	stringToSign := s.stringToSignV2(req)
	hm := hmac.New(sha1.New, []byte(secretKey))
	hm.Write([]byte(stringToSign))

	// 准备 Authorization 头
	authHeader := new(bytes.Buffer)
	fmt.Fprintf(authHeader, "%s %s:", signV2Algorithm, accessKey)
	encoder := base64.NewEncoder(base64.StdEncoding, authHeader)
	encoder.Write(hm.Sum(nil))
	encoder.Close()

	// 设置 Authorization 头
	req.Header.Set("Authorization", authHeader.String())

	return req
}

// Presign 使用 V2 算法预签名请求
// https://${S3_BUCKET}.s3.amazonaws.com/${S3_OBJECT}?AWSAccessKeyId=${S3_ACCESS_KEY}&Expires=${TIMESTAMP}&Signature=${SIGNATURE}
func (s *V2Signer) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	// 匿名凭证不需要签名
	if accessKey == "" || secretKey == "" {
		return req
	}

	d := time.Now().UTC()
	// 计算过期时间（Unix 时间戳）
	epochExpires := d.Unix() + int64(expires.Seconds())

	// 添加 Expires 头（如果不存在）
	if expiresStr := req.Header.Get("Expires"); expiresStr == "" {
		req.Header.Set("Expires", strconv.FormatInt(epochExpires, 10))
	}

	// 获取预签名字符串
	stringToSign := s.preStringToSignV2(req)
	hm := hmac.New(sha1.New, []byte(secretKey))
	hm.Write([]byte(stringToSign))

	// 计算签名
	signature := base64.StdEncoding.EncodeToString(hm.Sum(nil))

	query := req.URL.Query()
	// 处理 Google Cloud Storage 特殊情况
	if strings.Contains(getHostAddr(req), ".storage.googleapis.com") {
		query.Set("GoogleAccessId", accessKey)
	} else {
		query.Set("AWSAccessKeyId", accessKey)
	}

	// 填充 Expires 查询参数
	query.Set("Expires", strconv.FormatInt(epochExpires, 10))

	// 编码查询参数并保存
	req.URL.RawQuery = queryEncode(query)

	// 最后保存签名
	req.URL.RawQuery += "&Signature=" + encodePath(signature)

	return req
}

// stringToSignV2 生成待签名字符串
// StringToSign = HTTP-Verb + "\n" +
//
//	Content-Md5 + "\n" +
//	Content-Type + "\n" +
//	Date + "\n" +
//	CanonicalizedProtocolHeaders +
//	CanonicalizedResource
func (s *V2Signer) stringToSignV2(req *http.Request) string {
	buf := new(bytes.Buffer)
	// 写入标准头部
	s.writeSignV2Headers(buf, req)
	// 写入规范化协议头部（如果有）
	s.writeCanonicalizedHeaders(buf, req)
	// 写入规范化查询资源（如果有）
	s.writeCanonicalizedResource(buf, req)
	return buf.String()
}

// preStringToSignV2 生成预签名待签名字符串
func (s *V2Signer) preStringToSignV2(req *http.Request) string {
	buf := new(bytes.Buffer)
	// 写入标准头部
	s.writePreSignV2Headers(buf, req)
	// 写入规范化协议头部（如果有）
	s.writeCanonicalizedHeaders(buf, req)
	// 写入规范化查询资源（如果有）
	s.writeCanonicalizedResource(buf, req)
	return buf.String()
}

// writeSignV2Headers 写入 V2 签名所需的标准头部
func (s *V2Signer) writeSignV2Headers(buf *bytes.Buffer, req *http.Request) {
	buf.WriteString(req.Method + "\n")
	buf.WriteString(req.Header.Get("Content-Md5") + "\n")
	buf.WriteString(req.Header.Get("Content-Type") + "\n")
	buf.WriteString(req.Header.Get("Date") + "\n")
}

// writePreSignV2Headers 写入 V2 预签名所需的标准头部
func (s *V2Signer) writePreSignV2Headers(buf *bytes.Buffer, req *http.Request) {
	buf.WriteString(req.Method + "\n")
	buf.WriteString(req.Header.Get("Content-Md5") + "\n")
	buf.WriteString(req.Header.Get("Content-Type") + "\n")
	buf.WriteString(req.Header.Get("Expires") + "\n")
}

// writeCanonicalizedHeaders 写入规范化头部
func (s *V2Signer) writeCanonicalizedHeaders(buf *bytes.Buffer, req *http.Request) {
	var protoHeaders []string
	vals := make(map[string][]string)
	for k, vv := range req.Header {
		// 所有 AMZ 头部应该小写
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-amz") {
			protoHeaders = append(protoHeaders, lk)
			vals[lk] = vv
		}
	}
	sort.Strings(protoHeaders)
	for _, k := range protoHeaders {
		buf.WriteString(k)
		buf.WriteByte(':')
		for idx, v := range vals[k] {
			if idx > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(v)
		}
		buf.WriteByte('\n')
	}
}

// writeCanonicalizedResource 写入规范化资源
// CanonicalizedResource = [ "/" + Bucket ] +
//
//	<HTTP-Request-URI, from the protocol name up to the query string> +
//	[ subresource, if present ]
func (s *V2Signer) writeCanonicalizedResource(buf *bytes.Buffer, req *http.Request) {
	// 获取编码的路径
	path := s.encodeURL2Path(req)
	buf.WriteString(path)

	// 处理查询参数中的子资源
	query := req.URL.Query()
	var resourceList []string
	for _, resource := range v2ResourceList {
		if query.Get(resource) != "" {
			resourceList = append(resourceList, resource+"="+query.Get(resource))
		} else if _, ok := query[resource]; ok {
			resourceList = append(resourceList, resource)
		}
	}

	if len(resourceList) > 0 {
		buf.WriteByte('?')
		buf.WriteString(strings.Join(resourceList, "&"))
	}
}

// encodeURL2Path 编码 URL 路径
func (s *V2Signer) encodeURL2Path(req *http.Request) string {
	if s.virtualHost {
		reqHost := getHostAddr(req)
		dotPos := strings.Index(reqHost, ".")
		if dotPos > -1 {
			bucketName := reqHost[:dotPos]
			path := "/" + bucketName
			path += req.URL.Path
			path = encodePath(path)
			return path
		}
	}
	return encodePath(req.URL.Path)
}
