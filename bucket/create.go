// Package bucket bucket/create.go
package bucket

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Create 创建桶
func (s *bucketService) Create(ctx context.Context, bucketName string, opts ...CreateOption) error {
	// 验证桶名
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	// 应用选项
	options := applyCreateOptions(opts)

	// 如果区域为空，使用默认区域
	if options.Region == "" {
		options.Region = "us-east-1"
	}

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:     bucketName,
		BucketLocation: options.Region,
		CustomHeader:   make(http.Header),
	}

	// 设置对象锁定头
	if options.ObjectLocking {
		meta.CustomHeader.Set("x-amz-bucket-object-lock-enabled", "true")
	}

	// 设置强制创建头（RustFS 扩展）
	if options.ForceCreate {
		meta.CustomHeader.Set("x-rustfs-force-create", "true")
	}

	// 如果区域不是 us-east-1，需要发送 CreateBucketConfiguration
	if options.Region != "us-east-1" && options.Region != "" {
		config := createBucketConfiguration{
			Location: options.Region,
		}

		configBytes, err := xml.Marshal(config)
		if err != nil {
			return err
		}

		meta.ContentBody = bytes.NewReader(configBytes)
		meta.ContentLength = int64(len(configBytes))
		meta.ContentSHA256Hex = sumSHA256Hex(configBytes)
	}

	// 创建请求
	req := core.NewRequest(ctx, http.MethodPut, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, "")
	}

	// 成功后缓存桶位置
	if s.locationCache != nil {
		s.locationCache.Set(bucketName, options.Region)
	}

	return nil
}

// createBucketConfiguration 创建桶配置
type createBucketConfiguration struct {
	XMLName  xml.Name `xml:"CreateBucketConfiguration"`
	Location string   `xml:"LocationConstraint"`
}
