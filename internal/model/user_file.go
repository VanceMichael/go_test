package model

import "time"

type UserFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	FileName  string    `gorm:"size:500;not null" json:"fileName"`
	BOSPath   string    `gorm:"size:500" json:"bosPath"`
	FileSize  int64     `json:"fileSize"`
	IsPublic  bool      `gorm:"default:false" json:"isPublic"`
	CreatedAt time.Time `json:"createdAt"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserFile) TableName() string {
	return "user_files"
}
