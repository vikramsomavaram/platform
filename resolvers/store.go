/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/gqlerror"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

//Stores gives a list of stores
func (r *queryResolver) Stores(ctx context.Context, storeType *models.StoreType, text *string, storeStatus *models.StoreStatus, storeCategory *models.StoreCategory, after *string, before *string, first *int, last *int) (*models.StoreConnection, error) {
	var items []*models.Store
	var edges []*models.StoreEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetStores(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.StoreEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.StoreConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//Store returns a store by its ID
func (r *queryResolver) Store(ctx context.Context, id primitive.ObjectID) (*models.Store, error) {
	store := models.GetStoreByID(id.Hex())
	return store, nil
}

//AddStore adds a new store
func (r *mutationResolver) AddStore(ctx context.Context, input models.AddStoreInput) (*models.Store, error) {
	store := &models.Store{}
	_ = copier.Copy(&store, &input)
	store, err := models.CreateStore(*store)
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	store.CreatedBy = user.ID
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), store.ID.Hex(), "store", store, nil, ctx)
	return store, nil
}

//UpdateStore updates an existing store
func (r *mutationResolver) UpdateStore(ctx context.Context, input models.UpdateStoreInput) (*models.Store, error) {
	store := &models.Store{}
	store = models.GetStoreByID(input.ID.Hex())
	_ = copier.Copy(&store, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	store.CreatedBy = user.ID
	store, err = models.UpdateStore(store)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), store.ID.Hex(), "store", store, nil, ctx)
	return store, nil
}

//DeleteStore deletes an existing store
func (r *mutationResolver) DeleteStore(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteStoreByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "store", nil, nil, ctx)
	return &res, err
}

//ActivateStore activates a store by its ID
func (r *mutationResolver) ActivateStore(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	store := models.GetStoreByID(id.Hex())
	store.IsActive = true
	_, err := models.UpdateStore(store)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "store", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateStore deactivates a store by its ID
func (r *mutationResolver) DeactivateStore(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	store := models.GetStoreByID(id.Hex())
	store.IsActive = false
	_, err := models.UpdateStore(store)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "store", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// storeResolver is of type struct.
type storeResolver struct{ *Resolver }

func (r storeResolver) ServiceCategory(ctx context.Context, obj *models.Store) (models.StoreCategory, error) {
	panic("implement me")
}

func (r storeResolver) StoreLocations(ctx context.Context, obj *models.Store) ([]*models.StoreLocation, error) {
	panic("implement me")
}

// storeReviewResolver is of type struct.
type storeReviewResolver struct{ *Resolver }

func (r *mutationResolver) AddStoreLocation(ctx context.Context, input models.AddStoreLocationInput) (*models.StoreLocation, error) {
	storeLocation := &models.StoreLocation{}
	_ = copier.Copy(&storeLocation, &input)
	storeLocation, err := models.CreateStoreLocation(*storeLocation)
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	storeLocation.CreatedBy = user.ID
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), storeLocation.ID.Hex(), "storeLocation", storeLocation, nil, ctx)
	return storeLocation, nil
}

func (r *mutationResolver) UpdateStoreLocation(ctx context.Context, input models.UpdateStoreLocationInput) (*models.StoreLocation, error) {
	storeLocation := &models.StoreLocation{}
	storeLocation, err := models.GetStoreLocationByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&storeLocation, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	storeLocation.CreatedBy = user.ID
	storeLocation, err = models.UpdateStoreLocation(storeLocation)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), storeLocation.ID.Hex(), "storeLocation", storeLocation, nil, ctx)
	return storeLocation, nil
}

func (r *mutationResolver) DeleteStoreLocation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteStoreLocationByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "storeLocation", nil, nil, ctx)
	return &res, err
}

func (r *queryResolver) StoreLocations(ctx context.Context, storeLocationID *primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.StoreLocationConnection, error) {
	var items []*models.StoreLocation
	var edges []*models.StoreLocationEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetStoreLocations(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.StoreLocationEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.StoreLocationConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

func (r *queryResolver) StoreLocation(ctx context.Context, id primitive.ObjectID) (*models.StoreLocation, error) {
	storeLocation, err := models.GetStoreLocationByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return storeLocation, nil
}

//BankAccountDetails gives a list of bank account details
func (r *storeResolver) BankAccountDetails(ctx context.Context, obj *models.Store) (*models.BankAccountDetails, error) {
	bankAccount, err := models.GetBankAccountByID(obj.BankAccountDetails.Hex())
	if err != nil {
		return nil, err
	}
	bankAccountDetails := &models.BankAccountDetails{}
	_ = copier.Copy(&bankAccountDetails, &bankAccount)
	return bankAccountDetails, nil
}

func (r *mutationResolver) StoreSignUp(ctx context.Context, input models.StoreSignUpInput) (*models.Store, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	if !user.IsMobileVerified {
		return nil, &gqlerror.Error{Message: "mobile number verification required", Extensions: map[string]interface{}{"code": "mobile_number_not_verified"}}
	}
	if !user.IsEmailVerified {
		return nil, &gqlerror.Error{Message: "email address verification required", Extensions: map[string]interface{}{"code": "email_address_not_verified"}}
	}
	address := &models.Address{}
	_ = copier.Copy(&address, input.StoreAddress)
	storeLocation := &models.StoreLocation{}
	_ = copier.Copy(&storeLocation, input.StoreLocation)
	store := &models.Store{}
	store.StoreName = input.StoreName
	store.StoreLocation = *storeLocation
	store.ServiceCategory = input.ServiceCategory
	store.Email = input.Email
	store.MobileNumber = input.MobileNumber
	store.ContactPersonName = input.ContactPersonName
	store.ZipCode = input.ZipCode
	store.Country = input.Country
	store.State = input.State
	store.StoreAddress = *address
	store.StoreLogo = input.StoreLogo
	store.BankAccountDetails = input.BankAccountDetails
	store.Language = input.Language
	store, err = models.CreateStore(*store)
	if err != nil && store.ID.IsZero() {
		return nil, err
	}
	bankAccount := &models.BankAccount{}
	_ = copier.Copy(&bankAccount, &input.BankAccountDetails)
	bankAccount, err = models.CreateBankAccount(bankAccount)
	if err != nil {
		return store, err
	}
	store.BankAccountDetails = bankAccount.ID
	store, err = models.UpdateStore(store)
	if err != nil {
		return store, err
	}
	//TODO send welcome email,sms and push notification to service provider
	err = models.SendEmail("no-reply@tribe.cab", store.Email, "store.signup.welcome", store.Language, nil, nil)
	if err != nil {
		log.Errorln(err)
	}
	//sent, err := msg91.SendMessage("Welcome to Tribe! "+user.OTP+" is your Verification OTP.", true, strings.TrimPrefix(user.MobileNo, "+"))
	//if !sent || err != nil {
	//	log.Errorln(err)
	//}
	return store, &gqlerror.Error{Message: "profile activation pending", Extensions: map[string]interface{}{"code": "profile_activation_pending"}}
}

func (r *mutationResolver) UnblockStore(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	store := models.GetStoreByID(id.Hex())
	store.Blocked = false
	res, err := models.UpdateStore(store)
	if err != nil {
		return utils.PointerBool(false), err
	}
	if res.Blocked == false {
		return utils.PointerBool(true), nil
	}
	return utils.PointerBool(true), nil
}

func (r *mutationResolver) ApproveStore(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	store := models.GetStoreByID(id.Hex())
	if store.ApprovedAt != nil && store.ApprovedBy != nil {
		return utils.PointerBool(false), &gqlerror.Error{Message: "store already approved", Extensions: map[string]interface{}{"code": "store_already_approved"}}
	}
	approvedAt := time.Now()
	store.ApprovedAt = &approvedAt
	store.ApprovedBy = &user.ID
	res, err := models.UpdateStore(store)
	if err != nil {
		return utils.PointerBool(false), err
	}
	if res.ApprovedAt != nil {
		//TODO send welcome email,sms and push notification to service provider
		err = models.SendEmail("no-reply@tribe.cab", user.Email, "store.profile.approved", user.Language, nil, nil)
		if err != nil {
			log.Errorln(err)
		}
		//sent, err := msg91.SendMessage("Welcome to Tribe! "+user.OTP+" is your Verification OTP.", true, strings.TrimPrefix(user.MobileNo, "+"))
		//if !sent || err != nil {
		//	log.Errorln(err)
		//}
		return utils.PointerBool(true), nil
	}
	return utils.PointerBool(true), nil
}

func (r *mutationResolver) BlockStore(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	store := models.GetStoreByID(id.Hex())
	store.Blocked = true
	res, err := models.UpdateStore(store)
	if err != nil {
		return utils.PointerBool(false), err
	}
	if res.Blocked == true {
		return utils.PointerBool(true), nil
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Blocked, user.ID.Hex(), id.Hex(), "store", nil, nil, ctx)
	return utils.PointerBool(true), nil
}
