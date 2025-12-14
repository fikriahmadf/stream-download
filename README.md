# S3 File Upload & Streaming Download Service

Service Go untuk **upload file ke S3** dan **download multiple files sebagai ZIP secara streaming**.

## Highlights

* **Streaming ZIP Download** – File di-zip on-the-fly dan langsung di-stream ke client tanpa perlu menunggu seluruh ZIP selesai dibuat
* **No Content-Length** – Client tidak perlu tahu ukuran ZIP, download langsung mulai
* **Memory Efficient** – Server tidak menyimpan ZIP di memory/disk, langsung stream ke client
* **Network Throttle Simulation** – Testing kondisi jaringan lambat (3G/4G) menggunakan `tc`

---

## Arsitektur

### Upload Flow
```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Client    │────▶│   Go Server     │────▶│   LocalStack    │
│  (Upload)   │     │    (Fiber)      │     │   (S3 API)      │
└─────────────┘     └─────────────────┘     └─────────────────┘
```

### Streaming Download Flow
```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Client    │◀────│   Go Server     │◀────│  S3 File URLs   │
│  (Browser)  │     │  (Stream ZIP)   │     │  (HTTP GET)     │
└─────────────┘     └─────────────────┘     └─────────────────┘
      │                     │
      │   Chunked Transfer  │
      │◀────────────────────│
      │   (No Content-Length)
```

**Bagaimana cara kerjanya:**
1. Client mengirim list URL file yang ingin di-download
2. Server fetch file satu per satu dari URL (S3)
3. Setiap file langsung di-compress dan di-stream ke client
4. Client menerima ZIP secara streaming (download langsung mulai)

---

## Tech Stack

* **Go** – Backend server
* **Fiber** – HTTP framework
* **Swagger** – API documentation & testing UI
* **AWS SDK Go v2** – S3 client
* **LocalStack (Docker)** – AWS emulator
* **tc (Traffic Control)** – Network throttling (bandwidth & latency)

---

## Prerequisites

* Go **1.21+**
* Docker & Docker Compose
* LocalStack (via Docker image)
* AWS CLI / awslocal

---

## Struktur Project

```
stream-download/
├── main.go              # Entry point
├── go.mod               # Go modules
├── go.sum               # Dependencies checksum
├── .env                 # Environment variables
├── docker-compose.yml   # LocalStack setup
├── config/
│   └── config.go        # Configuration loader
├── handler/
│   ├── upload.go        # Upload handler
│   └── download.go      # Streaming ZIP download handler
├── service/
│   └── s3.go            # S3 service layer
└── README.md            # Dokumentasi
```

---

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

---

### 2. Jalankan LocalStack (Docker Compose)

```bash
docker compose up -d
```

Contoh `docker-compose.yml`:

```yaml
services:
  localstack:
    image: localstack/localstack:latest
    container_name: localstack-main
    ports:
      - "4566:4566"
    environment:
      - SERVICES=s3
      - LOCALSTACK_PERSISTENCE=1
    volumes:
      - localstack_data:/var/lib/localstack
      - /var/run/docker.sock:/var/run/docker.sock
    cap_add:
      - NET_ADMIN

volumes:
  localstack_data:
```

> `NET_ADMIN` diperlukan untuk simulasi network throttle menggunakan `tc`.

---

### 3. Buat S3 Bucket

```bash
awslocal s3 mb s3://my-bucket
```

---

## Network Throttling (Simulasi Jaringan Lambat)

Untuk mensimulasikan kondisi jaringan lambat (misalnya 3G / 4G / Wi-Fi buruk), gunakan **`tc (Traffic Control)`** di dalam container LocalStack.

### Masuk ke container

```bash
docker exec -it localstack-main bash
```

Install `tc` (sekali saja):

```bash
apt-get update
apt-get install -y iproute2
```

---

### Contoh Preset Throttle

#### Wi-Fi normal

```bash
tc qdisc add dev eth0 root tbf rate 10mbit burst 256kbit latency 50ms
```

* ~ **1.25 MB/s**
* Latency ringan

#### 4G lambat

```bash
tc qdisc add dev eth0 root tbf rate 2mbit burst 128kbit latency 200ms
```

#### 3G

```bash
tc qdisc add dev eth0 root tbf rate 500kbit burst 64kbit latency 600ms
```

#### Jaringan sangat buruk

```bash
tc qdisc add dev eth0 root tbf rate 300kbit burst 32kbit latency 800ms
```

---

### Matikan Throttle

```bash
tc qdisc del dev eth0 root
```

> Throttle akan otomatis hilang jika container direstart.

---

## API Endpoints

### 1. Upload File

**POST** `/api/upload`

**Request:**

* Content-Type: `multipart/form-data`
* Body:
  * `file` – File yang akan diupload
  * `filePath` (optional) – Custom path di S3
  * `fileName` (optional) – Custom filename

**Response:**

```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "filename": "example.pdf",
    "filePath": "documents/2024",
    "url": "http://localhost:4566/my-bucket/documents/2024/example.pdf",
    "size": 12345
  }
}
```

**cURL Example:**

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@/path/to/your/file.pdf" \
  -F "filePath=documents/2024" \
  -F "fileName=report.pdf"
```

---

### 2. Streaming ZIP Download

**POST** `/api/download`

**Request:**

* Content-Type: `application/json`
* Body:

```json
{
  "urls": [
    "http://localhost:4566/my-bucket/file1.pdf",
    "http://localhost:4566/my-bucket/file2.pdf",
    "http://localhost:4566/my-bucket/file3.pdf"
  ]
}
```

**Response:**

* Content-Type: `application/zip`
* Transfer-Encoding: `chunked` (streaming)
* File: `download.zip`

**cURL Example:**

```bash
curl -X POST http://localhost:8080/api/download \
  -H "Content-Type: application/json" \
  -d '{"urls": ["http://localhost:4566/my-bucket/file1.pdf", "http://localhost:4566/my-bucket/file2.pdf"]}' \
  --output download.zip
```

**Keuntungan Streaming:**

| Aspek | Traditional | Streaming |
|-------|-------------|----------|
| Memory usage | Tinggi (buffer semua file) | Rendah (stream per chunk) |
| Response time | Lambat (tunggu ZIP selesai) | Cepat (langsung mulai) |
| Content-Length | Diketahui | Tidak perlu |
| File besar | Bisa timeout | Aman |

---

## Flow

### Upload
1. Client mengirim file via `multipart/form-data`
2. Handler memvalidasi file
3. Service upload ke S3 (LocalStack)
4. Return URL file yang diupload

### Streaming Download
1. Client mengirim array of URLs (JSON)
2. Server fetch file pertama dari URL via HTTP GET
3. File langsung di-compress ke ZIP stream
4. Chunk dikirim ke client (Transfer-Encoding: chunked)
5. Ulangi untuk file berikutnya
6. Client menerima ZIP lengkap setelah semua file selesai

> **Note:** Client tidak perlu tahu ukuran ZIP karena menggunakan chunked transfer encoding. Download langsung mulai tanpa menunggu.

---

## Testing

### Via Swagger UI

1. Jalankan server: `go run main.go`
2. Buka: `http://localhost:8080/swagger/`
3. Upload file
4. Download file dari S3 URL untuk melihat efek throttle

### Via CLI

```bash
awslocal s3 ls s3://my-bucket/
```

---

## Notes

* Network throttle **tidak efektif jika hanya menggunakan `localhost` tanpa Docker**
* `tc` bekerja di level kernel Linux (container)
* Cocok untuk testing:

    * retry & timeout
    * streaming download
    * UX progress bar

---

## Next Improvement

* Preset throttle via script (`tc-on.sh`, `tc-off.sh`)
* UUID filename untuk avoid collision
* Progress tracking per file dalam ZIP
* Retry & timeout untuk failed downloads
* Parallel file fetching (dengan limit)
* Custom ZIP filename dari request
