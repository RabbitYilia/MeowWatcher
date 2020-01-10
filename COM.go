package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const MAXRWLEN = 128000

func SendCommand(Handler io.ReadWriteCloser, Command string) (string, string, error) {
	buffer := make([]byte, MAXRWLEN)
	//发命令之前清空缓冲区
	num, err := Handler.Read(buffer)
	if err != nil {
		return "", "", err
	}
	num, err = Handler.Write([]byte(Command + "\r\n"))
	if err != nil {
		return "", "", err
	}
	StrBuffer := ""
	ReturnBuffer := ""
	StatusBuffer := ""
	start := time.Now().Unix()
	for i := 0; i < 128000; i++ {
		num, err = Handler.Read(buffer)
		if num > 0 {
			StrBuffer += fmt.Sprintf("%s", string(buffer[:num]))
		}
		pos := strings.LastIndex(StrBuffer, "\r\nCOMMAND NOT SUPPORT\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\nOK\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\nERROR\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\n>\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		if time.Now().Unix()-start > 5 {
			break
		}
	}
	ReturnBuffer = strings.Replace(ReturnBuffer, Command+"\r", Command+"\r\n", -1)
	ReturnBuffer = strings.Replace(ReturnBuffer, "\r\n\r\n", "\r\n", -1)
	ReturnBuffer = strings.Replace(ReturnBuffer, Command+"\r\n", "", -1)
	ReturnBuffer = strings.TrimRight(ReturnBuffer, "\r\n")
	StatusBuffer = strings.Replace(StatusBuffer, Command+"\r", Command+"\r\n", -1)
	StatusBuffer = strings.Replace(StatusBuffer, "\r\n\r\n", "\r\n", -1)
	StatusBuffer = strings.Replace(StatusBuffer, Command+"\r\n", "", -1)
	StatusBuffer = strings.TrimRight(StatusBuffer, "\r\n")
	StatusBuffer = strings.TrimLeft(StatusBuffer, "\r\n")
	DebugOutput(3, Command, ReturnBuffer, StatusBuffer)
	return ReturnBuffer, StatusBuffer, nil
}

func DeivceSendCommand(DeviceName string, Command string) (string, string, error) {
	Handler := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser)
	buffer := make([]byte, MAXRWLEN)
	//发命令之前清空缓冲区
	num, err := Handler.Read(buffer)
	if err != nil {
		return "", "", err
	}
	num, err = Handler.Write([]byte(Command + "\r\n"))
	if err != nil {
		return "", "", err
	}
	StrBuffer := ""
	ReturnBuffer := ""
	StatusBuffer := ""
	start := time.Now().Unix()
	for i := 0; i < 128000; i++ {
		num, err = Handler.Read(buffer)
		if num > 0 {
			StrBuffer += fmt.Sprintf("%s", string(buffer[:num]))
		}
		pos := strings.LastIndex(StrBuffer, "\r\nCOMMAND NOT SUPPORT\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\nOK\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\nERROR\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\n>\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		if time.Now().Unix()-start > 5 {
			break
		}
	}
	ReturnBuffer = strings.Replace(ReturnBuffer, Command+"\r", Command+"\r\n", -1)
	ReturnBuffer = strings.Replace(ReturnBuffer, "\r\n\r\n", "\r\n", -1)
	ReturnBuffer = strings.Replace(ReturnBuffer, Command+"\r\n", "", -1)
	ReturnBuffer = strings.TrimRight(ReturnBuffer, "\r\n")
	StatusBuffer = strings.Replace(StatusBuffer, Command+"\r", Command+"\r\n", -1)
	StatusBuffer = strings.Replace(StatusBuffer, "\r\n\r\n", "\r\n", -1)
	StatusBuffer = strings.Replace(StatusBuffer, Command+"\r\n", "", -1)
	StatusBuffer = strings.TrimRight(StatusBuffer, "\r\n")
	StatusBuffer = strings.TrimLeft(StatusBuffer, "\r\n")
	DebugOutput(3, Command, ReturnBuffer, StatusBuffer)
	return ReturnBuffer, StatusBuffer, nil
}
func DeivceSendPDU(DeviceName string, PDU string) (string, string, error) {
	Handler := Config["Devices"].(map[string]interface{})[DeviceName].(map[string]interface{})["MDMPortHandler"].(io.ReadWriteCloser)
	buffer := make([]byte, MAXRWLEN)
	//发命令之前清空缓冲区
	num, err := Handler.Read(buffer)
	if err != nil {
		return "", "", err
	}
	num, err = Handler.Write([]byte(PDU + "\x1a"))
	if err != nil {
		return "", "", err
	}
	StrBuffer := ""
	ReturnBuffer := ""
	StatusBuffer := ""
	start := time.Now().Unix()
	for i := 0; i < 128000; i++ {
		num, err = Handler.Read(buffer)
		if num > 0 {
			StrBuffer += fmt.Sprintf("%s", string(buffer[:num]))
		}
		pos := strings.LastIndex(StrBuffer, "\r\nCOMMAND NOT SUPPORT\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\nOK\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\nERROR\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		pos = strings.LastIndex(StrBuffer, "\r\n>\r\n")
		if pos > 0 {
			ReturnBuffer = StrBuffer[:pos]
			StatusBuffer = StrBuffer[pos:]
			break
		}
		if time.Now().Unix()-start > 5 {
			break
		}
	}
	ReturnBuffer = strings.Replace(ReturnBuffer, PDU+"\r", PDU+"\r\n", -1)
	ReturnBuffer = strings.Replace(ReturnBuffer, "\r\n\r\n", "\r\n", -1)
	ReturnBuffer = strings.Replace(ReturnBuffer, PDU+"\r\n", "", -1)
	ReturnBuffer = strings.TrimRight(ReturnBuffer, "\r\n")
	StatusBuffer = strings.Replace(StatusBuffer, PDU+"\r", PDU+"\r\n", -1)
	StatusBuffer = strings.Replace(StatusBuffer, "\r\n\r\n", "\r\n", -1)
	StatusBuffer = strings.Replace(StatusBuffer, PDU+"\r\n", "", -1)
	StatusBuffer = strings.TrimRight(StatusBuffer, "\r\n")
	StatusBuffer = strings.TrimLeft(StatusBuffer, "\r\n")
	DebugOutput(3, PDU, ReturnBuffer, StatusBuffer)
	return ReturnBuffer, StatusBuffer, nil
}
