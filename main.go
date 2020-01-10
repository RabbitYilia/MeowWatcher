package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"time"
)

var Config map[string]interface{}

func main() {
	Config = ReadConfigFromFile("conf.json")
	InitDB()
	SetProxy()
	InitTGBotAPI()

	for {
		select {
		case update := <-Config["TGUpdatesChan"].(tgbotapi.UpdatesChannel):
			if update.Message == nil { // ignore any non-Message Updates
				break
			}
			if int64(update.Message.Date) < Config["TGStartTime"].(int64) {
				break
			}
			if update.Message.ReplyToMessage == nil {
				ProcessTGCommand(update.Message.MessageID, update.Message.Chat.ID, update.Message.Text)
			} else {
				//ProcessTGReply()
			}
		default:
			SeekOfflineDevice()
			for DeviceName, _ := range Config["Devices"].(map[string]interface{}) {
				DeviceStatus := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"]
				if DeviceStatus == "ON" {
					DeviceStatusUpdate(DeviceName)
					DeviceGetSMS(DeviceName)
				}
			}
			log.Println(Config)
			time.Sleep(time.Second * 5)
		}
	}
}
