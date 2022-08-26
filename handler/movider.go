package handler

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/movider/movider-go/client"
	"github.com/movider/movider-go/sms"
)

func sendSMS(phoneNumber, OTP string) (result string, err error) {
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client := client.New(os.Getenv("MOVIDER_API_KEY"), os.Getenv("MOVIDER_API_SECRET"))
	message := "SMS dengan Movider.\n Kode OTP: " + OTP

	d, err := sms.Send(client, []string{
		phoneNumber[1:],
	}, message, &sms.Params{})
	if err != nil {
		return "", err
	}
	fmt.Println(d.Result)
	return "success", nil
}
