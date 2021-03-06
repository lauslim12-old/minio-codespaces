# MinIO Codespaces

Trying out MinIO with Docker Compose + Go in GitHub Codespaces using `aws-sdk-go`.

## Topics

- Connecting to S3.
- Listing buckets.
- Creating a single bucket if bucket does not exist yet.
- Normal upload.
- Uploads a single file with presigned URLs.
- Fetches a single file with presigned URLs.
- Downloads a single file with the presigned URLs.

Attempting to access the presigned URL directly in GitHub Codespaces (copying the `localhost:9000` URL and opening them in the browser, taking advantage of GitHub Codespaces's port forwarding in the process in order to turn that `localhost:9000` into `*.githubpreview.dev`) will cause `SignatureDoesNotMatch` error. I do not know the exact cause of this error, but it is probably because of the different host and domain names. At first, I thought it was because I placed `Content-Disposition` in the file metadata during the upload process. MinIO does not support illegal characters which comes from `Content-Disposition` option.

After researching multiple times, it seems that this is the expected behavior. S3 storages are supposed to be deployed in a dedicated endpoint, taking example from AWS, GCP, and DigitalOcean respectively (`s3.amazonaws.com`, `storage.googleapis.com`, and `<SPACE_NAME>.<REGION>.digitaloceanspaces.com`). We connect our API/app to that endpoint, and that endpoint is solely used as an S3 compatible storage in production. When we deploy our S3 storages like that (or in `localhost`), we would not have any problem with presigned URLs (as they will not aim at our `localhost` anymore, thus not causing any problems when we need to generate presigned URLs).

It also seems that `aws-sdk-go` and `minio-go` creates their presigned URL based on the passed endpoint in the S3 session object. If the endpoint URL is `localhost:9000` then the presigned URL will start with `localhost:9000`, and so on and so forth.

To reiterate, if you attempt to access the URL while running this project locally, it will work just fine and as expected. However, if you try to access this project in GitHub Codespaces, it will not work for reasons mentioned above.

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

To try the project locally, you can clone the repository, and then:

```bash
docker-compose up -d
go run .
```

Please examine the output and the code to get the general idea of how MinIO works in both Docker and Codespaces environment!

## License

MIT License. Image credit is [Unsplash Image](https://unsplash.com/photos/2wcfY2qeFFE).
