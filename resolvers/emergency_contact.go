/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//AddEmergencyContact adds emergency contacts.
func (r *mutationResolver) AddEmergencyContact(ctx context.Context, input models.AddEmergencyContactInput) (*models.EmergencyContact, error) {
	emergencyContact := &models.EmergencyContact{}
	_ = copier.Copy(&emergencyContact, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	emergencyContact.CreatedBy = user.ID
	emergencyContact, err = models.CreateEmergencyContact(*emergencyContact)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), emergencyContact.ID.Hex(), "emergency contact", emergencyContact, nil, ctx)
	return emergencyContact, nil
}

//UpdateEmergencyContact updates emergency contacts.
func (r *mutationResolver) UpdateEmergencyContact(ctx context.Context, input models.UpdateEmergencyContactInput) (*models.EmergencyContact, error) {
	emergencyContact := &models.EmergencyContact{}
	emergencyContact, err := models.GetEmergencyContactByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&emergencyContact, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	emergencyContact.CreatedBy = user.ID
	emergencyContact, err = models.UpdateEmergencyContact(emergencyContact)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), emergencyContact.ID.Hex(), "emergency contact", emergencyContact, nil, ctx)
	return emergencyContact, nil
}

//DeleteEmergencyContact deletes emergency contacts.
func (r *mutationResolver) DeleteEmergencyContact(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteEmergencyContactByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "emergency contact", nil, nil, ctx)
	return &res, err
}

//EmergencyContacts returns list of emergency contacts.
func (r *queryResolver) EmergencyContacts(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.EmergencyContactConnection, error) {
	var items []*models.EmergencyContact
	var edges []*models.EmergencyContactEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetEmergencyContacts(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.EmergencyContactEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.EmergencyContactConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//EmergencyContact returns emergency contact by id.
func (r *queryResolver) EmergencyContact(ctx context.Context, id primitive.ObjectID) (*models.EmergencyContact, error) {
	emergencyContact, err := models.GetEmergencyContactByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return emergencyContact, nil
}
