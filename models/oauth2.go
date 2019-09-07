/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/utils/webhooks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// AuthorizedGrantType represents authorized grant type.
type AuthorizedGrantType string

const (
	// ImplicitGrant represents authorized grant type as implicit.
	ImplicitGrant AuthorizedGrantType = "implicit"
	// RefreshTokenGrant represents authorized grant type as refresh token.
	RefreshTokenGrant AuthorizedGrantType = "refreshToken"
	//PasswordGrant represents authorized grant type as password.
	PasswordGrant AuthorizedGrantType = "password"
	//AuthCodeGrant represents authorized grant type as authorization code.
	AuthCodeGrant AuthorizedGrantType = "authorizationCode"
)

//OAuthApplicationStatistics represents oauth application statistics.
type OAuthApplicationStatistics struct {
	Stats string `json:"stats" bson:"stats"`
}

// OAuthScope ...
type OAuthScope struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	Scope       string             `json:"scope" bson:"scope"`
	Description *string            `json:"description" bson:"description"`
	IsDefault   bool               `json:"isDefault" bson:"isDefault"`
}

// GetOAuthScopesByFilter gives requested oauth application by id.
func GetOAuthScopesByFilter(filter bson.D, limit int, after *string, before *string, first *int, last *int) (oAuthScopes []*OAuthScope, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(OAuthScopesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(OAuthScopesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		oAuthScope := &OAuthScope{}
		err = cur.Decode(&oAuthScope)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		oAuthScopes = append(oAuthScopes, oAuthScope)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return oAuthScopes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// OauthAuthorizationCode ...
type AuthorizationCode struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ClientID    string             `json:"clientId" bson:"clientId"`
	UserID      string             `json:"userId" bson:"userId"`
	Client      *OAuthApplication
	User        *User
	Code        string `json:"code" bson:"code"`
	RedirectURL *string
	ExpiresAt   time.Time
	Scope       string
}

// OAuthApplication represents a oauth application.
type OAuthApplication struct {
	ID                      primitive.ObjectID    `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt               time.Time             `json:"createdAt" bson:"createdAt"`
	DeletedAt               *time.Time            `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt               time.Time             `json:"updatedAt" bson:"updatedAt"`
	CreatedBy               primitive.ObjectID    `json:"createdBy" bson:"createdBy"`
	ClientID                string                `json:"clientId" bson:"clientId"`
	ClientSecret            string                `json:"clientSecret" bson:"clientSecret"`
	ClientDescription       string                `json:"clientDescription" bson:"clientDescription"`
	AccessTokenValiditySecs uint32                `json:"accessTokenValiditySecs" bson:"accessTokenValiditySecs"`
	Authorities             []Authority           `json:"authorities" bson:"authorities"`
	AuthorizedGrantTypes    []AuthorizedGrantType `json:"authorizedGrantTypes" bson:"authorizedGrantTypes"`
	AppName                 string                `json:"appName" bson:"appName"`
	PublisherName           string                `json:"publisherName" bson:"publisherName"`
	RedirectURL             string                `json:"redirectURLs" bson:"redirectURLs"`
	Scopes                  []string              `json:"scopes" bson:"scopes"`
	Developers              []string              `json:"developers" bson:"developers"`
	DevelopmentUsers        []string              `json:"developmentUsers" bson:"developmentUsers"`
	AppIcon                 string                `json:"appIcon" bson:"appIcon"`
	AllowImplicitGrant      AllowImplicitGrant    `json:"allowImplicitGrant" bson:"allowImplicitGrant"`
	WhiteListedDomains      []string              `json:"whiteListedDomains" bson:"whiteListedDomains"`
	TermsOfServiceURL       string                `json:"termsOfServiceURL" bson:"termsOfServiceURL"`
	PrivacyURL              string                `json:"privacyURL" bson:"privacyURL"`
	Website                 string                `json:"website" bson:"website"`
	ContactEmail            string                `json:"contactEmail" bson:"contactEmail"`
	IsActive                bool                  `json:"isActive" bson:"isActive"`
}

//UnmarshalBinary required for the redis cache to work
func (app *OAuthApplication) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, app); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (app *OAuthApplication) MarshalBinary() ([]byte, error) {
	return json.Marshal(app)
}

// CreateOAuthApplication creates new oauth application.
func CreateOAuthApplication(oAuthApplication OAuthApplication) (*OAuthApplication, error) {
	oAuthApplication.CreatedAt = time.Now()
	oAuthApplication.UpdatedAt = time.Now()
	oAuthApplication.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(OAuthApplicationsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &oAuthApplication)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("oauth_application.created", &oAuthApplication)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(oAuthApplication.ID.Hex(), oAuthApplication, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &oAuthApplication, nil
}

// GetOAuthApplicationByID gives requested oauth application by id.
func GetOAuthApplicationByID(ID string) (*OAuthApplication, error) {
	db := database.MongoDB
	oAuthApplication := &OAuthApplication{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(oAuthApplication)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(OAuthApplicationsCollection).FindOne(ctx, filter).Decode(&oAuthApplication)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, oAuthApplication, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return oAuthApplication, nil
}

// GetOAuthApplications gives a list of oauth applications.
func GetOAuthApplications(filter bson.D, limit int, after *string, before *string, first *int, last *int) (oAuthApps []*OAuthApplication, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(OAuthApplicationsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(OAuthApplicationsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		oAuthApp := &OAuthApplication{}
		err = cur.Decode(&oAuthApp)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		oAuthApps = append(oAuthApps, oAuthApp)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return oAuthApps, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateOAuthApplication updates oauth applications.
func UpdateOAuthApplication(c *OAuthApplication) (*OAuthApplication, error) {
	oAuthApplication := c
	oAuthApplication.UpdatedAt = time.Now()
	filter := bson.D{{"_id", oAuthApplication.ID}}
	db := database.MongoDB
	oAuthApplicationsCollection := db.Collection(OAuthApplicationsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := oAuthApplicationsCollection.FindOneAndReplace(context.Background(), filter, oAuthApplication, findRepOpts).Decode(&oAuthApplication)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("oauth_application.updated", &oAuthApplication)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(oAuthApplication.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return oAuthApplication, nil
}

// DeleteOAuthApplicationByID deletes oauth application.
func DeleteOAuthApplicationByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, _ := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	oAuthApplicationsCollection := db.Collection(OAuthApplicationsCollection)
	res, err := oAuthApplicationsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("oauth_application.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// Authority represents authority.
type Authority struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	Name      string
}

// RefreshToken represents refresh token.
type RefreshToken struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt      *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updatedAt"`
	ExpiresAt      time.Time          `json:"expiresAt" bson:"expiresAt"`
	ClientID       string             `json:"clientId" bson:"clientId"`
	UserID         string             `json:"userId" bson:"userId"`
	Token          string             `json:"token"`
	Scope          string             `json:"scope"`
	Authentication string             `json:"authentication"`
}

func (r *RefreshToken) Client() *OAuthApplication {
	return GetOAuthApplicationByFilter(bson.D{{"clientId", r.ClientID}})
}

func (r *RefreshToken) User() *User {
	return GetUserByID(r.UserID)
}

// CreateRefreshToken creates new refresh tokens.
func CreateRefreshToken(refreshToken *RefreshToken) (*RefreshToken, error) {
	refreshToken.CreatedAt = time.Now()
	refreshToken.UpdatedAt = time.Now()
	db := database.MongoDB
	collection := db.Collection(OAuthRefreshTokensCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &refreshToken)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("refresh_token.created", &refreshToken)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(refreshToken.ID.Hex(), refreshToken, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return refreshToken, nil
}

// GetRefreshTokenByFilter gives the requested refresh token.
func GetRefreshTokenByFilter(filter bson.D) *RefreshToken {
	db := database.MongoDB
	refreshToken := &RefreshToken{}
	err, filterHash := genBsonHash(filter)
	if err != nil {
		log.Error(err)
	}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err = cacheClient.Get(filterHash).Scan(refreshToken)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	ctx := context.Background()
	err = db.Collection(OAuthRefreshTokensCollection).FindOne(ctx, filter).Decode(&refreshToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return refreshToken
		}
		log.Errorln(err)
		return refreshToken
	}
	//set cache item
	err = cacheClient.Set(filterHash, refreshToken, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return refreshToken
}

// GetRefreshTokenByID gives the requested refresh token.
func GetRefreshTokenByID(ID string) (*RefreshToken, error) {
	db := database.MongoDB
	refreshToken := &RefreshToken{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(refreshToken)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(OAuthRefreshTokensCollection).FindOne(ctx, filter).Decode(&refreshToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, refreshToken, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return refreshToken, nil
}

// GetRefreshTokens gives a list of refresh tokens.
func GetRefreshTokens(filter bson.D, limit int, after *string, before *string, first *int, last *int) (refreshTokens []*RefreshToken, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(TokensCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(TokensCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		refreshToken := &RefreshToken{}
		err = cur.Decode(&refreshToken)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		refreshTokens = append(refreshTokens, refreshToken)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return refreshTokens, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateRefreshToken updates refresh tokens.
func UpdateRefreshToken(c *RefreshToken) (*RefreshToken, error) {
	refreshToken := c
	refreshToken.UpdatedAt = time.Now()
	filter := bson.D{{"_id", refreshToken.ID}}
	db := database.MongoDB
	oAuthRefreshTokensCollection := db.Collection(OAuthRefreshTokensCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := oAuthRefreshTokensCollection.FindOneAndReplace(context.Background(), filter, refreshToken, findRepOpts).Decode(&refreshToken)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("refresh_token.updated", &RefreshToken{})
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(refreshToken.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return refreshToken, nil
}

// DeleteAccessTokensByFilter deletes refresh tokens by filtering.
func DeleteAccessTokensByFilter(filter bson.D) (bool, error) {
	db := database.MongoDB
	oAuthRefreshTokensCollection := db.Collection(OAuthRefreshTokensCollection)
	res, err := oAuthRefreshTokensCollection.UpdateMany(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("refresh_tokens.deleted", &res)
	//Delete cache item
	//TODO handle the delete
	cacheClient := cache.RedisClient
	err = cacheClient.Del("").Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// DeleteRefreshTokenByFilter deletes refresh tokens by filtering.
func DeleteRefreshTokensByFilter(filter bson.D) (bool, error) {
	db := database.MongoDB
	oAuthRefreshTokensCollection := db.Collection(OAuthRefreshTokensCollection)
	res, err := oAuthRefreshTokensCollection.UpdateMany(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("refresh_tokens.deleted", &res)
	//Delete cache item
	//TODO handle the delete
	cacheClient := cache.RedisClient
	err = cacheClient.Del("").Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// DeleteRefreshTokenByID deletes refresh tokens by id.
func DeleteAuthorizationCodeByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(OAuthAuthorizationCodeCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("authorization_code.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// DeleteRefreshTokenByID deletes refresh tokens by id.
func DeleteRefreshTokenByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	oAuthRefreshTokensCollection := db.Collection(OAuthRefreshTokensCollection)
	res, err := oAuthRefreshTokensCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("refresh_token.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (refreshToken *RefreshToken) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, refreshToken); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (refreshToken *RefreshToken) MarshalBinary() ([]byte, error) {
	return json.Marshal(refreshToken)
}

// AccessToken represents access token.
type AccessToken struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	ExpiresAt        time.Time          `json:"expiresAt" bson:"expiresAt"`
	AccessToken      string             `json:"accessToken" bson:"accessToken"`
	TokenType        string             `json:"tokenType" bson:"tokenType"`
	RefreshToken     *RefreshToken      `json:"refreshToken" bson:"refreshToken,omitempty"`
	ExpiresIn        int                `json:"expiresIn" bson:"expiresIn"`
	AuthenticationID string             `json:"authenticationId" bson:"authenticationId"`
	Authentication   string             `json:"authentication" bson:"authentication"`
	ClientID         string             `json:"clientId" bson:"clientId"`
	UserID           string             `json:"userId" bson:"userId"`
	Scope            string             `json:"scope" bson:"scope"`
	Metadata         interface{}        `json:"metadata" bson:"metadata"` //User Device, Client, Grant Type, IP Address etc.,
}

//UnmarshalBinary required for the redis cache to work
func (accessToken *AccessToken) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, accessToken); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (accessToken *AccessToken) MarshalBinary() ([]byte, error) {
	return json.Marshal(accessToken)
}

// GetOAuthApplicationByFilter gives the oauth application by filter.
func GetOAuthApplicationByFilter(filter bson.D) *OAuthApplication {
	db := database.MongoDB
	oauthApp := &OAuthApplication{}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}}) // we dont want deleted documents
	err := db.Collection(OAuthApplicationsCollection).FindOne(context.TODO(), filter).Decode(&oauthApp)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
			return oauthApp
		}
	}
	return oauthApp
}

// CreateAccessToken creates new access tokens.
func CreateAccessToken(accessToken *AccessToken) (*AccessToken, error) {
	accessToken.CreatedAt = time.Now()
	accessToken.UpdatedAt = time.Now()
	accessToken.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(OAuthAccessTokensCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &accessToken)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("access_token.created", &accessToken)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(accessToken.ID.Hex(), accessToken, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return accessToken, nil
}

// CreateAuthorizationCode creates new access tokens.
func CreateAuthorizationCode(authCode *AuthorizationCode) (*AuthorizationCode, error) {
	authCode.CreatedAt = time.Now()
	authCode.UpdatedAt = time.Now()
	db := database.MongoDB
	collection := db.Collection(OAuthAuthorizationCodeCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &authCode)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("authorization_code.created", &authCode)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(authCode.ID.Hex(), authCode, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return authCode, nil
}

// GetAccessTokenByID gives requested access tokens by id.
func GetAccessTokenByID(ID string) (*AccessToken, error) {
	db := database.MongoDB
	accessToken := &AccessToken{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(accessToken)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(OAuthAccessTokensCollection).FindOne(ctx, filter).Decode(&accessToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, accessToken, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return accessToken, nil
}

// GetAccessTokenByFilter gives requested access tokens by id.
func GetAuthorizationCodeByFilter(filter bson.D) *AuthorizationCode {
	db := database.MongoDB
	authcode := &AuthorizationCode{}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	ctx := context.Background()
	err := db.Collection(OAuthAuthorizationCodeCollection).FindOne(ctx, filter).Decode(&authcode)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
			return authcode
		}
	}
	return authcode
}

// GetAccessTokenByFilter gives requested access tokens by id.
func GetAccessTokenByFilter(filter bson.D) (*AccessToken, error) {
	db := database.MongoDB
	accessToken := &AccessToken{}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	ctx := context.Background()
	err := db.Collection(OAuthAccessTokensCollection).FindOne(ctx, filter).Decode(&accessToken)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
			return nil, nil
		}
		return nil, err
	}
	return accessToken, nil
}

// GetAccessTokens gives a list of access tokens.
func GetAccessTokens(filter bson.D, limit int, after *string, before *string, first *int, last *int) (accessTokens []*AccessToken, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(TokensCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(TokensCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		accessToken := &AccessToken{}
		err = cur.Decode(&accessToken)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		accessTokens = append(accessTokens, accessToken)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return accessTokens, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateAccessToken updates the access tokens.
func UpdateAccessToken(accessToken *AccessToken) (*AccessToken, error) {
	accessToken.UpdatedAt = time.Now()
	filter := bson.D{{"_id", accessToken.ID}}
	db := database.MongoDB
	oAuthAccessTokensCollection := db.Collection(OAuthAccessTokensCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := oAuthAccessTokensCollection.FindOneAndReplace(context.Background(), filter, accessToken, findRepOpts).Decode(&accessToken)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("access_token.updated", &accessToken)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(accessToken.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return accessToken, nil
}

// DeleteAccessTokenByID deletes the access tokens by id.
func DeleteAccessToken(token string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"accessToken", token}}
	oAuthAccessTokensCollection := db.Collection(OAuthAccessTokensCollection)
	res, err := oAuthAccessTokensCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("access_token.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(token).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// DeleteAccessTokenByID deletes the access tokens by id.
func DeleteAccessTokenByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	oAuthAccessTokensCollection := db.Collection(OAuthAccessTokensCollection)
	res, err := oAuthAccessTokensCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("access_token.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// DeleteAccessTokenByFilter deletes the access tokens by filter.
func DeleteAccessTokenByFilter(filter bson.D) (bool, error) {
	db := database.MongoDB
	oAuthAccessTokensCollection := db.Collection(OAuthAccessTokensCollection)
	res, err := oAuthAccessTokensCollection.UpdateMany(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	return true, nil
}

// NewOAuthRefreshToken creates new OauthRefreshToken instance
func NewOAuthRefreshToken(client *OAuthApplication, user *User, expiresIn int, scope string) *RefreshToken {
	refreshToken := &RefreshToken{
		CreatedAt: time.Now().UTC(),
		ClientID:  client.ID.Hex(),
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		Scope:     scope,
	}
	if user != nil {
		refreshToken.UserID = user.ID.Hex()
	}
	return refreshToken
}

// NewOauthAccessToken creates new OauthAccessToken instance
func NewOAuthAccessToken(client *OAuthApplication, user *User, expiresIn int, scope string) *AccessToken {
	accessToken := &AccessToken{
		ID:          primitive.NewObjectID(),
		CreatedAt:   time.Now().UTC(),
		ClientID:    client.ID.Hex(),
		AccessToken: uuid.New().String(),
		ExpiresAt:   time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		Scope:       scope,
	}
	if user != nil {
		accessToken.UserID = user.ID.Hex()
	}
	return accessToken
}

// NewOauthAuthorizationCode creates new OauthAuthorizationCode instance
func NewOAuthAuthorizationCode(client *OAuthApplication, user *User, expiresIn int, redirectURI, scope string) *AuthorizationCode {
	return &AuthorizationCode{
		ID:          primitive.NewObjectID(),
		CreatedAt:   time.Now().UTC(),
		ClientID:    client.ID.Hex(),
		UserID:      user.ID.Hex(),
		Code:        uuid.New().String(),
		ExpiresAt:   time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		RedirectURL: &redirectURI,
		Scope:       scope,
	}
}
