/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// AdminReport represents a admin report.
type AdminReport struct {
	ID                 primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt          time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt          *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy          primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ServiceType        string             `json:"serviceType" bson:"serviceType"`
	OrderNo            string             `json:"orderNo" bson:"orderNo"`
	OrderDate          time.Time          `json:"orderDate" bson:"orderDate"`
	OrderAmount        string             `json:"orderAmount" bson:"orderAmount"`
	SiteCommission     string             `json:"siteCommission" bson:"siteCommission"`
	DeliveryCharges    string             `json:"deliveryCharges" bson:"deliveryCharges"`
	OutstandingAmount  string             `json:"outstandingAmount" bson:"outstandingAmount"`
	DriverAmount       string             `json:"driverAmount" bson:"driverAmount"`
	AdminEarningAmount string             `json:"adminEarningAmount" bson:"adminEarningAmount"`
	OrderStatus        string             `json:"orderStatus" bson:"orderStatus"`
	PaymentMethod      PaymentType        `json:"paymentMethod" bson:"paymentMethod"`
	IsActive           bool               `json:"isActive" bson:"isActive"`
}

// GetAdminReportByID gives the admin report by id.
func GetAdminReportByID(ID string) (*AdminReport, error) {
	db := database.MongoDB
	adminReport := &AdminReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(adminReport)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(AdminReportCollection).FindOne(ctx, filter).Decode(&adminReport)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, adminReport, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return adminReport, nil
}

// GetAdminReports gives a list of admin reports.
func GetAdminReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (adminReports []*AdminReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(AdminReportCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(AdminReportCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		adminReport := &AdminReport{}
		err = cur.Decode(&adminReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		adminReports = append(adminReports, adminReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return adminReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (adminReport *AdminReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, adminReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (adminReport *AdminReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(adminReport)
}

// JobRequestAcceptanceReport represents a job request acceptance report.
type JobRequestAcceptanceReport struct {
	ID                          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ProviderName                string             `json:"providerName" bson:"providerName"`
	TotalJobRequests            string             `json:"totalJobRequests" bson:"totalJobRequests"`
	RequestsAcceptedJobREquests string             `json:"requestsAcceptedJobREquests" bson:"requestsAcceptedJobREquests"`
	RequestsDeclined            string             `json:"requestsDeclined" bson:"requestsDeclined"`
	RequestsTimedOut            string             `json:"requestsTimedOut" bson:"requestsTimedOut"`
	MissedAttempts              string             `json:"missedAttempts" bson:"missedAttempts"`
	InProcessRequests           string             `json:"inProcessRequests" bson:"inProcessRequests"`
	AcceptancePercentage        string             `json:"acceptancePercentage" bson:"acceptancePercentage"`
	IsActive                    bool               `json:"isActive" bson:"isActive"`
}

// GetJobRequestAcceptanceReportByID gives a job request acceptance report by id.
func GetJobRequestAcceptanceReportByID(ID string) (*JobRequestAcceptanceReport, error) {
	db := database.MongoDB
	jobRequestAcceptanceReport := &JobRequestAcceptanceReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(jobRequestAcceptanceReport)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(JobRequestAcceptanceReportCollection).FindOne(ctx, filter).Decode(&jobRequestAcceptanceReport)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, jobRequestAcceptanceReport, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return jobRequestAcceptanceReport, nil
}

// GetJobRequestAcceptanceReports gives a list of job request acceptance reports.
func GetJobRequestAcceptanceReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (jobRequestAcceptanceReports []*JobRequestAcceptanceReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(JobRequestAcceptanceReportCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(JobRequestAcceptanceReportCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		jobRequestAcceptanceReport := &JobRequestAcceptanceReport{}
		err = cur.Decode(&jobRequestAcceptanceReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		jobRequestAcceptanceReports = append(jobRequestAcceptanceReports, jobRequestAcceptanceReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return jobRequestAcceptanceReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (jobRequestAcceptanceReport *JobRequestAcceptanceReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, jobRequestAcceptanceReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (jobRequestAcceptanceReport *JobRequestAcceptanceReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(jobRequestAcceptanceReport)
}

// JobTimeVariance represents a job time variance.
type JobTimeVariance struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt     *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy     primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	BookingNo     string             `json:"bookingNo" bson:"bookingNo"`
	Address       string             `json:"address" bson:"address"`
	JobDate       string             `json:"jobDate" bson:"jobDate"`
	Provider      string             `json:"provider" bson:"provider"`
	EstimatedTime string             `json:"estimatedTime" bson:"estimatedTime"`
	ActualTime    string             `json:"actualTime" bson:"actualTime"`
	Variance      string             `json:"variance" bson:"variance"`
	IsActive      bool               `json:"isActive" bson:"isActive"`
}

// GetJobTimeVarianceByID gives a job time variance by id.
func GetJobTimeVarianceByID(ID string) (*JobTimeVariance, error) {
	db := database.MongoDB
	jobTimeVariance := &JobTimeVariance{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(jobTimeVariance)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(JobTimeVarianceCollection).FindOne(ctx, filter).Decode(&jobTimeVariance)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, jobTimeVariance, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return jobTimeVariance, nil
}

// GetJobTimeVariances gives a list of job time variances.
func GetJobTimeVariances(filter bson.D, limit int, after *string, before *string, first *int, last *int) (jobTimeVariances []*JobTimeVariance, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(JobTimeVarianceCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(JobTimeVarianceCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		jobTimeVariance := &JobTimeVariance{}
		err = cur.Decode(&jobTimeVariance)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		jobTimeVariances = append(jobTimeVariances, jobTimeVariance)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return jobTimeVariances, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (jobTimeVariance *JobTimeVariance) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, jobTimeVariance); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (jobTimeVariance *JobTimeVariance) MarshalBinary() ([]byte, error) {
	return json.Marshal(jobTimeVariance)
}

// ProviderLogReport represents a provider lob report.
type ProviderLogReport struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name            string             `json:"name" bson:"name"`
	EMail           string             `json:"eMail" bson:"eMail"`
	OnlineTime      string             `json:"onlineTime" bson:"onlineTime"`
	OfflineTime     string             `json:"offlineTime" bson:"offlineTime"`
	TotalHoursLogIn string             `json:"totalHoursLogIn" bson:"totalHoursLogIn"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
}

// GetProviderLogReportByID gives a provider log report by id.
func GetProviderLogReportByID(ID string) (*ProviderLogReport, error) {
	db := database.MongoDB
	providerLogReport := &ProviderLogReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(providerLogReport)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProviderLogReportCollection).FindOne(ctx, filter).Decode(&providerLogReport)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, providerLogReport, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return providerLogReport, nil
}

// GetProviderLogReports gives a list of provider log reports.
func GetProviderLogReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (providerLogReports []*ProviderLogReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProviderLogReportCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProviderLogReportCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		providerLogReport := &ProviderLogReport{}
		err = cur.Decode(&providerLogReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		providerLogReports = append(providerLogReports, providerLogReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return providerLogReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (providerLogReport *ProviderLogReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, providerLogReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (providerLogReport *ProviderLogReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(providerLogReport)
}

// ProviderPaymentReport represents a provider payment report.
type ProviderPaymentReport struct {
	ID                                                   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                                            time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                                            *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                                            time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                                            primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ProviderName                                         string             `json:"providerName" bson:"providerName"`
	ProviderBankDetails                                  string             `json:"providerBankDetails" bson:"providerBankDetails"`
	TotalJobCommissionTakeFromProviderForCashJobs        string             `json:"totalJobCommissionTakeFromProviderForCashJobs" bson:"totalJobCommissionTakeFromProviderForCashJobs"`
	TotalJobAmountPayToProviderForCardJobs               string             `json:"totalJobAmountPayToProviderForCardJobs" bson:"totalJobAmountPayToProviderForCardJobs"`
	TotalFare                                            string             `json:"totalFare" bson:"totalFare"`
	TotalCashReceived                                    string             `json:"totalCashReceived" bson:"totalCashReceived"`
	TotalTaxAmountPayToProvider                          string             `json:"totalTaxAmountPayToProvider" bson:"totalTaxAmountPayToProvider"`
	TotalTipAmountPayToProvider                          string             `json:"totalTipAmountPayToProvider" bson:"totalTipAmountPayToProvider"`
	TotalWalletAdjustmentAmountPaytoProviderForCashJobs  string             `json:"totalWalletAdjustmentAmountPaytoProviderForCashJobs" bson:"totalWalletAdjustmentAmountPaytoProviderForCashJobs"`
	TotalCouponDiscountAmountPayToProviderForCashJobs    string             `json:"totalCouponDiscountAmountPayToProviderForCashJobs" bson:"totalCouponDiscountAmountPayToProviderForCashJobs"`
	TotalJobOutstandingAmountTakeFromProviderForCashJobs string             `json:"totalJobOutstandingAmountTakeFromProviderForCashJobs" bson:"totalJobOutstandingAmountTakeFromProviderForCashJobs"`
	TotalJobBookingFeeForCash                            string             `json:"totalJobBookingFeeForCash" bson:"totalJobBookingFeeForCash"`
	FinalAmountPayToProvider                             string             `json:"finalAmountPayToProvider" bson:"finalAmountPayToProvider"`
	FinalAmountToTakebackFromProvider                    string             `json:"finalAmountToTakebackFromProvider" bson:"finalAmountToTakebackFromProvider"`
	ProviderPaymentStatus                                string             `json:"providerPaymentStatus" bson:"providerPaymentStatus"`
	IsActive                                             bool               `json:"isActive" bson:"isActive"`
}

// GetProviderPaymentReportByID gives provider payment report by id.
func GetProviderPaymentReportByID(ID string) (*ProviderPaymentReport, error) {
	db := database.MongoDB
	providerPaymentReport := &ProviderPaymentReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(providerPaymentReport)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProviderPaymentReportCollection).FindOne(ctx, filter).Decode(&providerPaymentReport)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, providerPaymentReport, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return providerPaymentReport, nil
}

// GetProviderPaymentReports gives a list of provider payment reports.
func GetProviderPaymentReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (providerPaymentReports []*ProviderPaymentReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProviderPaymentReportCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProviderPaymentReportCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		providerPaymentReport := &ProviderPaymentReport{}
		err = cur.Decode(&providerPaymentReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		providerPaymentReports = append(providerPaymentReports, providerPaymentReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return providerPaymentReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (providerPaymentReport *ProviderPaymentReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, providerPaymentReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (providerPaymentReport *ProviderPaymentReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(providerPaymentReport)
}

// StorePaymentReport represents a store payment report.
type StorePaymentReport struct {
	ID                primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt         *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy         primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	SelectStore       string             `json:"selectStore" bson:"selectStore"`
	ServiceType       string             `json:"serviceType" bson:"serviceType"`
	StoreName         string             `json:"storeName" bson:"storeName"`
	StoreAccountName  string             `json:"storeAccountName" bson:"storeAccountName"`
	BankName          string             `json:"bankName" bson:"bankName"`
	AccountNumber     string             `json:"accountNumber" bson:"accountNumber"`
	SortCode          string             `json:"sortCode" bson:"sortCode"`
	FinalAmount       string             `json:"finalAmount" bson:"finalAmount"`
	PaymentStatus     PaymentStatus      `json:"paymentStatus" bson:"paymentStatus"`
	PaymentMethod     PaymentType        `json:"paymentMethod" bson:"paymentMethod"`
	OrderNumber       string             `json:"orderNumber" bson:"orderNumber"`
	ProviderName      string             `json:"providerName" bson:"providerName"`
	UserName          string             `json:"userName" bson:"userName"`
	OrderDate         time.Time          `json:"orderDate" bson:"orderDate"`
	OrderStatus       string             `json:"orderStatus" bson:"orderStatus"`
	OrderAmount       string             `json:"orderAmount" bson:"orderAmount"`
	SiteCommission    string             `json:"siteCommission" bson:"siteCommission"`
	DeliveryCharges   string             `json:"deliveryCharges" bson:"deliveryCharges"`
	OfferAmount       string             `json:"offerAmount" bson:"offerAmount"`
	OutstandingAmount string             `json:"outstandingAmount" bson:"outstandingAmount"`
	StoreAmount       string             `json:"storeAmount" bson:"storeAmount"`
	IsActive          bool               `json:"isActive" bson:"isActive"`
}

// GetStorePaymentReportByID gives a store payment report by id.
func GetStorePaymentReportByID(ID string) (*StorePaymentReport, error) {
	db := database.MongoDB
	storePaymentReport := &StorePaymentReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(storePaymentReport)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(StorePaymentReportCollection).FindOne(ctx, filter).Decode(&storePaymentReport)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, storePaymentReport, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return storePaymentReport, nil
}

// GetStorePaymentReports gives a list of store payment reports.
func GetStorePaymentReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (storePaymentReports []*StorePaymentReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(StorePaymentReportCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(StorePaymentReportCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		storePaymentReport := &StorePaymentReport{}
		err = cur.Decode(&storePaymentReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		storePaymentReports = append(storePaymentReports, storePaymentReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return storePaymentReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (storePaymentReport *StorePaymentReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, storePaymentReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (storePaymentReport *StorePaymentReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(storePaymentReport)
}

// CancelledReport represents a cancelled report.
type CancelledReport struct {
	ID                  primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt           time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt           *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt           time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy           primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ServiceType         string             `json:"serviceType" bson:"serviceType"`
	OrderNumber         string             `json:"orderNumber" bson:"orderNumber"`
	OrderDate           time.Time          `json:"orderDate" bson:"orderDate"`
	PayoutStore         string             `json:"payoutStore" bson:"payoutStore"`
	PayoutDriver        string             `json:"payoutDriver" bson:"payoutDriver"`
	CancellationCharges string             `json:"cancellationCharges" bson:"cancellationCharges"`
	OrderStatus         string             `json:"orderStatus" bson:"orderStatus"`
	PaymentMethod       PaymentType        `json:"paymentMethod" bson:"paymentMethod"`
	Action              PaymentStatus      `json:"action" bson:"action"`
	IsActive            bool               `json:"isActive" bson:"isActive"`
}

// GetCancelledReportByID gives a cancelled report by id.
func GetCancelledReportByID(ID string) (*CancelledReport, error) {
	db := database.MongoDB
	cancelledReport := &CancelledReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(cancelledReport)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(CancelledReportCollection).FindOne(ctx, filter).Decode(&cancelledReport)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, cancelledReport, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return cancelledReport, nil
}

// GetCancelledReports gives a list of cancelled report.
func GetCancelledReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (cancelledReports []*CancelledReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CancelledReportCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CancelledReportCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		cancelledReport := &CancelledReport{}
		err = cur.Decode(&cancelledReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		cancelledReports = append(cancelledReports, cancelledReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return cancelledReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (cancelledReport *CancelledReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, cancelledReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (cancelledReport *CancelledReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(cancelledReport)
}

// UserWalletReport represents a user wallet report.
type UserWalletReport struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Description     string             `json:"description" bson:"description"`
	Amount          string             `json:"amount" bson:"amount"`
	BookingNo       string             `json:"bookingNo" bson:"bookingNo"`
	TransactionDate string             `json:"transactionDate" bson:"transactionDate"`
	BalanceType     string             `json:"balanceType" bson:"balanceType"`
	Type            string             `json:"type" bson:"type"`
	Balance         string             `json:"balance" bson:"balance"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
}

// GetUserWalletReportByID gives a user wallet report by id.
func GetUserWalletReportByID(ID string) (*UserWalletReport, error) {
	db := database.MongoDB
	userWalletReport := &UserWalletReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(userWalletReport)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(UserWalletReportCollection).FindOne(ctx, filter).Decode(&userWalletReport)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, userWalletReport, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return userWalletReport, nil
}

// GetUserWalletReports gives a list of user wallet reports.
func GetUserWalletReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (userWalletReports []*UserWalletReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(UserWalletReportCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(UserWalletReportCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		userWalletReport := &UserWalletReport{}
		err = cur.Decode(&userWalletReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		userWalletReports = append(userWalletReports, userWalletReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return userWalletReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (userWalletReport *UserWalletReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, userWalletReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (userWalletReport *UserWalletReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(userWalletReport)
}

// AdminDashboard represents a admin dashboard.
type AdminDashboard struct {
	ID                primitive.ObjectID       `json:"id,omitempty" bson:"_id,omitempty"`
	SiteStatistics    SiteStatistics           `json:"siteStatistics" bson:"siteStatistics"`
	JobStatistics     JobStatistics            `json:"jobStatistics" bson:"jobStatistics"`
	JobsOverview      JobsOverview             `json:"jobsOverview" bson:"jobsOverview"`
	ProvidersOverview ProvidersOverview        `json:"providersOverview" bson:"providersOverview"`
	LatestJobs        []*LatestJobs            `json:"latestJobs" bson:"latestJobs"`
	NotificationPanel []*DashboardNotification `json:"notificationPanel" bson:"notificationPanel"`
}
