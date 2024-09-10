package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Email    string `gorm:"type:varchar;size:254;uniqueIndex"`
	Password string `gorm:"type:varchar;size:32"`
	Name     string `gorm:"type:varchar;size:32"`
}
