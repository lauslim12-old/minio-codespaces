# MinIO Codespaces

Trying out MinIO with Docker Compose + Go in GitHub Codespaces using `aws-sdk-go`.

## Topics

- Connecting to S3.
- Listing buckets.
- Creating a single bucket if bucket does not exist yet.
- Normal upload.
- Uploads with presigned URLs.
- Fetches with presigned URLs.

Presigned URLs are still in progress as I am still trying to figure out how to not cause `SignatureNotMatch` error when accessing the images through Codespaces's port forwarding feature.

## Credentials

Default credentials are as follows:

- `AWS_ACCESS_KEY` is `minioadmin`.
- `AWS_SECRET_ACCESS_KEY` is `minioadmin`.
- `AWS_REGION` is `ap-northeast-1`.

## Steps

In order to try out this project, open it in GitHub Codespaces, and then after the automated bootstrapping process is done, do the following commands:

```bash
go run .
```

Please examine the output and the code to get the general idea of how MinIO works in both Docker and Codespaces environment!

## License

MIT License.