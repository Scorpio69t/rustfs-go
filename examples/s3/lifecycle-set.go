//go:build example
// +build example

// Example: Set bucket lifecycle policy
// Demonstrates how to configure object expiration and transition rules
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "rustfsadmin"
	secretKey = "rustfsadmin"
	bucket    = "mybucket"
)

func main() {
	// Create client
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	service := client.Bucket()

	// 定义生命周期配置（XML 格式）
	// 规则：temp/ 目录下的对象 30 天后自动删除
	lifecycleConfig := `<LifecycleConfiguration>
	<Rule>
		<ID>expire-temp-files</ID>
		<Status>Enabled</Status>
		<Filter>
			<Prefix>temp/</Prefix>
		</Filter>
		<Expiration>
			<Days>30</Days>
		</Expiration>
	</Rule>
	<Rule>
		<ID>expire-old-logs</ID>
		<Status>Enabled</Status>
		<Filter>
			<Prefix>logs/</Prefix>
		</Filter>
		<Expiration>
			<Days>90</Days>
		</Expiration>
	</Rule>
</LifecycleConfiguration>`

	fmt.Printf("Setting lifecycle policy for bucket '%s'...\n", bucket)

	// Set lifecycle configuration
	err = service.SetLifecycle(ctx, bucket, []byte(lifecycleConfig))
	if err != nil {
		log.Fatalf("Failed to set lifecycle configuration: %v\n", err)
	}

	fmt.Println("✅ Lifecycle configuration set successfully")
	fmt.Println("\nLifecycle rules:")
	fmt.Println("  1. temp/ : expire after 30 days")
	fmt.Println("  2. logs/ : expire after 90 days")
}
