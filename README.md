# Go RBAC & Dynamic Menu System

Sistem Role-Based Access Control (RBAC) dan Manajemen Menu Dinamis (Hierarkis/Tree) berperforma tinggi yang dibangun menggunakan framework **Breeze v2** (event-driven worker pool) dan **GORM** dengan database **PostgreSQL**.

---

## 🚀 Fitur Utama

- **Framework Breeze v2 (`github.com/nelthaarion/breeze/v2`):**
  - Menggunakan arsitektur event-loop dan worker pool.
  - Penanganan route database menggunakan `HandleBlocking` sehingga tidak menghambat (stall) event reactor.
- **ORM & Database:**
  - GORM dengan PostgreSQL.
  - Auto-migration otomatis saat aplikasi dijalankan.
- **ID berbasis ULID (Universally Unique Lexicographically Sortable Identifier):**
  - 26 karakter string (Base32), sortable secara waktu, ramah B-Tree index PostgreSQL.
  - Auto-generate via GORM `BeforeCreate` hook.
- **Aturan Ketat `is_active = 1`:**
  - Semua entitas (`users`, `roles`, `menus`, `permissions`, `menu_permissions`, `role_menu_permissions`) harus bernilai `is_active = 1` agar diakui dan dapat diakses. Item dengan `is_active = 0` otomatis diabaikan.
- **Role Standar:**
  - `dev`: Akses penuh (bypass permission check).
  - `admin`: Pengelolaan sistem operasional.
  - `user`: Pengguna standar.
- **Hierarki Menu Dinamis (Tree):**
  - Mendukung struktur bertingkat (*Parent-Child*) tanpa batas kedalaman.
  - Tiap item menu menyertakan permission/action yang berhak diakses oleh user yang sedang login (contoh: `["view", "create", "edit", "delete"]`).
  - Endpoint `GET /api/v1/nav-menu` mengembalikan tree menu yang siap dikonsumsi oleh frontend (Vue/React/Angular/Next.js/Nuxt).
- **Tabel Penghubung Granular:**
  - `menu_permissions`: Mengikat menu dengan daftar aksi permission.
  - `role_menu_permissions`: Menentukan aksi spesifik yang dimiliki role untuk tiap menu.

---

## 📁 Struktur Direktori

```text
go-rbac-menu/
├── cmd/
│   └── api/
│       └── main.go                  # Entry point aplikasi
├── config/
│   ├── config.go                    # Konfigurasi environment (.env)
│   └── database.go                  # Setup koneksi GORM & PostgreSQL
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/             # Auth, Menu, Role, User Handlers
│   │       ├── middleware/          # JWT Auth & RBAC Permission Guards
│   │       └── router/              # Breeze router & route grouping
│   ├── domain/                      # Struct entitas (User, Role, Menu, dll)
│   ├── repository/                  # Query database & GORM operations
│   ├── seeder/                      # Initial seeder (roles, menus, default accounts)
│   └── service/                     # Business logic (Nav-menu tree builder, Auth)
├── pkg/
│   ├── response/                    # Standard format response JSON
│   └── utils/                       # ULID, Bcrypt Password, JWT Helper
├── .env.example
├── .gitignore
├── go.mod
└── go.sum
```

---

## 🛠️ Persiapan & Menjalankan Aplikasi

### 1. Salin Environment
```bash
cp .env.example .env
```
Sesuaikan konfigurasi koneksi database PostgreSQL di `.env`:
```env
APP_PORT=8080
APP_ENV=development

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_rbac_menu
DB_SSLMODE=disable

JWT_SECRET=super-secret-jwt-key-change-in-production
JWT_EXPIRE_HOURS=24
```

### 2. Buat Database di PostgreSQL
Pastikan database sudah dibuat di PostgreSQL Anda:
```sql
CREATE DATABASE go_rbac_menu;
```

### 3. Jalankan Aplikasi
```bash
go run cmd/api/main.go
```
*Catatan: Saat aplikasi pertama kali berjalan, GORM akan otomatis melakukan migrasi tabel dan menjalankan seeder data awal.*

---

## 🔑 Akun Default (Hasil Seeder)

| Username | Password | Role | Akses Menu |
|---|---|---|---|
| `dev` | `dev123` | `dev` | Akses penuh seluruh menu & aksi |
| `admin` | `admin123` | `admin` | Dashboard, User Mgmt (CRUD), Role Mgmt (View), Menu Mgmt (View) |
| `user` | `user123` | `user` | Dashboard (View) |

---

## 📡 Daftar Endpoint API Utama

### Public:
- `GET /health` : Cek status server.
- `POST /api/v1/auth/login` : Login user & generate JWT token.
- `POST /api/v1/auth/register` : Registrasi user baru (default role: `user`).

### Protected (Membutuhkan Header `Authorization: Bearer <token>`):
- `GET /api/v1/auth/profile` : Mengambil data user yang sedang login.
- `GET /api/v1/nav-menu` : **Dynamic Hierarchical Menu Tree** untuk sidebar navigasi frontend.

#### Menu Management:
- `GET /api/v1/menus` : Daftar semua menu aktif.
- `GET /api/v1/menus/:id` : Detail menu.
- `POST /api/v1/menus` : Buat menu baru.
- `PUT /api/v1/menus/:id` : Update menu.
- `DELETE /api/v1/menus/:id` : Nonaktifkan menu (`is_active = 0`).
- `POST /api/v1/menus/:id/permissions` : Hubungkan permission ke menu (`menu_permissions`).
- `GET /api/v1/menus/:id/permissions` : Daftar permission yang tersedia di menu.

#### Role Management:
- `GET /api/v1/roles` : Daftar role aktif.
- `POST /api/v1/roles` : Buat role baru.
- `PUT /api/v1/roles/:id` : Update role.
- `DELETE /api/v1/roles/:id` : Nonaktifkan role.
- `POST /api/v1/roles/menu-permissions` : Berikan hak akses menu permission ke role (`role_menu_permissions`).
- `DELETE /api/v1/roles/menu-permissions` : Cabut hak akses menu permission dari role.

#### User Management:
- `GET /api/v1/users` : Daftar user aktif.
- `GET /api/v1/users/:id` : Detail user.
- `POST /api/v1/users/:id/roles` : Pasang role ke user.
- `DELETE /api/v1/users/:id/roles` : Cabut role dari user.
- `PATCH /api/v1/users/:id/status` : Aktifkan / nonaktifkan user (`is_active = 1 / 0`).
