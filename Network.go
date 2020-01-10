package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func InitTGBotAPI() {
	if Config["TGBotToken"] == nil {
		return
	}
	APIToken := strings.Replace(Config["TGBotToken"].(string), "bot", "", 1)
	TGBot, err := tgbotapi.NewBotAPIWithClient(APIToken, Config["Client"].(*http.Client))
	if err != nil {
		log.Fatal(err)
	}
	if DebugLevel >= 3 {
		TGBot.Debug = true
	} else {
		TGBot.Debug = false
	}
	Config["TGBot"] = TGBot
	DebugOutput(0, "Authorized on account", TGBot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := Config["TGBot"].(*tgbotapi.BotAPI).GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}
	Config["TGUpdatesChan"] = updates
	Config["TGStartTime"] = time.Now().Unix()
}

func SetProxy() {
	if Config["Proxy"].(string) != "" {
		tr := http.Transport{}
		parseProxyUrl, _ := url.Parse(Config["Proxy"].(string))
		tr.Proxy = http.ProxyURL(parseProxyUrl)
		Config["Client"] = &http.Client{Transport: &tr}
	} else {
		Config["Client"] = &http.Client{}
	}
}

func PushTG(DeviceName string, Tittle string, Content string) {
	if Config["TGBot"] == nil || Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["TGChatID"] == nil {
		return
	}
	TGChatID := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["TGChatID"].(string)
	TGBot := Config["TGBot"].(*tgbotapi.BotAPI)
	if TGChatID != "" {
		TGChatIDInt64, _ := strconv.ParseInt(TGChatID, 10, 64)
		msg := tgbotapi.NewMessage(TGChatIDInt64, Content)
		count := 0
		for {
			count++
			if count > 15 {
				log.Fatal(fmt.Sprintf("%s Push Failed", DeviceName))
			}
			_, err := TGBot.Send(msg)
			if err == nil {
				break
			}
		}
	}
}

func PushSC(DeviceName string, Tittle string, Content string) {
	if Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["SCToken"] == nil {
		return
	}
	Client := Config["Client"].(*http.Client)
	SCToken := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["SCToken"].(string)
	URL := fmt.Sprintf("https://sc.ftqq.com/%s.send", SCToken)
	Body := fmt.Sprintf("text=%s&desp=%s", url.QueryEscape(Tittle), url.QueryEscape(Content))
	count := 0
	for {
		count++
		if count > 15 {
			log.Fatal(fmt.Sprintf("%s Push Failed", DeviceName))
		}
		req, err := http.NewRequest("POST", URL, strings.NewReader(Body))
		if err != nil {
			log.Println(err)
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := Client.Do(req)
		if err != nil {
			log.Println(err)
			continue
		} else {
			defer resp.Body.Close()
			break
		}
	}
}
