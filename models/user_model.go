package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Email    string `gorm:"size:254;uniqueIndex"`
	Password string `gorm:"type:varchar"`
}
