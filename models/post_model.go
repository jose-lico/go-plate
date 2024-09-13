package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model

	Title   string `gorm:"type:varchar(255);not null"`
	Summary string `gorm:"type:varchar(255);"`
	Content string `gorm:"type:varchar(1000);not null"`
	UserID  uint   `gorm:"not null"`
	User    User   `gorm:"foreignKey:UserID"`
}
