# Go RBAC & Dynamic Menu System

Sistem Role-Based Access Control (RBAC) dan Manajemen Menu Dinamis (Hierarkis/Tree) berperforma tinggi.

Repo ini memiliki **2 fungsi sekaligus**:
1. **Sebagai Reusable Go Package (Library):** Dapat di-import langsung (`go get`) ke dalam proyek Golang lain (seperti proyek `ppid`, `erp`, dll) tanpa peduli framework web yang dipakai (Gin, Fiber, Echo, Breeze, atau net/http).
2. **Sebagai Aplikasi Standalone (API Server):** Berjalan dengan framework **Breeze v2** dan **PostgreSQL + GORM**.

---

## 📦 Menggunakan Sebagai Library di Proyek Lain (Misal: Proyek PPID)

Jika Anda sudah memiliki proyek Golang (misal `ppid`) dan ingin mengimplementasikan sistem menu hierarkis ini:

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
- `menus` (struktur hierarkis bertingkat dengan `parent_id`)
- `permissions` (`view`, `create`, `edit`, `delete`)
- `menu_permissions` (aksi yang tersedia di tiap menu)
- `roles` (`dev`, `admin`, `user`, dll)
- `role_menu_permissions` (hak akses role terhadap aksi di menu)
- `user_roles` (relasi user ke role)

### 3. Mengambil Menu Navigasi (Tree) untuk User yang Login
Di handler/controller proyek Anda (Gin/Fiber/Echo):

```go
import "github.com/HumaCode/go-rbac-menu/rbacmenu"

func GetUserMenuHandler(c *gin.Context) {
    // Inisialisasi manager RBAC Menu
    menuManager := rbacmenu.New(ppidDB)

    // Ambil daftar role_id milik user yang sedang login
    userRoleIDs := []string{"01HQ0000000000000000000001"} 

    // Ambil struktur Tree Menu (Hanya yang is_active = 1)
    navTree, err := menuManager.GetNavMenuTree(userRoleIDs)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Mengembalikan JSON Tree siap pakai untuk sidebar frontend
    c.JSON(200, gin.H{
        "success": true,
        "data": navTree,
    })
}
```

### 4. Mengecek Permission Aksi di Endpoint Bisnis PPID
```go
// Contoh middleware di project PPID:
hasAccess, err := menuManager.CheckPermission(userRoleIDs, "/ppid/permohonan", "delete")
if !hasAccess {
    // Tolak akses (403 Forbidden)
}
```

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
- `GET /api/v1/menus` : Manajemen menu.
- `POST /api/v1/menus` : Buat menu baru.
- `POST /api/v1/menus/:id/permissions` : Hubungkan aksi ke menu (`menu_permissions`).
- `POST /api/v1/roles/menu-permissions` : Berikan izin aksi menu ke role (`role_menu_permissions`).
- `GET /api/v1/roles` : Manajemen role.
- `GET /api/v1/users` : Manajemen user.
