package domain

type MenuPermission struct {
	BaseModel
	MenuID       string `gorm:"type:varchar(26);not null;index" json:"menu_id"`
	PermissionID string `gorm:"type:varchar(26);not null;index" json:"permission_id"`
	IsActive     int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	// Relations
	Menu       Menu       `gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE" json:"menu,omitempty"`
	Permission Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnDelete:CASCADE" json:"permission,omitempty"`
}

func (MenuPermission) TableName() string {
	return "menu_permissions"
}

