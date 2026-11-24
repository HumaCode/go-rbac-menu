# Go RBAC & Dynamic Menu System

Sistem Role-Based Access Control (RBAC) dan Manajemen Menu Dinamis (Hierarkis/Tree) berperforma tinggi untuk ekosistem Golang.

Proyek ini memiliki **2 fungsi sekaligus**:
1. **Sebagai Reusable Go Package (Library):** Dapat di-install dan di-import langsung (`go get`) ke dalam proyek Golang apa pun (seperti proyek `ppid`, `erp`, sistem informasi pemerintahan/kantor, dll) tanpa peduli framework web yang Anda gunakan (Gin, Fiber, Echo, Breeze, atau standar `net/http`).
2. **Sebagai Aplikasi Standalone (API Server):** Berjalan mandiri dengan framework **Breeze v2** dan **PostgreSQL + GORM**.

---

## 📥 Instalasi Package

Di dalam folder proyek Golang Anda (misal proyek `ppid`), jalankan perintah:

```bash
go get github.com/HumaCode/go-rbac-menu@v1.3.0
```

---

## 🚀 Panduan Penggunaan di Proyek Anda

### 1. Inisialisasi Database (Auto-Migration)

Cukup panggil fungsi `rbacmenu.AutoMigrate` menggunakan instance `*gorm.DB` yang sudah ada di proyek Anda:

```go
package main

import (
    "log"
    "github.com/HumaCode/go-rbac-menu/rbacmenu"
    "gorm.io/gorm"
)

func initRBAC(db *gorm.DB) {
    // Otomatis membuat tabel-tabel RBAC & Menu di database Anda
    if err := rbacmenu.AutoMigrate(db); err != nil {
        log.Fatalf("Gagal migrate tabel menu: %v", err)
    }
}
```

Tabel-tabel yang otomatis dibuat menggunakan ID berbasis **ULID (26 karakter)**:
- `menus` (struktur hierarkis bertingkat dengan `parent_id`, `category`, `type`, `order_no`, `is_active`)
- `permissions` (`view`, `create`, `read`, `update`, `delete`, dll)
- `menu_permissions` (pivot aksi permission yang tersedia di tiap menu)
- `roles` (`dev`, `admin`, `user`, dll)
- `role_menu_permissions` (hak akses role terhadap aksi di menu)
- `user_roles` (relasi user ke role)

---

### 2. Spatie-Style Permission Check (Padanan `@can` Laravel)

Package ini menyediakan fungsi pengecekan izin yang identik dengan Laravel Spatie Permission:

#### A. Single Permission Check
```go
menuManager := rbacmenu.New(db)

// Cek berdasarkan daftar role_id user:
if menuManager.Can(userRoleIDs, "delete /ppid/permohonan") {
    // User berhak menghapus permohonan
}

// Atau langsung menggunakan user_id:
if menuManager.CanUser(currentUserID, "create /ppid/permohonan") {
    // User berhak membuat permohonan baru
}
```

#### B. Multiple Permissions Check (OR Logic ala `@canany` / `@can([...])`)
Jika user memiliki **salah satu** dari izin yang ditentukan, maka bernilai `true`:

```go
// 1. Melewatkan beberapa izin berjejer (variadic):
if menuManager.Can(userRoleIDs, "menu /admin/user", "menu /admin/profile", "menu /admin/settings") {
    // User memiliki minimal salah satu dari ketiga izin di atas
}

// 2. Melewatkan dalam bentuk Slice / Array:
daftarIzin := []string{
    "menu /admin/user", 
    "menu /admin/profile",
}
if menuManager.Can(userRoleIDs, daftarIzin...) {
    // User memiliki akses ke salah satu menu
}

// 3. Menggunakan UserID langsung:
if menuManager.CanUser(currentUserID, "menu /admin/user", "menu /admin/profile") {
    // ...
}
```

#### C. Multiple Permissions Check (AND Logic: Harus Punya Semua)
Jika user diwajibkan memiliki **seluruh izin tersebut sekaligus**:

```go
if menuManager.CanAll(userRoleIDs, "create /ppid/permohonan", "update /ppid/permohonan") {
    // User HARUS punya izin 'create' DAN 'update' sekaligus
}

// Menggunakan UserID:
if menuManager.CanUserAll(currentUserID, "create /ppid/permohonan", "update /ppid/permohonan") {
    // ...
}
```

*Catatan: Jika user memiliki role `dev`, seluruh pengecekan `Can` akan otomatis bernilai `true` (superadmin bypass).*

---

### 3. Otomatis Generate & Assign Permission Menu (Seperti Trait `HasMenuPermission`)

Anda tidak perlu membuat data permission manual satu per satu. Gunakan fungsi `AttachMenuPermissions`:

```go
// Otomatis:
// 1. Membuat permission: 'menu', 'read', 'create', 'update', 'delete'
// 2. Menghubungkan ke menu ID tersebut
// 3. Meng-assign izin langsung ke role 'admin' dan 'dev'
err := menuManager.AttachMenuPermissions(
    menuID, 
    []string{"menu", "read", "create", "update", "delete"}, 
    []string{"admin", "dev"},
)
```
*Jika parameter aksi dikosongkan (`nil`), otomatis menggunakan default: `["menu", "create", "read", "update", "delete"]`.*

---

### 4. Auto-Authorize Middleware (Seperti Trait `HasPermission` Interceptor)

Middleware ini secara otomatis membaca method HTTP dan mencocokkan URL menu terdekat:
- `GET` ➔ `read`
- `POST` ➔ `create`
- `PUT` / `PATCH` ➔ `update`
- `DELETE` ➔ `delete`

Lalu otomatis memvalidasi apakah user memiliki izin `"{action} {matched_url}"`. Jika tidak, server langsung me-return `403 Forbidden`.

```go
// Contoh di Framework Breeze:
router.Use(menuManager.AutoAuthorizeMiddleware())

// Contoh di Framework Gin:
func GinAutoAuthorize(manager *rbacmenu.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRoleIDs := c.GetStringSlice("role_ids")
        method := c.Request.Method
        action := rbacmenu.DefaultAbilities[method]

        var menuURLs []string
        manager.DB().Model(&rbacmenu.Menu{}).Where("is_active = 1").Pluck("path", &menuURLs)

        matchedURL := rbacmenu.MatchMenuURL(c.Request.URL.Path, menuURLs)
        if matchedURL != "" {
            if !manager.Can(userRoleIDs, action+" "+matchedURL) {
                c.AbortWithStatusJSON(403, gin.H{"message": "Akses ditolak: " + action + " " + matchedURL})
                return
            }
        }
        c.Next()
    }
}
```

---

### 5. Mengambil Menu Navigasi Berdasarkan Kategori untuk Frontend

Menu tersusun secara dinamis dengan nesting bertingkat tak terbatas (*Unlimited Depth*):
- `[Kategori]`
  - `[icon] Menu Level 1`
    - `[icon] Submenu Level 2`
      - `[icon] Childmenu Level 3`
        - `... dst`

```go
func GetNavMenuHandler(c *gin.Context) {
    userRoleIDs := c.GetStringSlice("role_ids")
    menuManager := rbacmenu.New(db)

    // Mengambil menu yang sudah dikelompokkan rapi per Kategori:
    categories, err := menuManager.GetNavMenuCategorized(userRoleIDs)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "success": true,
        "data":    categories,
    })
}
```

#### Struktur JSON yang Diterima Frontend:
```json
[
  {
    "category": "UTAMA",
    "order_no": 1,
    "menus": [
      {
        "title": "Dashboard",
        "path": "/dashboard",
        "icon": "layout-dashboard",
        "children": []
      }
    ]
  },
  {
    "category": "LAYANAN PPID",
    "order_no": 2,
    "menus": [
      {
        "title": "Layanan Permohonan",
        "path": "/ppid",
        "icon": "folder",
        "children": [
          {
            "title": "Permohonan Informasi",
            "path": "/ppid/permohonan",
            "icon": "file-text",
            "children": [
              {
                "title": "Permohonan Baru",
                "path": "/ppid/permohonan/masuk",
                "icon": "inbox",
                "permissions": ["read", "create"]
              },
              {
                "title": "Dalam Proses",
                "path": "/ppid/permohonan/proses",
                "icon": "clock",
                "permissions": ["read"]
              }
            ]
          }
        ]
      }
    ]
  }
]
```

---

### 6. Helper Berperforma Tinggi (`Menus()` & `URLMenu()`) - Zero N+1 & In-Memory Cache

Sistem ini memiliki helper yang mengadopsi persis pola caching & in-memory filtering Laravel:

#### A. Helper `URLMenu()` (Padanan `urlMenu()` Laravel):
- Mengambil seluruh path menu aktif yang di-cache di memori (RAM).
- Otomatis diurutkan dari yang karakternya **paling panjang ke paling pendek** agar Submenu / Childmenu lebih diprioritaskan daripada Parent Menu.
```go
urls := menuManager.URLMenu()
// Hasil terurut: ["/ppid/permohonan/masuk/detail", "/ppid/permohonan/masuk", "/ppid/permohonan", "/ppid"]
```

#### B. Helper `Menus()` (Padanan `menus()` Laravel - Zero N+1 Query):
- **100% In-Memory Filtering:** Struktur menu mentah di-cache secara global di memori.
- **Single Eager Query:** Hak akses role di-load dalam **1 query tunggal**, lalu penyaringan menu, submenu, dan childmenu dilakukan murni oleh CPU di memori.
- **Otomatis Bersih:** Kategori yang semua menunya tidak boleh diakses akan otomatis disaring keluar (tidak akan muncul kategori kosong).

```go
// 1. Ambil menu lengkap (menyesuaikan dinamis: menu, submenu, hingga childmenu):
categories := menuManager.Menus(userRoleIDs...)

// 2. Atau jika HANYA ingin Menu dan Submenu saja (membatasi max depth = 2):
categories := menuManager.MenusWithMaxDepth(2, userRoleIDs...)

// 3. Atau langsung menggunakan User ID:
categories := menuManager.MenusForUser(currentUserID)
```

#### C. Cache Invalidation:
Kapan pun ada penambahan/perubahan menu atau permission, cache dapat di-clear secara instan:
```go
menuManager.ClearCache()
```

---

### 7. Penggunaan di Template HTML Go (SSR Ala Blade)

Jika Anda merender file `.html` langsung dari Golang:

```go
funcMap := template.FuncMap{
    "can": func(abilities ...string) bool {
        return menuManager.Can(userRoleIDs, abilities...)
    },
}
tmpl := template.Must(template.New("index.html").Funcs(funcMap).ParseFiles("index.html"))
```

Di file `.html`:
```html
<!-- Single Permission -->
{{ if can "create /ppid/permohonan" }}
    <a href="/ppid/permohonan/create" class="btn btn-primary">+ Tambah</a>
{{ end }}

<!-- Multiple Permissions (Akan tampil jika memiliki salah satu izin) -->
{{ if can "menu /admin/user" "menu /admin/profile" }}
    <li class="nav-item">Pengaturan Akun</li>
{{ end }}
```

---

## 🖥️ Menjalankan Sebagai Standalone Server (Template Breeze v2)

Jika Anda ingin menjalankan repository ini sebagai REST API server mandiri:

### 1. Setup Environment
```bash
cp .env.example .env
```
Sesuaikan konfigurasi database PostgreSQL di `.env`:
```env
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_rbac_menu
DB_SSLMODE=disable
```

### 2. Jalankan Server
```bash
go run cmd/api/main.go
```
*GORM otomatis membuat skema tabel dan mengisi data awal (seeder).*

### 3. Akun Bawaan (Default Seeder)
- `dev` / `dev123` (Role: `dev` - Akses Penuh / Superadmin)
- `admin` / `admin123` (Role: `admin`)
- `user` / `user123` (Role: `user`)

---

## 📡 Endpoint API Standalone Server

- `POST /api/v1/auth/login` : Login user & generate JWT.
- `GET /api/v1/nav-menu` : Mengambil pohon menu pengguna.
- `GET /api/v1/nav-menu?grouped=true` : Mengambil menu yang dikelompokkan berdasarkan kategori.
- `GET /api/v1/menus` : Manajemen daftar menu.
- `POST /api/v1/menus` : Tambah menu baru (`category`, `parent_id`, `order_no`).
- `POST /api/v1/menus/:id/permissions` : Pasang aksi permission ke menu.
- `POST /api/v1/roles/menu-permissions` : Berikan izin aksi menu ke role.
- `GET /api/v1/roles` : Manajemen role.
- `GET /api/v1/users` : Manajemen pengguna.
