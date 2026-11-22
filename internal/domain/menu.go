package domain

type Menu struct {
	BaseModel
	ParentID *string `gorm:"type:varchar(26);index" json:"parent_id"` // null for root menus
	Title    string  `gorm:"type:varchar(100);not null" json:"title"`
	Path     string  `gorm:"type:varchar(150);not null" json:"path"`
	Icon     string  `gorm:"type:varchar(100)" json:"icon"`
	OrderNo  int     `gorm:"type:int;default:0;not null" json:"order_no"`
	IsActive int     `gorm:"type:smallint;default:1;not null;index" json:"is_active"` // 1 = active, 0 = inactive

	// Relations
	Parent          *Menu            `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:SET NULL" json:"parent,omitempty"`
	Children        []*Menu          `gorm:"-" json:"children,omitempty"`
	MenuPermissions []MenuPermission `gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE" json:"menu_permissions,omitempty"`
	Permissions     []string         `gorm:"-" json:"permissions,omitempty"` // populated during nav-menu tree building
}

func (Menu) TableName() string {
	return "menus"
}

