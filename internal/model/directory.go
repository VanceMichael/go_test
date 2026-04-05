package model

import "time"

type DirectoryType string

const (
	DirTypeWinClient DirectoryType = "WIN_CLIENT"
	DirTypeMacClient DirectoryType = "MAC_CLIENT"
	DirTypeServer    DirectoryType = "SERVER"
	DirTypeCustom    DirectoryType = "CUSTOM"
	DirTypeOther     DirectoryType = "OTHER"
)

type ListType string

const (
	ListTypeFull     ListType = "FULL"
	ListTypeBaseline ListType = "BASELINE"
)

type Directory struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	Name      string        `gorm:"size:200;not null" json:"name"`
	ParentID  *uint         `gorm:"index" json:"parentId"`
	Type      DirectoryType `gorm:"size:50" json:"type"`
	ListType  ListType      `gorm:"size:50" json:"listType"`
	SortOrder int           `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time     `json:"createdAt"`

	Children []Directory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Parent   *Directory  `gorm:"foreignKey:ParentID" json:"-"`
}

func (Directory) TableName() string {
	return "directories"
}
