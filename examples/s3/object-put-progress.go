//go:build example
// +build example

// 示例：带进度显示的对象上传
// 演示如何在上传过程中显示进度
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

// ProgressReader 包装一个 io.Reader 并显示读取进度
type ProgressReader struct {
	reader      io.Reader
	total       int64
	current     int64
	lastPercent int64
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

	// 计算当前进度百分比
	if pr.total > 0 {
		percent := pr.current * 100 / pr.total
		// 每变化 5% 打印一次
		if percent >= pr.lastPercent+5 || err == io.EOF {
			fmt.Printf("\r上传进度: %d%% (%d/%d 字节)", percent, pr.current, pr.total)
			pr.lastPercent = percent
			if err == io.EOF || percent >= 100 {
				fmt.Println() // 换行
			}
		}
	}

	return n, err
}

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

	objectName := "progress-upload.txt"

	// 创建测试数据（约 5MB）
	data := strings.Repeat("这是一个进度显示测试。", 100000)
	dataSize := int64(len(data))

	fmt.Printf("准备上传对象 '%s' (大小: %.2f MB)...\n", objectName, float64(dataSize)/1024/1024)

	// 使用进度读取器包装数据
	reader := strings.NewReader(data)
	progressReader := NewProgressReader(reader, dataSize)

	// 上传对象
	uploadInfo, err := service.Put(
		ctx,
		bucket,
		objectName,
		progressReader,
		dataSize,
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("\n上传失败: %v\n", err)
	}

	fmt.Println("\n✅ 上传成功")
	fmt.Printf("对象名: %s\n", uploadInfo.Key)
	fmt.Printf("ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("大小: %d 字节 (%.2f MB)\n", uploadInfo.Size, float64(uploadInfo.Size)/1024/1024)
}
