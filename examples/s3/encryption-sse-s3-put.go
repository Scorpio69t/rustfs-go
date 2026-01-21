// 示例：使用 SSE-S3 加密上传对象
//
// SSE-S3 使用 S3 服务器管理的密钥进行加密，无需客户端管理密钥。
// 这是最简单的服务端加密方式，适合大多数场景。
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// 从环境变量获取配置
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := os.Getenv("S3_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("S3_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucketName := "test-encryption"
	objectName := "encrypted-object.txt"

	// 创建 RustFS 客户端
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("初始化客户端失败: %v", err)
	}

	ctx := context.Background()

	// 创建存储桶（如果不存在）
	bucketSvc := client.Bucket()
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("检查存储桶失败: %v", err)
	}
	if !exists {
		err = bucketSvc.Create(ctx, bucketName)
		if err != nil {
			log.Fatalf("创建存储桶失败: %v", err)
		}
		fmt.Printf("✓ 创建存储桶: %s\n", bucketName)
	}

	// 准备上传数据
	content := "这是使用 SSE-S3 加密的敏感数据"
	reader := strings.NewReader(content)
	size := int64(len(content))

	// 使用 SSE-S3 加密上传对象
	objectSvc := client.Object()
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, size,
		object.WithSSES3(), // 启用 SSE-S3 加密
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("上传对象失败: %v", err)
	}

	fmt.Printf("✓ 使用 SSE-S3 加密上传成功\n")
	fmt.Printf("  存储桶: %s\n", uploadInfo.Bucket)
	fmt.Printf("  对象名: %s\n", uploadInfo.Key)
	fmt.Printf("  ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("  大小: %d 字节\n", uploadInfo.Size)

	// 下载对象（服务器会自动解密）
	downloadReader, info, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("下载对象失败: %v", err)
	}
	defer downloadReader.Close()

	// 读取内容
	buf := make([]byte, info.Size)
	n, err := downloadReader.Read(buf)
	if err != nil && err.Error() != "EOF" {
		log.Fatalf("读取对象失败: %v", err)
	}

	fmt.Printf("\n✓ 下载成功（服务器自动解密）\n")
	fmt.Printf("  内容: %s\n", string(buf[:n]))

	// 获取对象元数据，验证加密信息
	stat, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("获取对象元数据失败: %v", err)
	}

	fmt.Printf("\n✓ 对象元数据\n")
	fmt.Printf("  大小: %d 字节\n", stat.Size)
	fmt.Printf("  ETag: %s\n", stat.ETag)
	fmt.Printf("  最后修改: %s\n", stat.LastModified)

	// SSE-S3 加密信息通常在响应头中，可以通过自定义头获取
	fmt.Printf("\n📌 提示:\n")
	fmt.Printf("  - SSE-S3 使用服务器管理的密钥加密\n")
	fmt.Printf("  - 数据在服务器端自动加密和解密\n")
	fmt.Printf("  - 客户端无需管理加密密钥\n")
	fmt.Printf("  - 适合大多数加密需求场景\n")

	// 清理（可选）
	// err = objectSvc.Delete(ctx, bucketName, objectName)
	// if err != nil {
	// 	log.Printf("警告: 删除对象失败: %v", err)
	// }
}
