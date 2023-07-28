# Developer Guide: Using Go with Minio and SAONetwork

This guide will walk you through the process of using Go with Minio and SAONetwork to upload and download data.

## Prerequisites

- Go: You can download and install it from [here](https://golang.org/dl/).
- AWS SDK for Go: This provides a library of APIs and services for you to use with AWS services from Go. You can install it using the command `go get github.com/aws/aws-sdk-go`.

## Step 1: Create a new Go file

Create a new Go file, for example `main.go`, and open it in your favorite text editor. Import the necessary packages:

```go
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)
```

## Step 2: Configure the AWS SDK

In `main.go`, configure the AWS SDK with your Minio server's URL and credentials:

```go
func main() {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("minioadmin", "minioadmin", ""),
		Endpoint:         aws.String("http://localhost:9000"),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
}
```

Replace `"http://localhost:9000"`, `"minioadmin"`, and `"minioadmin"` with your Minio server's URL and credentials.

## Step 3: Upload a File

Add the following code to `main.go` to upload a file:

```go
file, err := os.Open("test.txt")
if err != nil {
	log.Fatal(err)
}
defer file.Close()

fileInfo, _ := file.Stat()
var size int64 = fileInfo.Size()
buffer := make([]byte, size)
file.Read(buffer)

_, err = s3Client.HeadBucket(&s3.HeadBucketInput{
	Bucket: aws.String("platformId"), // This is the platformId in SAO Network
})

if err != nil {
	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String("platformId"), // This is the platformId in SAO Network
	})
	if err != nil {
		log.Fatal(err)
	}
}

_, err = s3Client.PutObject(&s3.PutObjectInput{
	Bucket:               aws.String("platformId"), // This is the platformId in SAO Network
	Key:                  aws.String(file.Name()),
	Body:                 bytes.NewReader(buffer),
	ContentLength:        aws.Int64(size),
	ContentType:          aws.String(http.DetectContentType(buffer)),
	ContentDisposition:   aws.String("attachment"),
})
if err != nil {
	log.Fatal(err)
}
```

This code first checks if a bucket named 'mybucket' exists, and if it doesn't, it creates one. Then it uploads a file named 'test.txt' to the 'mybucket' bucket. In SAO Network, the 'bucket' concept is represented as 'platformId', which doesn't need to be created beforehand.

## Step 4: Download a File

Add the following code to `main.go` to download the file:

```go
output, err := s3Client.GetObject(&s3.GetObjectInput{
	Bucket: aws.String("platformId"), // This is the platformId in SAO Network
	Key:    aws.String("test.txt"),
})
if err != nil {
	log.Fatal(err)
}
defer output.Body.Close()

body, err := ioutil.ReadAll(output.Body)
if err != nil {
	log.Fatal(err)
}

fmt.Println(string(body))
```

This code downloads the 'test.txt' file from the 'mybucket' bucket and prints its content.

## Step 5: Run the Go file

You can run the Go file with the following command:

```bash
go run main.go
```

## Conclusion

You should now have a working Go application that interacts with a Minio server configured to work with SAONetwork. You can use this application to upload and download data. In SAO Network, the 'bucket' concept is represented as 'platformId', which doesn't need to be created beforehand. However, the bucket creation step is still necessary for Minio to function correctly.