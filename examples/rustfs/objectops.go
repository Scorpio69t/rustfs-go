//go:build example
// +build example

package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		YOURACCESSKEYID     = "4UYIdunFNM0viXm1w6eg"
		YOURSECRETACCESSKEY = "WBINTZ41oP8pic5QjOEbMh09Ynx3ymfU2JvKARSw"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket" // 'mc mb play/mybucket' if it does not exist.
	)

	// 初始化客户端
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Creds:  credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET
	objectName := "test-object.txt"

	// 上传对象（从字符串）
	data := strings.NewReader("Hello, RustFS!")
	uploadInfo, err := client.PutObject(ctx, bucketName, objectName, data, data.Size(), rustfs.PutObjectOptions{
		ContentType: "text/plain",
		UserMetadata: map[string]string{
			"author": "rustfs-go",
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("成功上传对象: %s (ETag: %s, 大小: %d bytes)\n",
		uploadInfo.Key, uploadInfo.ETag, uploadInfo.Size)

	// 从文件上传对象
	uploadInfo, err = client.FPutObject(ctx, bucketName, "file-object.txt", "./file-object.txt", rustfs.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("成功从文件上传对象: %s\n", uploadInfo.Key)

	// 获取对象信息
	objInfo, err := client.StatObject(ctx, bucketName, objectName, rustfs.StatObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("对象信息: %s (大小: %d bytes, 类型: %s, 修改时间: %s)\n",
		objInfo.Key, objInfo.Size, objInfo.ContentType, objInfo.LastModified.Format("2006-01-02 15:04:05"))

	// 下载对象
	obj, err := client.GetObject(ctx, bucketName, objectName, rustfs.GetObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	defer func(obj *rustfs.Object) {
		err := obj.Close()
		if err != nil {

		}
	}(obj)

	buf := make([]byte, 1024)
	n, err := obj.Read(buf)
	if err != nil && err.Error() != "EOF" {
		log.Fatalln(err)
	}
	log.Printf("下载的对象内容: %s\n", string(buf[:n]))

	// 下载对象到文件
	// err = client.FGetObject(ctx, bucketName, objectName, "/path/to/local/download.txt", rustfs.GetObjectOptions{})
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// log.Println("成功下载对象到文件")

	// 删除对象
	//err = client.RemoveObject(ctx, bucketName, objectName, rustfs.RemoveObjectOptions{})
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//log.Printf("成功删除对象: %s\n", objectName)
}
