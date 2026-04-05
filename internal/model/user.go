package model

import "time"

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
	DisplayName string    `gorm:"size:200" json:"displayName"`
	Email       string    `gorm:"size:200" json:"email"`
	IsAdmin     bool      `gorm:"default:false" json:"isAdmin"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}
