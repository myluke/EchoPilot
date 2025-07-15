package cfr2

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mylukin/EchoPilot/helper"
)

var client *s3.Client

// new s3 client
func init() {
	var accountId = helper.Config("R2_ACCOUNT_ID")
	var accessKeyId = helper.Config("R2_ACCESS_KEY_ID")
	var accessKeySecret = helper.Config("R2_ACCESS_KEY_SECRET")

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	client = s3.NewFromConfig(cfg)
}

// upload file
func Upload(bucket, key, filePath string) (*s3.PutObjectOutput, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建足够大的缓冲区来存储文件的前512字节
	buffer := make([]byte, 512)
	if _, err = file.Read(buffer); err != nil {
		return nil, err
	}

	// 检测文件内容类型
	contentType := http.DetectContentType(buffer)

	// 重置文件指针到文件起始位置
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// get file
func Get(bucket, key string) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	// Read all data from resp.Body
	content, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return content, nil
}
