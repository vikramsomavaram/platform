/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
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

// User represents a user.
type User struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	FirstName        string             `json:"firstName,required" bson:"firstName"`
	LastName         string             `json:"lastName,required" bson:"lastName"`
	Email            string             `json:"email" bson:"email"`
	MobileNo         string             `json:"mobileNo,required" bson:"mobileNo"`
	Roles            []string           `json:"roles" bson:"roles"`
	RoleGroups       []string           `json:"roleGroups" bson:"roleGroups"`
	Password         string             `json:"password,required" bson:"password"`
	Country          string             `json:"country,required" bson:"country,required"`
	State            string             `json:"state,required" bson:"state,required"`
	City             string             `json:"city,required" bson:"city,required"`
	OTP              string             `json:"otp" bson:"otp"`
	IsActive         bool               `json:"isActive" bson:"isActive"`
	IsMobileVerified bool               `json:"isMobileVerified" bson:"isMobileVerified"`
	IsEmailVerified  bool               `json:"isEmailVerified" bson:"isEmailVerified"`
	IsLocked         bool               `json:"isLocked" bson:"isLocked"`
	Gender           Gender             `json:"gender,required" bson:"gender,required"`
	DateOfBirth      time.Time          `json:"dateOfBirth" bson:"dateOfBirth"`
	ReferralCode     string             `json:"referralCode" bson:"referralCode"`
	Language         string             `json:"language" bson:"language"`
}

//UnmarshalBinary required for the redis cache to work
func (user *User) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, user); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (user *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(user)
}

// CreateUser creates new user.
func CreateUser(user *User) (*User, error) {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	db := database.MongoDB
	Collection := db.Collection(UsersCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := Collection.InsertOne(ctx, &user)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(user.ID.Hex(), user, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user.created", &user)
	return user, nil
}

// GetUsers gives  a list of users.
func GetUsers(filter bson.D, limit int, after *string, before *string, first *int, last *int) (users []*User, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(UsersCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(UsersCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		user := &User{}
		err = cur.Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		users = append(users, user)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return users, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// GetUserByID gives user by id.
func GetUserByID(ID string) *User {
	user := &User{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(user)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	if !user.ID.IsZero() {
		return user
	}

	db := database.MongoDB
	oid, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	filter := bson.D{{"_id", oid}, {"deletedAt", bson.M{"$exists": false}}}
	err = db.Collection(UsersCollection).FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
			return nil
		}
		return nil
	}

	//set cache item
	err = cacheClient.Set(ID, user, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}

	return user
}

//GetUserByFilter returns user by given filter
func GetUserByFilter(filter bson.D) *User {
	db := database.MongoDB
	dbUser := new(User)
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	err := db.Collection(UsersCollection).FindOne(context.Background(), filter).Decode(dbUser)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
		}
		return nil
	}
	return dbUser
}

// UpdateUser updates user.
func UpdateUser(user *User) (*User, error) {
	user.UpdatedAt = time.Now()
	filter := bson.D{{"_id", user.ID}}
	db := database.MongoDB
	usersCollection := db.Collection(UsersCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := usersCollection.FindOneAndReplace(context.Background(), filter, user, findRepOpts).Decode(&user)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(user.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user.updated", &user)
	return user, nil
}

// DeleteUserByID deletes user by id.
func DeleteUserByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	usersCollection := db.Collection(UsersCollection)
	res, err := usersCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user.deleted", &res)
	return true, nil
}

//GetRoleGroupRoles returns role group roles.
func GetRoleGroupRoles(roleGroupName string) *[]*UserRole {
	roleGroup := GetUserRoleGroupByFilter(bson.D{{"name", roleGroupName}})
	var roles []*UserRole
	if len(roleGroup.Roles) > 0 {
		for _, role := range roleGroup.Roles {
			dbRole := GetRoleByName(role)
			if dbRole != nil {
				roles = append(roles, dbRole)
			}
		}
	}
	return &roles
}

//GetRoleByName returns role by name.
func GetRoleByName(roleName string) *UserRole {
	db := database.MongoDB
	dbRole := new(UserRole)
	filter := bson.D{{"name", roleName}, {"deletedAt", bson.M{"$exists": false}}}
	err := db.Collection(UserRolesCollection).FindOne(context.Background(), filter).Decode(dbRole)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
		}
		return nil
	}
	return dbRole
}

//GetUserRoleGroupsByUserID
func GetUserRoleGroupsByUserID(userId string) *[]*UserRoleGroup {
	user := GetUserByID(userId)
	userRoleGroups := &[]*UserRoleGroup{}
	for _, roleGroupID := range user.RoleGroups {
		userRoleGroupRoles := GetUserRoleGroupByFilter(bson.D{{"_id", roleGroupID}})
		*userRoleGroups = append(*userRoleGroups, userRoleGroupRoles)
	}
	return userRoleGroups
}

func GetMergedUserRoles(userId string) *[]*UserRole {
	user := GetUserByID(userId)
	userRoles := GetUserRoles(userId)
	for _, roleGroupName := range user.RoleGroups {
		userRoleGroupRoles := GetRoleGroupRoles(roleGroupName)
		*userRoles = append(*userRoles, *userRoleGroupRoles...)
	}
	return userRoles
}

func GetMergedUserPermissions(userId string) *[]string {
	userRoles := GetMergedUserRoles(userId)

	encountered := map[string]bool{}
	result := []string{}

	for _, userRole := range *userRoles {
		for _, permission := range userRole.Permissions {
			if encountered[permission] == true {
				// Do not add duplicate.
			} else {
				// Record this element as an encountered element.
				encountered[permission] = true
				// Append to result slice.
				result = append(result, permission)
			}
		}
	}
	return &result
}

// UserRole represents a user role.
type UserRole struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Permissions []string           `json:"permissions" bson:"permissions"`
}

// CreateUserRole creates new user role.
func CreateUserRole(userRole UserRole) (*UserRole, error) {
	userRole.CreatedAt = time.Now()
	userRole.UpdatedAt = time.Now()
	userRole.ID = primitive.NewObjectID()
	db := database.MongoDB
	Collection := db.Collection(UserRoleCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := Collection.InsertOne(ctx, &userRole)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(userRole.ID.Hex(), userRole, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role.created", &userRole)
	return &userRole, nil
}

// GetUserRoleByID gives the requested user role by id.
func GetUserRoleByID(ID string) (*UserRole, error) {
	db := database.MongoDB
	userRole := &UserRole{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(userRole)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(UserRoleCollection).FindOne(ctx, filter).Decode(&userRole)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return userRole, nil
		}
		log.Errorln(err)
		return userRole, err
	}
	//set cache item
	err = cacheClient.Set(ID, userRole, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return userRole, nil
}

// GetAllUserRoles gives a list of user roles.
func GetAllUserRoles(filter bson.D, limit int, after *string, before *string, first *int, last *int) (allUserRoles []*UserRole, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(UserRoleCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(UserRoleCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		userRole := &UserRole{}
		err = cur.Decode(&userRole)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		allUserRoles = append(allUserRoles, userRole)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return allUserRoles, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//GetUserRoles returns user roles.
func GetUserRoles(userID string) (userRoles *[]*UserRole) {
	user := GetUserByID(userID)
	var roles []*UserRole
	for _, role := range user.Roles {
		dbRole := GetRoleByName(role)
		if dbRole != nil {
			roles = append(roles, dbRole)
		}
	}
	return &roles
}

// UpdateUserRole updates the user role.
func UpdateUserRole(userRole *UserRole) (*UserRole, error) {
	userRole.UpdatedAt = time.Now()
	filter := bson.D{{"_id", userRole.ID}}
	db := database.MongoDB
	collection := db.Collection(UserRoleCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, userRole, findRepOpts).Decode(&userRole)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role.updated", &userRole)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(userRole.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return userRole, nil
}

// DeleteUserRoleByID deletes user role by id.
func DeleteUserRoleByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(UserRoleCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role.deleted", &res)
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (userRole *UserRole) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, userRole); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (userRole *UserRole) MarshalBinary() ([]byte, error) {
	return json.Marshal(userRole)
}

//UserRoleGroup represents user role group.
type UserRoleGroup struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name        string             `json:"name" bson:"name"`
	Roles       []string           `json:"roles" bson:"roles"`
	Description string             `json:"description" bson:"description"`
}

// CreateUserRoleGroup creates new user role group.
func CreateUserRoleGroup(userRoleGroup *UserRoleGroup) (*UserRoleGroup, error) {
	userRoleGroup.CreatedAt = time.Now()
	userRoleGroup.UpdatedAt = time.Now()
	userRoleGroup.ID = primitive.NewObjectID()
	db := database.MongoDB
	Collection := db.Collection(UserRoleGroupCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := Collection.InsertOne(ctx, &userRoleGroup)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(userRoleGroup.ID.Hex(), userRoleGroup, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role_group.created", &userRoleGroup)
	return userRoleGroup, nil
}

//GetUserRoleGroupByFilter returns user role group by id.
func GetUserRoleGroupByFilter(filter bson.D) *UserRoleGroup {
	db := database.MongoDB
	dbRoleGroup := new(UserRoleGroup)
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	err := db.Collection(UserRolesCollection).FindOne(context.Background(), filter).Decode(dbRoleGroup)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
		}
		return nil
	}
	return dbRoleGroup
}

// UpdateUserRoleGroup updates the user role group.
func UpdateUserRoleGroup(userRoleGroup *UserRoleGroup) (*UserRoleGroup, error) {
	userRoleGroup.UpdatedAt = time.Now()
	filter := bson.D{{"_id", userRoleGroup.ID}}
	db := database.MongoDB
	collection := db.Collection(UserRoleGroupCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, userRoleGroup, findRepOpts).Decode(&userRoleGroup)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(userRoleGroup.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role_group.updated", &userRoleGroup)
	return userRoleGroup, nil
}

// DeleteUserRoleGroupByID deletes user role group by id.
func DeleteUserRoleGroupByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(UserRoleGroupCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role_group.deleted", &res)
	return true, nil
}

func GetUserRoleGroups(filter bson.D, limit int, after *string, before *string, first *int, last *int) (userRoleGroups []*UserRoleGroup, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(UserRoleGroupCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(UserRoleGroupCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		userRoleGroup := &UserRoleGroup{}
		err = cur.Decode(&userRoleGroup)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		userRoleGroups = append(userRoleGroups, userRoleGroup)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return userRoleGroups, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (userRoleGroup *UserRoleGroup) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, userRoleGroup); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (userRoleGroup *UserRoleGroup) MarshalBinary() ([]byte, error) {
	return json.Marshal(userRoleGroup)
}

// UserRolePermissions represents a user role permissions.
type UserRolePermissions struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string             `json:"createdBy" bson:"createdBy"`
	Service   string             `json:"service" bson:"service"`
	Actions   []string           `json:"actions" bson:"actions"`
}

// CreateUserRolePermission creates new user role permission.
func CreateUserRolePermission(userRolePermission UserRolePermissions) (*UserRolePermissions, error) {
	userRolePermission.CreatedAt = time.Now()
	userRolePermission.UpdatedAt = time.Now()
	userRolePermission.ID = primitive.NewObjectID()
	db := database.MongoDB
	Collection := db.Collection(UserRolePermissionsCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := Collection.InsertOne(ctx, &userRolePermission)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(userRolePermission.ID.Hex(), userRolePermission, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role_permission.created", &userRolePermission)
	return &userRolePermission, nil
}

// GetUserRolePermissionByID gives requested user role permission by id.
func GetUserRolePermissionByID(ID string) (*UserRolePermissions, error) {
	db := database.MongoDB
	userRolePermission := &UserRolePermissions{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(userRolePermission)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(UserRolePermissionsCollection).FindOne(ctx, filter).Decode(&userRolePermission)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, userRolePermission, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return userRolePermission, nil
}

// UpdateUserRolePermission updates the user role group.
func UpdateUserRolePermission(userRolePermission *UserRolePermissions) *UserRolePermissions {
	userRolePermission.UpdatedAt = time.Now()
	filter := bson.D{{"_id", userRolePermission.ID}}
	db := database.MongoDB
	collection := db.Collection(UserRolePermissionsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, userRolePermission, findRepOpts).Decode(&userRolePermission)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(userRolePermission.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role_permission.updated", &userRolePermission)
	return userRolePermission
}

// DeleteUserRolePermissionByID deletes user role Permissions by id.
func DeleteUserRolePermissionByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(UserRolePermissionsCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_role_permission.deleted", &res)
	return true, nil
}

func GetUserRolePermissionsByFilter(filter bson.D) *UserRolePermissions {
	db := database.MongoDB
	dbUserRolePermission := new(UserRolePermissions)
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	err := db.Collection(UserRolePermissionsCollection).FindOne(context.Background(), filter).Decode(dbUserRolePermission)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
		}
		return nil
	}
	return dbUserRolePermission
}

func GetUserRolePermissions(filter bson.D, limit int, after *string, before *string, first *int, last *int) (userRolePermissions []*UserRolePermissions, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(UserRolePermissionsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(UserRolePermissionsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		userRolePermission := &UserRolePermissions{}
		err = cur.Decode(&userRolePermission)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		userRolePermissions = append(userRolePermissions, userRolePermission)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return userRolePermissions, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (userRolePermission *UserRolePermissions) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, userRolePermission); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (userRolePermission *UserRolePermissions) MarshalBinary() ([]byte, error) {
	return json.Marshal(userRolePermission)
}
