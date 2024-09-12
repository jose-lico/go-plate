package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Email    string `gorm:"type:varchar(255);uniqueIndex"`
	Password string `gorm:"type:varchar(64);not null"`
	Name     string `gorm:"type:varchar(32);not null"`
}
