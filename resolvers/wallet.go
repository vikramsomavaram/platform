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

//Wallet Payments & Bill Payments
type walletTransactionResolver struct{ *Resolver }

func (r *walletTransactionResolver) Balance(ctx context.Context, obj *models.WalletTransaction) (float64, error) {
	return obj.RemainingBalance, nil
}

func (r *walletTransactionResolver) UpdatedAt(ctx context.Context, obj *models.WalletTransaction) (string, error) {
	return obj.UpdatedAt.String(), nil
}

//CreatedAt
func (r *walletTransactionResolver) CreatedAt(ctx context.Context, obj *models.WalletTransaction) (string, error) {
	return obj.CreatedAt.String(), nil
}

//Withdrawal returns a withdrawal by its ID
func (r *queryResolver) Withdrawal(ctx context.Context, id primitive.ObjectID) (*models.Withdrawal, error) {
	withdrawal, err := models.GetWithdrawalByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return withdrawal, nil
}

//Withdrawals returns a list of withdrawals
func (r *queryResolver) Withdrawals(ctx context.Context, after *string, before *string, first *int, last *int) (*models.WithdrawalConnection, error) {
	var items []*models.Withdrawal
	var edges []*models.WithdrawalEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetWithdrawals(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.WithdrawalEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.WithdrawalConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//UserWalletTransaction returns a user wallet transaction of given ID
func (r *queryResolver) UserWalletTransaction(ctx context.Context, id primitive.ObjectID) (*models.WalletTransaction, error) {
	transaction, err := models.GetWalletTransactionByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return transaction, nil
}

//UserWalletTransactions gives a list of user wallet transactions
func (r *queryResolver) UserWalletTransactions(ctx context.Context, fromDate *time.Time, toDate *time.Time, after *string, before *string, first *int, last *int) (*models.WalletTransactionConnection, error) {
	var items []*models.WalletTransaction
	var edges []*models.WalletTransactionEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetWalletTransactions(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.WalletTransactionEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.WalletTransactionConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//ProviderWalletTransaction returns a provider wallet transaction of given ID
func (r *queryResolver) ProviderWalletTransaction(ctx context.Context, id primitive.ObjectID) (*models.ProviderWalletTransaction, error) {
	transaction, err := models.GetProviderWalletTransactionByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return transaction, nil
}

//ProviderWalletTransactions gives a list of provider wallet transactions
func (r *queryResolver) ProviderWalletTransactions(ctx context.Context, fromDate *time.Time, toDate *time.Time, after *string, before *string, first *int, last *int) (*models.ProviderWalletTransactionConnection, error) {
	var items []*models.ProviderWalletTransaction
	var edges []*models.ProviderWalletTransactionEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetProviderWalletTransactions(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProviderWalletTransactionEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ProviderWalletTransactionConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

// withdrawalResolver is of type struct.
type withdrawalResolver struct{ *Resolver }

//PaymentMethod returns payment method used for withdrawal.
func (r *withdrawalResolver) PaymentMethod(ctx context.Context, obj *models.Withdrawal) (models.PaymentMethodType, error) {
	paymentMethod, err := models.GetPaymentMethodByID(obj.ID.Hex())
	if err != nil {
		return "", err
	}
	return paymentMethod.Type, nil
}
