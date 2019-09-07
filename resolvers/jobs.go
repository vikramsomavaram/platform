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
	"time"
)

//TODO
//RecentUnpaidJobs gives a list of recent unpaid jobs
func (r *queryResolver) RecentUnpaidJobs(ctx context.Context, after *string, before *string, first *int, last *int) (*models.RecentUnpaidEarningConnection, error) {
	panic("implement me")
}

//TODO
//PaidJobs gives a list of paid jobs
func (r *queryResolver) PaidJobs(ctx context.Context, after *string, before *string, first *int, last *int) (*models.PaidEarningConnection, error) {
	panic("implement me")
}

//CancelledJobs gives a list of cancelled jobs
func (r *queryResolver) CancelledJobs(ctx context.Context, fromDate *time.Time, toDate *time.Time, provider *string, cancelledJobserviceType *models.CancelledJobServiceType, text *string, after *string, before *string, first *int, last *int) (*models.JobConnection, error) {
	var items []*models.Job
	var edges []*models.JobEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetJobs(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.JobEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.JobConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//Job returns a job by its ID
func (r *queryResolver) Job(ctx context.Context, jobID string) (*models.Job, error) {
	job, err := models.GetJobByID(jobID)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return job, nil
}

type jobResolver struct{ *Resolver }

func (r *jobResolver) JobType(ctx context.Context, obj *models.Job) (string, error) {
	panic("implement me")
}

func (r *jobResolver) CompanyID(ctx context.Context, obj *models.Job) (*string, error) {
	return &obj.CompanyID, nil
}

//Jobs gives a list of jobs
func (r *queryResolver) Jobs(ctx context.Context, fromDate *time.Time, toDate *time.Time, jobStatus *models.JobStatus, jobNumber *string, company *string, provider *string, user *string, serviceType *models.JobServiceType, laterJobsOnly *bool, after *string, before *string, first *int, last *int) (*models.JobConnection, error) {
	var items []*models.Job
	var edges []*models.JobEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetJobs(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.JobEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.JobConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}
