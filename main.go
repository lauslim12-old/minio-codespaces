package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
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

// downloadImagePresignedURL downloads an image from a presigned URL and places them in an
// arbitrary location at the filesystem
func downloadImagePresignedURL(filePath, url string) (string, error) {
	// create the file
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// get the data
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// check server response
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloadImagePresignedURL: download fails with status %s", response.Status)
	}

	// write the resulting body to that file
	_, err = io.Copy(out, response.Body)
	if err != nil {
		return "", err
	}

	// return the status code
	return response.Status, nil
}

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

	// return file information as described in the godoc comment
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
		Bucket:        aws.String(BUCKET),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(length),
		ContentType:   aws.String(kind),
	})
	if err != nil {
		return nil, err
	}

	return output, err
}

// presignedUpload uploads files with presigned URLs
func presignedUpload(s3Client *s3.S3, key string, length int64, kind string, buffer []byte) (string, string, error) {
	// preparation process to generate a request to put the object using presigned URLs
	request, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:        aws.String("bucket"),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(length),
		ContentType:   aws.String(kind),
	})

	// generate presigned URL for 15 minutes
	url, err := request.Presign(15 * time.Minute)
	if err != nil {
		return "", "", err
	}

	// prepare HTTP client to upload to the server
	client := &http.Client{}
	uploader, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(buffer))
	uploader.ContentLength = length
	uploader.Header.Set("Content-Type", kind)
	if err != nil {
		return "", "", err
	}

	// do perform the upload
	response, err := client.Do(uploader)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()

	// return the presigned url and status code
	return url, response.Status, nil
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
	fileName, fileSize, fileType, fileBuffer, err := getFileInformation(path.Join("images", "image.png"))
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
	putPresignedURL, statusText, err := presignedUpload(s3Client, fileName, fileSize, fileType, fileBuffer)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Uploaded file with `presignedUpload` function with status: %s\n", statusText)
	fmt.Println()

	// get files with presigned URLs
	getPresignedURL, err := getBucketItemPresigned(s3Client, path.Join("images", "image.png"))
	if err != nil {
		panic(err.Error())
	}

	// download file from that presigned URL
	status, err := downloadImagePresignedURL(path.Join("images", "downloaded.png"), getPresignedURL)
	if err != nil {
		panic(err.Error())
	}

	// print collection of presigned URLs
	fmt.Printf("Presigned URL for PUT: %s\n", putPresignedURL)
	fmt.Println()
	fmt.Printf("Presigned URL for GET: %s\n", getPresignedURL)
	fmt.Println()
	fmt.Printf("Status text for Presigned GET process: %s\n", status)
	fmt.Println()

	// create simple webserver to test the `Host` header
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// try to generate another S3 client
		s3HTTPClient := getS3()

		// load template
		template, err := template.ParseFiles(path.Join("views", "index.html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get another presigned url for our file
		getPresignedURL, err := getBucketItemPresigned(s3HTTPClient, path.Join("images", "image.png"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// execute template as a response
		err = template.Execute(w, map[string]interface{}{"url": getPresignedURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// serve html file to test `Host` header
	fmt.Println("Server starts on port 8080!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
