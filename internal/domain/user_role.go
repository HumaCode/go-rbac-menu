package domain

type UserRole struct {
	BaseModel
	UserID string `gorm:"type:varchar(26);not null;index" json:"user_id"`
	RoleID string `gorm:"type:varchar(26);not null;index" json:"role_id"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Role Role `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

