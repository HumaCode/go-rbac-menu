package domain

type Role struct {
	BaseModel
	Code        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"` // 'dev', 'admin', 'user'
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	IsActive    int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	// Relations
	Users           []User           `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:UserID" json:"users,omitempty"`
	MenuPermissions []MenuPermission `gorm:"many2many:role_menu_permissions;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:MenuPermissionID" json:"menu_permissions,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

