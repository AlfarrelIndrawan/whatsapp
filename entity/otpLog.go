package entity

import "gorm.io/gorm"

type OtpLog struct {
	gorm.Model
	UserWaId uint   `jason:"userWaId" validate:"required"`
	OTP      string `json:"otp" validate:"required"`
	Status   string `json:"status" validate:"required"`
}
