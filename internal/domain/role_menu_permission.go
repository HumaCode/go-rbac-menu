package domain

type RoleMenuPermission struct {
	BaseModel
	RoleID           string `gorm:"type:varchar(26);not null;index" json:"role_id"`
	MenuPermissionID string `gorm:"type:varchar(26);not null;index" json:"menu_permission_id"`
	IsActive         int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	// Relations
	Role           Role           `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE" json:"role,omitempty"`
	MenuPermission MenuPermission `gorm:"foreignKey:MenuPermissionID;references:ID;constraint:OnDelete:CASCADE" json:"menu_permission,omitempty"`
}

func (RoleMenuPermission) TableName() string {
	return "role_menu_permissions"
}

