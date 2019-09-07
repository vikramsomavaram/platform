package main

import (
	"context"
	"encoding/json"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/i18n"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"time"
)

func main() {
	//Populate Countries & States
	//migrateCities(db)
	//migrateRoles(database.ConnectMongo())
	//db :=
	cache.ConnectRedis()
	database.ConnectMongo()
	//migrateRoles(db)
	//readEmailTemplateFiles("./data/email_templates_inputs/")
	//readSMSTemplateFiles("./data/sms_templates_inputs/")
}

// LocationCollection returns location.
func locationCollection(db *mongo.Database) {
	indexes := mongo.IndexModel{Keys: bsonx.Doc{{"location", bsonx.String("2dsphere")}}, Options: options.Index().SetUnique(false)}
	_, err := db.Collection("location_log").Indexes().CreateOne(context.Background(), indexes)
	if err != nil {
		log.Errorln(err)
	}
}

func migrateStates(db *mongo.Database) {
	indexes := mongo.IndexModel{Keys: bsonx.Doc{{"code", bsonx.Int32(1)}, {"continent", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}
	_, err := db.Collection(models.CountryCollection).Indexes().CreateOne(context.Background(), indexes)
	if err != nil {
		log.Errorln(err)
	}

	//query list of countries from db
	//append the states
}

func migratePermissions(db *mongo.Database) {
	perms := []models.UserRolePermissions{{CreatedAt: time.Now(), UpdatedAt: time.Now(), Service: "Job", Actions: []string{"create", "read", "update"}},
		{CreatedAt: time.Now(), UpdatedAt: time.Now(), Service: "ServiceCompany", Actions: []string{"create", "read", "update"}}}
	userRolePermissionsCollection := db.Collection(models.UserRolePermissionsCollection)
	for _, perm := range perms {
		_, err := userRolePermissionsCollection.InsertOne(context.TODO(), &perm)
		if err != nil {
			log.Errorln(err)
		}
	}
}

func migrateRoles(db *mongo.Database) {
	indexes := mongo.IndexModel{Keys: bsonx.Doc{{"name", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}
	_, err := db.Collection(models.UserRolesCollection).Indexes().CreateOne(context.Background(), indexes)
	if err != nil {
		log.Errorln(err)
	}
	indexes = mongo.IndexModel{Keys: bsonx.Doc{{"name", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}
	_, err = db.Collection(models.UserRoleGroupCollection).Indexes().CreateOne(context.Background(), indexes)
	if err != nil {
		log.Errorln(err)
	}
	perms := []models.UserRole{{CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "Admin", Description: "Super Admin role having access to entire system modules and features", Permissions: []string{"*:*"}},
		{CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "User", Description: "User account operations role", Permissions: []string{"User:*"}},
	}

	collection := db.Collection(models.UserRolesCollection)
	for _, perm := range perms {
		_, err := collection.InsertOne(context.TODO(), &perm)
		if err != nil {
			log.Errorln(err)
		}
	}
}

func migrateGroups(db *mongo.Database) {
	perms := []models.UserRoleGroup{{CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "SuperAdminsGroup", Roles: []string{}},
		{CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "UserAdminGroup", Roles: []string{}},
	}
	collection := db.Collection(models.UserRoleGroups)
	for _, perm := range perms {
		_, err := collection.InsertOne(context.TODO(), &perm)
		if err != nil {
			log.Errorln(err)
		}
	}
}

func migrateCities(db *mongo.Database) {
	indexes := mongo.IndexModel{Keys: bsonx.Doc{{"code", bsonx.Int32(1)}, {"continent", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}
	_, err := db.Collection(models.CitiesCollection).Indexes().CreateOne(context.Background(), indexes)
	if err != nil {
		log.Errorln(err)
	}
}

func migrateCountries(db *mongo.Database) {
	indexes := mongo.IndexModel{Keys: bsonx.Doc{{"code", bsonx.Int32(1)}, {"continent", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}
	_, err := db.Collection(models.CountryCollection).Indexes().CreateOne(context.Background(), indexes)
	if err != nil {
		log.Errorln(err)
	}

	for continent, val := range i18n.GetContinents() {
		continentData := reflect.ValueOf(val)
		for _, countrySlice := range continentData.MapKeys() {
			countriesRef := continentData.MapIndex(countrySlice)
			switch t := countriesRef.Interface().(type) {
			case string:
				_ = countriesRef.Interface().(string)
				logrus.Debug(t)
			case []string:
				println(continent)
				countries := countriesRef.Interface().([]string)
				for _, countryCode := range countries {
					country := &models.Country{}
					country.CreatedAt = time.Now()
					country.UpdatedAt = time.Now()
					country.Code = countryCode
					country.CountryName = i18n.GetCountryName(countryCode)
					country.PhoneCode = i18n.GetCountryPhoneCode(countryCode)
					country.Continent = continent
					country.ID = primitive.NewObjectID()
					installationCollection := db.Collection(models.CountryCollection)
					_, err := installationCollection.InsertOne(context.Background(), &country)
					if err != nil {
						log.Errorln(err)
					}
				}
			default:

			}
		}
	}
}

type SMSJsonTemplate struct {
	Purpose    string `json:"purpose"`
	TemplateID string `json:"template_id"`
	Templates  []struct {
		Language string `json:"language"`
		Body     string `json:"body"`
	} `json:"templates"`
}

func readSMSTemplateFiles(filePath string) {
	files, err := filepath.Glob(filePath + "*.json")
	if err != nil {
		log.Error(err)
	}
	sMSTemplate := &SMSJsonTemplate{}
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Error(err)
		}
		err = json.Unmarshal(content, &sMSTemplate)
		if err != nil {
			log.Error(err)
		}
		for _, template := range sMSTemplate.Templates {
			smsTemplate := &models.SMSTemplate{
				Body:       template.Body,
				Language:   template.Language,
				TemplateID: sMSTemplate.TemplateID,
				Purpose:    sMSTemplate.Purpose,
			}
			_, err := models.CreateSMSTemplate(smsTemplate)
			if err != nil {
				log.Error(err)
			}

		}
	}
}

type EmailJsonTemplate struct {
	Purpose    string `json:"purpose"`
	TemplateID string `json:"template_id"`
	Templates  []struct {
		Language string `json:"language"`
		Subject  string `json:"subject"`
		Body     string `json:"body"`
	} `json:"templates"`
}

func readEmailTemplateFiles(filePath string) {
	files, err := filepath.Glob(filePath + "*.json")
	if err != nil {
		log.Error(err)
	}
	emailTemplate := &EmailJsonTemplate{}
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Error(err)
		}
		err = json.Unmarshal(content, &emailTemplate)
		if err != nil {
			log.Error(err)
		}
		for _, template := range emailTemplate.Templates {
			mailTemplate := &models.EmailTemplate{
				Subject:    template.Subject,
				HTMLBody:   template.Body,
				TextBody:   template.Body,
				Language:   template.Language,
				TemplateID: emailTemplate.TemplateID,
				Purpose:    emailTemplate.Purpose,
			}
			_, err := models.CreateEmailTemplate(mailTemplate)
			if err != nil {
				log.Error(err)
			}
		}
	}
}
