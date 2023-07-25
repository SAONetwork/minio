# Developer Guide: Using Node.js with Minio and SAONetwork

This guide will walk you through the process of using Node.js with Minio and SAONetwork.

## Prerequisites

- Node.js and npm: You can download and install them from [here](https://nodejs.org/en/download/).
- AWS SDK for JavaScript: This provides a library of APIs and services for you to use with AWS services from Node.js. You can install it using npm with the command `npm install aws-sdk`.

## Step 1: Initialize a new Node.js project

Create a new directory for your project and initialize a new Node.js project:

```bash
mkdir s3-app && cd s3-app
npm init -y
```

## Step 2: Configure the AWS SDK

Create a new JavaScript file, for example `app.js`, and open it in your favorite text editor. Import the AWS SDK and configure it with your Minio server's URL and credentials:

```javascript
const AWS = require('aws-sdk');

AWS.config.update({
    accessKeyId: 'minioadmin',
    secretAccessKey: 'minioadmin',
    region: 'us-east-1',
    endpoint: new AWS.Endpoint('http://localhost:9000'),
});

const s3 = new AWS.S3();
```

Replace `"http://localhost:9000"`, `"minioadmin"`, and `"minioadmin"` with your Minio server's URL and credentials.

## Step 3: Create a Bucket

Create a new JavaScript file named `createBucket.js`:

```javascript
const AWS = require('aws-sdk');

AWS.config.update({
  accessKeyId: 'minioadmin',
  secretAccessKey: 'minioadmin',
  region: 'us-east-1',
  s3ForcePathStyle: true,
  endpoint: new AWS.Endpoint('http://localhost:9000'),
});

const s3 = new AWS.S3();
const bucketName = 'nodejs';

s3.headBucket({ Bucket: bucketName }, function(err, data) {
  if (err) {
    if (err.code === 'NotFound') {
      s3.createBucket({ Bucket: bucketName }, function(err, data) {
        if (err) {
          console.log('Error creating bucket', err);
        } else {
          console.log('Bucket created successfully', data.Location);
        }
      });
    } else {
      console.log('Error occurred', err);
    }
  } else {
    console.log('Bucket already exists');
  }
});
```

Run the script with `node createBucket.js`. This script checks if a bucket named 'nodejs' exists, and if it doesn't, it creates one.

## Step 4: Upload a File

Create a new JavaScript file named `uploadFile.js`:

```javascript
const AWS = require('aws-sdk');

AWS.config.update({
  accessKeyId: 'minioadmin',
  secretAccessKey: 'minioadmin',
  region: 'us-east-1',
  s3ForcePathStyle: true,
  endpoint: new AWS.Endpoint('http://localhost:9000'),
});

const s3 = new AWS.S3();

const uploadParams = {
  Bucket: 'nodejs',
  Key: 'sao-test',
  Body: 'Hello, world!',
};

s3.upload(uploadParams, function(err, data) {
  if (err) {
    console.log("Error", err);
  } if (data) {
    console.log("Upload Success", data.Location);
  }
});
```

Run the script with `node uploadFile.js`. This script uploads a file named 'sao-test' with the content 'Hello, world!' to the 'nodejs' bucket.

## Step 5: Download a File

Create a new JavaScript file named `downloadFile.js`:

```javascript
const AWS = require('aws-sdk');

AWS.config.update({
  accessKeyId: 'minioadmin',
  secretAccessKey: 'minioadmin',
  region: 'us-east-1',
  s3ForcePathStyle: true,
  endpoint: new AWS.Endpoint('http://localhost:9000'),
});

const s3 = new AWS.S3();

const downloadParams = {
  Bucket: 'nodejs',
  Key: 'sao-test',
};

s3.getObject(downloadParams, function(err, data) {
  if (err) {
    console.log("Error", err);
  } if (data) {
    console.log("Download Success", data.Body.toString());
  }
});
```

Run the script with `node downloadFile.js`. This script downloads the 'sao-test' file from the 'nodejs' bucket and prints its content.

## Conclusion

You should now have a working Node.js application that interacts with a Minio server configured to work with SAONetwork. You can use this application to create buckets, upload files, and download files.