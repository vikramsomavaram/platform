/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package msg91

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

//Msg91 msg91.com sms API provider
type Msg91 struct {
	AuthKey    string
	SenderName string
}

//http://api.msg91.com/api/sendhttp.php?authkey=YourAuthKey&mobiles=919999999990,919999999999&message=message&sender=ABCDEF&route=4&country=0

type msg91Request struct {
	Sender  string     `json:"sender"`
	Route   string     `json:"route"`
	Unicode int        `json:"unicode"`
	Country string     `json:"country"`
	Flash   int        `json:"flash"`
	Sms     []msg91Sms `json:"sms"`
}
type msg91Sms struct {
	Message string   `json:"message"`
	To      []string `json:"to"`
}

type msg91Response struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

//NewMsg91 creates new instance of auth key
func NewMsg91(Authkey string, Sender string) *Msg91 {
	return &Msg91{AuthKey: Authkey, SenderName: Sender}
}

//SendMessage send a text/sms via msg91
func SendMessage(message string, transactionSms bool, mobileNos ...string) (bool, error) {

	m := &Msg91{os.Getenv("MSG91_KEY"), "MTRIBE"}
	url := "https://api.msg91.com/api/v2/sendsms"
	log.Println("URL:>", url)
	msgreq := &msg91Request{
		//Country: "0",
		Sender:  m.SenderName,
		Flash:   0,
		Unicode: 1,
		Sms:     []msg91Sms{{message, mobileNos}},
	}
	if transactionSms {
		msgreq.Route = "4"
	} else {
		msgreq.Route = "1"
	}
	msgreqjson, err := json.Marshal(msgreq)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(msgreqjson))
	req.Header.Add("authkey", m.AuthKey)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
	log.Println("Response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln(err)
	}
	log.Println("Response Body:", string(body))
	msgresp := new(msg91Response)
	json.Unmarshal(body, msgresp)
	if msgresp.Type != "success" {
		err = errors.New("[" + msgresp.Code + "]" + msgresp.Message)
		return false, err
	}
	return true, nil
}

//SendOTP sends a random OTP
func (m *Msg91) SendOTP() (string, error) {
	return "", nil
}

//VerifyOTP Verifies the OTP
func (m *Msg91) VerifyOTP() (bool, error) {
	return false, nil
}
