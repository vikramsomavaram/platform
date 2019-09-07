/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package paytm

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/spacemonkeygo/openssl"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

// Crypter is structure of paytm
type Crypter struct {
	key    []byte
	iv     []byte
	cipher *openssl.Cipher
}

// NewCrypter is function of paytm
func NewCrypter(key []byte, iv []byte) (*Crypter, error) {
	cipher, err := openssl.GetCipherByName("aes-128-cbc")
	if err != nil {
		return nil, err
	}

	return &Crypter{key, iv, cipher}, nil
}

// Encrypt is function of encryption algorithm
func (c *Crypter) Encrypt(input []byte) ([]byte, error) {
	ctx, err := openssl.NewEncryptionCipherCtx(c.cipher, nil, c.key, c.iv)
	if err != nil {
		return nil, err
	}

	cipherbytes, err := ctx.EncryptUpdate(input)
	if err != nil {
		return nil, err
	}

	finalbytes, err := ctx.EncryptFinal()
	if err != nil {
		return nil, err
	}

	cipherbytes = append(cipherbytes, finalbytes...)
	return cipherbytes, nil
}

// Decrypt is function of decryption key
func (c *Crypter) Decrypt(input []byte) ([]byte, error) {
	ctx, err := openssl.NewDecryptionCipherCtx(c.cipher, nil, c.key, c.iv)
	if err != nil {
		return nil, err
	}

	cipherbytes, err := ctx.DecryptUpdate(input)
	if err != nil {
		return nil, err
	}

	finalbytes, err := ctx.DecryptFinal()
	if err != nil {
		return nil, err
	}

	cipherbytes = append(cipherbytes, finalbytes...)
	return cipherbytes, nil
}

// GetChecksumFromArray is function to generate checksum key
func GetChecksumFromArray(paramsMap map[string]string) (checksum string, err error) {
	var keys = make([]string, 0, 0)
	for k, v := range paramsMap {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var arrayList = make([]string, 0, 0)
	for _, key := range keys {
		if value, ok := paramsMap[key]; ok && value != "" {
			arrayList = append(arrayList, value)
		}
	}
	arrayStr := getArray2Str(arrayList)
	salt := generateSalt(4)
	finalString := arrayStr + "|" + salt
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(finalString)))
	hashString := hash + salt
	crypt, err := Encrypt([]byte(hashString))
	if err != nil {
		return
	}
	checksum = base64.StdEncoding.EncodeToString(crypt)
	return
}

// VerifyCheckum is function to verify checksum
func VerifyCheckum(paramsMap map[string]string, checksum string) (ok bool) {
	delete(paramsMap, "CHECKSUMHASH")
	var keys = make([]string, 0, 0)
	for k, v := range paramsMap {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var arrayList = make([]string, 0, 0)
	for _, key := range keys {
		if value, ok := paramsMap[key]; ok && value != "" {
			arrayList = append(arrayList, value)
		}
	}
	arrayStr := getArray2StrForVerify(arrayList)
	cs, err := base64.StdEncoding.DecodeString(checksum)
	if err != nil {
		fmt.Printf("base64 DecodeString err [%v]\n", err)
		return
	}
	paytmHash, err := Decrypt(cs)
	if err != nil {
		fmt.Printf("Decrypt err [%v]\n", err)
		return
	}
	paytmHashStr := string(paytmHash)
	salt := paytmHashStr[len(paytmHashStr)-4:]
	finalString := arrayStr + "|" + salt
	h := sha256.New()
	h.Write([]byte(finalString))
	finalStringHash := fmt.Sprintf("%x", h.Sum(nil))
	websiteHashStr := finalStringHash + salt
	if websiteHashStr == paytmHashStr {
		return true
	}
	return false
}

// Encrypt is function to encryption
func Encrypt(input []byte) (output []byte, err error) {
	iv := "@@@@&&&&####$$$$"
	crypter, _ := NewCrypter([]byte(os.Getenv("PAYTM_MERCHANT_KEY")), []byte(iv))
	output, err = crypter.Encrypt(input)
	return
}

// Decrypt is function to decryption
func Decrypt(input []byte) (output []byte, err error) {
	iv := "@@@@&&&&####$$$$"
	crypter, err := NewCrypter([]byte(os.Getenv("PAYTM_MERCHANT_KEY")), []byte(iv))
	output, err = crypter.Decrypt(input)
	return
}

// getArray2Str is function for convert array to string
func getArray2Str(arrayList []string) (str string) {
	findme := "REFUND"
	findmepipe := "|"
	flag := 1
	for _, v := range arrayList {
		pos := strings.Index(v, findme)
		pospipe := strings.Index(v, findmepipe)
		if pos != -1 || pospipe != -1 {
			continue
		}
		if flag > 0 {
			str += strings.TrimSpace(v)
			flag = 0
		} else {
			str += "|" + strings.TrimSpace(v)
		}
	}
	return
}

// getArray2StrForVerify is function for verify array to string
func getArray2StrForVerify(arrayList []string) (str string) {
	flag := 1
	for _, v := range arrayList {
		if flag > 0 {
			str += strings.TrimSpace(v)
			flag = 0
		} else {
			str += "|" + strings.TrimSpace(v)
		}
	}
	return
}

// generateSalt is function for generate salt
func generateSalt(length int) (salt string) {
	rand.Seed(time.Now().UnixNano())
	data := "AbcDE123IJKLMN67QRSTUVWXYZ"
	data += "aBCdefghijklmn123opq45rs67tuv89wxyz"
	data += "0FGH45OP89"
	for i := 0; i < length; i++ {
		salt += string(data[int(rand.Int()%len(data))])
	}
	//salt = "1234"
	return
}

// GetTransactionStatus is function for get transaction status

// TransactionStatus is function for check transaction status
type TransactionStatus struct {
	TxnID        string `json:"TXNID"`
	BankTxnID    string `json:"BANKTXNID"`
	OrderID      string `json:"ORDERID"`
	TxnAmount    string `json:"TXNAMOUNT"`
	Status       string `json:"STATUS"`
	TxnType      string `json:"TXNTYPE"`
	GatewayName  string `json:"GATEWAYNAME"`
	RespCode     string `json:"RESPCODE"`
	RespMsg      string `json:"RESPMSG"`
	BankName     string `json:"BANKNAME"`
	MID          string `json:"MID"`
	PaymentMode  string `json:"PAYMENTMODE"`
	RefundAmount string `json:"REFUNDAMT"`
	TxnDate      string `json:"TXNDATE"`
}
