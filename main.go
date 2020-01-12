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
		if Config["TGUpdatesChan"] == nil {
			DoOfflineJob()
		} else {
			DoOnlineJob()
		}
		PrintOnlineDevice()
		PrintNonOnlineDevice()
	}
}

func DoOfflineJob() {
	SeekOfflineDevice()
	for DeviceName, _ := range Config["Devices"].(map[string]interface{}) {
		DeviceStatus := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"]
		if DeviceStatus == "ON" {
			log.Println("===[" + DeviceName + "] Processing===")
			DeviceStatusUpdate(DeviceName)
			DeviceGetSMS(DeviceName)
		}
	}
	time.Sleep(time.Second * 5)
}

func DoOnlineJob() {
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
		DoOfflineJob()
	}
}
func PrintOnlineDevice(){
	log.Println("==OnLine Device==")
	for DeviceName, _ := range Config["Devices"].(map[string]interface{}) {
		DeviceStatus := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"]
		if DeviceStatus == "ON" {
			log.Println("===[" + DeviceName + "] Online===")
		}
	}
}

func PrintNonOnlineDevice(){
	log.Println("==No-ONLine Device==")
	for DeviceName, _ := range Config["Devices"].(map[string]interface{}) {
		DeviceStatus := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"]
		if DeviceStatus != "ON" {
			log.Println("===[" + DeviceName + "] ===")
		}
	}
}