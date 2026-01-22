//go:build example
// +build example

// Example: Get bucket CORS configuration
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/cors"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "cors-bucket"
	)

	// Initialize RustFS client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()

	config, err := bucketSvc.GetCORS(ctx, YOURBUCKET)
	if err != nil {
		if err == cors.ErrNoCORSConfig {
			fmt.Printf("No CORS configuration found for %s\n", YOURBUCKET)
			return
		}
		log.Fatalf("Failed to get CORS: %v", err)
	}

	fmt.Printf("CORS configuration for %s:\n", YOURBUCKET)
	for i, rule := range config.CORSRules {
		fmt.Printf("  Rule %d:\n", i+1)
		fmt.Printf("    ID: %s\n", rule.ID)
		fmt.Printf("    AllowedOrigin: %v\n", rule.AllowedOrigin)
		fmt.Printf("    AllowedMethod: %v\n", rule.AllowedMethod)
		fmt.Printf("    AllowedHeader: %v\n", rule.AllowedHeader)
		fmt.Printf("    ExposeHeader: %v\n", rule.ExposeHeader)
		fmt.Printf("    MaxAgeSeconds: %d\n", rule.MaxAgeSeconds)
	}
}
