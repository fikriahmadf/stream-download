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

## Load Testing dengan k6

### Prerequisites

Install k6:

```bash
# macOS
brew install k6

# Linux
sudo apt install k6

# atau download dari https://k6.io/docs/getting-started/installation/
```

### Memory Monitoring

Aplikasi menyediakan endpoint untuk monitoring memory secara real-time:

1. **Memory Stats API**: `GET /debug/memory`
   ```bash
   curl http://localhost:8080/debug/memory
   ```

2. **Memory Monitor UI**: `http://localhost:8080/monitor.html`
    - Grafik real-time untuk Alloc, Heap, Sys Memory
    - Goroutines & GC Cycles tracking
    - Configurable time window (1m, 5m, 10m, 30m)

3. **pprof Profiling**: `http://localhost:8080/debug/pprof/`
   ```bash
   go tool pprof http://localhost:8080/debug/pprof/heap
   ```

### Menjalankan Load Test

1. **Jalankan server**
   ```bash
   go run main.go
   ```

2. **Buka Memory Monitor** (opsional)
   ```
   http://localhost:8080/monitor.html
   ```
   Klik "Start Monitoring" untuk melihat grafik memory real-time.

3. **Jalankan k6 load test**
   ```bash
   k6 run k6/load-test.js
   ```

### Konfigurasi Load Test

Script `k6/load-test.js` menggunakan skenario berikut:

| Stage | Durasi | Virtual Users | Tujuan |
|-------|--------|---------------|--------|
| Ramp up | 30s | 0 → 10 | Pemanasan |
| Steady | 1m | 10 | Stabilitas beban normal |
| Ramp up | 30s | 10 → 20 | Naikkan beban |
| Steady | 1m | 20 | Stabilitas beban tinggi |
| Ramp down | 30s | 20 → 0 | Turun bertahap |

**Total durasi: ~3.5 menit**

### Custom Base URL

```bash
k6 run -e BASE_URL=http://192.168.1.100:8080 k6/load-test.js
```

### Hasil Load Test

Hasil test otomatis disimpan di folder `k6/results/`:

| File | Format | Isi |
|------|--------|-----|
| `summary-{timestamp}.json` | JSON | Data mentah untuk analisis |
| `summary-{timestamp}.txt` | Text | Summary readable |

### Interpretasi Memory Stats

| Metric | Penjelasan |
|--------|------------|
| **Alloc** | Memory yang sedang dialokasikan (aktif) |
| **Heap Alloc** | Heap memory yang digunakan |
| **Sys Memory** | Total memory dari OS |
| **Total Alloc** | Total memory yang pernah dialokasikan (kumulatif) |
| **GC Cycles** | Jumlah Garbage Collection yang sudah berjalan |
| **Goroutines** | Jumlah goroutine aktif |

**Tanda tidak ada memory leak:**
- Alloc/Heap kembali rendah setelah test selesai
- Goroutines kembali ke jumlah baseline
- GC Cycles bertambah (menunjukkan GC aktif membersihkan memory)


---

## Notes

* Network throttle **tidak efektif jika hanya menggunakan `localhost` tanpa Docker**
* `tc` bekerja di level kernel Linux (container)
* Cocok untuk testing:

    * retry & timeout
    * streaming download
    * UX progress bar

---



## Example S3 Data
<img width="1274" height="868" alt="Screenshot 2025-12-24 at 18 09 27" src="https://github.com/user-attachments/assets/6b546480-9919-4de5-a0da-0cfdb84316d0" />

## Example Download Processing
https://github.com/user-attachments/assets/1e5883b1-e84c-4d9d-bbea-d31c3a3918e7
https://github.com/user-attachments/assets/2dde4e50-f758-435d-aef6-10b9b58ff6bb

## Proof Of Load Test
<img width="1010" height="952" alt="Screenshot 2025-12-24 at 20 42 21" src="https://github.com/user-attachments/assets/0b70b5d0-745f-4199-a803-6c8aec93a3fe" />




