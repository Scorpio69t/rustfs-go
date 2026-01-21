// 示例：配置存储桶默认加密
//
// 设置存储桶的默认加密配置后，所有新上传的对象将自动加密。
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	rustfs "github.com/Scorpio69t/rustfs-go"
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

	bucketName := "encrypted-bucket"

	// 创建 RustFS 客户端
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("初始化客户端失败: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()

	// 创建存储桶（如果不存在）
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

	// 创建 SSE-S3 加密配置
	encryptionConfig := sse.NewConfiguration()

	fmt.Printf("\n📝 设置存储桶默认加密配置...\n")
	fmt.Printf("  加密算法: %s\n", encryptionConfig.Rules[0].ApplySSEByDefault.SSEAlgorithm)

	// 设置存储桶加密
	err = bucketSvc.SetEncryption(ctx, bucketName, *encryptionConfig)
	if err != nil {
		log.Fatalf("设置存储桶加密失败: %v", err)
	}
	fmt.Printf("✓ 成功设置存储桶默认加密\n")

	// 获取存储桶加密配置
	fmt.Printf("\n📥 获取存储桶加密配置...\n")
	retrievedConfig, err := bucketSvc.GetEncryption(ctx, bucketName)
	if err != nil {
		log.Fatalf("获取存储桶加密失败: %v", err)
	}

	fmt.Printf("✓ 存储桶加密配置:\n")
	for i, rule := range retrievedConfig.Rules {
		fmt.Printf("  规则 %d:\n", i+1)
		fmt.Printf("    算法: %s\n", rule.ApplySSEByDefault.SSEAlgorithm)
		fmt.Printf("    Bucket Key: %v\n", rule.BucketKeyEnabled)
		if rule.ApplySSEByDefault.KMSMasterKeyID != "" {
			fmt.Printf("    KMS Key ID: %s\n", rule.ApplySSEByDefault.KMSMasterKeyID)
		}
	}

	// 演示：使用 SSE-KMS 配置（可选）
	fmt.Printf("\n🔑 演示：设置 SSE-KMS 加密配置\n")
	kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
	kmsConfig := sse.NewKMSConfiguration(kmsKeyID)

	fmt.Printf("  KMS Key ID: %s\n", kmsConfig.Rules[0].ApplySSEByDefault.KMSMasterKeyID)
	fmt.Printf("  算法: %s\n", kmsConfig.Rules[0].ApplySSEByDefault.SSEAlgorithm)

	// 注意：实际设置 KMS 需要有效的 KMS 密钥
	// err = bucketSvc.SetEncryption(ctx, bucketName, *kmsConfig)
	// if err != nil {
	// 	log.Printf("警告: 设置 KMS 加密失败（可能需要有效的 KMS 密钥）: %v", err)
	// }

	// 删除加密配置
	fmt.Printf("\n🗑️  删除存储桶加密配置...\n")
	err = bucketSvc.DeleteEncryption(ctx, bucketName)
	if err != nil {
		log.Fatalf("删除存储桶加密失败: %v", err)
	}
	fmt.Printf("✓ 成功删除存储桶加密配置\n")

	// 验证删除
	fmt.Printf("\n📥 验证加密配置已删除...\n")
	_, err = bucketSvc.GetEncryption(ctx, bucketName)
	if err != nil {
		if err == sse.ErrNoEncryptionConfig {
			fmt.Printf("✓ 确认：存储桶无加密配置\n")
		} else {
			log.Printf("获取加密配置时出错: %v", err)
		}
	} else {
		fmt.Printf("⚠️  警告：删除后仍能获取到加密配置\n")
	}

	fmt.Printf("\n📌 存储桶加密提示:\n")
	fmt.Printf("  ✓ 默认加密对所有新上传的对象生效\n")
	fmt.Printf("  ✓ 不影响已存在的对象\n")
	fmt.Printf("  ✓ 支持 SSE-S3 和 SSE-KMS 两种模式\n")
	fmt.Printf("  ✓ SSE-C 不支持作为存储桶默认加密\n")
	fmt.Printf("  ✓ 建议对包含敏感数据的存储桶启用默认加密\n")

	// 清理（可选）
	// err = bucketSvc.Delete(ctx, bucketName)
	// if err != nil {
	// 	log.Printf("警告: 删除存储桶失败: %v", err)
	// }
}
