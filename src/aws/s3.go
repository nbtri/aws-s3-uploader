package aws

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	AWS_REGION = "ap-northeast-2"
)

type S3Client struct {
	s      *session.Session
	config *S3ClientConfig
}

type S3ClientConfig struct {
	Bucket string
}

func NewS3Service(config *S3ClientConfig) *S3Client {
	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION),
		Credentials: credentials.NewEnvCredentials(),
	})

	client := &S3Client{
		s:      s,
		config: config,
	}

	if err != nil {
		log.Fatal(err)
	}

	return client
}

func (client *S3Client) createTmpFile() (*os.File, error) {
	tempFilePath := os.TempDir() + string(os.PathSeparator)
	file, err := ioutil.TempFile(tempFilePath, "m-file-")
	if err != nil {
		log.Fatal(err)
	}

	return file, err
}

func (client *S3Client) DownloadFile(fileItem string) (string, error) {
	file, err := client.createTmpFile()

	//file, err := os.Create(tmpFile)
	if err != nil {
		log.Printf("Unable to open file %q, %v\n", file.Name(), err)
		return "", err
	}

	defer file.Close()

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	})

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(client.config.Bucket),
			Key:    aws.String(fileItem),
		})
	if err != nil {
		log.Printf("Unable to download item %q, %v\n", fileItem, err)
	}

	log.Println("Downloaded", file.Name(), numBytes, "bytes")

	return file.Name(), err
}

func (client *S3Client) UploadFiles(basePath, insideDir string, files []string) {
	for i, file := range files {
		var per float64
		per = float64(i) / float64(len(files))
		log.Printf("Uploaded %d of %d - %.4f percentages of files", i, len(files), per)
		bucketPath := client.getBucketPath(basePath, file)
		if len(bucketPath) > 0 {
			log.Println("Dest: " + insideDir + bucketPath)
			//client.UploadFile(file, client.getBucketPath(basePath, file))
			client.UploadFile(file, insideDir+bucketPath)
		}
	}
}

func (client *S3Client) getBucketPath(basePath, filePath string) string {
	if strings.Compare(basePath, filePath) == 0 {
		return ""
	}

	log.Printf("Build bucket path for: %s\n", filePath)

	extension := filepath.Ext(filePath)

	// folder detected
	if len(extension) == 0 {
		return ""
	}

	pos := strings.Index(filePath, basePath) + len(basePath) + 1 // +1 for slash
	runes := []rune(filePath)
	bucketPath := string(runes[pos:len(filePath)])
	return bucketPath
}

//TODO: update ext https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Complete_list_of_MIME_types
func (client *S3Client) getContentType(filePath string, buffer []byte) string {
	extension := filepath.Ext(filePath)

	if strings.Compare(".css", extension) == 0 {
		return "text/css"
	} else if strings.Compare(".js", extension) == 0 {
		return "application/javascript"
	} else if strings.Compare(".html", extension) == 0 {
		return "text/html"
	} else if strings.Compare(".svg", extension) == 0 {
		return "image/svg+xml"
	} else {
		return http.DetectContentType(buffer)
	}
}

func (client *S3Client) UploadFile(filePath string, bucketPath string) (string, error) {
	// Open the file for use
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	//fileDes := GetFilenameFromPath(filePath)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	p, err := s3.New(client.s).PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(client.config.Bucket),
		Key:           aws.String(bucketPath),
		ACL:           aws.String("public-read"),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(client.getContentType(filePath, buffer)),
		//ContentDisposition:   aws.String("attachment"),
		//ServerSideEncryption: aws.String("AES256"),
	})

	if p != nil {
		log.Println("Success: " + p.String())
	}

	if err != nil {
		log.Println(err.Error())
		log.Printf("Failed: " + filePath)
		return "", err
	}

	return p.String(), err
}

/**
 * Get file name & ext
 */
func GetFilenameFromPath(filepath string) string {
	re := regexp.MustCompile(`^(.*/)?(?:$|(.+?)(?:(\.[^.]*$)|$))`)

	match2 := re.FindStringSubmatch(filepath)

	filename := match2[2]

	return filename + path.Ext(filepath)
}
