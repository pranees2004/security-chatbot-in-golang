package src

import (
	ctx "context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	wp "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var client *whatsmeow.Client
var db *gorm.DB

type MessageLog struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    string `gorm:"index"`
	Role      string
	Message   string
	Timestamp time.Time
}

func SaveImageFromBytes(imageData []byte, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %v", err)
	}
	defer out.Close()
	_, err = out.Write(imageData)
	if err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}
	return nil
}

func eventHandler(evt interface{}) {
	if evt, ok := evt.(*events.Message); ok {
		go HandleConversation(evt)
	}
}

func MakeWaClient() *whatsmeow.Client {
	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New("sqlite3", "file:aiasst.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "INFO", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(ctx.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}
	return client
}

func HandleConversation(evt *events.Message) {
	sender := evt.Info.Sender.User
	phoneNumber := evt.Info.Sender.ToNonAD().String()
	var userMessage string

	if evt.Message.GetConversation() != "" {
		userMessage = evt.Message.GetConversation()
		LogMessage(phoneNumber, "User", userMessage)
		response := GenerateResponse(sender, userMessage)
		client.SendMessage(ctx.Background(), evt.Info.Chat, &wp.Message{
			Conversation: proto.String(response),
		})
		LogMessage(phoneNumber, "Bot", response)

	} else if imageMsg := evt.Message.GetImageMessage(); imageMsg != nil {
		imageData, err := client.Download(imageMsg)
		if err != nil {
			fmt.Printf("Error downloading image: %v\n", err)
			return
		}
		userImageFileName := fmt.Sprintf("chat_logs/%s_reconstructed_image.jpg", phoneNumber)
		err = SaveImageFromBytes(imageData, userImageFileName)
		if err != nil {
			fmt.Printf("Failed to save the image to file: %v\n", err)
		}
		userMessage = "[Image Message]"
		LogMessage(phoneNumber, "User", userMessage)

	} else {
		userMessage = "[Unsupported Message Type]"
		LogMessage(phoneNumber, "User", userMessage)
	}

	if userStates[sender] == "" {
		response := GenerateResponse(sender, userMessage)
		client.SendMessage(ctx.Background(), evt.Info.Chat, &wp.Message{
			Conversation: proto.String(response),
		})
		LogMessage(phoneNumber, "Bot", response)
	}
}

func GenerateResponse(userID, message string) string {
	intent := GetIntentFromMessage(message)
	SetUserState(userID, intent)
	response := GetResponseByState(userID)
	return response
}

func GetIntentFromMessage(message string) string {
	for _, intent := range intents {
		for _, example := range intent.Examples {
			if strings.Contains(strings.ToLower(message), strings.ToLower(example)) {
				return intent.Name
			}
		}
	}
	return "default"
}

func LogMessage(userID, role, message string) {
	logEntry := MessageLog{
		UserID:    userID,
		Role:      role,
		Message:   message,
		Timestamp: time.Now(),
	}
	db.Create(&logEntry)
}

func SetupDatabase() error {
	var err error
	db, err = gorm.Open(sqlite.Open("chat_logs.db"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&MessageLog{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}
