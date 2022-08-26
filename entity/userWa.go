package entity

import "gorm.io/gorm"

type UserWa struct {
	gorm.Model
	Name        string `json:"name" validate:"required"`
	PhoneCode   string `json:"phoneCode" validate:"required"`
	PhoneNumber string `json:"phoeNumber" gorm:"unique" validate:"required"`
	Password    []byte `json:"-"`
	Status      string `json:"status" validate:"required"`
}
