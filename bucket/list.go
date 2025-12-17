// Package bucket bucket/list.go
package bucket

import (
	"context"
	"encoding/xml"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// List 列出所有桶
func (s *bucketService) List(ctx context.Context) ([]types.BucketInfo, error) {
	// 构建请求元数据（无桶名表示列出所有桶）
	meta := core.RequestMetadata{}

	// 创建 GET 请求
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp, "", "")
	}

	// 解析响应
	var result listAllMyBucketsResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Buckets.Bucket, nil
}

// GetLocation 获取桶位置
func (s *bucketService) GetLocation(ctx context.Context, bucketName string) (string, error) {
	// 验证桶名
	if err := validateBucketName(bucketName); err != nil {
		return "", err
	}

	// 先检查缓存
	if s.locationCache != nil {
		if location, ok := s.locationCache.Get(bucketName); ok {
			return location, nil
		}
	}

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName: bucketName,
		QueryValues: map[string][]string{
			"location": {""},
		},
	}

	// 创建 GET 请求
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return "", parseErrorResponse(resp, bucketName, "")
	}

	// 解析响应
	var result locationConstraint
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	location := result.Location
	if location == "" {
		location = "us-east-1"
	}

	// 缓存位置
	if s.locationCache != nil {
		s.locationCache.Set(bucketName, location)
	}

	return location, nil
}

// listAllMyBucketsResult 列出所有桶的响应
type listAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Owner   owner    `xml:"Owner"`
	Buckets buckets  `xml:"Buckets"`
}

// owner 所有者信息
type owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// buckets 桶列表
type buckets struct {
	Bucket []types.BucketInfo `xml:"Bucket"`
}

// locationConstraint 桶位置约束
type locationConstraint struct {
	XMLName  xml.Name `xml:"LocationConstraint"`
	Location string   `xml:",chardata"`
}
