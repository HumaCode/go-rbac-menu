package domain

type Permission struct {
	BaseModel
	Code        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"` // 'view', 'create', 'edit', 'delete'
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	IsActive    int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive
}

func (Permission) TableName() string {
	return "permissions"
}

