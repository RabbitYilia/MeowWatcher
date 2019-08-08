package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

func SetProxy(config *Config) {
	if config.Proxy != "" {
		tr := http.Transport{}
		parseProxyUrl, _ := url.Parse("http://127.0.0.1:8080")
		tr.Proxy = http.ProxyURL(parseProxyUrl)
		config.client = &http.Client{Transport: &tr}
	} else {
		config.client = &http.Client{}
	}
}

func PushByPost(client *http.Client, URL string, Content string) {
	count := 0
	for {
		count++
		if count > 15 {
			log.Fatal("Push Failed.")
		}
		req, err := http.NewRequest("POST", URL, strings.NewReader(Content))
		if err != nil {
			log.Println(err)
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			continue
		} else {
			defer resp.Body.Close()
			break
		}
	}
}
