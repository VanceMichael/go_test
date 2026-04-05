package model

import "time"

type OperationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"userId"`
	Action     string    `gorm:"size:100;not null" json:"action"`
	TargetType string    `gorm:"size:100" json:"targetType"`
	TargetID   uint      `json:"targetId"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IPAddress  string    `gorm:"size:50" json:"ipAddress"`
	CreatedAt  time.Time `json:"createdAt"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
