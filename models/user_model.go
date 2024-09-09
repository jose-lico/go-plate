package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Email    string `gorm:"size:254;uniqueIndex"`
	Password string `gorm:"type:varchar"`
	Name     string `gorm:"size:32;type:varchar"`
}
