# Developer Guide: Using Java with Minio and SAONetwork

This guide will walk you through the process of using Java with Minio and SAONetwork.

## Prerequisites

- Java: You can download and install it from [here](https://www.oracle.com/java/technologies/javase-jdk11-downloads.html).
- Spring Boot: This guide assumes you have some familiarity with Spring Boot. If you're new to Spring Boot, you can learn more about it [here](https://spring.io/projects/spring-boot).
- AWS SDK for Java: This provides a library of APIs and services for you to use with AWS services from Java. You can add it to your project using Maven or Gradle.

## Step 1: Create a new Spring Boot project

Create a new Spring Boot project. You can use [Spring Initializr](https://start.spring.io/) to generate a project with 'Web' and 'Amazon S3' dependencies.

## Step 2: Configure your application

Add the following properties to your `application.properties` file:

```properties
cloud.aws.credentials.accessKey=minioadmin
cloud.aws.credentials.secretKey=minioadmin
cloud.aws.region.static=us-east-1
cloud.aws.stack.auto=false
cloud.s3.path.style.access=true
cloud.aws.endpoint.static=http://localhost:9000
```

Replace `"http://localhost:9000"`, `"minioadmin"`, and `"minioadmin"` with your Minio server's URL and credentials.

## Step 3: Configure the AWS SDK

Create a `S3ClientConfig` class to configure the `AmazonS3` bean:

```java
@Configuration
public class S3ClientConfig {

    @Value("${cloud.aws.endpoint.static}")
    private String endpoint;

    @Value("${cloud.aws.region.static}")
    private String region;

    @Value("${cloud.aws.credentials.accessKey}")
    private String accessKey;

    @Value("${cloud.aws.credentials.secretKey}")
    private String secretKey;

    @Bean
    public AmazonS3 s3client() {
        AWSCredentials credentials = new BasicAWSCredentials(accessKey, secretKey);

        return AmazonS3ClientBuilder
                .standard()
                .withEndpointConfiguration(new AwsClientBuilder.EndpointConfiguration(endpoint, region))
                .withPathStyleAccessEnabled(true)
                .withCredentials(new AWSStaticCredentialsProvider(credentials))
                .build();
    }
}
```

## Step 4: Create a Service for S3 operations

Create a `S3Service` class to handle S3 operations:

```java
@Service
public class S3Service {

    private final AmazonS3 s3Client;

    @Autowired
    public S3Service(AmazonS3 s3Client) {
        this.s3Client = s3Client;
    }

    public void createBucket(String bucketName) {
        if (!s3Client.doesBucketExistV2(bucketName)) {
            s3Client.createBucket(bucketName);
        }
    }

    public void uploadFile(String bucketName, String fileName, File file) {
        s3Client.putObject(bucketName, fileName, file);
    }

    public S3Object downloadFile(String bucketName, String fileName) {
        return s3Client.getObject(bucketName, fileName);
    }

    public Map<String, String> getObjectMetadata(String bucketName, String objectKey) {
        ObjectMetadata objectMetadata = s3Client.getObjectMetadata(bucketName, objectKey);
        Map<String, String> userMetadata = objectMetadata.getUserMetadata();
        return userMetadata;
    }
}
```

## Step 5: Use the Service in your application

You can now use the `S3Service` in your controllers to create buckets, upload files, download files, and retrieve object metadata.

## Conclusion

You should now have a working Java application that interacts with a Minio server configured to work with SAONetwork. You can use this application to create buckets, upload files, download files, and retrieve object metadata. Please note that you may need to adjust the code to fit your needs.