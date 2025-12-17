// Package object object/list.go
package object

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// ListBucketV2Result 列出存储桶 V2 响应结果
type ListBucketV2Result struct {
	XMLName               xml.Name           `xml:"ListBucketResult"`
	Name                  string             `xml:"Name"`
	Prefix                string             `xml:"Prefix"`
	KeyCount              int                `xml:"KeyCount"`
	MaxKeys               int                `xml:"MaxKeys"`
	Delimiter             string             `xml:"Delimiter"`
	IsTruncated           bool               `xml:"IsTruncated"`
	Contents              []types.ObjectInfo `xml:"Contents"`
	CommonPrefixes        []CommonPrefix     `xml:"CommonPrefixes"`
	ContinuationToken     string             `xml:"ContinuationToken"`
	NextContinuationToken string             `xml:"NextContinuationToken"`
	StartAfter            string             `xml:"StartAfter"`
}

// CommonPrefix 公共前缀
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// List 列出对象（实现）
func (s *objectService) List(ctx context.Context, bucketName string, opts ...ListOption) <-chan types.ObjectInfo {
	// 创建对象信息通道
	objectCh := make(chan types.ObjectInfo)

	// 启动后台 goroutine 进行列表操作
	go func() {
		defer close(objectCh)

		// 验证参数
		if err := validateBucketName(bucketName); err != nil {
			objectCh <- types.ObjectInfo{Err: err}
			return
		}

		// 应用选项
		options := applyListOptions(opts)

		// 设置分隔符
		delimiter := "/"
		if options.Recursive {
			// 递归列出，不使用分隔符
			delimiter = ""
		}

		// 保存 ContinuationToken 用于下一次请求
		var continuationToken string

		for {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				objectCh <- types.ObjectInfo{Err: ctx.Err()}
				return
			default:
			}

			// 查询对象列表（最多 1000 个）
			result, err := s.listObjectsV2Query(ctx, bucketName, &options, delimiter, continuationToken)
			if err != nil {
				objectCh <- types.ObjectInfo{Err: err}
				return
			}

			// 发送内容对象
			for _, object := range result.Contents {
				// 移除 ETag 引号
				object.ETag = trimETag(object.ETag)

				select {
				case objectCh <- object:
				case <-ctx.Done():
					objectCh <- types.ObjectInfo{Err: ctx.Err()}
					return
				}
			}

			// 发送公共前缀（仅在使用分隔符时）
			for _, prefix := range result.CommonPrefixes {
				select {
				case objectCh <- types.ObjectInfo{Key: prefix.Prefix}:
				case <-ctx.Done():
					objectCh <- types.ObjectInfo{Err: ctx.Err()}
					return
				}
			}

			// 如果有下一个 ContinuationToken，保存它
			if result.NextContinuationToken != "" {
				continuationToken = result.NextContinuationToken
			}

			// 如果列表未截断，结束
			if !result.IsTruncated {
				return
			}

			// 防止无限循环（某些 S3 实现可能有 bug）
			if continuationToken == "" {
				objectCh <- types.ObjectInfo{
					Err: fmt.Errorf("list is truncated without continuation token"),
				}
				return
			}
		}
	}()

	return objectCh
}

// listObjectsV2Query 查询对象列表 V2
func (s *objectService) listObjectsV2Query(ctx context.Context, bucketName string, options *ListOptions, delimiter, continuationToken string) (ListBucketV2Result, error) {
	// 构建查询参数
	queryValues := url.Values{}

	// 设置 list-type=2 (V2)
	queryValues.Set("list-type", "2")

	// 设置 encoding-type
	queryValues.Set("encoding-type", "url")

	// 设置前缀
	if options.Prefix != "" {
		queryValues.Set("prefix", options.Prefix)
	}

	// 设置分隔符
	if delimiter != "" {
		queryValues.Set("delimiter", delimiter)
	}

	// 设置 start-after
	if options.StartAfter != "" {
		queryValues.Set("start-after", options.StartAfter)
	}

	// 设置 continuation-token
	if continuationToken != "" {
		queryValues.Set("continuation-token", continuationToken)
	}

	// 设置 max-keys
	maxKeys := options.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 1000 // 默认最大值
	}
	queryValues.Set("max-keys", strconv.Itoa(maxKeys))

	// 设置 fetch-owner
	queryValues.Set("fetch-owner", "true")

	// 设置 metadata
	if options.WithMetadata {
		queryValues.Set("metadata", "true")
	}

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		QueryValues:  queryValues,
		CustomHeader: options.CustomHeaders,
	}

	// 创建 GET 请求
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return ListBucketV2Result{}, err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return ListBucketV2Result{}, parseErrorResponse(resp, bucketName, "")
	}

	// 解析 XML 响应
	var result ListBucketV2Result
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return ListBucketV2Result{}, fmt.Errorf("failed to decode list objects response: %w", err)
	}

	// URL 解码对象名称（因为使用了 encoding-type=url）
	for i := range result.Contents {
		if decodedKey, err := url.QueryUnescape(result.Contents[i].Key); err == nil {
			result.Contents[i].Key = decodedKey
		}
	}

	// URL 解码公共前缀
	for i := range result.CommonPrefixes {
		if decodedPrefix, err := url.QueryUnescape(result.CommonPrefixes[i].Prefix); err == nil {
			result.CommonPrefixes[i].Prefix = decodedPrefix
		}
	}

	return result, nil
}

// trimETag 移除 ETag 的引号
func trimETag(etag string) string {
	if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
		return etag[1 : len(etag)-1]
	}
	return etag
}
