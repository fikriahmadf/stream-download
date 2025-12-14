# S3 File Upload Service

Service Go untuk upload file ke Amazon S3 menggunakan LocalStack sebagai emulator lokal.

## Arsitektur

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Client    │────▶│   Go Server     │────▶│   LocalStack    │
│  (Upload)   │     │    (Fiber)      │     │   (S3 API)      │
└─────────────┘     └─────────────────┘     └─────────────────┘
```

## Tech Stack

- **Go** - Backend server
- **Fiber** - HTTP framework
- **AWS SDK Go v2** - S3 client
- **LocalStack** - AWS emulator untuk development

## Prerequisites

- Go 1.21+
- LocalStack CLI (`pip install localstack`)
- awslocal CLI (sudah termasuk dalam LocalStack)

## Struktur Project

```
stream-download/
├── main.go              # Entry point
├── go.mod               # Go modules
├── go.sum               # Dependencies checksum
├── .env                 # Environment variables
├── config/
│   └── config.go        # Configuration loader
├── handler/
│   └── upload.go        # Upload handler
├── service/
│   └── s3.go            # S3 service layer
└── README.md            # Dokumentasi
```

## Setup & Konfigurasi

### 1. Environment Variables

Buat file `.env`:

```env
AWS_REGION=ap-southeast-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
S3_ENDPOINT=http://localhost:4566
S3_BUCKET=my-bucket
SERVER_PORT=8080
```

### 2. Jalankan LocalStack

```bash
localstack start
```

> **Note:** Data LocalStack akan tersimpan secara global, bisa digunakan oleh project lain.

### 3. Buat S3 Bucket

```bash
awslocal s3 mb s3://my-bucket
```

## API Endpoint

### Upload File

**POST** `/api/upload`

**Request:**
- Content-Type: `multipart/form-data`
- Body: `file` (file yang akan diupload)

**Response:**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "filename": "example.pdf",
    "url": "http://localhost:4566/my-bucket/example.pdf",
    "size": 12345
  }
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@/path/to/your/file.pdf"
```

## Flow Upload

1. **Client** mengirim file via `multipart/form-data`
2. **Handler** menerima dan validasi file (size, type)
3. **Service** upload file ke S3 menggunakan AWS SDK
4. **Response** dikembalikan dengan URL file yang diupload

## Langkah Implementasi

| Step | Deskripsi | Status |
|------|-----------|--------|
| 1 | Setup project structure | Done   |
| 2 | Install dependencies (Fiber, AWS SDK) | ⬜      |
| 3 | Buat config loader | ⬜      |
| 4 | Buat S3 service | ⬜      |
| 5 | Buat upload handler | ⬜      |
| 6 | Setup router & main | ⬜      |
| 7 | Testing dengan LocalStack | ⬜      |

## Dependencies

```bash
go get github.com/gofiber/fiber/v2
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/joho/godotenv
```

## Testing

### Manual Test

```bash
# Start server
go run main.go

# Upload file
curl -X POST http://localhost:8080/api/upload \
  -F "file=@test.txt"
```

### Verify di LocalStack

```bash
# List files in bucket
awslocal s3 ls s3://my-bucket/
```

## Notes

- LocalStack menyediakan S3 API yang kompatibel dengan AWS
- Untuk production, ganti `S3_ENDPOINT` ke AWS endpoint yang sebenarnya
- File akan disimpan dengan nama asli, bisa ditambahkan UUID untuk uniqueness
