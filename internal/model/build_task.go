package model

import "time"

type BuildStatus string

const (
	BuildStatusPending  BuildStatus = "PENDING"
	BuildStatusBuilding BuildStatus = "BUILDING"
	BuildStatusSuccess  BuildStatus = "SUCCESS"
	BuildStatusFailed   BuildStatus = "FAILED"
)

type BuildTask struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	UserID       uint        `gorm:"index;not null" json:"userId"`
	YamlContent  string      `gorm:"type:text" json:"yamlContent"`
	Status       BuildStatus `gorm:"size:50;default:PENDING" json:"status"`
	ResultPath   string      `gorm:"size:500" json:"resultPath"`
	ErrorMessage string      `gorm:"type:text" json:"errorMessage"`
	CreatedAt    time.Time   `json:"createdAt"`
	CompletedAt  *time.Time  `json:"completedAt"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (BuildTask) TableName() string {
	return "build_tasks"
}
