# S3 Multipart Content Server Adapter

## Overview

This project implements a **S3-compatible content server adapter**. It allows you to:

* Upload files to S3-compatible storage (MinIO) via streaming.
* Download files by document ID.
* List objects in a bucket.
* Retrieve file metadata.
* Delete objects.

The server is built with **Go**, using **Fiber** for HTTP handling, and **AWS SDK v2** for S3 operations.

---

## Features

* **Multipart upload support** (large file streaming)
* **Secure HTTPS endpoints** using self-signed TLS certificates
* **Server memory metrics** (`/mem` endpoint)
* Fully **automated integration tests** with MinIO in Docker
* Supports **object metadata and tagging** for document management
* **Benchmark support** for upload/download/delete performance

---

## Prerequisites

* [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
* [Go 1.25+](https://golang.org/dl/)
* `make` (optional, for convenience)

---

## Project Structure

```
.
├── bench/                  # Benchmark test
├── build/                  # Dockerfile & docker-compose
├── certs/                  # TLS certificates
├── config/                 # Configuration for server & S3
├── docs/                   # Swagger/OpenAPI documentation
├── tests/                  # Integration & error tests
├── utils/                  # Helper functions: handlers, S3 client, logging
├── main.go                 # Entry point
└── README.md
```

---

## Configuration

The server reads its configuration from `config/config.yaml`:

```yaml
server:
  port: 8080
s3:
  url: "http://minio:9000"
  accessKey: "minio_user"
  secretKey: "minio_password"
  region: "aws-global"
  maxConnections: 100
  bucketName: "test-bucket"
```

* `url` – S3 endpoint (MinIO)
* `accessKey` / `secretKey` – credentials
* `bucketName` – default bucket for tests

---

## Running Locally
### 🔹 0. Generate TLS Certificates (required for local HTTPS testing):

```bash
openssl req -x509 -nodes -days 365 -new -key key.pem -out cert.pem -config san.cnf
```
> **Important** Put certificates into `certs` folder to create correct `server` and `worker` docker images

> **Important:** Ensure that `san.cnf` includes the correct Subject Alternative Names (SANs) for your local server (`localhost`, IP, etc.)

### 1. Start MinIO + Adapter via Docker Compose

```bash
cd build
docker-compose up --build
```

* MinIO UI: [http://localhost:9001](http://localhost:9001)
* Content Server Adapter: `https://localhost:8080/ContentServer/ContentServer.dll`

> **Important:** Don't forget to create `test-bucket` for test usage

### 2. TLS Certificates

Self-signed certs are included in `certs/`:

```text
cert.pem
key.pem
```

---

## API Endpoints

### Upload document (POST)

```bash
curl -k -X POST "https://localhost:8080/ContentServer/ContentServer.dll?contRep=test-bucket&docId=TEST1" \
  -F "file=@test.txt"
```

### Download document (GET)

```bash
curl -k -X GET "https://localhost:8080/ContentServer/ContentServer.dll?get&contRep=test-bucket&docId=TEST1" -O
```

### Delete document (DELETE)

```bash
curl -k -X DELETE "https://localhost:8080/ContentServer/ContentServer.dll?contRep=test-bucket&docId=TEST1"
```

### Get document info (GET)

```bash
curl -k -X GET "https://localhost:8080/ContentServer/ContentServer.dll?info&contRep=test-bucket&docId=TEST1"
```

### List objects (GET)

```bash
curl -k -X GET "https://localhost:8080/ContentServer/ContentServer.dll?list&contRep=test-bucket"
```

### Server memory stats (GET)

```bash
curl -k -X GET "https://localhost:8080/mem"
```

---

## Swagger / API Docs

* Swagger UI available at `/docs/*`
* OpenAPI JSON: `/docs/swagger.json`
* Auto-generated using [swaggo/swag](https://github.com/swaggo/swag)

---

## Running Tests

Integration tests use the same Docker setup and the `test.txt` sample file.

```bash
cd s3-multipart-request-adapter
go test ./tests -v -count=1
```

> Tests perform full lifecycle: **upload → download → delete**.

---

## Benchmarking

You can measure the performance of your content server using the included Go benchmark, just run `bench/benchmark.go`:

```bash
go run bench/benchmark.go
```

* Uploads, downloads, and deletes the file multiple times - you can change it in code
* Reports timing for each iteration
* Useful for comparing different S3 backends or network performance

Example output:

```
Iteration 1...
Iteration 1 completed in 129.8264ms
Iteration 2...
Iteration 2 completed in 123.4325ms
Iteration 3...
Iteration 3 completed in 102.4284ms
Iteration 4...
Iteration 4 completed in 112.4924ms
Iteration 5...
Iteration 5 completed in 112.7068ms
```
---

## Logging

Requests are logged with the following format:

```
[time] req=<requestID> <METHOD> <PATH> func=<funcName> ip=<clientIP> duration=<duration> <extra>
```

---

## Development / Contribution

### 1. Adding New Endpoints

1. Implement a new **Fiber handler** in `utils/handlers.go`.
2. Add a route in `main.go` for your endpoint:

```go
app.Get("/new-endpoint", utils.HandleNewFeature())
```

3. Add Swagger annotations above the handler function in `main.go`:

```go
// @Summary Describe your endpoint
// @Tags Feature
// @Accept json
// @Produce json
// @Param paramName query string true "Description"
// @Success 200 {object} map[string]interface{}
// @Router /new-endpoint [get]
```

### 2. Generating Swagger Documentation

Use [swag](https://github.com/swaggo/swag) to generate Swagger docs:

```bash
# Install swag CLI if not installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g main.go -o docs
```

This will update `docs/swagger.json` and `docs/swagger.yaml`.

### 3. Running Locally Without Docker

1. Ensure MinIO is running locally (or configure your S3 endpoint in `config.yaml`).
2. Run:

```bash
go run main.go
```

### 4. Contributing

* Fork the repo
* Add tests for any new features in `tests/`
* Make sure all integration tests pass
* Submit a pull request with clear description

---

## Troubleshooting

### TLS Issues

If `curl` or Go HTTP client fails with TLS errors:

* Use `-k` flag to skip TLS verification, or
* Add `cert.pem` to your trusted certificates.

### MinIO Connectivity

* Ensure the MinIO container is running on `http://minio:9000`.
* Adapter container depends on `minio` service in Docker Compose.

### File Content Issues

* Line endings may differ between OS (`\r\n` on Windows vs `\n` on Linux). Tests normalize line endings automatically.

---

## Notes

* Adapter is compatible with any S3-compatible storage, but currently tested with **MinIO**.
* Supports object tagging for metadata.
* All uploaded files are stored **with uppercase document IDs** to avoid conflicts.

---

## Author

**Almaz**  
📧 [alkadriev@gmail.com](mailto:alkadriev@gmail.com)  
📦 Repository: `s3-multipart-request-adapter`
