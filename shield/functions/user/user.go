package user

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aidarkhanov/nanoid"
	"github.com/appblocks-hub/SHIELD/functions/appreg"
	"github.com/appblocks-hub/SHIELD/functions/auth"
	"github.com/appblocks-hub/SHIELD/functions/general"
	"github.com/appblocks-hub/SHIELD/functions/mailer"
	"github.com/appblocks-hub/SHIELD/functions/pghandling"
	"github.com/appblocks-hub/SHIELD/functions/token"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Signup decodes the json payload into the struct.
// It passes the Password to HashPassword() for salt and hash.
// It called SetupDB() to nitialize database connection
// and insert an entry into DB using gorm Create().
// It returns json result with HTTP status code.

// swagger:route POST /signup SignupRequest
// Create a new user
//
// parameters:
//
//	SignupRequest
//
// responses:
//
//	201: Response
//	400: Response
func Signup(c *gin.Context) {
	// invalidate cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     general.Envs["ID_TOKEN_NAME"],
		Value:    "",
		Domain:   ".appblocks.com",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

	// fetching request url params for attach with request
	var urlParams general.AppUrlParams
	urlParams.ClientId = c.Query("client_id")
	urlParams.ResponseType = c.Query("response_type")
	urlParams.State = c.Query("state")
	urlParams.RedirectUri = c.Query("redirect_uri")

	//for app token exchange
	urlParams.GrantType = c.Query("grant_type")

	q := url.Values{}
	if len(urlParams.ClientId) != 0 && ((urlParams.ResponseType == "code" || urlParams.ResponseType == "device_code") && len(urlParams.State) != 0 && len(urlParams.RedirectUri) != 0) {
		q.Set("client_id", urlParams.ClientId)
		q.Set("response_type", urlParams.ResponseType)
		q.Set("state", urlParams.State)
		q.Set("redirect_uri", urlParams.RedirectUri)
	}
	if len(urlParams.ClientId) != 0 && urlParams.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", urlParams.GrantType)
		q.Set("client_id", urlParams.ClientId)
	}

	var newUser general.SignupRequestEvent

	// err := c.BindJSON(&newUser)
	// if err != nil {
	// 	log.Println(err)
	// 	q.Set("error", "Something went wrong")
	// 	location := url.URL{Path: "/signup", RawQuery: q.Encode()}
	// 	c.Redirect(http.StatusFound, location.RequestURI())
	// 	return
	// }

	if strings.ContainsAny(c.PostForm("password"), " ") || strings.ContainsAny(c.PostForm("username"), " ") {
		log.Println("Password or Username contains space")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Invalid Username or Password")
		return
	}
	newUser.Email = strings.TrimSpace(strings.ToLower(c.PostForm("email")))
	newUser.Password = strings.TrimSpace(c.PostForm("password"))
	newUser.UserName = strings.TrimSpace(c.PostForm("username"))
	newUser.FullName = strings.TrimSpace(c.PostForm("fullname"))
	newUser.Address1 = strings.TrimSpace(c.PostForm("address1"))
	newUser.Address2 = strings.TrimSpace(c.PostForm("address2"))
	newUser.Phone = strings.TrimSpace(c.PostForm("phone"))

	err := general.ValidateStruct(&newUser)
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Invalid request payload")
		return
	}

	hashPwd, err := general.HashPassword(newUser.Password)
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
		return
	}

	newUser.Password = hashPwd

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
		return
	}

	defer sqlDB.Close()

	var Users general.User

	result := db.Limit(2).Find(&Users, "user_name = ? OR email = ?", newUser.UserName, newUser.Email)
	if result.Error != nil {
		log.Println(result.Error)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
		return
	}

	if result.RowsAffected > 1 {
		// email and username used by different accounts
		log.Println("Email/Username is already taken")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Email/Username is already taken")
		return
	}

	var userid string
	var UserProviders general.UserProvider

	VerificationCode := general.GetRandomCode(6)
	VerificationExpiry := time.Now().Add(time.Minute * 10)

	if result.RowsAffected > 0 {
		// checking same email or not
		// if Users.Email != newUser.Email {
		// 	log.Println("Username is already taken")
		// 	general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Username is already taken")
		// 	return
		// }

		// user already exists in user table

		userid = Users.UserID

		result := db.Limit(1).Find(&UserProviders, "user_id = ? AND provider = ?", userid, general.PROVIDER_SHIELD)
		if result.Error != nil {
			log.Println(result.Error)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
			return
		}

		if result.RowsAffected > 0 {
			log.Println("Email is already taken")
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Email is already taken")
			return
		}

		// create entry to UserProvider table
		UserProviders.UserId = userid
		UserProviders.Provider = general.PROVIDER_SHIELD

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&general.User{}).Where("user_Id = ?", userid).Select("user_name", "email", "password", "email_verification_code", "email_verified", "email_verification_expiry").Updates(general.User{UserName: newUser.UserName, Email: newUser.Email, Password: newUser.Password, EmailVerificationCode: VerificationCode, EmailVerified: false, EmailVerificationExpiry: VerificationExpiry}).Error; err != nil {
				// return any error - will rollback
				return err
			}
			if err := tx.Create(&UserProviders).Error; err != nil {
				// return any error - will rollback
				return err
			}
			// return nil will commit the whole transaction
			return nil
		})

		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
			return
		}

	} else {

		userid = nanoid.New()

		Users.UserID = userid
		Users.UserName = newUser.UserName
		Users.FullName = newUser.FullName
		Users.Email = newUser.Email
		Users.Password = newUser.Password
		Users.Address1 = newUser.Address1
		Users.Address2 = newUser.Address2
		Users.Phone = newUser.Phone
		Users.EmailVerificationCode = VerificationCode
		Users.EmailVerified = false
		Users.EmailVerificationExpiry = VerificationExpiry

		UserProviders.UserId = userid
		UserProviders.Provider = general.PROVIDER_SHIELD

		var tx = db.Begin()

		err = auth.CreateUserSpace(tx, Users)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
			return
		}

		result := tx.Create(&UserProviders)

		if result.Error != nil {
			log.Println(result.Error)
			tx.Rollback()
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
			return
		}

		tx.Commit()

		// err = db.Transaction(func(tx *gorm.DB) error {
		// 	if err := tx.Create(&Users).Error; err != nil {
		// 		// return any error - will rollback
		// 		return err
		// 	}
		// 	if err := tx.Create(&UserProviders).Error; err != nil {
		// 		// return any error - will rollback
		// 		return err
		// 	}
		// 	// return nil will commit the whole transaction
		// 	return nil
		// })

		// if err != nil {
		// 	log.Println(err)
		// 	general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", "Something went wrong")
		// 	return
		// }

	}

	// Send emial verification =>
	go func() {

		// TODO optimize
		// split the verification code and adds a - at center
		vcSplit := strings.Split(VerificationCode, "")
		vcSlice := append(vcSplit[:3], vcSplit[2:]...)
		vcSlice[3] = " - "
		modifiedVerificationCode := strings.Join(vcSlice, "")

		verfiyEmailUserData := mailer.VerifyEmailUserStruct{
			UserName:         newUser.UserName,
			Email:            newUser.Email,
			VerificationCode: modifiedVerificationCode,
		}

		SendEmailVerification(verfiyEmailUserData)
	}()
	//  <= Send emial verification;

	log.Println(userid)

	q.Set("user_id", userid)
	q.Set("user_email", newUser.Email)

	location := url.URL{Path: "/verify-user-email", RawQuery: q.Encode()}
	c.Redirect(http.StatusFound, location.RequestURI())

	// general.RespondHandler(c.Writer, true, http.StatusCreated, map[string]string{"user_id": userid.String()})

}

// Login decodes the json payload into the struct.
// It validates the credentials.
// If credentials matched, it will return the token pair.

// swagger:route POST /login LoginRequest
// Login to a user account
//
// parameters:
//  LoginRequest
// responses:
//  200: Response
//  400: Response

func Login(c *gin.Context) {

	var urlParams general.AppUrlParams

	urlParams.ClientId = c.Query("client_id")
	urlParams.ResponseType = c.Query("response_type")
	urlParams.State = c.Query("state")
	urlParams.RedirectUri = c.Query("redirect_uri")

	//for app token exchange
	urlParams.GrantType = c.Query("grant_type")

	var appds *general.AppFromClientId
	if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(urlParams.ClientId, urlParams.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Not a valid client id")
			return
		}

		if err != nil {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Failed to fetch app details from db")
			return
		}

		if !Valid {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Given redirect uri not registered for this app")
			return
		}

	} else {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Access blocked: authorisation error: invalid_request")
		return
	}

	q := url.Values{}
	if len(urlParams.ClientId) != 0 && ((urlParams.ResponseType == "code" || urlParams.ResponseType == "device_code") && len(urlParams.State) != 0 && len(urlParams.RedirectUri) != 0) {
		q.Set("client_id", urlParams.ClientId)
		q.Set("response_type", urlParams.ResponseType)
		q.Set("state", urlParams.State)
		q.Set("redirect_uri", urlParams.RedirectUri)
	}
	if len(urlParams.ClientId) != 0 && urlParams.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", urlParams.GrantType)
		q.Set("client_id", urlParams.ClientId)
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Something went wrong")
		return
	}

	var userCreds general.LoginRequestEvent

	userCreds.EmailOrUserName = strings.TrimSpace(c.PostForm("email"))
	userCreds.Password = strings.TrimSpace(c.PostForm("password"))

	//TODO :err

	// if c.Request.Form == nil {
	// 	if err := c.Request.ParseForm(); err != nil {
	// 		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	// err := c.BindJSON(&userCreds)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Invalid JSON: "+err.Error())
	// 	log.Println(err)
	// 	return
	// }

	err := general.ValidateStruct(&userCreds)
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Invalid request payload")
		return
	}

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Something went wrong")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Something went wrong")
		return
	}

	defer sqlDB.Close()

	//var Users general.User

	var userwprovider general.UserWithProvider

	result := db.Model(&general.User{}).Select("users.user_id, users.user_name, users.email, users.password, users.email_verified, user_providers.provider").Joins("inner join user_providers on users.user_id = user_providers.user_id").Where("users.email = ? OR users.user_name = ?", userCreds.EmailOrUserName, userCreds.EmailOrUserName).Where("user_providers.provider = ?", general.PROVIDER_SHIELD).Limit(1).Scan(&userwprovider)
	if result.Error != nil {
		log.Println(result.Error)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Something went wrong")
		return
	}

	if result.RowsAffected <= 0 {
		log.Println("Incorrect Username/Password.")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Incorrect Username/Password.")
		return
	}

	log.Println("userwprovider:", userwprovider)

	err = bcrypt.CompareHashAndPassword([]byte(userwprovider.Password), []byte(userCreds.Password))

	if err != nil {
		log.Println("hash password error/mismatch")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Incorrect Username/Password.")
		return
	}

	// if email not verified, redirect to verify email
	if !(userwprovider.EmailVerified) {
		q.Set("user_id", userwprovider.UserId)
		q.Set("user_email", userwprovider.Email)

		vlocation := url.URL{Path: "/verify-user-email", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, vlocation.RequestURI())
		return

	}

	// generate device id
	DeviceId := general.RandPassword(8)

	//Complete auth
	if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {

		ad := &general.AccessDetails{
			AccessUuid:  "",
			RefreshUuid: "",
			UserId:      userwprovider.UserId,
			DeviceId:    DeviceId,
		}
		session := &general.ActiveSession{
			IsActiveu_sid: false,
			Sessionstring: "",
		}

		err = auth.CompleteAuth(c, RedisClient, &urlParams, ad, appds, session)
		if err != nil {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Failed to generate token")
			return
		}
	}
	// else {
	// 	err := auth.SetAppblockToken(c, RedisClient, userwprovider.UserId, DeviceId)
	// 	if err != nil {
	// 		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Failed to generate token")
	// 		return
	// 	}

	// 	general.OutputHTML(c.Writer, c.Request, "static/src/pages/success-msg-general.html")
	// 	return
	// }

}

// GetEmailForDevice returns email id of user from db using device access token,

// swagger:route GET /device/get-email GetEmailForDevice
// End point get email using device acceess token
//
// security:
// - Key: []
// responses:
//
//	200: Response
//	400: Response
func GetEmailForDevice(c *gin.Context) {
	authCodeValues, _, Status, err := token.ProcessDeviceAccessToken(c)
	if err != nil {
		general.RespondHandler(c.Writer, false, Status, err.Error())
		log.Println(err)
		return
	}

	// strArr := strings.Split(authCodeValues.AppUserPermission, ",")

	// log.Println(strArr)

	db, err := pghandling.SetupDB()
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	defer sqlDB.Close()

	var Permissions []general.Permission

	db.First(&Permissions, "permission_name = ?", "Email")

	log.Println(&Permissions)

	for _, v := range authCodeValues.AppUserPermission {
		log.Println("PermissionId String before validate:", Permissions[0].PermissionId)
		if v == Permissions[0].PermissionId {
			log.Println("Permission Matched")
			var Users general.User
			db.First(&Users, "user_id = ?", authCodeValues.UserId)

			c.JSON(http.StatusOK, map[string]string{"email": Users.Email})
			log.Println(Users.Email)
			return
		}
	}

	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed")

}

func VerifyTokenAndGetEmailForDevice(c *gin.Context) {
	ClientId := c.Request.Header.Get("Client-Id")

	if len(ClientId) == 0 {

		err := errors.New("client id missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

	authCodeValues, _, Status, err := token.ProcessDeviceAccessToken(c)
	if err != nil {
		general.RespondHandler(c.Writer, false, Status, err.Error())
		log.Println(err)
		return
	}

	if authCodeValues.ClientId != ClientId {
		err := errors.New("client id not matched")
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}

	db, err := pghandling.SetupDB()
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	defer sqlDB.Close()

	var Permissions []general.Permission

	db.First(&Permissions, "permission_name = ?", "Email")

	for _, v := range authCodeValues.AppUserPermission {
		if v == Permissions[0].PermissionId {
			var Users general.User
			db.First(&Users, "user_id = ?", authCodeValues.UserId)

			c.JSON(http.StatusOK, map[string]string{"email": Users.Email})
			log.Println(Users.Email)
			return
		}
	}

	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed")

}

// GetUserDetails returns user details using access token,

// swagger:route GET /get-user-details GetUserDetails
// End point get user details using acceess token
//
// security:
// - Key: []
// responses:
//
//	200: Response
//	400: Response
func GetUserDetails(c *gin.Context) {

	userid, Status, err := token.GetUserIdFromAccessToken(c)

	if err != nil {
		general.RespondHandler(c.Writer, false, Status, err.Error())
		log.Println(err)
		return
	}

	db, err := pghandling.SetupDB()
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	defer sqlDB.Close()

	var UserDetails general.UserDetails

	userquery := `SELECT u.user_id, 
	array_agg(
	case p.provider when 1 then 'Password'
	when 2 then 'Google'
	when 3 then 'Twitter'
	when 4 then 'LinkedIn'
	else ''
	end    ) AS provider,
	u.user_name,
	u.full_name,
	u.email,
	u.address1,
	u.address2,
	u.phone 
	FROM   users u
	JOIN   user_providers p  ON p.user_id = u.user_id
	WHERE u.user_id = ?
	GROUP BY u.user_id`

	res := db.Raw(userquery, userid).Scan(&UserDetails)
	if res.Error != nil {
		log.Println(res.Error)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, res.Error)
		return
	}

	general.RespondHandler(c.Writer, true, http.StatusOK, UserDetails)

}

// GetUser returns user details using access token with client cred validation,

// swagger:route POST /get-user GetUser after client cred and token validation
// End point get user details using acceess token
//
// security:
// - Key: []
// responses:
//
//	200: Response
//	400: Response
func GetUser(c *gin.Context) {

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

	//extract token string from request
	tokenString := token.ExtractToken(c)
	if len(tokenString) == 0 {
		err := errors.New("access token not found")
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}

	// log.Println("token string:", tokenString)

	tokenval, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}
	claims, ok := tokenval.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token_type from token claim")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return
		}

		log.Println("token type:", tokenType)

		var authCodeValues *general.AuthCodeValues
		var ad *general.AccessDetails
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, ad, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPACCESS:
			authCodeValues, ad, Status, err = token.ProcessAppAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPBACCESS:
			authCodeValues, ad, Status, err = token.ProcessAppblockAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		default:
			err := errors.New("token is not allowed to access this")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return

		}

		log.Println("Authcodevalues:", authCodeValues)
		log.Println("Access Details:", ad)

		//if client id passing not matched with client id in token
		if ClientId != authCodeValues.ClientId {
			err := errors.New("client id not matched")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return
		}

		db, err := pghandling.SetupDB()
		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
			log.Println(err)
			return
		}

		defer sqlDB.Close()

		var Apps general.ShieldApp

		result := db.First(&Apps, "client_id = ?", authCodeValues.ClientId)

		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
			log.Println(result.Error)
			return
		}

		log.Println("Apss:", Apps)

		//if client secret passing not matched with client secret from db
		if ClientSecret != Apps.ClientSecret {
			err := errors.New("client secret not matched")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return
		}

		//Fetch permission name and id from db
		var Permissions []general.Permission

		starr := []string{"Email", "Username", "Name", "Address", "Phone"}
		db.Where("permission_name IN ?", starr).Find(&Permissions)
		perarray := make(map[string]string)

		for _, p := range Permissions {
			switch p.PermissionName {
			case "Email":
				perarray[p.PermissionName] = p.PermissionId
			case "Username":
				perarray[p.PermissionName] = p.PermissionId
			case "Name":
				perarray[p.PermissionName] = p.PermissionId
			case "Address":
				perarray[p.PermissionName] = p.PermissionId
			case "Phone":
				perarray[p.PermissionName] = p.PermissionId
			}
		}

		var UserDetails general.UserDetailsWithProvider

		userquery := `users.user_id, case user_providers.provider when 1 then 'Password'
		when 2 then 'Google'
		when 3 then 'Twitter'
		when 4 then 'LinkedIn'
		else ''
		end AS provider, users.user_name, users.full_name, users.email, users.address1, users.address2, users.phone`
		res := db.Model(&general.User{}).Order("users.created_at").Select(userquery).Joins("inner join user_providers on user_providers.user_id = users.user_id").Where("users.user_id = ?", authCodeValues.UserId).First(&UserDetails)

		if res.Error != nil {
			log.Println(res.Error)
			return
		}

		var UseDetailsSend general.UserDetailsWithProvider

		UseDetailsSend.UserId = UserDetails.UserId
		UseDetailsSend.Provider = UserDetails.Provider

		// strArr := strings.Split(authCodeValues.AppUserPermission, ",")

		for _, v := range authCodeValues.AppUserPermission {
			if v == perarray["Email"] {
				UseDetailsSend.Email = UserDetails.Email
			}
			if v == perarray["Username"] {
				UseDetailsSend.UserName = UserDetails.UserName
			}
			if v == perarray["Name"] {
				UseDetailsSend.FullName = UserDetails.FullName
			}
			if v == perarray["Address"] {
				UseDetailsSend.Address1 = UserDetails.Address1
				UseDetailsSend.Address2 = UserDetails.Address2
			}
			if v == perarray["Phone"] {
				UseDetailsSend.Phone = UserDetails.Phone
			}
		}

		general.RespondHandler(c.Writer, true, http.StatusOK, UseDetailsSend)

		return

	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

// GetUserId returns user id using user access token,

// swagger:route GET /get-user-id GetUserId
// End point get user id using user acceess token
//
// security:
// - Key: []
// responses:
//
//	200: Response
//	400: Response
func GetUserId(c *gin.Context) {

	// extracts UserId from token
	// userid, Status, err := token.GetUserIdFromUserAccessToken(c)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, Status, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	userid, Status, err := token.GetUserIdFromAccessToken(c)
	if err != nil {
		general.RespondHandler(c.Writer, false, Status, err.Error())
		log.Println(err)
		return
	}

	general.RespondHandler(c.Writer, true, http.StatusOK, map[string]string{"user_id": userid})

}

// GetUserId returns user id using user access token,

// swagger:route GET /get-user-id GetUserId
// End point get user id using user acceess token
//
// security:
// - Key: []
// responses:
//
//	200: Response
//	400: Response
func GetUid(c *gin.Context) {

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

	//extract token string from request
	tokenString := token.ExtractToken(c)
	if len(tokenString) == 0 {
		err := errors.New("access token not found")
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}

	// log.Println("token string:", tokenString)

	tokenval, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}
	claims, ok := tokenval.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token_type from token claim")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return
		}

		log.Println("token type:", tokenType)

		var authCodeValues *general.AuthCodeValues
		var ad *general.AccessDetails
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, ad, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPACCESS:
			authCodeValues, ad, Status, err = token.ProcessAppAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPBACCESS:
			authCodeValues, ad, Status, err = token.ProcessAppblockAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		default:
			err := errors.New("token is not allowed to access this")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return

		}

		log.Println("Authcodevalues:", authCodeValues)
		log.Println("Access Details:", ad)

		//if client id passing not matched with client id in token
		if ClientId != authCodeValues.ClientId {
			err := errors.New("client id not matched")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return
		}

		db, err := pghandling.SetupDB()
		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
			log.Println(err)
			return
		}

		defer sqlDB.Close()

		var Apps general.ShieldApp

		result := db.First(&Apps, "client_id = ?", authCodeValues.ClientId)

		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
			log.Println(result.Error)
			return
		}

		//if client secret passing not matched with client secret from db
		if ClientSecret != Apps.ClientSecret {
			err := errors.New("client secret not matched")
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
			log.Println(err)
			return
		}

		general.RespondHandler(c.Writer, true, http.StatusOK, map[string]string{"user_id": authCodeValues.UserId})
		return

	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

// SendVerificationEmail
func SendEmailVerification(verfiyEmailUserData mailer.VerifyEmailUserStruct) (string, error) {

	emailMessage, err := mailer.EmailVerficationMailer(verfiyEmailUserData)
	if err != nil {
		log.Println("Email Not Send")
		log.Println(err)
		return "Error", err
	}
	log.Println(emailMessage)
	return emailMessage, err

}

// VerifyUserEmail verfies the email address and return success/error message,

// swagger:route GET /verify-user-email VerifyUserEmail

// responses:
//
//	200: Response
//	400: Response
func VerfiyUserEmail(c *gin.Context) {

	// fetching request url params
	var urlParams general.AppUrlParams
	urlParams.ClientId = c.Query("client_id")
	urlParams.ResponseType = c.Query("response_type")
	urlParams.State = c.Query("state")
	urlParams.RedirectUri = c.Query("redirect_uri")

	//for app token exchange
	urlParams.GrantType = c.Query("grant_type")

	var appds *general.AppFromClientId
	if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(urlParams.ClientId, urlParams.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Not a valid client id")
			return
		}

		if err != nil {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Failed to fetch app details from db")
			return
		}

		if !Valid {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Given redirect uri not registered for this app")
			return
		}

	} else {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Access blocked: authorisation error: invalid_request")
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Something went wrong")
		return
	}

	var verifyData general.VerfiyEmailRequestEvent

	// err := c.BindJSON(&verifyData)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	verifyData.UserId = c.PostForm("user_id")
	verifyData.VerificationCode = c.PostForm("verification_code")

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Something went wrong")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Something went wrong")
		return
	}

	defer sqlDB.Close()

	var userData general.User

	result := db.Model(&general.User{}).Where("user_Id = ?", verifyData.UserId).Scan(&userData)
	if result.Error != nil {
		log.Println(result.Error)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Something went wrong")
		return
	}

	if userData.EmailVerified {
		log.Println("email already verified")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Email already verified")
		return
	}

	if verifyData.VerificationCode == userData.EmailVerificationCode {
		if time.Now().Before(userData.EmailVerificationExpiry) {

			result := db.Model(&general.User{}).Where("user_Id = ?", verifyData.UserId).Update("email_verified", true)
			if result.Error != nil {
				log.Println(result.Error)
				general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Something went wrong")
				return
			}

			// generate device id
			DeviceId := general.RandPassword(8)

			//Complete auth
			if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {

				ad := &general.AccessDetails{
					AccessUuid:  "",
					RefreshUuid: "",
					UserId:      userData.UserID,
					DeviceId:    DeviceId,
				}
				session := &general.ActiveSession{
					IsActiveu_sid: false,
					Sessionstring: "",
				}

				err = auth.CompleteAuth(c, RedisClient, &urlParams, ad, appds, session)
				if err != nil {
					general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Failed to generate token")
					return
				}
			}
			//  else {
			// 	err := auth.SetAppblockToken(c, RedisClient, userData.UserID, DeviceId)
			// 	if err != nil {
			// 		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Failed to generate token")
			// 		return
			// 	}

			// 	general.OutputHTML(c.Writer, c.Request, "static/src/pages/success-msg-general.html")
			// 	return
			// }
		} else {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "OTP expired")
		}

	} else {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/verify-email-error.html", "Incorrect OTP")
	}
}

func ResendUserEmailOTP(c *gin.Context) {

	var RequestBody general.ResendOTPRequestEvent

	err := c.BindJSON(&RequestBody)
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = general.ValidateStruct(&RequestBody)
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
		return
	}

	defer sqlDB.Close()

	var userData general.User

	result := db.Model(&general.User{}).Where("user_Id = ?", RequestBody.UserId).Scan(&userData)

	if result.Error != nil {
		log.Println(result.Error)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if result.RowsAffected <= 0 {
		log.Println("invalid user_id:", RequestBody.UserId)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Invalid User Id")
		return
	}

	if userData.EmailVerified {
		log.Println("Email already verified")
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Email already verified")
		return

	}

	log.Printf("userData %v", userData)

	VerificationCode := general.GetRandomCode(6)
	VerificationExpiry := time.Now().Add(time.Minute * 10)

	// result = db.Model(&general.User{UserId: RequestBody.UserId}).Select("email_verification_code, email_verification_expiry").Updates(general.User{EmailVerificationCode: VerificationCode, EmailVerificationExpiry: VerificationExpiry})
	result = db.Exec("UPDATE users SET email_verification_code = ?, email_verification_expiry = ?  WHERE user_id = ?", VerificationCode, VerificationExpiry, RequestBody.UserId)
	if result.Error != nil {
		log.Println(result.Error)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
		return
	}

	log.Println("result.RowsAffected:", result.RowsAffected)

	// Send emial verification =>
	go func() {

		// TODO optimize
		// split the verification code and adds a - at center
		vcSplit := strings.Split(VerificationCode, "")
		log.Println("vcSplit:", vcSplit)
		vcSlice := append(vcSplit[:3], vcSplit[2:]...)
		vcSlice[3] = " - "
		modifiedVerificationCode := strings.Join(vcSlice, "")

		verfiyEmailUserData := mailer.VerifyEmailUserStruct{
			UserName:         userData.UserName,
			Email:            userData.Email,
			VerificationCode: modifiedVerificationCode,
		}

		SendEmailVerification(verfiyEmailUserData)
	}()
	//  <= Send emial verification;

	general.RespondHandler(c.Writer, true, http.StatusOK, "success")

}

func UpdateUserProfile(c *gin.Context) {

	updateValue := make(map[string]interface{})

	var req general.UpdateUserProfileRequestEvent
	var ok bool
	isUpdate := false

	tokenString := token.ExtractToken(c)
	if len(tokenString) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "Access token not found")
		return
	}

	tok, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		return
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "Failed to fetch token_type from token claim")
			return
		}

		var ad *general.AccessDetails
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			_, ad, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPACCESS:
			_, ad, Status, err = token.ProcessAppAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPBACCESS:
			_, ad, Status, err = token.ProcessAppblockAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		default:
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "Token is not allowed to access this")
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return

		}

		db, err := pghandling.SetupDB()
		if err != nil {
			log.Println(err)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Println(err)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
			return
		}

		defer sqlDB.Close()

		if req.UserName, ok = c.GetPostForm("username"); ok {

			result := db.Where("user_name = ?", req.UserName).First(&general.User{})

			if result.Error != nil && !(strings.Contains(result.Error.Error(), "record not found")) {
				log.Println(result.Error)
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
				return
			}

			if result.RowsAffected > 0 {
				log.Println(result.Error)
				general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Username already taken")
				return
			}
			updateValue["user_name"] = req.UserName

			isUpdate = true
		}

		if req.FullName, ok = c.GetPostForm("fullname"); ok {
			isUpdate = true
			updateValue["full_name"] = req.FullName
		}
		if req.Address1, ok = c.GetPostForm("address1"); ok {
			isUpdate = true
			updateValue["address1"] = req.Address1
		}
		if req.Address2, ok = c.GetPostForm("address2"); ok {
			isUpdate = true
			updateValue["address2"] = req.Address2
		}
		if req.Phone, ok = c.GetPostForm("phone"); ok {
			isUpdate = true
			updateValue["phone"] = req.Phone
		}

		if !isUpdate {

			log.Println("Nothing to update")
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Nothing to update")
			return
		}

		result := db.Model(&general.User{}).Where("user_id = ?", ad.UserId).Updates(&updateValue)

		if result.Error != nil {
			log.Println(result.Error)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
			return
		}
		if result.RowsAffected <= 0 {
			log.Println("invalid user id:", ad.UserId)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Update fails")
			return
		}

		general.RespondHandler(c.Writer, true, http.StatusOK, "User profile updated")

		return
	}

	err = errors.New("failed to fetch claims from token")

	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())

}
