// 示例：使用 SSE-C（客户提供密钥）加密上传和下载对象
//
// SSE-C 使用客户端提供的 256 位加密密钥进行加密。
// 密钥不会存储在服务器上，每次访问对象时都需要提供相同的密钥。
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/sse"
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
	objectName := "encrypted-with-customer-key.txt"

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

	// 生成 256 位（32 字节）加密密钥
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		log.Fatalf("生成加密密钥失败: %v", err)
	}
	fmt.Printf("✓ 生成 256 位加密密钥: %x...\n", encryptionKey[:8])

	// 创建 SSE-C 加密器
	sseEncrypter, err := sse.NewSSEC(encryptionKey)
	if err != nil {
		log.Fatalf("创建 SSE-C 加密器失败: %v", err)
	}

	// 准备上传数据
	content := "这是使用客户端密钥加密的高度敏感数据，密钥不会存储在服务器"
	reader := strings.NewReader(content)
	size := int64(len(content))

	// 使用 SSE-C 加密上传对象
	objectSvc := client.Object()
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, size,
		object.WithSSE(sseEncrypter), // 使用客户提供的密钥加密
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("上传对象失败: %v", err)
	}

	fmt.Printf("\n✓ 使用 SSE-C 加密上传成功\n")
	fmt.Printf("  存储桶: %s\n", uploadInfo.Bucket)
	fmt.Printf("  对象名: %s\n", uploadInfo.Key)
	fmt.Printf("  ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("  大小: %d 字节\n", uploadInfo.Size)

	// 下载对象（必须提供相同的密钥）
	fmt.Printf("\n📥 使用相同密钥下载对象...\n")
	downloadReader, info, err := objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetSSE(sseEncrypter), // 必须提供相同的密钥
	)
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

	fmt.Printf("✓ 下载成功（使用客户端密钥解密）\n")
	fmt.Printf("  内容: %s\n", string(buf[:n]))

	// 演示：使用错误的密钥无法下载
	fmt.Printf("\n🔒 测试：使用错误的密钥下载...\n")
	wrongKey := make([]byte, 32)
	rand.Read(wrongKey)
	wrongEncrypter, _ := sse.NewSSEC(wrongKey)

	_, _, err = objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetSSE(wrongEncrypter),
	)
	if err != nil {
		fmt.Printf("✓ 正确行为：使用错误密钥无法下载\n")
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("⚠️  警告：使用错误密钥也能下载（不应该发生）\n")
	}

	fmt.Printf("\n📌 SSE-C 重要提示:\n")
	fmt.Printf("  ✓ 密钥长度必须是 256 位（32 字节）\n")
	fmt.Printf("  ✓ 密钥不会存储在服务器上\n")
	fmt.Printf("  ✓ 每次访问对象都需要提供相同的密钥\n")
	fmt.Printf("  ✓ 丢失密钥意味着永久失去数据访问权\n")
	fmt.Printf("  ✓ 适合需要完全控制加密密钥的场景\n")
	fmt.Printf("  ⚠️  客户端需要安全管理密钥（推荐使用密钥管理系统）\n")

	// 清理（可选）
	// err = objectSvc.Delete(ctx, bucketName, objectName)
	// if err != nil {
	// 	log.Printf("警告: 删除对象失败: %v", err)
	// }
}
