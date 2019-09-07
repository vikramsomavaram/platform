/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/gqlerror"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/lib/msg91"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"strings"
	"time"
)

//BlockUser returns a block user by ID
func (r *mutationResolver) BlockUser(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	user := models.GetUserByID(id.Hex())
	if user.ID.IsZero() {
		return nil, errors.New("invalid user")
	}
	user.IsLocked = true
	_, err := models.UpdateUser(user)
	if err != nil {
		return utils.PointerBool(false), err
	}
	return utils.PointerBool(false), nil
}

//Me returns current logged in user details.
func (r *queryResolver) Me(ctx context.Context) (*models.User, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		log.Errorln(err)
	}
	if user != nil {
		return user, nil
	}
	return nil, errors.New("invalid authentication")
}

func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	accessToken := auth.JwtToken(ctx)
	ok, err := models.DeleteAccessToken(accessToken)
	if err != nil && !ok {
		log.Error(err)
		return false, err
	}
	return true, nil
}

//User returns a user by ID
func (r *queryResolver) User(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	user := models.GetUserByID(id.Hex())
	return user, nil
}

type userResolver struct{ *Resolver }

//Roles returns assigned roles of a given user.
func (r *userResolver) Roles(ctx context.Context, obj *models.User) ([]*models.UserRole, error) {
	roles := models.GetUserRoles(obj.ID.Hex())
	return *roles, nil
}

//RoleGroups returns assigned role groups of given user.
func (r *userResolver) RoleGroups(ctx context.Context, obj *models.User) ([]*models.UserRoleGroup, error) {
	roleGroups := models.GetUserRoleGroupsByUserID(obj.ID.Hex())
	return *roleGroups, nil
}

//Users gives a list of users
func (r *queryResolver) Users(ctx context.Context, userType *models.UserSearchType, text *string, userStatus *models.UserStatus, after *string, before *string, first *int, last *int) (*models.UserConnection, error) {
	var items []*models.User
	var edges []*models.UserEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetUsers(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.UserEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)
	itemList := &models.UserConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AddUser adds a new user
func (r *mutationResolver) AddUser(ctx context.Context, input models.AddUserInput) (*models.User, error) {

	db := database.MongoDB
	var user models.User
	if input.MobileNo != "" && input.Password != "" {
		filter := bson.D{{"mobile_no", input.MobileNo}}
		err := db.Collection(models.UsersCollection).FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, errors.New("internal server error")
			}
		}
	}

	if input.Email != "" && input.Password != "" {
		filter := bson.D{{"email", input.Email}}
		err := db.Collection(models.UsersCollection).FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, errors.New("internal server error")
			}
		}
	}

	if user.ID.IsZero() {
		_ = copier.Copy(&user, &input)
		user.ID = primitive.NewObjectID()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.IsEmailVerified = false
		user.IsActive = true
		user.IsMobileVerified = false
		//user.OTP = utils.RandomInt(6)
		user.Password = utils.HashPassword(input.Password)
		userCollection := db.Collection(models.UsersCollection)
		_, err := userCollection.InsertOne(context.TODO(), user)
		if err != nil {
			return nil, errors.New("internal server error")
		}

		//Send a email
		err = models.SendEmail("", "", "user.welcome_email", user.Language, nil, nil)
		if err != nil {
			log.Errorln(err)
		}
		//Send verification SMS
		sent, err := msg91.SendMessage("Welcome to Tribe! "+user.OTP+" is your Verification OTP.", true, strings.TrimPrefix(user.MobileNo, "+"))
		if !sent || err != nil {
			log.Errorln(err)
		}

		return &user, nil
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), user.ID.Hex(), "user", user, nil, ctx)
	return nil, errors.New("bad request")
}

//SignUpWithEmail lets you sign up with email address
func (r *mutationResolver) SignUpWithEmail(ctx context.Context, input models.UserSignUpDetails) (*models.AuthPayload, error) {
	db := database.MongoDB
	var user models.User
	if input.MobileNo != nil && input.Password != "" {
		filter := bson.D{{"mobileNo", input.MobileNo}}
		err := db.Collection(models.UsersCollection).FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, errors.New("internal server error")
			}
		}
	}

	if input.Email != "" && input.Password != "" {
		filter := bson.D{{"email", input.Email}}
		err := db.Collection(models.UsersCollection).FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, errors.New("internal server error")
			}
		}
	}

	if user.ID.IsZero() {
		_ = copier.Copy(&user, &input)
		user.ID = primitive.NewObjectID()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.IsEmailVerified = false
		user.IsActive = true
		user.IsMobileVerified = false
		user.OTP = utils.RandomInt(6)
		user.Password = utils.HashPassword(input.Password)
		user.Roles = []string{"user"}
		userCollection := db.Collection(models.UsersCollection)
		_, err := userCollection.InsertOne(context.TODO(), user)
		if err != nil {
			return nil, errors.New("internal server error")
		}

		//Send a email
		err = models.SendEmail("no-reply@tribe.cab", user.Email, "user.account.email_verification", user.Language, nil, nil)
		if err != nil {
			log.Errorln(err)
		}
		////Send verification SMS
		//sent, err := msg91.SendMessage("Welcome to Tribe! "+user.OTP+" is your Verification OTP.", true, strings.TrimPrefix(user.MobileNo, "+"))
		//if !sent || err != nil {
		//	log.Errorln(err)
		//}

		// Create token
		token := jwt.New(jwt.SigningMethodHS256)
		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["id"] = user.ID
		claims["exp"] = time.Now().Add(time.Hour * 1440).Unix() //60days
		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte("jwtsecret"))
		if err != nil {
			return nil, errors.New("internal server error")
		}
		authPayload := &models.AuthPayload{
			Token: t,
			User:  &user,
		}
		return authPayload, &gqlerror.Error{Message: "please verify your email address", Extensions: map[string]interface{}{"code": "user_verify_email"}}
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), user.ID.Hex(), "sign up with email", user, nil, ctx)
	return nil, errors.New("bad request")
}

//UpdateUser updates an existing user
func (r *mutationResolver) UpdateUser(ctx context.Context, input models.UpdateUserInput) (*models.User, error) {
	user := &models.User{}
	_ = copier.Copy(&user, &input)
	user.UpdatedAt = time.Now()
	user, err := models.UpdateUser(user)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), user.ID.Hex(), "user", user, nil, ctx)
	return user, nil
}

//LoginWithCredentials checks your credentials and logs in
func (r *mutationResolver) LoginWithCredentials(ctx context.Context, email string, password string) (*models.AuthPayload, error) {
	db := database.MongoDB

	var user models.User

	if email != "" && password != "" {
		filter := bson.D{{"email", email}}
		err := db.Collection(models.UsersCollection).FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			return nil, errors.New("please check your username or password")
		}
	}

	if user.ID.Hex() != "" && !user.IsActive {
		return nil, errors.New("your account is not active, please contact our support team for more details")
	}

	if comparePasswords(user.Password, []byte(password)) {
		// Create token
		token := jwt.New(jwt.SigningMethodHS256)
		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["id"] = user.ID
		claims["exp"] = time.Now().Add(time.Hour * 1440).Unix() //60days
		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte("jwtsecret"))
		if err != nil {
			return nil, err
		}

		//utils.SendLoginMail(db, utils.LoginTemplate{Name: user.FirstName, IPAddress: ctx.Request().RemoteAddr, Client: ctx.Request().UserAgent(), TimeStamp: time.Now().Format(utils.TimeLayout)}, utils.EmailRequest{To: user.Email, From: utils.SUPPORT_EMAIL})

		authPayload := &models.AuthPayload{
			Token: t,
			User:  &user,
		}
		return authPayload, nil
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.LoggedIn, user.ID.Hex(), user.ID.Hex(), "user", user, nil, ctx)
	return nil, errors.New("invalid user credentials, lets try again")

}

func generateSecret(s int) string {
	b := make([]byte, s)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		return false
	}
	return true
}

//RequestLoginOtp requests for a password
func (r *mutationResolver) RequestLoginOtp(ctx context.Context, countryCode string, mobileNo string) (*bool, error) {

	var user models.User
	db := database.ConnectMongo()
	if mobileNo != "" {

		filter := bson.D{{"mobileNo", countryCode + mobileNo}}
		err := db.Collection(models.UsersCollection).FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				log.Errorln(err)
				return nil, errors.New("internal server error")
			}
			return nil, errors.New("invalid mobile number")

		}

		if !user.ID.IsZero() && !user.IsActive {
			return nil, errors.New("your account is not active, please contact our support for more details")
		}

		user.OTP = utils.RandomInt(6)
		idFilter := bson.D{{"_id", user.ID}}
		update := bson.D{{"$set", bson.D{{"otp", user.OTP}}}}
		usersCollection := db.Collection(models.UsersCollection)
		res := usersCollection.FindOneAndUpdate(context.TODO(), idFilter, update)
		if res.Err() != nil {
			log.Errorln(res.Err())
			return nil, errors.New("internal server error")
		}
		sent, err := msg91.SendMessage("Welcome to Tribe! "+user.OTP+" is your One Time Pin.", true, strings.TrimPrefix(user.MobileNo, "+"))
		if !sent || err != nil {
			log.Errorln(err)
		}
		return utils.PointerBool(true), nil
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Requested, user.ID.Hex(), user.ID.Hex(), "login otp", user.OTP, nil, ctx)
	return nil, errors.New("mobile number is required")

}

//VerifyLoginOtp verifies your one time password
func (r *mutationResolver) VerifyLoginOtp(ctx context.Context, countryCode string, mobileNo string, otp string) (*models.AuthPayload, error) {
	var user models.User
	db := database.ConnectMongo()
	if mobileNo != "" && otp != "" {

		filter := bson.D{{"mobileNo", countryCode + mobileNo}}
		err := db.Collection(models.UsersCollection).FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				log.Errorln(err)
				return nil, errors.New("internal server error")
			}
			return nil, errors.New("invalid mobile number")
		}

		if !user.ID.IsZero() && user.OTP == otp {
			filter := bson.D{{"_id", user.ID}}
			update := bson.D{{"$set", bson.D{{"otp", ""}}}}
			usersCollection := db.Collection(models.UsersCollection)
			res := usersCollection.FindOneAndUpdate(context.TODO(), filter, update)
			if res.Err() != nil {
				log.Errorln(res.Err())
				return nil, errors.New("internal server error")
			}

			// Create token
			token := jwt.New(jwt.SigningMethodHS256)
			// Set claims
			claims := token.Claims.(jwt.MapClaims)
			claims["id"] = user.ID
			claims["exp"] = time.Now().Add(time.Hour * 1440).Unix()
			// Generate encoded token and send it as response.
			t, err := token.SignedString([]byte("jwtsecret"))
			if err != nil {
				return nil, errors.New("internal server error")
			}
			//utils.SendLoginMail(db, utils.LoginTemplate{Name: user.FirstName, IPAddress: ctx.Request().RemoteAddr, Client: ctx.Request().UserAgent(), TimeStamp: time.Now().Format(utils.TimeLayout)}, utils.EmailRequest{To: user.Email, From: utils.SUPPORT_EMAIL})
			authPayload := &models.AuthPayload{
				Token: t,
				User:  &user,
			}
			return authPayload, nil
		}
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Verified, user.ID.Hex(), user.ID.Hex(), "login otp", user.OTP, nil, ctx)
	return nil, errors.New("invalid mobile number or otp")
}

//LoginWithSocialAuth lets you login using social media
func (r *mutationResolver) LoginWithSocialAuth(ctx context.Context, provider models.SocialAuthProvder, accessToken string, accessSecret *string) (*models.AuthPayload, error) {
	switch provider {
	case "FACEBOOK":
		//TODO implement facebook login
		return nil, errors.New("facebook to be implemented")
	case "GOOGLE":
		//TODO implement google login
		return nil, errors.New("google to be implemented")
	case "MICROSOFT":
		//TODO implement google login
		return nil, errors.New("microsoft to be implemented")
	case "AMAZON":
		//TODO implement google login
		return nil, errors.New("amazon to be implemented")
	default:
		return nil, errors.New("invalid oauth2 provider")
	}
}

//DeleteUser deletes an existing user
func (r *mutationResolver) DeleteUser(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteUserByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "user", nil, nil, ctx)
	return &res, err
}

//DeactivateUser deactivates a user by ID
func (r *mutationResolver) DeactivateUser(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	user := models.GetUserByID(id.Hex())
	if user == nil {
		return utils.PointerBool(false), errors.New("invalid user")
	}
	user.IsActive = false
	_, err := models.UpdateUser(user)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err = auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "user", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ActivateUser activates a user by ID
func (r *mutationResolver) ActivateUser(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	user := models.GetUserByID(id.Hex())
	if user == nil {
		return utils.PointerBool(false), errors.New("invalid user")
	}
	user.IsActive = true
	_, err := models.UpdateUser(user)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err = auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "user", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//TODO
//ChangeUserProfilePassword changes user's password
func (r *mutationResolver) ChangeUserProfilePassword(ctx context.Context, currentPassword *string, newPassword *string, confirmNewPassword *string) (*bool, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		log.Errorln(err)
	}

	if *currentPassword == user.Password && newPassword == confirmNewPassword {
		user.Password = utils.HashPassword(*newPassword)
		_, err := models.UpdateUser(user)
		if err != nil {
			log.Errorln(err)
		}
	}
	user, err = auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), user.ID.Hex(), "user password", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//TODO: send mail
//ResetUserProfilePassword resets the existing password of user
func (r *mutationResolver) ResetUserProfilePassword(ctx context.Context, token string, newPassword, confirmNewPassword string) (bool, error) {
	passwordToken, err := models.GetPasswordToken(token)
	if err != nil {
		return false, err
	}
	if passwordToken != nil && passwordToken.ExpiresAt.Sub(time.Now()).Minutes() < 15 {
		if newPassword == confirmNewPassword {
			user := models.GetUserByID(passwordToken.UserID.Hex())
			user.Password = utils.HashPassword(newPassword)
			_, err := models.UpdateUser(user)
			if err != nil {
				log.Errorln(err)
				return false, errors.New("internal server error")
			}
			err = models.SendEmail("noreply@tribe.cab", user.Email, "user.update.changed_password", user.Language, passwordToken, nil)
			if err != nil {
				log.Errorln(err)
				return false, errors.New("internal server error")
			}
			user, err = auth.ForContext(ctx)
			if err != nil {
				return false, err
			}
			//Update audit log
			go audit_log.NewAuditLogWithCtx(models.Reset, user.ID.Hex(), user.ID.Hex(), "user password", nil, nil, ctx)
			return true, nil
		}
		return false, errors.New("passwords do not match")
	}
	return false, errors.New("invalid password token")
}

//RequestResetPassword sends a token to the user
func (r *mutationResolver) RequestResetPassword(ctx context.Context, mobileNumber *string, email *string) (bool, error) {
	user := &models.User{}
	if mobileNumber != nil {
		user = models.GetUserByFilter(bson.D{{"mobileNo", *mobileNumber}})
		if user == nil {
			return false, errors.New("invalid user account")
		}
	}
	if email != nil {
		user = models.GetUserByFilter(bson.D{{"email", *email}})
		if user == nil {
			return false, errors.New("invalid user account")
		}
	}
	//TODO Update metadata of password token with user access details (IP ADDR, Browser, etc.)
	passwordToken := &models.PasswordToken{UserID: user.ID, TokenType: models.RequestPassword, Token: utils.RandomIDGen(25), ExpiresAt: time.Now().Add(time.Minute * 15)}
	passwordToken, err := models.CreatePasswordToken(passwordToken)
	if err != nil {
		return false, errors.New("internal server error")
	}
	err = models.SendEmail("noreply@tribe.cab", user.Email, "user.request.reset_password", user.Language, passwordToken, nil)
	if err != nil {
		log.Errorln(err)
		return false, errors.New("internal server error")
	}
	user, err = auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Requested, user.ID.Hex(), user.ID.Hex(), "user password reset", nil, nil, ctx)
	return true, nil
}

//TODO
//ChangeUserMobileNumber changes user's mobile number
func (r *mutationResolver) ChangeUserMobileNumber(ctx context.Context, countryCode string, mobileNumber int, changeAuthToken string) (*bool, error) {
	panic("not implemented")
}

//AssignRoleToUser assigns role to user.
func (r *mutationResolver) AssignRoleToUser(ctx context.Context, roleID primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	user := models.GetUserByID(userID.Hex())
	role, err := models.GetUserRoleByID(roleID.Hex())
	if err != nil {
		return false, err
	}
	user.Roles = appendToSlice(user.Roles, role.Name)
	updatedUser, err := models.UpdateUser(user)
	if err != nil {
		return false, err
	}
	if reflect.DeepEqual(updatedUser.Roles, user.Roles) {
		return true, nil
	}
	user, err = auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Assigned, user.ID.Hex(), roleID.Hex(), "user role", nil, nil, ctx)
	return false, nil
}

func appendToSlice(slice []string, element string) []string {
	for _, ele := range slice {
		if ele == element {
			return slice
		}
	}
	return append(slice, element)

}

//AssignRoleGroupToUser assigns role group to user.
func (r *mutationResolver) AssignRoleGroupToUser(ctx context.Context, roleGroupID primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	user := models.GetUserByID(userID.Hex())
	roleGroup := models.GetUserRoleGroupByFilter(bson.D{{"_id", roleGroupID}})
	user.RoleGroups = appendToSlice(user.RoleGroups, roleGroup.Name)
	updatedUser, err := models.UpdateUser(user)
	if err != nil {
		return false, err
	}
	if reflect.DeepEqual(updatedUser.RoleGroups, user.RoleGroups) {
		return true, nil
	}
	user, err = auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Assigned, user.ID.Hex(), roleGroupID.Hex(), "user role group", nil, nil, ctx)
	return false, nil
}

//AssignRoleToRoleGroup assigns role to role group.
func (r *mutationResolver) AssignRoleToRoleGroup(ctx context.Context, roleID primitive.ObjectID, roleGroupID primitive.ObjectID) (bool, error) {
	roleGroup := models.GetUserRoleGroupByFilter(bson.D{{"_id", roleGroupID}})
	role, err := models.GetUserRoleByID(roleID.Hex())
	if err != nil {
		return false, err
	}
	roleGroup.Roles = appendToSlice(roleGroup.Roles, role.Name)
	updatedRoleGroup, err := models.UpdateUserRoleGroup(roleGroup)
	if err != nil {
		return false, err
	}
	if reflect.DeepEqual(updatedRoleGroup.Roles, roleGroup.Roles) {
		return true, nil
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Assigned, user.ID.Hex(), roleGroupID.Hex(), "role group role", nil, nil, ctx)
	return false, nil
}

//AssignPermissionToUserRole assigns permission to user role.
func (r *mutationResolver) AssignPermissionToUserRole(ctx context.Context, permission string, roleID primitive.ObjectID) (bool, error) {
	userRole, err := models.GetUserRoleByID(roleID.Hex())
	if err != nil {
		return false, err
	}
	userRole.Permissions = appendToSlice(userRole.Permissions, permission)
	updatedUserRole, err := models.UpdateUserRole(userRole)
	if err != nil {
		return false, err
	}
	if reflect.DeepEqual(updatedUserRole.Permissions, userRole.Permissions) {
		return true, nil
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Assigned, user.ID.Hex(), roleID.Hex(), "permission", nil, nil, ctx)
	return false, nil
}

//AddUserRole adds new user role.
func (r *mutationResolver) AddUserRole(ctx context.Context, input models.AddUserRoleInput) (*models.UserRole, error) {
	userRole := &models.UserRole{}
	_ = copier.Copy(&userRole, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	userRole.CreatedBy = user.ID
	userRole, err = models.CreateUserRole(*userRole)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), userRole.ID.Hex(), "user role", userRole, nil, ctx)
	return userRole, nil
}

//UpdateUserRole updates the user role.
func (r *mutationResolver) UpdateUserRole(ctx context.Context, input models.UpdateUserRoleInput) (*models.UserRole, error) {
	userRole := &models.UserRole{}
	userRole, err := models.GetUserRoleByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&userRole, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	userRole.CreatedBy = user.ID
	userRole, err = models.UpdateUserRole(userRole)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), userRole.ID.Hex(), "user role", userRole, nil, ctx)
	return userRole, nil
}

//DeleteUserRole deletes user role.
func (r *mutationResolver) DeleteUserRole(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteUserRoleByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "user role", nil, nil, ctx)
	return &res, err
}

//AddUserRoleGroup adds user role group.
func (r *mutationResolver) AddUserRoleGroup(ctx context.Context, input models.AddUserRoleGroupInput) (*models.UserRoleGroup, error) {
	userRoleGroup := &models.UserRoleGroup{}
	_ = copier.Copy(&userRoleGroup, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	userRoleGroup.CreatedBy = user.ID
	userRoleGroup, err = models.CreateUserRoleGroup(userRoleGroup)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), userRoleGroup.ID.Hex(), "user role group", userRoleGroup, nil, ctx)
	return userRoleGroup, nil
}

//UpdateUserRoleGroup updates user role group.
func (r *mutationResolver) UpdateUserRoleGroup(ctx context.Context, input models.UpdateUserRoleGroupInput) (*models.UserRoleGroup, error) {
	userRoleGroup := &models.UserRoleGroup{}
	userRoleGroup = models.GetUserRoleGroupByFilter(bson.D{{"_id", input.ID}})
	_ = copier.Copy(&userRoleGroup, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	userRoleGroup.CreatedBy = user.ID
	userRoleGroup, err = models.UpdateUserRoleGroup(userRoleGroup)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), userRoleGroup.ID.Hex(), "user role group", userRoleGroup, nil, ctx)
	return userRoleGroup, nil
}

//DeleteUserRoleGroup deletes user role group by id.
func (r *mutationResolver) DeleteUserRoleGroup(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteUserRoleGroupByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "user role group", nil, nil, ctx)
	return &res, err
}

//UserRoles returns list of user roles.
func (r *queryResolver) UserRoles(ctx context.Context, after *string, before *string, first *int, last *int) (*models.UserRoleConnection, error) {
	var items []*models.UserRole
	var edges []*models.UserRoleEdge
	filter := bson.D{}
	limit := 300
	items, totalCount, hasPrevious, hasNext, err := models.GetAllUserRoles(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.UserRoleEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.UserRoleConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//UserRole returns user role by id.
func (r *queryResolver) UserRole(ctx context.Context, id primitive.ObjectID) (*models.UserRole, error) {
	userRole, err := models.GetUserRoleByID(id.Hex())
	if err != nil {
		return nil, err
	}
	return userRole, nil
}

//UserRoleGroups returns a list of user role groups.
func (r *queryResolver) UserRoleGroups(ctx context.Context, after *string, before *string, first *int, last *int) (*models.UserRoleGroupConnection, error) {
	var items []*models.UserRoleGroup
	var edges []*models.UserRoleGroupEdge
	filter := bson.D{}
	limit := 300
	items, totalCount, hasPrevious, hasNext, err := models.GetUserRoleGroups(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.UserRoleGroupEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.UserRoleGroupConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

func getPageInfo(startCursor string, endCursor string, edgeLength int, hasNext bool, hasPrevious bool) *models.PageInfo {
	pageInfo := &models.PageInfo{}
	if edgeLength > 1 {
		pageInfo.StartCursor = startCursor
		pageInfo.EndCursor = endCursor
		pageInfo.HasNextPage = hasNext
		pageInfo.HasPreviousPage = hasPrevious
	}
	return pageInfo
}

//UserRoleGroup returns user role group by id.
func (r *queryResolver) UserRoleGroup(ctx context.Context, id primitive.ObjectID) (*models.UserRoleGroup, error) {
	userRoleGroup := models.GetUserRoleGroupByFilter(bson.D{{"_id", id}})
	return userRoleGroup, nil
}

func remove(slice []string, key string) []string {
	for i, v := range slice {
		if v == key {
			slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	return slice
}

//UnAssignUserRole unassigns user role.
func (r *mutationResolver) UnAssignUserRole(ctx context.Context, roleId primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	user := models.GetUserByID(userID.Hex())
	role, err := models.GetUserRoleByID(roleId.Hex())
	if err != nil {
		return false, err
	}
	user.Roles = remove(user.Roles, role.Name)
	updatedUser, err := models.UpdateUser(user)
	if err != nil {
		return false, err
	}
	for _, n := range updatedUser.Roles {
		if n == role.Name {
			return false, nil
		}
		return true, nil
	}
	user, err = auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Unassigned, user.ID.Hex(), roleId.Hex(), "user role", nil, nil, ctx)
	return true, nil
}

//UnAssignUserRoleGroup un assigns user role group.
func (r *mutationResolver) UnAssignUserRoleGroup(ctx context.Context, roleGroupID primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	user := models.GetUserByID(userID.Hex())
	roleGroup := models.GetUserRoleGroupByFilter(bson.D{{"_id", roleGroupID}})
	user.RoleGroups = remove(user.RoleGroups, roleGroup.Name)
	updatedUser, err := models.UpdateUser(user)
	if err != nil {
		return false, err
	}
	for _, n := range updatedUser.RoleGroups {
		if n == roleGroup.Name {
			return false, nil
		}
		return true, nil
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Unassigned, user.ID.Hex(), roleGroupID.Hex(), "user role group", nil, nil, ctx)
	return true, nil
}

//UnAssignRoleGroupRole un assigns the group roles.
func (r *mutationResolver) UnAssignRoleGroupRole(ctx context.Context, roleID primitive.ObjectID, roleGroupID primitive.ObjectID) (bool, error) {
	roleGroup := models.GetUserRoleGroupByFilter(bson.D{{"_id", roleGroupID}})
	role, err := models.GetUserRoleByID(roleID.Hex())
	if err != nil {
		return false, nil
	}
	roleGroup.Roles = remove(roleGroup.Roles, role.Name)
	updatedRoleGroup, err := models.UpdateUserRoleGroup(roleGroup)
	if err != nil {
		return false, err
	}
	for _, n := range updatedRoleGroup.Roles {
		if n == role.Name {
			return false, nil
		}
		return true, nil
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Unassigned, user.ID.Hex(), roleID.Hex(), "role group role", nil, nil, ctx)
	return true, nil
}

//UnAssignUserRolePermission un assigns the permissions.
func (r *mutationResolver) UnAssignUserRolePermission(ctx context.Context, permission string, roleID primitive.ObjectID) (bool, error) {
	userRole, err := models.GetUserRoleByID(roleID.Hex())
	if err != nil {
		return false, err
	}
	userRole.Permissions = remove(userRole.Permissions, permission)
	updatedUserRole, err := models.UpdateUserRole(userRole)
	if err != nil {
		return false, err
	}
	for _, n := range updatedUserRole.Permissions {
		if n == permission {
			return false, nil
		}
		return true, nil
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Unassigned, user.ID.Hex(), roleID.Hex(), "user role permission", nil, nil, ctx)
	return true, nil
}
