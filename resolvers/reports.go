/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

//AdminReports gives a list of admin reports
func (r *queryResolver) AdminReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, storeSearch *string, paymentType *models.PaymentType, serviceType *string, text *string, after *string, before *string, first *int, last *int) (*models.AdminReportConnection, error) {
	var items []*models.AdminReport
	var edges []*models.AdminReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetAdminReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.AdminReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.AdminReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AdminReport returns an admin report by its ID
func (r *queryResolver) AdminReport(ctx context.Context, id primitive.ObjectID) (*models.AdminReport, error) {
	report, err := models.GetAdminReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

//StorePaymentReports gives a list of store payment reports
func (r *queryResolver) StorePaymentReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, storeSearch *string, serviceType *string, paymentStatus *models.PaymentStatus, text *string, after *string, before *string, first *int, last *int) (*models.StorePaymentReportConnection, error) {
	var items []*models.StorePaymentReport
	var edges []*models.StorePaymentReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetStorePaymentReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.StorePaymentReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.StorePaymentReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//StorePaymentReport returns a store payment report by its ID
func (r *queryResolver) StorePaymentReport(ctx context.Context, id primitive.ObjectID) (*models.StorePaymentReport, error) {
	report, err := models.GetStorePaymentReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

//ProviderPaymentReports gives a list of provider payment reports
func (r *queryResolver) ProviderPaymentReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, report *string, provider *string, paymentType *models.PaymentType, paymentStatus *models.PaymentStatus, serviceType *string, text *string, after *string, before *string, first *int, last *int) (*models.ProviderPaymentReportConnection, error) {
	var items []*models.ProviderPaymentReport
	var edges []*models.ProviderPaymentReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetProviderPaymentReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProviderPaymentReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ProviderPaymentReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//ProviderPaymentReport returns a provider payment report by its ID
func (r *queryResolver) ProviderPaymentReport(ctx context.Context, id primitive.ObjectID) (*models.ProviderPaymentReport, error) {
	report, err := models.GetProviderPaymentReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

//CancelledReports gives a list of cancelled reports
func (r *queryResolver) CancelledReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, paymentType *models.PaymentType, serviceType *string, text *string, after *string, before *string, first *int, last *int) (*models.CancelledReportConnection, error) {
	var items []*models.CancelledReport
	var edges []*models.CancelledReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetCancelledReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CancelledReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.CancelledReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//CancelledReport returns a cancelled report by its ID
func (r *queryResolver) CancelledReport(ctx context.Context, id primitive.ObjectID) (*models.CancelledReport, error) {
	report, err := models.GetCancelledReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

//UserWalletReports gives a list of user wallet reports
func (r *queryResolver) UserWalletReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, searchByUser *models.UserWalletReportSearchByUserType, searchByPayment *models.WalletTransactionType, searchByBalance *models.UserWalletReportSearchByBalanceType, selectProvider *string, selectUser *string, text *string, after *string, before *string, first *int, last *int) (*models.UserWalletReportConnection, error) {
	var items []*models.UserWalletReport
	var edges []*models.UserWalletReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetUserWalletReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.UserWalletReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.UserWalletReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//UserWalletReport returns a user wallet report by its ID
func (r *queryResolver) UserWalletReport(ctx context.Context, id primitive.ObjectID) (*models.UserWalletReport, error) {
	report, err := models.GetUserWalletReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

type adminReportResolver struct{ *Resolver }

type providerLogReportResolver struct{ *Resolver }

//ProviderLogReports give a list of provider log reports
func (r *queryResolver) ProviderLogReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, provider *string, text *string, after *string, before *string, first *int, last *int) (*models.ProviderLogReportConnection, error) {
	var items []*models.ProviderLogReport
	var edges []*models.ProviderLogReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetProviderLogReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProviderLogReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ProviderLogReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//ProviderLogReport returns a provider log report by its ID
func (r *queryResolver) ProviderLogReport(ctx context.Context, id primitive.ObjectID) (*models.ProviderLogReport, error) {
	report, err := models.GetProviderLogReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

//JobRequestAcceptanceReports gives a list of job request acceptance reports
func (r *queryResolver) JobRequestAcceptanceReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, provider *string, text *string, after *string, before *string, first *int, last *int) (*models.JobRequestAcceptanceReportConnection, error) {
	var items []*models.JobRequestAcceptanceReport
	var edges []*models.JobRequestAcceptanceReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetJobRequestAcceptanceReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.JobRequestAcceptanceReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.JobRequestAcceptanceReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//JobRequestAcceptanceReport returns a job request acceptance report by its ID
func (r *queryResolver) JobRequestAcceptanceReport(ctx context.Context, id primitive.ObjectID) (*models.JobRequestAcceptanceReport, error) {
	report, err := models.GetJobRequestAcceptanceReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

//JobTimeVariances gives a list of job time variances
func (r *queryResolver) JobTimeVariances(ctx context.Context, fromDate *time.Time, toDate *time.Time, driver *string, text *string, after *string, before *string, first *int, last *int) (*models.JobTimeVarianceConnection, error) {
	var items []*models.JobTimeVariance
	var edges []*models.JobTimeVarianceEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetJobTimeVariances(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.JobTimeVarianceEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.JobTimeVarianceConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//JobTimeVariance returns a job time variance by its ID
func (r *queryResolver) JobTimeVariance(ctx context.Context, id primitive.ObjectID) (*models.JobTimeVariance, error) {
	report, err := models.GetJobTimeVarianceByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

//MarketStatistics gives a list of market statistics
func (r *queryResolver) MarketStatistics(ctx context.Context, city *string, service *string, country string, state *string, currency *string) (*models.MarketStatistics, error) {
	marketStatistics := &models.MarketStatistics{}
	return marketStatistics, nil
}

//AdminDashboard returns the admin dashboard
func (r *queryResolver) AdminDashboard(ctx context.Context, city *string) (*models.AdminDashboard, error) {
	adminDashboard := &models.AdminDashboard{}
	return adminDashboard, nil
}
