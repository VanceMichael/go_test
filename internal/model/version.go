package model

import "time"

type Version struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DirectoryID uint      `gorm:"index;not null" json:"directoryId"`
	VersionName string    `gorm:"size:200;not null" json:"versionName"`
	Description string    `gorm:"type:text" json:"description"`
	BOSPath     string    `gorm:"size:500" json:"bosPath"`
	InternalURL string    `gorm:"size:500" json:"internalUrl"`
	ExternalURL string    `gorm:"size:500" json:"externalUrl"`
	FileSize    int64     `json:"fileSize"`
	UploaderID  uint      `gorm:"index" json:"uploaderId"`
	CreatedAt   time.Time `json:"createdAt"`

	Directory Directory `gorm:"foreignKey:DirectoryID" json:"-"`
	Uploader  User      `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
}

func (Version) TableName() string {
	return "versions"
}

type BaselineConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	BaselineDirID uint      `gorm:"index;not null" json:"baselineDirId"`
	VersionID     uint      `gorm:"index;not null" json:"versionId"`
	CreatedAt     time.Time `json:"createdAt"`
	CreatedBy     uint      `json:"createdBy"`

	BaselineDir Directory `gorm:"foreignKey:BaselineDirID" json:"-"`
	Version     Version   `gorm:"foreignKey:VersionID" json:"version,omitempty"`
}

func (BaselineConfig) TableName() string {
	return "baseline_configs"
}
