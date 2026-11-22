package domain

type User struct {
	BaseModel
	Username     string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	IsActive     int    `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	// Relations
	Roles []Role `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID" json:"roles,omitempty"`
}

func (User) TableName() string {
	return "users"
}

