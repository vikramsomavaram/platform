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
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrServiceSubCategoryNotFound = errors.New("service sub category not found")
var ErrServiceProviderNotFound = errors.New("service provider not found")

//CancelReason returns a cancel reason by ID
func (r *queryResolver) CancelReason(ctx context.Context, id primitive.ObjectID) (*models.CancelReason, error) {
	cancelReason := models.GetCancelReasonByID(id.Hex())
	return cancelReason, nil
}

//JobLaterBooking returns a job later booking by ID
func (r *queryResolver) JobLaterBooking(ctx context.Context, id primitive.ObjectID) (*models.JobLaterBooking, error) {
	booking := models.GetJobLaterBookingByID(id.Hex())
	return booking, nil
}

//CancelReasons gives a list of cancel reasons
func (r *queryResolver) CancelReasons(ctx context.Context, searchCancelReasonType *models.SearchCancelReasonType, text *string, searchCancelReasonStatus *models.SearchCancelReasonStatus, cancelReasonServiceType *models.CancelReasonServiceType, after *string, before *string, first *int, last *int) (*models.CancelReasonConnection, error) {
	var items []*models.CancelReason
	var edges []*models.CancelReasonEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetCancelReasons(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CancelReasonEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.CancelReasonConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//JobLaterBookings returns a job later booking by ID
func (r *queryResolver) JobLaterBookings(ctx context.Context, jobLaterSearchServiceType *models.JobLaterSearchServiceType, jobLaterType *models.JobLaterType, text *string, after *string, before *string, first *int, last *int) (*models.JobLaterBookingConnection, error) {
	var items []*models.JobLaterBooking
	var edges []*models.JobLaterBookingEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetJobLaterBookings(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.JobLaterBookingEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.JobLaterBookingConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//TODO
//EstimateBookingFare gives an estimate booking fare
func (r *mutationResolver) EstimateBookingFare(ctx context.Context, input models.BookingInput) (*models.BookingFareEstimate, error) {
	//Just do the dry run and report back the fare to the user
	panic("implement me")
}

//CreateBooking creates a new booking
func (r *mutationResolver) CreateBooking(ctx context.Context, input models.BookingInput) (*models.Booking, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	couponValidity := models.GetCouponByFilter(bson.D{{"code", input.Coupon}})
	if couponValidity.ID.IsZero() {
		return nil, errors.New("invalid coupon code")
	}
	//TODO
	//check for validity
	//check for usage limit, previous use by the user
	//check for minimum amount and maximum amount applicable to the coupon
	serviceSubCategory := models.GetServiceSubCategoryByID(input.ServiceSubCategoryID.Hex())
	if serviceSubCategory.ID.IsZero() {
		return nil, ErrServiceSubCategoryNotFound
	}
	service := models.GetServiceByID(serviceSubCategory.ServiceID)
	if service.ID.IsZero() {
		return nil, ErrServiceNotFound
	}
	//TODO check for service in that service location
	var provider models.ServiceProvider
	if input.ProviderID != "" {
		provider = *models.GetServiceProviderByID(input.ProviderID)
		if provider.ID.IsZero() {
			return nil, ErrServiceProviderNotFound
		}
	}
	emailTemplateID := ""
	//smsTemplateID:=""
	//pushNotificationTemplateID:=""

	job := &models.Job{}

	switch service.Category {
	case models.ServiceCategoryTaxiService:
	case models.ServiceCategoryDeliveryService:
	case models.ServiceCategoryRentalService:
	case models.ServiceCategoryProfessionalService:
		//filter for service providers in that service location
		//check for service provider schedule(personal availability + existing job bookings)
		//if availability is true the create job
		//create a job
		address := &models.Address{}
		_ = copier.Copy(&address, &input.OtherServiceDetails.DeliveryAddress)

		job.CreatedBy = user.ID
		job.JobType = service.Category
		job.ToAddress = *address
		job.JobDate = *input.OtherServiceDetails.Schedule
		job.CompanyID = provider.CompanyID.Hex()
		job.ProviderID = provider.ID.Hex()
		job.UserID = user.ID.Hex()
		job.ServiceOrderItems = &input.OtherServiceDetails.ServiceOrderItems

		job, err := models.CreateJob(job)
		if err != nil {
			return nil, err
		}
		if job.ID.IsZero() {
			return nil, errors.New("internal server error")
		}

		emailTemplateID = "provider.professional_job.created"
		//smsTemplateID="provider.professional_job.created"
		//pushNotificationTemplateID="provider.professional_job.created"

	default:
		return nil, errors.New("internal server error")

	}
	//notify service provider via email sms and push notification
	//Send a email
	err = models.SendEmail("noreply@tribe.cab", "", emailTemplateID, user.Language, nil, nil)
	if err != nil {
		log.Errorln(err)
	}
	//Send verification SMS
	//sent, err := msg91.SendMessage("Welcome to Tribe! "+user.OTP+" is your Verification OTP.", true, strings.TrimPrefix(user.MobileNo, "+"))
	//if !sent || err != nil {
	//	log.Errorln(err)
	//}
	//TODO write push notification logic for service providers

	// TODO
	// get the service user is trying to book
	//create temporary user if user doesn't exists
	//check for service provider availability
	//check for promocode validity
	//check for serviceable location
	//check for prepaid or postpaid
	//send back the response
	booking := &models.Booking{
		JobID:     &job.ID,
		VehicleID: job.ServiceVehicleID,
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), job.ID.Hex(), "job", job, nil, ctx)
	return booking, nil
}

type cancelReasonResolver struct{ *Resolver }

// AddCancelReason adds cancel reason.
func (r *mutationResolver) AddCancelReason(ctx context.Context, input models.AddCancelReasonInput) (*models.CancelReason, error) {
	cancelReason := &models.CancelReason{}
	_ = copier.Copy(&cancelReason, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	cancelReason.CreatedBy = user.ID
	cancelReason, err = models.CreateCancelReason(*cancelReason)
	if err != nil {
		return nil, err
	}

	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), cancelReason.ID.Hex(), "cancel reason", cancelReason, nil, ctx)
	return cancelReason, nil
}

// UpdateCancelReason updates the cancel reason.
func (r *mutationResolver) UpdateCancelReason(ctx context.Context, input models.UpdateCancelReasonInput) (*models.CancelReason, error) {
	cancelReason := &models.CancelReason{}
	cancelReason = models.GetCancelReasonByID(input.ID)
	_ = copier.Copy(&cancelReason, &input)

	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	cancelReason.CreatedBy = user.ID
	cancelReason, err = models.UpdateCancelReason(cancelReason)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), cancelReason.ID.Hex(), "cancel reason", cancelReason, nil, ctx)
	return cancelReason, nil
}

// DeleteCancelReason deletes cancel reason.
func (r *mutationResolver) DeleteCancelReason(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteCancelReasonByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "cancel reason", nil, nil, ctx)
	return &res, err
}

// DeactivateCancelReason deactivates cancel reason.
func (r *mutationResolver) DeactivateCancelReason(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	cancelReason := models.GetCancelReasonByID(id.Hex())
	if cancelReason.ID.IsZero() {
		return utils.PointerBool(false), errors.New("cancel reason not found")
	}
	cancelReason.IsActive = false
	_, err := models.UpdateCancelReason(cancelReason)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "cancel reason", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// ActivateCancelReason activates cancel reason.
func (r *mutationResolver) ActivateCancelReason(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	cancelReason := models.GetCancelReasonByID(id.Hex())
	if cancelReason.ID.IsZero() {
		return utils.PointerBool(false), errors.New("cancel reason not found")
	}
	cancelReason.IsActive = true
	_, err := models.UpdateCancelReason(cancelReason)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "cancel reason", nil, nil, ctx)
	return utils.PointerBool(true), nil
}
