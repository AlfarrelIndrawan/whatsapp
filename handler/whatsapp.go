package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

func Test(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"message": "test",
	})
}

func WAConnect() (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, err
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, err
	}
	client := whatsmeow.NewClient(deviceStore, waLog.Noop)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func sendMessage(phoneNumber, OTP string) (check string, err error) {
	wac, err := WAConnect()
	if err != nil {
		return "", err
	}

	var fullPhoneNumber []string
	fullPhoneNumber = append(fullPhoneNumber, phoneNumber)

	response, err := wac.IsOnWhatsApp(fullPhoneNumber)
	if err != nil {
		return "", err
	}
	if response[0].IsIn == false {
		return sendSMS(phoneNumber, OTP)
	}
	message := "Ini adalah kode OTP: " + OTP + "\nJangan beri tau siapapun!"
	_, err = wac.SendMessage(context.Background(), types.JID{
		User:   phoneNumber[1:],
		Server: types.DefaultUserServer,
	}, "", &waProto.Message{
		Conversation: proto.String(message),
	})
	return "Success", nil
}
