/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

//Package webhookworker webhook's delivery worker utilizes GoogleCloud PubSub to poll and delivery events as they happen to clients.
package webhookworker

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/lib/log/log_formatter"
	"io/ioutil"
	"net/http"
)

// WebhookEvent represents webhook event.
type WebhookEvent struct {
	PayloadURL    string `json:"payload_url"`
	PayloadData   string `json:"payload_data"` // base64 encoded payload
	WebhookSecret string `json:"webhook_secret"`
}


func init() {
	//Logrus Settings
	log.SetFormatter(log_formatter.NewFormatter())
	log.SetReportCaller(true)
	database.ConnectMongo() //Connect to MongoDB
}


//func main() {
//	runtime.GOMAXPROCS(runtime.NumCPU())
//	ctx := context.Background()
//	projectID := ""
//	psclient, err := pubsub.NewClient(ctx, projectID)
//	if err != nil {
//		log.Fatalf("Failed to create client: %v", err)
//	}
//	//Pull from Google Cloud PubSub
//	deliverySub := psclient.Subscription("webhook_delivery")
//	deliverySub.ReceiveSettings = pubsub.ReceiveSettings{MaxOutstandingMessages: 20, MaxExtension: pubsub.DefaultReceiveSettings.MaxExtension}
//	// deliveryTopic := psclient.Topic("webhook_delivery")
//	err = deliverySub.Receive(ctx, DeliverPayload)
//	if err != nil {
//		log.Fatalln(err)
//	}
//}

// DeliverPayload delivers webhook.
func DeliverPayload(ctx context.Context, msg *pubsub.Message) error {
	deliveryTask := &WebhookEvent{}
	err := json.Unmarshal(msg.Data, deliveryTask)
	if err != nil {
		log.Fatalln(err)
	}
	payloadData, err := base64.StdEncoding.DecodeString(deliveryTask.PayloadData)
	if err != nil {
		log.Fatalln(err)
	}
	payloadSignature := computeHmac256(string(payloadData), deliveryTask.WebhookSecret)
	req, err := http.NewRequest("POST", deliveryTask.PayloadURL, bytes.NewBuffer(payloadData))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("X-Webhook-Signature", payloadSignature)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	// TODO Store and check if the URL is constantly failing
	// may be hash the url and put it in a cache along with an retried incrementer and if exceeds max incrementer value just disable web hook and send a alert email to developer.
	if resp.StatusCode == 200 {
		msg.Ack()
	} else {
		msg.Nack()
	}
	body, err := ioutil.ReadAll(resp.Body)
	log.Println("Response: ", string(body))
	resp.Body.Close()
	return nil
}

func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
