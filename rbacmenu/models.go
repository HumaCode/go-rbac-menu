package rbacmenu

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel provides ULID primary key and timestamps
type BaseModel struct {
	ID        string    `gorm:"type:varchar(26);primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = NewULID()
	}
	return nil
}

// Role entity (e.g. 'dev', 'admin', 'user')
type Role struct {
	BaseModel
	Code        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	IsActive    int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive
}

func (Role) TableName() string {
	return "roles"
}

// UserRole pivot
type UserRole struct {
	BaseModel
	UserID string `gorm:"type:varchar(26);not null;index" json:"user_id"`
	RoleID string `gorm:"type:varchar(26);not null;index" json:"role_id"`

	Role Role `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE" json:"role,omitempty"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// Permission action catalog (e.g. 'view', 'create', 'edit', 'delete')
type Permission struct {
	BaseModel
	Code        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	IsActive    int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive
}

func (Permission) TableName() string {
	return "permissions"
}

// Menu entity supporting recursive parent-child hierarchy
type Menu struct {
	BaseModel
	ParentID *string `gorm:"type:varchar(26);index" json:"parent_id"` // null for root menus
	Title    string  `gorm:"type:varchar(100);not null" json:"title"`
	Path     string  `gorm:"type:varchar(150);not null" json:"path"`
	Icon     string  `gorm:"type:varchar(100)" json:"icon"`
	OrderNo  int     `gorm:"type:int;default:0;not null" json:"order_no"`
	IsActive int     `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	// Relations & Dynamic Tree fields
	Parent          *Menu            `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:SET NULL" json:"parent,omitempty"`
	Children        []*Menu          `gorm:"-" json:"children,omitempty"`
	MenuPermissions []MenuPermission `gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE" json:"menu_permissions,omitempty"`
	Permissions     []string         `gorm:"-" json:"permissions,omitempty"` // populated during tree generation
}

func (Menu) TableName() string {
	return "menus"
}

// MenuPermission binds available actions to a specific menu
type MenuPermission struct {
	BaseModel
	MenuID       string `gorm:"type:varchar(26);not null;index" json:"menu_id"`
	PermissionID string `gorm:"type:varchar(26);not null;index" json:"permission_id"`
	IsActive     int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	Menu       Menu       `gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE" json:"menu,omitempty"`
	Permission Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnDelete:CASCADE" json:"permission,omitempty"`
}

func (MenuPermission) TableName() string {
	return "menu_permissions"
}

// RoleMenuPermission assigns specific menu actions to a role
type RoleMenuPermission struct {
	BaseModel
	RoleID           string `gorm:"type:varchar(26);not null;index" json:"role_id"`
	MenuPermissionID string `gorm:"type:varchar(26);not null;index" json:"menu_permission_id"`
	IsActive         int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	Role           Role           `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE" json:"role,omitempty"`
	MenuPermission MenuPermission `gorm:"foreignKey:MenuPermissionID;references:ID;constraint:OnDelete:CASCADE" json:"menu_permission,omitempty"`
}

func (RoleMenuPermission) TableName() string {
	return "role_menu_permissions"
}
