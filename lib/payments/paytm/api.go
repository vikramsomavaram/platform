/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package paytm

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// Paytm represents payment.
type Paytm struct {
	MerchantMID             string
	MerchantKey             string
	MerchantWebsite         string
	TransactionStatusAPIURL string //https://securegw.paytm.in/order/status
	SendOTPAPIURL           string //https://accounts.paytm.com/signin/otp
}

const (
//https://securegw-stage.paytm.in/theia/api/v1/initiateTransaction?mid=<mid>&orderId=<orderId>
//https://securegw.paytm.in/theia/api/v1/initiateTransaction?mid=<mid>&orderId=<orderId>
)

var errorCodes = map[string]string{
	"k": "",
	"":  "",
}

// SendOTPRequest represents send otp request.
type SendOTPRequest struct {
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	ClientID     string `json:"clientId"`
	Scope        string `json:"scope"`
	ResponseType string `json:"responseType"`
}

// SendOTPResponse represents send otp response.
type SendOTPResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	ResponseCode string `json:"responseCode"`
	State        string `json:"state"`
}

// IInitiateTransactionRequestHeaders initiates transaction request headers.
type IInitiateTransactionRequestHeaders struct {
	Version          string
	ChannelID        string
	RequestTimestamp string
	ClientID         string
	Signature        string
}

// InitiateTransactionRequest initiates transaction request.
type InitiateTransactionRequest struct {
	Body struct {
		RequestType string `json:"requestType"`
		Mid         string `json:"mid"`
		WebsiteName string `json:"websiteName"`
		OrderID     string `json:"orderId"`
		TxnAmount   struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"txnAmount"`
		UserInfo struct {
			CustID string `json:"custId"`
		} `json:"userInfo"`
		CallbackURL string `json:"callbackUrl"`
	} `json:"body"`
	Head struct {
		ClientID         string `json:"clientId"`
		Version          string `json:"version"`
		RequestTimestamp int    `json:"requestTimestamp"`
		ChannelID        string `json:"channelId"`
		Signature        string `json:"signature"`
	} `json:"head"`
}

// InitiateTransactionResponse initiates transaction response.
type InitiateTransactionResponse struct {
	Head struct {
		ResponseTimestamp string `json:"responseTimestamp"`
		Version           string `json:"version"`
		ClientID          string `json:"clientId"`
		Signature         string `json:"signature"`
	} `json:"head"`
	Body struct {
		ResultInfo struct {
			ResultStatus string `json:"resultStatus"`
			ResultCode   string `json:"resultCode"`
			ResultMsg    string `json:"resultMsg"`
		} `json:"resultInfo"`
		TxnToken      string `json:"txnToken"`
		IsCouponValid bool   `json:"isCouponValid"`
		Authenticated bool   `json:"authenticated"`
	} `json:"body"`
}

// InitiatePayment initiates payment.
func (p *Paytm) InitiatePayment() {

}

//SendOTP sends otp ... https://developer.paytm.com/docs/send-otp-api/
func (p *Paytm) SendOTP(mobileNo string, clientID string) (bool, SendOTPResponse, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cache-Control"] = "no-cache"
	var otpResp SendOTPResponse

	sendOtp := &SendOTPRequest{Phone: mobileNo, ClientID: clientID, Scope: "Wallet", ResponseType: "Token"}
	sendOtpReq, err := json.Marshal(sendOtp)
	if err != nil {
		return false, otpResp, err
	}

	resp, err := p.call(p.SendOTPAPIURL, "POST", sendOtpReq, nil, headers)
	if err != nil {
		return false, otpResp, err
	}

	if err = json.Unmarshal(resp, &otpResp); err != nil {
		return false, otpResp, err
	}
	if otpResp.Status == "SUCCESS" {
		return true, otpResp, nil
	}
	return false, otpResp, err
}

// VerifyOTPResponse verifies otp response.
type VerifyOTPResponse struct {
}

// VerifyOTPRequest verifies otp request.
type VerifyOTPRequest struct {
}

// VerifyOTP verifies otp ... https://developer.paytm.com/docs/validate-otp-api/
func (p *Paytm) VerifyOTP(otp string, state string) (VerifyOTPResponse, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cache-Control"] = "no-cache"
	var otpResp VerifyOTPResponse

	verifyOtp := &VerifyOTPRequest{}
	verifyOtpReq, err := json.Marshal(verifyOtp)
	if err != nil {
		return otpResp, err
	}

	resp, err := p.call(p.SendOTPAPIURL, "POST", verifyOtpReq, nil, headers)
	if err != nil {
		return otpResp, err
	}

	if err = json.Unmarshal(resp, &otpResp); err != nil {
		return otpResp, err
	}

	return otpResp, nil
}

// RevokeAccess ceases access.
func (p *Paytm) RevokeAccess() {

}

// ValidateToken validates tokens.
func (p *Paytm) ValidateToken() {

}

// AddMoney adds money.
func (p *Paytm) AddMoney() {

}

// CheckBalance checks balance.
func (p *Paytm) CheckBalance() {

}

//TransactionStatus returns transaction status ... https://developer.paytm.com/docs/transaction-status-api/
func (p *Paytm) TransactionStatus(orderID string, checksum string) (bool, TransactionStatus, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cache-Control"] = "no-cache"
	var txnStatus TransactionStatus

	txnStatusReq := fmt.Sprintf(`{"MID":"%s","ORDERID":"%s","CHECKSUMHASH":"%s"}`, p.MerchantMID, orderID, checksum)
	resp, err := p.call(p.TransactionStatusAPIURL, "POST", []byte(txnStatusReq), nil, headers)
	if err != nil {
		return false, txnStatus, err
	}

	if err = json.Unmarshal(resp, &txnStatus); err != nil {
		return false, txnStatus, err
	}
	if txnStatus.Status == "TXN_SUCCESS" {
		return true, txnStatus, nil
	}
	return false, txnStatus, err
}

func (p *Paytm) call(url string, method string, reqbody []byte, queryparams map[string]string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqbody))
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return respbody, nil
}
