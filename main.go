package main

import (
	"fmt"
	"github.com/nbtri/aws-s3-uploader/aws"
	"github.com/nbtri/aws-s3-uploader/utils"
	"math/rand"
	"os"
	"time"
)

func config() (string, string, string, string, string, string) {
	return os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_S3_BUCKET"),
		os.Getenv("SITE"),
		os.Getenv("INSIDE_DIR"),
		os.Getenv("AWS_DEFAULT_REGION")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Reading aws credentials")

	id, secret, bucket, site, insideDir, region := config()

	fmt.Printf("Id: %s\nSecret: %s\nBucket: %s\n", id, secret, bucket)

	files := utils.Dig(site)

	s3Client := aws.NewS3Service(&aws.S3ClientConfig{Bucket: bucket})

	s3Client.UploadFiles(site, insideDir, files)

}
