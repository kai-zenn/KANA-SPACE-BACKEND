# KANA Backend

Backend Go untuk KANA, platform Waste-to-Value yang menghubungkan penghasil limbah industri kreatif dengan pengrajin dan pembeli di sekitarnya. Aplikasi ini menyediakan autentikasi, KANA Space, marketplace Lapak KANA, percakapan dan penawaran harga, transaksi, push notification, serta integrasi ke layanan NLP eksternal untuk matching material.

Dokumen ini menjelaskan kondisi kode yang saat ini terpasang. Beberapa bagian masih bersifat parsial dan dicatat secara eksplisit di bagian [Batasan yang diketahui](#batasan-yang-diketahui).

## Ringkasan teknis

- Bahasa: Go 1.25.1
- HTTP framework: Gin
- Database: PostgreSQL melalui GORM
- Autentikasi: JWT dan bcrypt
- ID: UUID
- File storage: filesystem lokal pada `uploads/v1/photos`
- Push notification: Firebase Cloud Messaging
- Matching: HTTP client ke service NLP pada endpoint `/v1/embed` dan `/v1/match`
- Scheduler: `robfig/cron`, auto-expire transaksi setiap 10 menit
- Container: Docker dan Docker Compose

## Fitur yang tersedia

### User dan autentikasi

- Register dengan nama, username, nomor telepon, email, dan password.
- Login username/password.
- Endpoint login Google sudah terdaftar, tetapi verifier Google belum di-wire di `cmd/main.go`, sehingga belum dapat digunakan pada konfigurasi sekarang.
- Profil publik berdasarkan username.
- Update profil, password, dan foto profil.
- Upgrade akun menjadi seller.
- Follow dan unfollow user.
- JWT membawa identitas user dan role: `user`, `seller`, atau `admin`.

### KANA Space

- Membuat posting dengan maksimal empat gambar.
- Tag yang tersedia: `CariMaterial`, `PajangKarya`, `TipsTrick`, `DapurHijau`, `Diskusi`, `KabarKomunitas`, dan `Lifestyle`.
- Feed dengan filter tag dan cursor pagination.
- Like/unlike idempotent.
- Komentar flat dengan cursor pagination.
- Post `CariMaterial` wajib menyertakan latitude dan longitude.
- Post `CariMaterial` memicu proses embedding dan matching secara asynchronous.

### Lapak

- Kategori bertingkat dengan dua branch: `RAW_MATERIAL` dan `FINISHED_GOODS`.
- Membuat, mengubah, dan menghapus produk.
- Listing type: `HIBAH`, `JUAL_BORONGAN`, dan `DIJUAL`.
- Maksimal empat gambar produk.
- Filter produk berdasarkan branch, kategori, harga, lokasi, radius, dan cursor.
- Pencarian produk di sekitar lokasi.
- Transaksi langsung.
- Checkout berdasarkan offer yang sudah diterima seller.
- Alur transaksi berbeda untuk bahan mentah dan barang jadi.

### Chat dan negosiasi

- Satu conversation unik untuk kombinasi produk dan buyer.
- Pesan `TEXT` dan `OFFER`.
- Offer mempunyai harga dan status `PENDING`, `ACCEPTED`, atau `REJECTED`.
- Seller dapat menerima atau menolak offer.
- Offer hanya dapat dibuat pada produk yang dapat dinegosiasikan, bukan `HIBAH`.

### Notifikasi

- Register device token untuk platform `android`, `ios`, atau `web`.
- Melihat notifikasi milik user.
- Menandai notifikasi sebagai sudah dibaca.
- Matching dan event bisnis dapat mengirim push notification melalui FCM.

## Arsitektur dan struktur kode

Repository ini menggunakan modular monolith. Semua modul berjalan dalam satu binary dan satu database, tetapi domain dipisahkan dalam package berikut:

```text
cmd/main.go                         bootstrap aplikasi
internal/
  adapters/                         adapter lintas modul
  configs/                          pembacaan .env dan environment
  database/                         migrasi dan seeding
  middlewares/                      autentikasi dan otorisasi
  modules/
    user/                           akun dan autentikasi
    space/                          posting, feed, like, komentar
    lapak/                          kategori, produk, matching, transaksi
    chat/                           conversation, pesan, offer
    notification/                   notifikasi dan device token
  pkgs/
    bcrypt/                         hashing password
    fcm/                            Firebase Cloud Messaging
    geo/                            utilitas geografis
    google/                         package Google, saat ini belum berisi verifier
    jwt/                            pembuatan dan validasi JWT
    nlpclient/                      client HTTP service NLP
    storage/                        upload file lokal
  rest/                             wiring dependency dan route Gin
  workers/                          scheduler dan auto-cancel transaksi
uploads/v1/photos/                  file upload lokal
```

Konvensi modul umumnya adalah `entity.go` untuk model database, `model.go` untuk DTO, `repository.go` untuk akses data, `usecase*.go` untuk business logic, dan `handler.go` untuk HTTP layer. Dependensi lintas modul dipasang melalui adapter di `internal/adapters` untuk menghindari import domain yang saling mengikat.

## Menjalankan secara lokal

### Prasyarat

- Go 1.25.1 atau kompatibel.
- PostgreSQL yang dapat diakses aplikasi.
- File service account Firebase untuk fitur notifikasi dan startup aplikasi.
- Service NLP opsional untuk embedding dan ranking matching.

### Konfigurasi

Salin `.env.example` menjadi `.env`, lalu sesuaikan nilainya. Konfigurasi yang dibaca aplikasi adalah:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=my_db

JWT_SECRET=ganti-dengan-secret-yang-kuat
JWT_EXPIRY=30d

FIREBASE_CREDENTIALS_PATH=sesuaikan-dengan-adminsdk-kalian

NLP_BASE_URL=http://localhost:8001
NLP_API_KEY=
```

Catatan konfigurasi:

- `JWT_EXPIRY` menerima format hari seperti `30d` atau format duration Go seperti `24h`.
- `FIREBASE_CREDENTIALS_PATH` harus menunjuk ke file JSON service account yang tersedia dari working directory aplikasi. Tanpa file ini server berhenti saat startup.
- `NLP_BASE_URL` digunakan sebagai prefix untuk request ke `POST <NLP_BASE_URL>/v1/embed` dan `POST <NLP_BASE_URL>/v1/match`.
- Server saat ini listen pada port `9090` secara hard-coded. Nilai `PORT` di `.env.example` belum digunakan oleh bootstrap.
- `DB_URL`, `SSL_MODE`, dan `CORS_ORIGIN` yang ada di `.env.example` belum digunakan oleh kode saat ini.
- Jangan commit secret JWT, API key, atau file credential Firebase ke repository produksi.

### Migrasi, seeding, dan server

Jalankan dari root repository:

```bash
go mod download
go run cmd/main.go migrate
go run cmd/main.go seed
go run cmd/main.go
```

Perintah:

- `migrate`: menjalankan GORM `AutoMigrate` untuk seluruh entity.
- `seed`: membuat admin, kategori, user contoh, produk contoh, dan post contoh.
- tanpa argumen: menyalakan HTTP server pada `http://localhost:9090`.

Database harus sudah ada sebelum perintah dijalankan; aplikasi tidak membuat database PostgreSQL secara otomatis.

### Menjalankan test

```bash
go test ./...
```

Test yang tersedia saat ini terutama mencakup usecase user. Package lain tetap dikompilasi oleh perintah tersebut, tetapi belum mempunyai cakupan integration test yang setara.

## Menjalankan dengan Docker Compose

`docker-compose.yaml` menyediakan PostgreSQL dan service aplikasi:

```bash
docker compose up --build
```

Compose memetakan aplikasi ke port `9090` dan volume `./uploads` ke container. PostgreSQL menggunakan variable `DB_USER`, `POSTGRES_PASSWORD`, dan `DB_NAME` untuk inisialisasi container, sedangkan aplikasi menggunakan `DB_PASSWORD`. Pastikan variable yang dibutuhkan kedua service tersedia di `.env`; khususnya `POSTGRES_PASSWORD` perlu didefinisikan untuk container PostgreSQL.

Setelah database siap, migrasi dan seeding dapat dijalankan di container aplikasi:

```bash
docker compose run --rm app ./main migrate
docker compose run --rm app ./main seed
```

File credential Firebase dan `.env` juga harus tersedia di lokasi yang digunakan oleh konfigurasi container. Dockerfile hanya menyalin binary dan folder konfigurasi; ia tidak otomatis menyalin file `.env` atau credential Firebase.

## API

Base URL lokal adalah `http://localhost:9090/api`. Tidak ada prefix `/api/v1` pada router yang terpasang sekarang. Endpoint yang berada di bawah `/user`, `/lapak` yang dilindungi, `/space`, `/chat`, dan `/notifications` memerlukan header:

```http
Authorization: Bearer <jwt>
```

Response sukses dan error diformat oleh handler sebagai object JSON dengan field status/message/data jika handler tersebut menggunakannya. Bentuk error dan HTTP status harus diperlakukan sebagai kontrak endpoint, bukan diasumsikan selalu sama untuk semua handler.

### Auth

| Method | Path                 | Auth               |
| ------ | -------------------- | ------------------ |
| POST   | `/api/auth/register` | Tidak              |
| POST   | `/api/auth/login`    | Tidak              |
| POST   | `/api/auth/google`   | Tidak, belum aktif |

### User dan device

| Method | Path                          | Auth          |
| ------ | ----------------------------- | ------------- |
| GET    | `/api/user/profile/:username` | Ya            |
| PATCH  | `/api/user/profile`           | Ya            |
| PUT    | `/api/user/profile/password`  | Ya            |
| POST   | `/api/user/profile/photo`     | Ya, multipart |
| POST   | `/api/user/upgrade`           | Ya            |
| POST   | `/api/user/:id/follow`        | Ya            |
| POST   | `/api/user/:id/unfollow`      | Ya            |
| POST   | `/api/user/devices`           | Ya            |

### Lapak

| Method | Path                                     | Auth               |
| ------ | ---------------------------------------- | ------------------ |
| GET    | `/api/lapak/products`                    | Tidak              |
| GET    | `/api/lapak/products/:id`                | Tidak              |
| GET    | `/api/lapak/categories`                  | Tidak              |
| GET    | `/api/lapak/products/nearby`             | Tidak              |
| POST   | `/api/lapak/products`                    | Ya, multipart      |
| PATCH  | `/api/lapak/products/:id`                | Ya                 |
| DELETE | `/api/lapak/products/:id`                | Ya                 |
| POST   | `/api/lapak/products/:id/transactions`   | Ya                 |
| POST   | `/api/lapak/transactions/checkout-offer` | Ya                 |
| POST   | `/api/lapak/transactions/:id/confirm`    | Ya, finished goods |
| POST   | `/api/lapak/transactions/complete-qr`    | Ya, raw material   |
| POST   | `/api/lapak/transactions/:id/complete`   | Ya, finished goods |
| POST   | `/api/lapak/transactions/:id/cancel`     | Ya                 |

Query list produk mendukung `branch`, `category_id`, `min_price`, `max_price`, `cursor`, `limit`, `lat`, `lng`, dan `radius_meters`. Endpoint nearby menggunakan `lat`, `lng`, `radius_meters`, `limit`, dan `offset`.

### Space

| Method | Path                                    | Auth          |
| ------ | --------------------------------------- | ------------- |
| POST   | `/api/space/posts`                      | Ya, multipart |
| GET    | `/api/space/posts`                      | Ya            |
| DELETE | `/api/space/posts/:id`                  | Ya            |
| POST   | `/api/space/posts/:id/like`             | Ya            |
| POST   | `/api/space/posts/:id/unlike`           | Ya            |
| POST   | `/api/space/posts/:id/comments`         | Ya            |
| GET    | `/api/space/posts/:id/comments`         | Ya            |
| DELETE | `/api/space/posts/comments/:comment_id` | Ya            |

Feed mendukung query `tag`, `cursor`, dan `limit`. Feed dan komentar memakai cursor berbasis timestamp RFC3339. Batas default feed adalah 15 item dan maksimum 20 item.

### Chat

| Method | Path                                         | Auth |
| ------ | -------------------------------------------- | ---- |
| GET    | `/api/chat/conversations`                    | Ya   |
| POST   | `/api/chat/products/:productId/conversation` | Ya   |
| POST   | `/api/chat/conversations/:id/messages`       | Ya   |
| GET    | `/api/chat/conversations/:id/messages`       | Ya   |
| POST   | `/api/chat/messages/:id/respond-offer`       | Ya   |

Pesan `OFFER` wajib mengisi `offer_price`. History message memakai `cursor` dan `limit`; default message limit adalah 10 dan maksimum 20.

### Notifications

| Method | Path                          | Auth |
| ------ | ----------------------------- | ---- |
| GET    | `/api/notifications`          | Ya   |
| PATCH  | `/api/notifications/:id/read` | Ya   |

File upload dapat diakses melalui static route `/uploads/*path`, misalnya URL yang dikembalikan storage lokal.

## Alur bisnis utama

### Matching `CariMaterial`

1. User membuat post dengan tag `CariMaterial` dan koordinat.
2. Post disimpan dan embedding diminta secara asynchronous melalui service NLP.
3. Matching mencari kandidat produk `RAW_MATERIAL` yang masih `AVAILABLE` dalam radius tetap 20 km.
4. Kandidat dikirim ke service NLP dengan konfigurasi BM25 top-k 10, final top-k 3, dan semantic threshold 0.80.
5. Hasil disimpan ke tabel `matches` dan notifikasi dikirim ke pembuat post.
6. Match `HIBAH` mengarahkan user ke klaim langsung; listing lain mengarahkan user ke chat.

Jika service NLP gagal, kode masuk ke fungsi fallback keyword search, tetapi implementasi fallback saat ini belum menghasilkan match. Karena itu NLP service diperlukan agar fitur matching menghasilkan rekomendasi.

### Transaksi raw material

- `RAW_MATERIAL` selalu diproses sebagai satu lot, sehingga quantity menjadi 1.
- Listing dikunci ke status `LOCKED` ketika checkout.
- Jarak sampai 4 km menggunakan `SELF_PICKUP` dan koordinat meetup dibulatkan untuk mengurangi presisi lokasi asli.
- Jarak di atas 4 km menggunakan `3PL`.
- QR code dibuat saat checkout dan transaksi kedaluwarsa setelah 24 jam.
- Seller menyelesaikan transaksi melalui endpoint `complete-qr`.
- Scheduler menjalankan expire worker setiap 10 menit.
- Checkout dari chat hanya dapat dilakukan menggunakan offer berstatus `ACCEPTED` oleh buyer pemilik offer.

### Transaksi finished goods

- Quantity mengikuti request dan harus tersedia di stock.
- Total harga adalah `price * quantity`.
- Status awal `PENDING` dan logistic type `ARRANGED_OFFLINE`.
- Seller mengonfirmasi transaksi menjadi `CONFIRMED`.
- Buyer atau seller dapat menyelesaikan transaksi secara manual.
- Pembatalan mengembalikan stock.

## Database

`migrate` membuat tabel untuk entity berikut:

- `users`
- `posts`, `post_images`, `comments`, `post_likes`
- `categories`, `products`, `product_images`, `matches`, `transactions`
- `conversations`, `messages`
- `notifications`, `user_devices`

Kategori menggunakan `parent_id` untuk struktur bertingkat. Product dan post dapat menyimpan embedding sebagai array `float8[]`. Conversation unik berdasarkan kombinasi `product_id` dan `buyer_id`, sedangkan post like unik berdasarkan kombinasi `post_id` dan `user_id`.

## Batasan yang diketahui

- Google Login belum berfungsi karena implementasi verifier belum tersedia dan `nil` dikirim saat dependency wiring.
- Fallback matching ketika NLP gagal masih berupa placeholder dan mengembalikan hasil kosong.
- Server dan scheduler belum menerima konfigurasi port atau interval dari environment; port server adalah `9090` dan interval worker adalah 10 menit.
- Storage upload masih local filesystem, sehingga deployment multi-instance membutuhkan shared storage atau object storage.
- Tidak ada migration versioning; skema dikelola dengan GORM `AutoMigrate`.
- Test end-to-end untuk PostgreSQL, Firebase, NLP, upload, transaksi, dan scheduler belum tersedia.
- Kredensial Google/Firebase dan validasi keamanan production perlu dilengkapi sebelum deployment publik.

## Arah pengembangan berikutnya

Prioritas teknis yang paling dekat dengan kondisi repository:

1. Implementasi dan wire Google ID token verifier.
2. Menambahkan test repository dan integration test untuk state machine transaksi.
3. Memindahkan konfigurasi port, worker interval, radius matching, dan threshold NLP ke config.
4. Menambahkan migration versioning, storage eksternal, dan observability untuk deployment production.

## Lisensi dan kontribusi

Ikuti struktur modular yang sudah ada ketika menambah fitur. Logic lintas domain harus melalui interface adapter, akses database berada di repository, dan error bisnis yang membutuhkan HTTP status tertentu sebaiknya menggunakan sentinel error di usecase terkait.
