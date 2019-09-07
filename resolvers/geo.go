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
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type countryResolver struct{ *Resolver }

type cityResolver struct{ *Resolver }

func (r *countryResolver) DistanceUnit(ctx context.Context, obj *models.Country) (models.DistanceUnits, error) {
	return models.DistanceUnits(obj.DistanceUnit), nil
}

func (r *countryResolver) Tax(ctx context.Context, obj *models.Country) (string, error) {
	//tax, _ := json.Marshal(obj.Tax)
	//return string(tax), nil
	//TODO
	return "", nil
}

//AddCountry adds a new country
func (r *mutationResolver) AddCountry(ctx context.Context, input models.AddCountryInput) (*models.Country, error) {
	country := &models.Country{}
	_ = copier.Copy(&country, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	country.CreatedBy = user.ID
	country, err = models.CreateCountry(*country)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), country.ID.Hex(), "country", country, nil, ctx)
	return country, nil
}

//UpdateCountry updates an existing country
func (r *mutationResolver) UpdateCountry(ctx context.Context, input models.UpdateCountryInput) (*models.Country, error) {
	country := &models.Country{}
	country, err := models.GetCountryByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&country, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	country.CreatedBy = user.ID
	country, err = models.UpdateCountry(country)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), country.ID.Hex(), "country", country, nil, ctx)
	return country, nil
}

//DeleteCountry deletes an existing country
func (r *mutationResolver) DeleteCountry(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteCountry(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "country", nil, nil, ctx)
	return &res, err
}

//ActivateCountry activates a country by its id
func (r *mutationResolver) ActivateCountry(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	country, err := models.GetCountryByID(id.Hex())
	if err != nil {
		return nil, err
	}
	country.IsActive = true
	_, err = models.UpdateCountry(country)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "country", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateCountry deactivates a country by its id
func (r *mutationResolver) DeactivateCountry(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	country, err := models.GetCountryByID(id.Hex())
	if err != nil {
		return nil, err
	}
	country.IsActive = false
	_, err = models.UpdateCountry(country)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "country", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//AddState adds a new state
func (r *mutationResolver) AddState(ctx context.Context, input models.AddStateInput) (*models.State, error) {
	state := &models.State{}
	_ = copier.Copy(&state, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	state.CreatedBy = user.ID
	state, err = models.CreateState(*state)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), state.ID.Hex(), "state", state, nil, ctx)
	return state, nil
}

//UpdateState updates an existing state
func (r *mutationResolver) UpdateState(ctx context.Context, input models.UpdateStateInput) (*models.State, error) {
	state := &models.State{}
	state, err := models.GetStateByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&state, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	state.CreatedBy = user.ID
	state, err = models.UpdateState(state)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), state.ID.Hex(), "state", state, nil, ctx)
	return state, nil
}

//DeleteState deletes an existing state
func (r *mutationResolver) DeleteState(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteState(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "state", nil, nil, ctx)
	return &res, err
}

//ActivateState activates a state by its ID
func (r *mutationResolver) ActivateState(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	state, err := models.GetStateByID(id.Hex())
	if err != nil {
		return nil, err
	}
	state.IsActive = true
	_, err = models.UpdateState(state)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "state", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateState deactivates a state by its ID
func (r *mutationResolver) DeactivateState(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	state, err := models.GetStateByID(id.Hex())
	if err != nil {
		return nil, err
	}
	state.IsActive = false
	_, err = models.UpdateState(state)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "state", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//AddCity adds a new city
func (r *mutationResolver) AddCity(ctx context.Context, input models.AddCityInput) (*models.City, error) {
	city := &models.City{}
	_ = copier.Copy(&city, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	city, err = models.CreateCity(*city)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), city.ID.Hex(), "city", city, nil, ctx)
	return city, nil
}

//UpdateCity updates an existing city
func (r *mutationResolver) UpdateCity(ctx context.Context, input models.UpdateCityInput) (*models.City, error) {
	city := &models.City{}
	city, err := models.GetCityByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&city, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	city.CreatedBy = user.ID
	city, err = models.UpdateCity(city)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), city.ID.Hex(), "city", city, nil, ctx)
	return city, nil
}

//DeleteCity deletes an existing city
func (r *mutationResolver) DeleteCity(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteCity(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "city", nil, nil, ctx)
	return &res, err
}

//ActivateCity activates a city by its ID
func (r *mutationResolver) ActivateCity(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	city, err := models.GetCityByID(id.Hex())
	if err != nil {
		return nil, err
	}
	city.IsActive = true
	_, err = models.UpdateCity(city)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "city", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateCity deactivates a city by its ID
func (r *mutationResolver) DeactivateCity(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	city, err := models.GetCityByID(id.Hex())
	if err != nil {
		return nil, err
	}
	city.IsActive = false
	_, err = models.UpdateCity(city)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "city", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//CountryStates gives a list of states
func (r *queryResolver) CountryStates(ctx context.Context, country *string, stateType *models.StateType, text *string, after *string, before *string, first *int, last *int) (*models.StateConnection, error) {
	var items []*models.State
	var edges []*models.StateEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetStates(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.StateEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.StateConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//CountryState returns a state by its ID
func (r *queryResolver) CountryState(ctx context.Context, countryCode string, stateCode string) (*models.State, error) {
	state, err := models.GetState(countryCode, stateCode)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return state, nil
}

//Cities gives a list of cities
func (r *queryResolver) Cities(ctx context.Context, cityType *models.CityType, text *string, after *string, before *string, first *int, last *int) (*models.CityConnection, error) {
	var items []*models.City
	var edges []*models.CityEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetCities(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CityEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.CityConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//City returns a city by its ID
func (r *queryResolver) City(ctx context.Context, id primitive.ObjectID) (*models.City, error) {
	city, err := models.GetCityByCode(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return city, nil
}

//Country returns a country by its ID
func (r *queryResolver) Country(ctx context.Context, countryCode string) (*models.Country, error) {
	country, err := models.GetCountryByCode(countryCode)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return country, nil
}

//Countries gives a list of countries
func (r *queryResolver) Countries(ctx context.Context, countryType *models.CountryType, countryStatus *models.CountryStatus, text *string, after *string, before *string, first *int, last *int) (*models.CountryConnection, error) {
	var items []*models.Country
	var edges []*models.CountryEdge
	filter := bson.D{}
	//filter := bson.D{{"phoneCode", bson.M{"$ne": ""}}}
	limit := 300

	items, totalCount, hasPrevious, hasNext, err := models.GetCountries(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CountryEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.CountryConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

func paginationUtility(after *string, before *string, first *int, last *int) (filter bson.D, limit int, err error) {

	if after != nil {
		afterID, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return filter, 0, err
		}
		afterCursor := bson.E{"_id", bson.M{"$gt": afterID}}
		filter = append(filter, afterCursor)
	}

	if before != nil {
		beforeID, err := base64.StdEncoding.DecodeString(*before)
		if err != nil {
			return filter, 0, err
		}
		afterCursor := bson.E{"_id", bson.M{"$lt": beforeID}}
		filter = append(filter, afterCursor)
	}

	if first != nil {
		limit = *first
	}

	if last != nil {
		limit = *last
	}

	return filter, limit, nil
}

// stateResolver is of type struct.
type stateResolver struct{ *Resolver }

//Code returns code.
func (r *stateResolver) Code(ctx context.Context, obj *models.State) (*string, error) {
	return &obj.StateCode, nil
}
