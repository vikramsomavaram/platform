/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"errors"
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

//CurrentServiceProviderProfile returns a service provider profile by its ID
func (r *queryResolver) CurrentServiceProviderProfile(ctx context.Context) (*models.ServiceProvider, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	providerProfile := models.GetServiceProviderByFilter(bson.D{{"user", user.ID}})
	if providerProfile.ID.IsZero() {
		return nil, errors.New("service provider not found")
	}
	if !providerProfile.IsActive {
		return providerProfile, &gqlerror.Error{Message: "your profile is currently inactive", Extensions: map[string]interface{}{"code": "provider_profile_inactive"}}
	}
	if providerProfile.Blocked {
		return providerProfile, &gqlerror.Error{Message: "your profile is currently blocked. please contact partner support.", Extensions: map[string]interface{}{"code": "provider_profile_blocked"}}
	}
	if providerProfile.ApprovedAt == nil {
		return providerProfile, &gqlerror.Error{Message: "your profile activation is pending", Extensions: map[string]interface{}{"code": "provider_profile_activation_pending"}}
	}
	return providerProfile, nil
}

//ServiceProvider returns a service provider by ID
func (r *queryResolver) ServiceProvider(ctx context.Context, id primitive.ObjectID) (*models.ServiceProvider, error) {
	serviceProvider := models.GetServiceProviderByID(id.Hex())
	if serviceProvider.ID.IsZero() {
		return nil, errors.New("service provider not found")
	}
	return serviceProvider, nil
}

func (r *mutationResolver) UnblockServiceProvider(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceProvider := models.GetServiceProviderByID(id.Hex())
	serviceProvider.Blocked = false
	res, err := models.UpdateServiceProvider(serviceProvider)
	if err != nil {
		return utils.PointerBool(false), err
	}
	if res.Blocked == false {
		return utils.PointerBool(true), nil
	}
	return utils.PointerBool(true), nil
}

func (r *mutationResolver) ApproveServiceProvider(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceProvider := models.GetServiceProviderByID(id.Hex())
	if serviceProvider.ApprovedAt != nil && serviceProvider.ApprovedBy != nil {
		return utils.PointerBool(false), &gqlerror.Error{Message: "profile already approved", Extensions: map[string]interface{}{"code": "profile_already_approved"}}
	}
	approvedAt := time.Now()
	serviceProvider.ApprovedAt = &approvedAt
	serviceProvider.ApprovedBy = &user.ID
	res, err := models.UpdateServiceProvider(serviceProvider)
	if err != nil {
		return utils.PointerBool(false), err
	}
	if res.ApprovedAt != nil {
		//TODO send welcome email,sms and push notification to service provider
		err = models.SendEmail("no-reply@tribe.cab", user.Email, "provider.profile.approved", user.Language, nil, nil)
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

type serviceProviderResolver struct{ *Resolver }

func (r *serviceProviderResolver) Company(ctx context.Context, obj *models.ServiceProvider) (primitive.ObjectID, error) {
	panic("implement me")
}

func (r *serviceProviderResolver) CompanyID(ctx context.Context, obj *models.ServiceProvider) (string, error) {
	panic("implement me")
}

func (r *serviceProviderResolver) Currency(ctx context.Context, obj *models.ServiceProvider) (*models.Currency, error) {
	currencyID := obj.Currency
	currency, err := models.GetCurrencyByID(currencyID)
	if err != nil {
		return nil, err
	}
	return currency, nil
}

func (r *serviceProviderResolver) User(ctx context.Context, obj *models.ServiceProvider) (*models.User, error) {
	user := models.GetUserByID(obj.User.Hex())
	return user, nil
}

//BankAccountDetails gives a list of bank account details
func (r *serviceProviderResolver) BankAccountDetails(ctx context.Context, obj *models.ServiceProvider) (*models.BankAccountDetails, error) {
	bankAccount, err := models.GetBankAccountByID(obj.BankDetails.Hex())
	if err != nil {
		return nil, err
	}
	bankAccountDetails := &models.BankAccountDetails{}
	_ = copier.Copy(&bankAccountDetails, &bankAccount)
	return bankAccountDetails, nil
}

//ServiceProviders gives a list of service provider
func (r *queryResolver) ServiceProviders(ctx context.Context, searchProvidersType *models.SearchProviderType, text *string, providerStatus *models.ProviderStatus, after *string, before *string, first *int, last *int) (*models.ServiceProviderConnection, error) {
	var items []*models.ServiceProvider
	var edges []*models.ServiceProviderEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetServiceProviders(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ServiceProviderEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ServiceProviderConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AddServiceProvider adds a new service provider
func (r *mutationResolver) AddServiceProvider(ctx context.Context, input models.AddServiceProviderInput) (*models.ServiceProvider, error) {
	serviceProvider := &models.ServiceProvider{}
	_ = copier.Copy(&serviceProvider, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceProvider.CreatedBy = user.ID
	bankAccount := &models.BankAccount{}
	bankAccount.UserID = user.ID
	bankAccount.CreatedBy = user.ID
	bankAccount, err = models.CreateBankAccount(bankAccount)
	if err != nil {
		return nil, err
	}
	serviceProvider.BankDetails = bankAccount.ID
	serviceProvider.User = user.ID
	serviceProvider, err = models.CreateServiceProvider(serviceProvider)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), serviceProvider.ID.Hex(), "service provider", serviceProvider, nil, ctx)
	return serviceProvider, nil
}

//UpdateServiceProvider updates an existing service provider
func (r *mutationResolver) UpdateServiceProvider(ctx context.Context, input models.UpdateServiceProviderInput) (*models.ServiceProvider, error) {
	serviceProvider := &models.ServiceProvider{}
	serviceProvider = models.GetServiceProviderByID(input.ID.Hex())
	_ = copier.Copy(&serviceProvider, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceProvider.CreatedBy = user.ID
	serviceProvider, err = models.UpdateServiceProvider(serviceProvider)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), serviceProvider.ID.Hex(), "service provider", serviceProvider, nil, ctx)
	return serviceProvider, nil
}

//DeleteServiceProvider deletes an existing service provider
func (r *mutationResolver) DeleteServiceProvider(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteServiceProviderByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "service provider", nil, nil, ctx)
	return &res, err
}

//DeactivateServiceProvider deactivates a service provider by ID
func (r *mutationResolver) DeactivateServiceProvider(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceProvider := models.GetServiceProviderByID(id.Hex())
	serviceProvider.IsActive = false
	_, err := models.UpdateServiceProvider(serviceProvider)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "service provider", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ActivateServiceProvider activates a service provider by ID
func (r *mutationResolver) ActivateServiceProvider(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceProvider := models.GetServiceProviderByID(id.Hex())
	if serviceProvider.ID.IsZero() {
		return utils.PointerBool(false), ErrServiceProviderNotFound
	}
	serviceProvider.IsActive = true
	_, err := models.UpdateServiceProvider(serviceProvider)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "service provider", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//UpdateServiceProviderProfile updates existing service provider profile
func (r *mutationResolver) UpdateServiceProviderProfile(ctx context.Context, input models.ServiceProviderProfileInput) (*models.ServiceProvider, error) {
	serviceProvider := &models.ServiceProvider{}
	serviceProvider = models.GetServiceProviderByID(input.ID.Hex())
	_ = copier.Copy(&serviceProvider, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceProvider.CreatedBy = user.ID
	serviceProvider, err = models.UpdateServiceProvider(serviceProvider)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), serviceProvider.ID.Hex(), "service provider profile", serviceProvider, nil, ctx)
	return serviceProvider, nil
}

//UpdateServiceProviderBankDetails updates existing service provider bank details
func (r *mutationResolver) UpdateServiceProviderBankDetails(ctx context.Context, input models.UpdateBankDetailsInput) (bool, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	bankAccount, err := models.GetBankAccountByID(input.ID.Hex())
	if err != nil {
		return false, err
	}
	_ = copier.Copy(&bankAccount, &input)
	bankAccount.CreatedBy = user.ID
	bankAccount, err = models.UpdateBankAccount(bankAccount)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), bankAccount.ID.Hex(), "provider bank details", bankAccount, nil, ctx)
	return !bankAccount.ID.IsZero(), nil
}

//BlockServiceProvider blocks service provider by ID
func (r *mutationResolver) BlockServiceProvider(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceProvider := models.GetServiceProviderByID(id.Hex())
	serviceProvider.Blocked = true
	res, err := models.UpdateServiceProvider(serviceProvider)
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
	go audit_log.NewAuditLogWithCtx(models.Blocked, user.ID.Hex(), id.Hex(), "service provider", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

func (r *mutationResolver) ServiceProviderSignUp(ctx context.Context, input models.ServiceProviderSignUpInput) (*models.ServiceProvider, error) {
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
	_ = copier.Copy(&address, &input.Address)
	provider := &models.ServiceProvider{}
	provider = models.GetServiceProviderByFilter(bson.D{{"user", user.ID}})
	if provider.ID.IsZero() {
		address := &models.Address{}
		_ = copier.Copy(&address, &input.Address)
		provider.User = user.ID
		provider.FirstName = user.FirstName
		provider.LastName = user.LastName
		provider.Email = user.Email
		provider.Password = user.Password
		provider.Gender = user.Gender
		provider.Country = user.Country
		provider.State = user.State
		provider.City = user.City
		provider.MobileNumber = user.MobileNo
		provider.Language = user.Language
		provider.Address = *address
		provider.ServiceCategory = []models.ServiceCategory{input.ServiceCategory}
		provider.ServiceSubCategory = []primitive.ObjectID{input.ServiceSubCategory}
		provider, err := models.CreateServiceProvider(provider)
		if err != nil && provider.ID.IsZero() {
			return nil, err
		}
		bankAccount := &models.BankAccount{}
		_ = copier.Copy(&bankAccount, &input.BankAccountDetails)
		bankAccount, err = models.CreateBankAccount(bankAccount)
		if err != nil {
			return provider, err
		}
		provider.BankDetails = bankAccount.ID
		provider, err = models.UpdateServiceProvider(provider)
		if err != nil {
			return provider, err
		}
		//TODO send welcome email,sms and push notification to service provider
		err = models.SendEmail("no-reply@tribe.cab", user.Email, "provider.signup.welcome", user.Language, nil, nil)
		if err != nil {
			log.Errorln(err)
		}
		//sent, err := msg91.SendMessage("Welcome to Tribe! "+user.OTP+" is your Verification OTP.", true, strings.TrimPrefix(user.MobileNo, "+"))
		//if !sent || err != nil {
		//	log.Errorln(err)
		//}
		return provider, &gqlerror.Error{Message: "profile activation pending", Extensions: map[string]interface{}{"code": "profile_activation_pending"}}
	}
	return provider, nil

}
