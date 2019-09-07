/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package database

//CreateIndexes create mongodb indexes as required
func CreateIndexes() {
	//Create required unique indexes
	//indexes := mongo.IndexModel{Keys: bsonx.Doc{{"code", bsonx.Int32(1)}, {"continent", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}
	//_, err := MongoDB.Collection(models.CountryCollection).Indexes().CreateOne(context.Background(), indexes)
	//if err != nil {
	//	log.Errorln(err)
	//}
	//indexes := mongo.IndexModel{Keys: bsonx.Doc{{"location", bsonx.String("2dsphere")}}, Options: options.Index().SetUnique(false)}
	//_, err := db.Collection("location_log").Indexes().CreateOne(context.Background(), indexes)
	//if err != nil {
	//	log.Errorln(err)
	//}

	/*
		db := MongoDB
		bindata := mongo.IndexModel{Keys: bsonx.Doc{{"bin", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
		_, err := db.Collection(models.BINDataCollection).Indexes().CreateOne(context.Background(), bindata)
		if err != nil {
			log.Errorln(err)
		}

		users := []mongo.IndexModel{{Keys: bsonx.Doc{{"mobile_no", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)},
			{Keys: bsonx.Doc{{"email", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}}
		_, err = db.Collection(models.UsersCollection).Indexes().CreateMany(context.Background(), users)
		if err != nil {
			log.Errorln(err)
		}

		merchants := []mongo.IndexModel{{Keys: bsonx.Doc{{"mid", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)},
			{Keys: bsonx.Doc{{"email", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}}
		_, err = db.Collection(models.MerchantsCollection).Indexes().CreateMany(context.Background(), merchants)
		if err != nil {
			log.Errorln(err)
		}

		axisVpa := []mongo.IndexModel{{Keys: bsonx.Doc{{"vpa", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)},
			{Keys: bsonx.Doc{{"mid", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}}
		_, err = db.Collection(models.UPIVPACollection).Indexes().CreateMany(context.Background(), axisVpa)
		if err != nil {
			log.Errorln(err)
		}

		oauthApps := []mongo.IndexModel{{Keys: bsonx.Doc{{"client_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)},
			{Keys: bsonx.Doc{{"client_secret", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)},
			//{Keys: bsonx.Doc{{"merchant_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
			}
		_, err = db.Collection(models.OAuthApplicationsCollection).Indexes().CreateMany(context.Background(), oauthApps)
		if err != nil {
			log.Errorln(err)
		}

		payments := mongo.IndexModel{Keys: bsonx.Doc{{"merchant_id", bsonx.Int32(1)}, {"order_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)}
		_, err = db.Collection(models.PaymentsCollection).Indexes().CreateOne(context.Background(), payments)
		if err != nil {
			log.Errorln(err)
		}

	*/

	//webhook := mongo.IndexModel{Keys: bsonx.Doc{{"merchant_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
	//_, err = db.Collection(models.WebhooksCollection).Indexes().CreateOne(context.Background(), webhook)
	//if err != nil {
	//	log.Errorln(err)
	//}
	//
	//settlement := mongo.IndexModel{Keys: bsonx.Doc{{"merchant_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
	//_, err = db.Collection(models.SettlementsCollection).Indexes().CreateOne(context.Background(), settlement)
	//if err != nil {
	//	log.Errorln(err)
	//}
	//
	//invoice := mongo.IndexModel{Keys: bsonx.Doc{{"merchant_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
	//_, err = db.Collection(models.InvoiceCollection).Indexes().CreateOne(context.Background(), invoice)
	//if err != nil {
	//	log.Errorln(err)
	//}
	//
	//userRoles := mongo.IndexModel{Keys: bsonx.Doc{{"merchant_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
	//_, err = db.Collection(models.UserRolesCollection).Indexes().CreateOne(context.Background(), userRoles)
	//if err != nil {
	//	log.Errorln(err)
	//}

	/*
		userAuditLog := mongo.IndexModel{Keys: bsonx.Doc{{"user_id", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
		_, err = db.Collection(models.UserAuditLogsCollection).Indexes().CreateOne(context.Background(), userAuditLog)
		if err != nil {
			log.Errorln(err)
		}

		install := mongo.IndexModel{Keys: bsonx.Doc{{"fcm_token", bsonx.Int32(1)}}, Options: options.Index().SetUnique(false)}
		_, err = db.Collection(models.InstallationsCollection).Indexes().CreateOne(context.Background(), install)
		if err != nil {
			log.Errorln(err)
		}
	*/

}
