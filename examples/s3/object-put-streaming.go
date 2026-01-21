//go:build example
// +build example

// 示例：流式上传对象
// 演示如何使用流式上传大文件
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "XhJOoEKn3BM6cjD2dVmx"
	secretKey = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
	bucket    = "mybucket"
)

func main() {
	// 创建客户端
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	service := client.Object()

	objectName := "streaming-upload.txt"

	// 创建一个大的数据流（模拟流式数据）
	// 实际应用中可以是文件流、网络流等
	data := strings.Repeat("This is a streaming upload test. ", 1000)
	reader := strings.NewReader(data)

	fmt.Printf("开始流式上传对象 '%s'...\n", objectName)
	fmt.Printf("数据大小: %d 字节\n", len(data))

	// 流式上传
	// 注意：-1 表示未知大小，SDK 会使用分块上传
	uploadInfo, err := service.Put(
		ctx,
		bucket,
		objectName,
		reader,
		int64(len(data)), // 如果已知大小可以指定，未知可用 -1
		object.WithContentType("text/plain"),
	)
	if err != nil {
		log.Fatalf("流式上传失败: %v\n", err)
	}

	fmt.Println("\n✅ 上传成功")
	fmt.Printf("对象名: %s\n", uploadInfo.Key)
	fmt.Printf("ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("大小: %d 字节\n", uploadInfo.Size)
}

// ProgressReader 是一个带进度显示的 Reader
type ProgressReader struct {
	reader    io.Reader
	total     int64
	current   int64
	lastPrint int64
}

// NewProgressReader 创建一个进度读取器
func NewProgressReader(r io.Reader, total int64) *ProgressReader {
	return &ProgressReader{
		reader: r,
		total:  total,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)

	// 每 10% 打印一次进度
	progress := pr.current * 100 / pr.total
	if progress >= pr.lastPrint+10 || err == io.EOF {
		fmt.Printf("上传进度: %d%%\n", progress)
		pr.lastPrint = progress
	}

	return n, err
}
