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
	"strconv"
)

type installationResolver struct{ *Resolver }

func (r *installationResolver) DeviceWidth(ctx context.Context, obj *models.Installation) (string, error) {
	return strconv.FormatFloat(obj.DeviceWidth, 'f', 6, 32), nil
}

func (r *installationResolver) DeviceType(ctx context.Context, obj *models.Installation) (models.DeviceType, error) {
	deviceType := models.DeviceType(obj.DeviceType)
	return deviceType, nil
}

func (r *installationResolver) DeviceHeight(ctx context.Context, obj *models.Installation) (string, error) {
	return strconv.FormatFloat(obj.DeviceHeight, 'f', 6, 32), nil
}

//Installations gives a list of installations
func (r *queryResolver) Installations(ctx context.Context, after *string, before *string, first *int, last *int) (*models.InstallationConnection, error) {
	var items []*models.Installation
	var edges []*models.InstallationEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetInstallations(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.InstallationEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.InstallationConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//UpdateInstallation updates an existing installation
func (r *mutationResolver) UpdateInstallation(ctx context.Context, input models.InstallationInput) (*models.Installation, error) {
	installation := &models.Installation{}
	_ = copier.Copy(&installation, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	installation.CreatedBy = user.ID
	installation, err = models.CreateInstallation(installation)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), installation.ID.Hex(), "installation", installation, nil, ctx)
	return installation, nil
}

//Installation returns an installation by its ID
func (r *queryResolver) Installation(ctx context.Context, id primitive.ObjectID) (*models.Installation, error) {
	installation, err := models.GetInstallationByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return installation, nil
}
