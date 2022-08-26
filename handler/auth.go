package handler

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/alfarrelindrawan/whatsapp/config"
	"github.com/alfarrelindrawan/whatsapp/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	jwt.StandardClaims
}

func Register(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Create User

	userWa := new(entity.UserWa)

	userWa.Name = data["name"]
	password, err := bcrypt.GenerateFromPassword([]byte(data["password"]), bcrypt.DefaultCost)
	userWa.Password = password
	userWa.PhoneCode = data["phoneCode"]
	userWa.PhoneNumber = data["phoneNumber"]
	userWa.Status = "0"

	if err := config.Database.Create(&userWa).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Create Token
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 0, 7).Unix(),
			Issuer:    strconv.Itoa(int(userWa.ID)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSigned, err := token.SignedString([]byte("secret"))
	if err != nil {
		return fmt.Errorf("Something error when generate JWT Token: %v", err)
	}

	// Create OTP
	var OTP string
	rand.NewSource(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		OTP += strconv.Itoa(rand.Intn(10))
	}
	OtpLog := new(entity.OtpLog)

	OtpLog.UserWaId = userWa.ID
	OtpLog.OTP = OTP
	OtpLog.Status = "0"

	if err := config.Database.Create(&OtpLog).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Send message
	fullPhoneNumber := data["phoneCode"] + data["phoneNumber"]
	response, err := sendMessage(fullPhoneNumber, OTP)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": err.Error(),
		})
	}
	return c.Status(200).JSON(fiber.Map{
		"code":    "200",
		"message": response,
		"token":   tokenSigned,
		"OTP":     OTP,
	})
}

func RegisterAuth(c *fiber.Ctx) error {
	// Parse Token
	tokenString := c.Get("Authorization")
	tokenString = strings.Fields(tokenString)[1]
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": err.Error(),
			"token":   tokenString,
		})
	}
	if !token.Valid {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": "Unauthorized",
		})
	}

	// Check OTP
	otp := struct {
		OTP    string `json:"otp"`
		Status string `jason:"status"`
	}{}
	if err := c.BodyParser(&otp); err != nil {
		return err
	}
	userWaId := token.Claims.(*Claims)
	otpLogStruct := new(entity.OtpLog)
	if err := config.Database.Where("user_wa_id = ? AND status = 0 AND otp = ?", userWaId.Issuer, otp.OTP).First(&otpLogStruct); err.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": "OTP already used. Request another OTP",
		})
	}
	if time.Now().After(otpLogStruct.CreatedAt.Add(time.Minute)) {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": "OTP already expired. Request another OTP",
		})
	}

	// Update user and OTP status
	otpLogStruct.Status = "1"
	config.Database.Where("user_wa_id = ? AND status = 0  AND otp = ?", userWaId.Issuer, otp.OTP).Updates(&otpLogStruct)
	userWaStruct := new(entity.UserWa)
	userWaStruct.Status = "1"
	config.Database.Where("id = ? AND status = 0", userWaId.Issuer).Updates(&userWaStruct)
	return c.Status(200).JSON(fiber.Map{
		"code":    "200",
		"message": "Auth success",
	})
}

func Login(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Check User
	var userWa entity.UserWa

	config.Database.Where("phone_code = ? AND phone_number = ?", data["phoneCode"], data["phoneNumber"]).First(&userWa)

	if userWa.ID == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"error": "User not found",
		})
	}

	err := bcrypt.CompareHashAndPassword(userWa.Password, []byte(data["password"]))

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	// Create Token
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 0, 7).Unix(),
			Issuer:    strconv.Itoa(int(userWa.ID)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSigned, err := token.SignedString([]byte("secret"))
	if err != nil {
		return fmt.Errorf("Something error when generate JWT Token: %v", err)
	}

	return c.JSON(fiber.Map{
		"status":  "200",
		"success": "success",
		"token":   tokenSigned,
	})
}

func ResendOTP(c *fiber.Ctx) error {
	// Check Token
	tokenString := c.Get("Authorization")
	tokenString = strings.Fields(tokenString)[1]
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": err.Error(),
			"token":   tokenString,
		})
	}
	if !token.Valid {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": "Unauthorized",
		})
	}

	// Get User
	userWaId := token.Claims.(*Claims)
	userWaStruct := new(entity.UserWa)
	if err := config.Database.Where("id = ? AND status = 0", userWaId.Issuer).First(&userWaStruct); err.Error != nil {
		fmt.Println(err.Error)
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": "Can't find user",
		})
	}

	// Create OTP
	var OTP string
	rand.NewSource(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		OTP += strconv.Itoa(rand.Intn(10))
	}
	OtpLog := new(entity.OtpLog)

	OtpLog.UserWaId = userWaStruct.ID
	OtpLog.OTP = OTP
	OtpLog.Status = "0"

	if err := config.Database.Create(&OtpLog).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Send Message
	fullPhoneNumber := userWaStruct.PhoneCode + userWaStruct.PhoneNumber
	response, err := sendMessage(fullPhoneNumber, OTP)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"code":    "404",
			"message": err.Error(),
		})
	}
	return c.Status(200).JSON(fiber.Map{
		"code":    "200",
		"message": response,
		"OTP":     OTP,
	})
}
