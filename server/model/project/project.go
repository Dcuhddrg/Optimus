package project

import (
	"optimus/server/global"
	"time"

	"gorm.io/gorm"
)

type Project struct {
	global.OPTIMUS_MODEL
	Name        string         `gorm:"size:128;not null" json:"name"`
	Code        string         `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Description string         `gorm:"size:512" json:"description"`
	OwnerID     uint           `gorm:"not null;index" json:"ownerId"`
	Status      string         `gorm:"size:16;not null;default:'active'" json:"status"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Project) TableName() string {
	return "projects"
}

type UserProject struct {
	UserID      uint      `gorm:"primaryKey;autoIncrement:false" json:"userId"`
	ProjectID   uint      `gorm:"primaryKey;autoIncrement:false" json:"projectId"`
	ProjectRole string    `gorm:"size:32;not null" json:"projectRole"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (UserProject) TableName() string {
	return "user_projects"
}
