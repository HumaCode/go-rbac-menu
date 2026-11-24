# Go RBAC & Dynamic Menu System

Sistem Role-Based Access Control (RBAC) dan Manajemen Menu Dinamis (Hierarkis/Tree) berperforma tinggi.

Repo ini memiliki **2 fungsi sekaligus**:
1. **Sebagai Reusable Go Package (Library):** Dapat di-import langsung (`go get`) ke dalam proyek Golang lain (seperti proyek `ppid`, `erp`, dll) tanpa peduli framework web yang dipakai (Gin, Fiber, Echo, Breeze, atau net/http).
2. **Sebagai Aplikasi Standalone (API Server):** Berjalan dengan framework **Breeze v2** dan **PostgreSQL + GORM**.

---

## ✨ Fitur Menu Dinamis & Kategori

- **Mendukung Pengelompokan Kategori (`Category`):**
  - Menu dapat dikelompokkan berdasarkan nama kategori (contoh: `UTAMA`, `LAYANAN PPID`, `PENGATURAN SISTEM`).
- **Nesting Bertingkat Tak Terbatas (*Unlimited Depth*):**
  - `[Kategori]`
    - `[icon] Menu (Level 1)`
      - `[icon] Submenu (Level 2)`
        - `[icon] Childmenu (Level 3)`
          - `... dst`
- **Opsi Output:**
  - `GetNavMenuTree`: Mengembalikan pohon menu murni (`[]*Menu`).
  - `GetNavMenuCategorized`: Mengembalikan menu yang sudah tersusun rapi berdasarkan kategori (`[]*NavCategory`) untuk kemudahan render di sidebar UI.

---

## 📦 Menggunakan Sebagai Library di Proyek Lain (Misal: Proyek PPID)

### 1. Install Package
Di terminal folder proyek Anda (`ppid`):
```bash
go get github.com/HumaCode/go-rbac-menu
```

### 2. Auto-Migrate Tabel Menu ke Database Proyek Anda
Cukup panggil fungsi `rbacmenu.AutoMigrate` menggunakan instance `*gorm.DB` proyek Anda:

```go
package main

import (
    "github.com/HumaCode/go-rbac-menu/rbacmenu"
    "gorm.io/gorm"
)

func main() {
    // ppidDB adalah koneksi *gorm.DB dari project PPID Anda
    err := rbacmenu.AutoMigrate(ppidDB)
    if err != nil {
        panic(err)
    }
}
```
Tabel-tabel berikut akan otomatis dibuat di database proyek Anda dengan tipe ID **ULID (26 karakter)**:
- `menus` (struktur hierarkis bertingkat dengan `parent_id`, `category`, `type`)
- `permissions` (`view`, `create`, `edit`, `delete`)
- `menu_permissions` (aksi yang tersedia di tiap menu)
- `roles` (`dev`, `admin`, `user`, dll)
- `role_menu_permissions` (hak akses role terhadap aksi di menu)
- `user_roles` (relasi user ke role)

### 3. Mengambil Menu Navigasi Berdasarkan Kategori untuk Frontend
Di handler/controller proyek Anda (Gin/Fiber/Echo):

```go
import "github.com/HumaCode/go-rbac-menu/rbacmenu"

func GetUserMenuHandler(c *gin.Context) {
    menuManager := rbacmenu.New(ppidDB)
    userRoleIDs := []string{"01HQ0000000000000000000001"} 

    // Opsi A: Dapatkan menu yang sudah dikelompokkan rapi per Kategori
    categorizedMenus, err := menuManager.GetNavMenuCategorized(userRoleIDs)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "success": true,
        "data": categorizedMenus,
    })
}
```

Contoh JSON response yang diterima Frontend:
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
        "title": "Permohonan Informasi",
        "path": "/ppid/permohonan",
        "icon": "file-text",
        "children": [
          {
            "title": "Permohonan Baru",
            "path": "/ppid/permohonan/masuk",
            "icon": "inbox",
            "children": []
          },
          {
            "title": "Dalam Proses",
            "path": "/ppid/permohonan/proses",
            "icon": "clock",
            "children": []
          }
        ]
      }
    ]
  }
]
```

### 4. Spatie-Style Permission Check (`@can` & `attachMenupermission` di Go)

Package ini menyediakan fungsi yang persis dengan konsep Laravel Spatie Permission:

#### A. Cek Izin Ala `@can('action url')` (Mendukung Single Maupun Multiple):
```go
// 1. Single Permission:
if menuManager.Can(userRoleIDs, "delete /ppid/permohonan") {
    // Boleh hapus
}

// 2. Multiple Permissions ala @can(['menu /admin/user', 'menu /admin/profile'])
// Otomatis bernilai TRUE jika memiliki SALAH SATU (OR logic ala @canany):
if menuManager.Can(userRoleIDs, "menu /admin/user", "menu /admin/profile") {
    // Punya izin ke salah satu menu
}

// Bisa juga passing slice / array:
listIzin := []string{"menu /admin/user", "menu /admin/profile"}
if menuManager.Can(userRoleIDs, listIzin...) {
    // ...
}

// 3. Jika ingin memastikan HARUS PUNYA SEMUA izin (AND logic):
if menuManager.CanAll(userRoleIDs, "create /ppid/permohonan", "update /ppid/permohonan") {
    // Memiliki kedua izin sekaligus
}

// 4. Langsung menggunakan UserID:
if menuManager.CanUser(currentUserID, "menu /admin/user", "menu /admin/profile") {
    // User memiliki salah satu izin
}
```

#### B. Otomatis Generate & Assign Permission Menu (Seperti Trait `HasMenuPermission`):
```go
// Menghasilkan aksi: 'menu', 'create', 'read', 'update', 'delete' untuk menu tsb,
// sekaligus meng-assign ke role 'admin' dan 'dev'
err := menuManager.AttachMenuPermissions(menuID, []string{"menu", "create", "read", "update", "delete"}, []string{"admin", "dev"})
```

#### C. Auto Authorize Middleware (Seperti Trait `HasPermission` Interceptor):
Secara otomatis mencocokkan URL request dengan URL menu terdekat dan memetakan method HTTP:
- `GET` ➔ `read`
- `POST` ➔ `create`
- `PUT`/`PATCH` ➔ `update`
- `DELETE` ➔ `delete`

Lalu otomatis mengecek permission `"{action} {matched_url}"`!

---

## 🚀 Menjalankan Sebagai Standalone Server (Template Breeze v2)

### 1. Salin Environment
```bash
cp .env.example .env
```
Sesuaikan konfigurasi database PostgreSQL di `.env`.

### 2. Jalankan Aplikasi
```bash
go run cmd/api/main.go
```

### 3. Akun Bawaan (Default Seeder)
- `dev` / `dev123` (Role: `dev` - Akses Penuh)
- `admin` / `admin123` (Role: `admin`)
- `user` / `user123` (Role: `user`)

---

## 📡 Endpoint API Standalone Server

- `POST /api/v1/auth/login` : Login user & generate JWT.
- `GET /api/v1/nav-menu` : Mengambil **Dynamic Hierarchical Menu Tree** untuk user yang login.
- `GET /api/v1/nav-menu?grouped=true` : Mengambil **Dynamic Menu yang dikelompokkan per Kategori**.
- `GET /api/v1/menus` : Manajemen menu.
- `POST /api/v1/menus` : Buat menu baru (bisa isi `category`, `parent_id`, `order_no`).
- `POST /api/v1/menus/:id/permissions` : Hubungkan aksi ke menu (`menu_permissions`).
- `POST /api/v1/roles/menu-permissions` : Berikan izin aksi menu ke role (`role_menu_permissions`).
- `GET /api/v1/roles` : Manajemen role.
- `GET /api/v1/users` : Manajemen user.
