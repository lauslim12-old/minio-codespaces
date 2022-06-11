package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// all constants
const (
	AWS_ACCESS_KEY        = "minioadmin"            // access key
	AWS_ENDPOINT          = "http://localhost:9000" // in minio this is set according to your localhost URL
	AWS_SECRET_ACCESS_KEY = "minioadmin"            // secret access key
	AWS_REGION            = "ap-northeast-1"        // region
	BUCKET                = "bucket"                // bucket name
)

// getBucketItemPresigned allows a user to fetch an item from a bucket with a
// presigned URL
func getBucketItemPresigned(s3Client *s3.S3, key string) (string, error) {
	// preparation process for the presigned URL
	request, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(BUCKET),
		Key:    aws.String(key),
	})

	// presign URL and return it properly
	url, err := request.Presign(15 * time.Minute)
	if err != nil {
		return "", err
	}

	return url, nil
}

// getFileInformation fetches all of the file's required information, which include and are limited
// to: name, size, type, and buffer
func getFileInformation(filePath string) (string, int64, string, []byte, error) {
	// open and fetch image
	file, err := os.Open(path.Clean(filePath))
	if err != nil {
		return "", 0, "", nil, err
	}
	defer file.Close()

	// get image information
	fileInformation, err := file.Stat()
	if err != nil {
		return "", 0, "", nil, err
	}

	// read file to buffer
	fileBuffer := make([]byte, fileInformation.Size())
	file.Read(fileBuffer)

	// return required variables
	return file.Name(), fileInformation.Size(), http.DetectContentType(fileBuffer), fileBuffer, nil
}

// getS3 gets the s3 session from minio
func getS3() *s3.S3 {
	return s3.New(session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(AWS_ACCESS_KEY, AWS_SECRET_ACCESS_KEY, ""),
		Endpoint:         aws.String(AWS_ENDPOINT),
		Region:           aws.String(AWS_REGION),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})))
}

// normalUpload uploads files the normal way without presigned URLs
func normalUpload(s3Client *s3.S3, key string, length int64, kind string, buffer []byte) (*s3.PutObjectOutput, error) {
	output, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(BUCKET),
		Key:                aws.String(key),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(length),
		ContentType:        aws.String(kind),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		return nil, err
	}

	return output, err
}

// presignedUpload uploads files with presigned URLs
// there is a bug: the presigned URL should not be 'localhost:9000', but instead, it should be '*.github.dev' as this
// is developed on github codespaces. this is a known issue and the fixes for this problem are:
// 1. replace 'localhost:9000' with the proper '*.github.dev' url
// 2. do not use presigned URLs, you can use the normal 'PutObjectInput' instead
func presignedUpload(s3Client *s3.S3, key string, length int64, kind string, buffer []byte) error {
	// preparation process to generate a request to put the object using presigned URLs
	request, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:             aws.String("bucket"),
		Key:                aws.String(key),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(length),
		ContentType:        aws.String(kind),
		ContentDisposition: aws.String("attachment"),
	})

	// generate presigned URL for 15 minutes
	url, err := request.Presign(15 * time.Minute)
	if err != nil {
		return err
	}

	// this url is bugged. see godoc in the function
	bug := fmt.Sprintf("presignedUpload: bugged presigned url, value: %s", url)
	fmt.Println(bug)
	fmt.Println()

	// prepare HTTP client to upload to the server
	client := &http.Client{}
	uploader, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(buffer))
	uploader.ContentLength = length
	uploader.Header.Set("Content-Type", kind)
	if err != nil {
		return err
	}

	// do perform the upload
	response, err := client.Do(uploader)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// the resulting will always be 400 as codespaces forces you to access `localhost:9000` on your computer,
	// resulting in desynchronization between the codespaces and your computer, as the image is actually hosted in a
	// port-forwarded minio, not in your localhost
	bug = fmt.Sprintf("presignedUpload: failure status code, value %s", response.Status)
	fmt.Println(bug)
	fmt.Println()

	return nil
}

func main() {
	// attempt to create a secure session with S3
	s3Client := getS3()

	// get all buckets
	result, err := s3Client.ListBuckets(nil)
	if err != nil {
		panic(err.Error())
	}

	// create a single bucket if it does not exists
	bucketExists := false
	for _, bucket := range result.Buckets {
		if *bucket.Name == BUCKET {
			bucketExists = true
			break
		}
	}

	if !bucketExists {
		_, err = s3Client.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(BUCKET)})
		if err != nil {
			panic(err.Error())
		}
	}

	// get required information from file
	fileName, fileSize, fileType, fileBuffer, err := getFileInformation("images/naruto.png")
	if err != nil {
		panic(err.Error())
	}

	// upload file to S3 bucket without presigned URL, straight to the endpoint
	_, err = normalUpload(s3Client, fileName, fileSize, fileType, fileBuffer)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Successfully uploaded with `normalUpload` to the server")
	fmt.Println()

	// upload file with presigned URL
	err = presignedUpload(s3Client, fileName, fileSize, fileType, fileBuffer)
	if err != nil {
		panic(err.Error())
	}

	// get files with presigned URLs
	url, err := getBucketItemPresigned(s3Client, "naruto.png")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(url)
	fmt.Println()

	fmt.Println("For now, presigned URLs with PUT requests and GET requests still end up in failures due to the localhost hardcoding.")
	fmt.Println()
}
