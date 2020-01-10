package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func ProcessTGCommand(MessageID int,ChatID int64,Msg string){
	Args:=strings.Split(Msg,"\n")
	if(len(Args)<1){
		return
	}
	switch Args[0] {
	case "PING":
		msg := tgbotapi.NewMessage(ChatID, "PONG")
		msg.ReplyToMessageID = MessageID
		Config["TGBot"].(*tgbotapi.BotAPI).Send(msg)
	case "SEND":
		if(len(Args)<4){
			msg := tgbotapi.NewMessage(ChatID, "Arg Not Enough")
			msg.ReplyToMessageID = MessageID
			Config["TGBot"].(*tgbotapi.BotAPI).Send(msg)
		}else{
			DeviceName:=Args[1]
			DstPhoneNumber:=Args[2]
			if(Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"]==nil){
				msg := tgbotapi.NewMessage(ChatID, "Failed:Device Not Ready")
				msg.ReplyToMessageID = MessageID
				Config["TGBot"].(*tgbotapi.BotAPI).Send(msg)
				return
			}else{
				if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["Status"].(string)!="ON"{
					msg := tgbotapi.NewMessage(ChatID, "Failed:Device Not Ready")
					msg.ReplyToMessageID = MessageID
					Config["TGBot"].(*tgbotapi.BotAPI).Send(msg)
					return
				}
			}
			err:=DeviceSendSMS(DeviceName,DstPhoneNumber,Args[3])
			if err!=nil{
				msg := tgbotapi.NewMessage(ChatID, "Failed:"+err.Error())
				msg.ReplyToMessageID = MessageID
				Config["TGBot"].(*tgbotapi.BotAPI).Send(msg)
			}else {
				msg := tgbotapi.NewMessage(ChatID, "SEND OKAY")
				msg.ReplyToMessageID = MessageID
				Config["TGBot"].(*tgbotapi.BotAPI).Send(msg)
			}
		}
	}

}
