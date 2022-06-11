package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	// attempt to create a secure session with S3
	s3Client := s3.New(session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("minioadmin", "minioadmin", ""),
		Endpoint:         aws.String("http://localhost:9000"), // in minio this is set according to your localhost URL
		Region:           aws.String("ap-northeast-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})))

	// validate s3 object
	fmt.Println(s3Client.Endpoint, s3Client.ServiceID, s3Client.ServiceName)
	fmt.Println()

	// create bucket
	// _, err := s3Client.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String("bucket")})
	// if err != nil {
	// 	fmt.Println(err.Error()) // our bucket may have been created by us, for now we do not check the errors
	// 	fmt.Println()

	// 	// if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
	// 	// 	// If the SDK can determine the request or retry delay was canceled
	// 	// 	// by a context the CanceledErrorCode error code will be returned.
	// 	// 	fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
	// 	// } else {
	// 	// 	fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
	// 	// }
	// }

	// get naruto image
	file, err := os.Open("images/naruto.png")
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	fmt.Println(file.Name())
	fmt.Println()

	// get naruto image information
	fileInformation, err := file.Stat()
	if err != nil {
		panic(err.Error())
	}
	fileSize := fileInformation.Size()

	// read file to buffer
	fileBuffer := make([]byte, fileInformation.Size())
	file.Read(fileBuffer)
	fileType := http.DetectContentType(fileBuffer)

	// upload file to S3 bucket without presigned URL, straight to the endpoint
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String("bucket"),
		Key:                aws.String(file.Name()),
		Body:               bytes.NewReader(fileBuffer),
		ContentLength:      aws.Int64(fileSize),
		ContentType:        aws.String(fileType),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		panic(err.Error())
	}

	// // preparation process to generate a request to put the object using presigned URLs
	// request, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
	// 	Bucket:             aws.String("bucket"),
	// 	Key:                aws.String(file.Name()),
	// 	Body:               bytes.NewReader(fileBuffer),
	// 	ContentLength:      aws.Int64(fileSize),
	// 	ContentType:        aws.String(fileType),
	// 	ContentDisposition: aws.String("attachment"),
	// })

	// // generate presigned URL for two minutes
	// url, err := request.Presign(15 * time.Minute)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// // the URL should not be 'localhost:9000', but instead, it should be '*.github.dev' as this is developed
	// // on github codespaces. this is a known issue and the fixes for this problem are:
	// // 1. replace 'localhost:9000' with the proper '*.github.dev' url
	// // 2. do not use presigned URLs, you can use the normal 'PutObjectInput' instead
	// fmt.Println(url)
	// fmt.Println()

	// // prepare HTTP client to upload to the server
	// client := &http.Client{}
	// uploader, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(fileBuffer))
	// uploader.ContentLength = fileSize
	// uploader.Header.Set("Content-Type", fileType)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// // do perform the upload
	// response, err := client.Do(uploader)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer response.Body.Close()
	// fmt.Println(response.Status, response.StatusCode)
	// fmt.Println()

	// get file by presigned url
	buckets, err := s3Client.ListBuckets(nil)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(buckets)
	fmt.Println()
}
