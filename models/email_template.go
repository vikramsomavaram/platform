/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/jordan-wright/email"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	cloudstorage "github.com/tribehq/platform/lib/storage"
	"github.com/tribehq/platform/utils/webhooks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"net/textproto"
	"os"
	"time"
)

// EmailTemplate represents a email template.
type EmailTemplate struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt  *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt  time.Time          `json:"updatedAt"bson:"updatedAt"`
	CreatedBy  primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Subject    string             `json:"subject" bson:"subject"`
	From       string             `json:"from" bson:"from"`
	HTMLBody   string             `json:"htmlBody" bson:"htmlBody"`
	TextBody   string             `json:"textBody" bson:"textBody"`
	Purpose    string             `json:"purpose" bson:"purpose" `
	Language   string             `json:"language" bson:"language"`
	TemplateID string             `json:"templateId" bson:"templateId"`
}

//UnmarshalBinary required for the redis cache to work
func (et *EmailTemplate) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, et); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (et *EmailTemplate) MarshalBinary() ([]byte, error) {
	return json.Marshal(et)
}

// EmailMessage represents a email message.
type EmailMessage struct {
	From        string      `json:"from" bson:"from"`
	To          string      `json:"to" bson:"to"`
	TemplateID  string      `json:"templateId" bson:"templateId"`
	Language    string      `json:"language" bson:"language"`
	Data        interface{} `json:"data" bson:"data"`
	Attachments []string    `json:"attachments" bson:"attachments"`
}

// CreateEmailTemplate creates new email templates.
func CreateEmailTemplate(emailTemplate *EmailTemplate) (*EmailTemplate, error) {
	emailTemplate.CreatedAt = time.Now()
	emailTemplate.UpdatedAt = time.Now()
	emailTemplate.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(EmailTemplateCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := installationCollection.InsertOne(ctx, &emailTemplate)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("email_template.created", &emailTemplate)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(emailTemplate.ID.Hex(), emailTemplate, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return emailTemplate, nil
}

// GetEmailTemplateByID gives the requested email template using id.
func GetEmailTemplateByID(ID string) (*EmailTemplate, error) {
	db := database.MongoDB
	emailTemplate := &EmailTemplate{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(emailTemplate)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(EmailTemplateCollection).FindOne(ctx, filter).Decode(&emailTemplate)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, emailTemplate, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return emailTemplate, nil
}

// GetEmailTemplateByFilter gives the email template after a specified filter.
func GetEmailTemplateByFilter(filter bson.D) (*EmailTemplate, error) {
	db := database.MongoDB
	emailTemplate := &EmailTemplate{}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}}) // we dont want deleted documents
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err := db.Collection(EmailTemplateCollection).FindOne(ctx, filter).Decode(&emailTemplate)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	return emailTemplate, nil
}

// GetEmailTemplates gives an array of email templates.
func GetEmailTemplates(filter bson.D, limit int, after *string, before *string, first *int, last *int) (emailTemplates []*EmailTemplate, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(EmailTemplateCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(EmailTemplateCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		emailTemplate := &EmailTemplate{}
		err = cur.Decode(&emailTemplate)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		emailTemplates = append(emailTemplates, emailTemplate)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return emailTemplates, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateEmailTemplate updates email templates.
func UpdateEmailTemplate(e *EmailTemplate) (*EmailTemplate, error) {
	emailTemplate := e
	emailTemplate.UpdatedAt = time.Now()
	filter := bson.D{{"_id", emailTemplate.ID}}
	db := database.MongoDB
	emailTemplateCollection := db.Collection(EmailTemplateCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := emailTemplateCollection.FindOneAndReplace(context.Background(), filter, emailTemplate, findRepOpts).Decode(&emailTemplate)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("email_template.updated", &emailTemplate)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(emailTemplate.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return emailTemplate, nil
}

// DeleteEmailTemplateByID deletes email templates by id.
func DeleteEmailTemplateByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	emailTemplateCollection := db.Collection(EmailTemplateCollection)
	res, err := emailTemplateCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("email_template.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// GetEmailContents returns email contents.
func GetEmailContents(templateID string, language string) *EmailTemplate {
	emailTemplate := &EmailTemplate{}
	filter := bson.D{{"language", language}, {"templateId", templateID}}
	emailTemplate, err := GetEmailTemplateByFilter(filter)
	if err != nil {
		log.Errorln(err)
	}
	return emailTemplate
}

//parseTemplateFile parses the template file.
func parseTemplateFile(templateFileName string, data interface{}) (string, error) {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func parseTemplate(templateID string, templateBody string, data interface{}) (string, error) {
	t, err := template.New(templateID).Parse(templateBody)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SendEmail sends email via smtp, attachments passed here are uploaded (already present) in Google Cloud Storage
func SendEmail(from string, to string, templateID string, language string, data interface{}, attachments []string) error {
	emailTemplate := GetEmailContents(templateID, language)
	textBody, err := parseTemplate(templateID+language, emailTemplate.TextBody, data)
	if err != nil {
		log.Errorln(err)
	}

	htmlBody, err := parseTemplate(templateID+language, emailTemplate.HTMLBody, data)
	if err != nil {
		log.Errorln(err)
	}
	e := &email.Email{
		To:      []string{to},
		From:    from,
		Subject: emailTemplate.Subject,
		Text:    []byte(textBody),
		HTML:    []byte(htmlBody),
		Headers: textproto.MIMEHeader{},
	}

	if len(attachments) > 0 {

		bucket := cloudstorage.StorageBucket
		bucketName := os.Getenv("GCS_BUCKET")
		if bucket == nil {
			log.Errorf("failed to get default GCS bucket name: %v", err)
		}

		for _, attachmentURL := range attachments {
			ctx := context.Background()
			file := bucket.Object(attachmentURL)
			fh, err := file.NewReader(ctx)
			if err != nil {
				log.Errorf("readFile: unable to open file from bucket %q, file %q: %v", bucketName, attachmentURL, err)
				return err
			}
			defer fh.Close()
			_, err = ioutil.ReadAll(fh)
			if err != nil {
				log.Errorf("readFile: unable to read data from bucket %q, file %q: %v", bucketName, attachmentURL, err)
				return err
			}

			attributes, err := file.Attrs(ctx)
			if err != nil {
				log.Errorln(err)
				return err
			}

			_, err = e.Attach(fh, attributes.Name, attributes.ContentType)
			if err != nil {
				log.Errorln(err)
				return err
			}

		}

	}

	err = e.Send(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_HOST")))
	if err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}
