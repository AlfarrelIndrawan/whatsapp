package entity

import "gorm.io/gorm"

type UserWa struct {
	gorm.Model
	Name        string `json:"name" validate:"required"`
	PhoneNumber string `json:"phoneNumber" gorm:"unique" validate:"required"`
	Password    []byte `json:"-"`
	Status      string `json:"status" validate:"required"`
}
