package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aidarkhanov/nanoid"
	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/kurrik/oauth1a"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/appblocks-hub/SHIELD/functions/appreg"
	"github.com/appblocks-hub/SHIELD/functions/general"
	"github.com/appblocks-hub/SHIELD/functions/mailer"
	"github.com/appblocks-hub/SHIELD/functions/pghandling"
	"github.com/appblocks-hub/SHIELD/functions/token"
	"github.com/gin-gonic/gin"
)

func init() {

}

// GetLogin validate user token and redirect to login page for invalid token
// swagger:route GET /login Login
// End point for device access code request
//
// security:
// - Bearer: []
// parameters:
//
//	AppUrlParams
//
// responses:
//
//	200: Response
//	400: Response
func GetLogin(c *gin.Context) {
	var urlParams general.AppUrlParams
	urlParams.ClientId = c.Query("client_id")
	urlParams.ResponseType = c.Query("response_type")
	urlParams.State = c.Query("state")
	urlParams.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	urlParams.GrantType = c.Query("grant_type")

	appds := &general.AppFromClientId{}

	if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(urlParams.ClientId, urlParams.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Not a valid client id")
			return
		}

		if err != nil {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Something went wrong")
			return
		}

		if !Valid {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Given redirect url not registered for this app")
			return
		}

	} else {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Access blocked: authorisation error: invalid_request")
		return
	}

	if len(c.Query("error")) != 0 {
		error_string := general.FetchErrorFormCode(c.Query("error"))
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", error_string)
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Something went wrong")
		return
	}

	// device auth flow or appblocks token generation flow

	ad, _, err := token.GetDetailsFromIDToken(c)

	if ad == nil {
		ad = &general.AccessDetails{
			AccessUuid:  "",
			RefreshUuid: "",
			UserId:      "",
			DeviceId:    "",
		}
	}

	session := &general.ActiveSession{
		IsActiveu_sid: false,
		Sessionstring: "",
	}

	if err != nil || len(ad.UserId) == 0 {

		//check for session and fetch user id from it
		session.Sessionstring, err = c.Cookie("u_sid")
		if err != nil {
			log.Println(err)
			log.Println("open login.html")
			general.OutputHTML(c.Writer, c.Request, "static/src/pages/login.html")
			return
		}

		valid, authcodevalues, err := ValidateKeyAndFetchDetails(RedisClient, session.Sessionstring, "u_sid")

		if err != nil {
			log.Println(err)
			log.Println("open login.html")
			general.OutputHTML(c.Writer, c.Request, "static/src/pages/login.html")
			return
		}

		if !valid {
			log.Println("session expired")
			log.Println("open login.html")
			general.OutputHTML(c.Writer, c.Request, "static/src/pages/login.html")
			return
		}

		ad.UserId = authcodevalues.UserId
		ad.DeviceId = authcodevalues.DeviceId

		session.IsActiveu_sid = true

		log.Println("Authcode values b4 in getLogin 1:", authcodevalues)
	}

	//Complete auth
	if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {

		err = CompleteAuth(c, RedisClient, &urlParams, ad, appds, session)
		if err != nil {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Failed to generate token")
			return
		}
	}
	// else {

	// 	err := SetAppblockToken(c, RedisClient, ad.UserId, ad.DeviceId)
	// 	if err != nil {
	// 		general.OutputHTMLTemplate(c.Writer, "static/src/pages/login-error.html", "Failed to generate token")
	// 		return
	// 	}

	// 	general.OutputHTML(c.Writer, c.Request, "static/src/pages/success-msg-general.html")
	// 	return
	// }
}

func SetAppblockToken(c *gin.Context, RedisClient *redis.Client, UserId, DeviceId string) error {
	tp, err := token.CreateTokenPair(RedisClient, UserId, DeviceId)

	if err != nil {
		log.Println(err)
		return err
	}

	expStr := general.Envs["ACCESS_TOKEN_EXPIRY"]

	if len(expStr) == 0 {
		log.Println("using default expiry for access token")
		expStr = "604800" // 7 days
	}

	expInt, err := strconv.Atoi(expStr)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return err
	}

	rtexpStr := general.Envs["REFRESH_TOKEN_EXPIRY"]

	if len(rtexpStr) == 0 {
		log.Println("using default expiry for refresh token")
		rtexpStr = "604800"
	}

	rtexpInt, err := strconv.Atoi(rtexpStr)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return err
	}

	c.SetCookie("u_at", tp.AccessToken, expInt, "/", ".appblocks.com", false, false)
	c.SetCookie("u_rt", tp.RefreshToken, rtexpInt, "/", ".appblocks.com", false, false)

	return nil
}

func SetIDToken(c *gin.Context, RedisClient *redis.Client, UserId, DeviceId string) error {

	expStr := general.Envs["ID_TOKEN_EXPIRY"]

	if len(expStr) == 0 {
		log.Println("using default expiry for id token")
		expStr = "604800"
	}

	expInt64, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return err
	}

	idtexpires := time.Now().Add(time.Second * time.Duration(expInt64)).Unix()

	idt, err := token.CreateToken(general.IDTOKEN, idtexpires)
	if err != nil {
		log.Println(err)
		return err
	}

	idtd := &general.UserAccessDetails{}
	idtd.PairTokenUuid = ""
	idtd.UserId = UserId
	idtd.DeviceId = DeviceId

	err = token.SaveToken(idt.Uuid+":"+DeviceId+":"+UserId, idt.Expires, RedisClient, idtd)
	if err != nil {
		log.Println(err)
		return err
	}

	expInt, err := strconv.Atoi(expStr)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return err
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     general.Envs["ID_TOKEN_NAME"],
		Value:    idt.Token,
		Domain:   ".appblocks.com",
		Path:     "/",
		MaxAge:   expInt,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

	return nil
}

func CompleteAuth(c *gin.Context, RedisClient *redis.Client, urlParams *general.AppUrlParams, ad *general.AccessDetails, appds *general.AppFromClientId, session *general.ActiveSession) error {

	// Ckeck for permissions already set for this user
	PermissionAlreadySet, err := appreg.IsPermissionSetforUser(ad.UserId, appds.AppId)

	if err != nil {
		log.Println(err)
		return err
	}

	if !PermissionAlreadySet {

		appPermissions, err := appreg.FetchAppPermission(appds.AppId)
		if err != nil {
			err := errors.New("something went wrong")
			return err
		}

		var AppUserPermission []general.AppUserPermission

		for _, appperm := range *appPermissions {
			AppUserPermission = append(AppUserPermission, general.AppUserPermission{
				UserId:       ad.UserId,
				AppId:        appperm.AppId,
				PermissionId: appperm.PermissionId,
			})

		}

		err = appreg.CreateAppUserPermissionsDBEntry(&AppUserPermission)

		if err != nil {
			err := errors.New("something went wrong")
			return err
		}

	}

	// if urlParams.ResponseType == "code" && appds.AppType != 3 && session.IsActiveu_sid {

	// 	err := SetAppblockToken(c, RedisClient, ad.UserId, ad.DeviceId)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	err = GenerateAndRedirectAuthCode(c, RedisClient, urlParams, ad, appds.AppId, urlParams.ResponseType)
	if err != nil {
		log.Println(err)
		return err
	}

	if session.IsActiveu_sid {

		codeuuid, err := base64.StdEncoding.DecodeString(session.Sessionstring)

		if err != nil {
			log.Println(err)
			return err
		}

		_, err = token.DeleteToken(string(codeuuid)+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
		if err != nil {
			log.Println(err)
			return err
		}

		// invalidate cookie
		c.SetCookie("u_sid", "", -1, "/", "", true, true)
	}

	return nil

	// // creating user session

	// sesssionuuid := uuid.New().String()
	// sessionstring := base64.StdEncoding.EncodeToString([]byte(sesssionuuid))
	// log.Println(sessionstring)

	// Expires := time.Now().Add(time.Minute * 10)

	// //for token exchange
	// var v []byte
	// if len(urlParams.ClientId) != 0 && urlParams.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
	// 	v, err = json.Marshal(&general.AuthCodeValues{
	// 		UserId:            ad.UserId,
	// 		ClientId:          general.APPBLOCKCLIENTID,
	// 		AppSname:          "",
	// 		AppUserPermission: nil,
	// 		CodeName:          "u_sid",
	// 		DeviceId:          ad.DeviceId,
	// 	})
	// 	if err != nil {
	// 		log.Println(err)
	// 		return err
	// 	}
	// } else {
	// 	v, err = json.Marshal(&general.AuthCodeValues{
	// 		UserId:            ad.UserId,
	// 		ClientId:          urlParams.ClientId,
	// 		AppSname:          "",
	// 		AppUserPermission: nil,
	// 		CodeName:          "u_sid",
	// 		DeviceId:          ad.DeviceId,
	// 	})
	// 	if err != nil {
	// 		log.Println(err)
	// 		return err
	// 	}
	// }

	// err = RedisClient.Set(sesssionuuid+":"+ad.DeviceId+":"+ad.UserId, v, time.Until(Expires)).Err()
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	// c.SetCookie("u_sid", sessionstring, 600, "/", "", false, true)

	// q := url.Values{}
	// if len(urlParams.ClientId) != 0 && ((urlParams.ResponseType == "code" || urlParams.ResponseType == "device_code") && len(urlParams.State) != 0 && len(urlParams.RedirectUri) != 0) {
	// 	q.Set("client_id", urlParams.ClientId)
	// 	q.Set("response_type", urlParams.ResponseType)
	// 	q.Set("state", urlParams.State)
	// 	q.Set("redirect_uri", urlParams.RedirectUri)
	// }
	// if len(urlParams.ClientId) != 0 && urlParams.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
	// 	q.Set("grant_type", urlParams.GrantType)
	// 	q.Set("client_id", urlParams.ClientId)
	// }

	// if appds.AppType != 3 && appds.AppType != 4 {

	// 	log.Println("open authorize-appname-permission.html")
	// 	q.Set("app_sname", appds.AppSname)
	// 	location := url.URL{Path: "/authorize-appname-permission", RawQuery: q.Encode()}
	// 	c.Redirect(http.StatusFound, location.RequestURI())
	// 	return nil

	// } else {
	// 	log.Println("open permission")
	// 	q.Set("app_sname", appds.AppSname)
	// 	location := url.URL{Path: "/authorize-appname", RawQuery: q.Encode()}
	// 	c.Redirect(http.StatusFound, location.RequestURI())
	// 	return nil

	// }

}

func GenerateAndRedirectAuthCode(c *gin.Context, RedisClient *redis.Client, urlParams *general.AppUrlParams, ad *general.AccessDetails, appid string, codename string) (err error) {

	//Fetching AppUserPermissions
	appUserPermissions, err := appreg.FetchAppUserPermissions(ad.UserId, appid)
	if err != nil {
		log.Println(err)
		return err
	}

	// appUserPermissionString := ""
	// for _, appperm := range *appUserPermissions {
	// 	appUserPermissionString += appperm.PermissionId.String() + ","
	// }

	// if last := len(appUserPermissionString) - 1; last >= 0 {
	// 	appUserPermissionString = appUserPermissionString[:last]
	// }

	var appUserPermissionSlice []string
	for _, appperm := range *appUserPermissions {
		appUserPermissionSlice = append(appUserPermissionSlice, appperm.PermissionId)
	}

	codeuuid := uuid.New().String()
	codestring := base64.StdEncoding.EncodeToString([]byte(codeuuid))
	log.Println(codestring)

	u, err := url.Parse(urlParams.RedirectUri)
	if err != nil {
		log.Println("error parse:", err)
		return err
	}

	qp := u.Query()

	qp.Set(codename, codestring)
	qp.Set("state", urlParams.State)

	u.RawQuery = qp.Encode()

	log.Println(u.String())

	Expires := time.Now().Add(time.Minute * 10)

	v, err := json.Marshal(&general.AuthCodeValues{
		UserId:            ad.UserId,
		ClientId:          urlParams.ClientId,
		AppSname:          "",
		AppUserPermission: appUserPermissionSlice,
		CodeName:          codename,
		DeviceId:          ad.DeviceId,
	})
	if err != nil {

		return err
	}

	err = RedisClient.Set(codeuuid+":"+ad.DeviceId+":"+ad.UserId, v, time.Until(Expires)).Err()
	if err != nil {
		return err
	}

	c.Writer.Header().Set("Location", u.String())
	// c.Writer.Header().Set("Access-Control-Allow-Origin", urlParams.RedirectUri)
	// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	// c.Writer.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Origin, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	// c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	//general.RespondHandler(c.Writer, true, http.StatusFound, "success")

	//temp
	c.Writer.WriteHeader(http.StatusFound)

	return
}

// CreateShieldAppUserPermissions assign device app permissions to user
// swagger:route POST /allow-permissions AllowPermissions
// To Allow permissions of device app
//
// security:
// - Key: []
// parameters:
//
//	AppUrlParams
//
// responses:
//
//	200: Response
//	400: Response
func CreateShieldAppUserPermissions(c *gin.Context) {

	// fetching request url params
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

	var appds *general.AppFromClientId
	if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(urlParams.ClientId, urlParams.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Not a valid client id")
			return
		}

		if err != nil {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Failed to fetch app details from db")
			return
		}

		if !Valid {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Given redirect uri not registered for this app")
			return
		}

	} else {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Access blocked: authorisation error: invalid_request")
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Something went wrong")
		return
	}

	if len(urlParams.ClientId) != 0 && (urlParams.ResponseType == "code" || urlParams.ResponseType == "device_code") && len(urlParams.State) != 0 && len(urlParams.RedirectUri) != 0 {
		ad, _, err := token.GetDetailsFromIDToken(c)

		if ad == nil {
			ad = &general.AccessDetails{
				AccessUuid:  "",
				RefreshUuid: "",
				UserId:      "",
				DeviceId:    "",
			}
		}

		IsActiveu_sid := false
		Sessionstring := ""

		if err != nil || len(ad.UserId) == 0 {
			//check for session and fetch user id from it
			Sessionstring, err := c.Cookie("u_sid")
			if err != nil {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			valid, authcodevalues, err := ValidateKeyAndFetchDetails(RedisClient, Sessionstring, "u_sid")

			if err != nil {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			if !valid {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			ad.UserId = authcodevalues.UserId
			ad.DeviceId = authcodevalues.DeviceId

			IsActiveu_sid = true
		}

		appPermissions, err := appreg.FetchAppPermission(appds.AppId)
		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Something went wrong")
			return
		}

		var AppUserPermission []general.AppUserPermission

		for _, appperm := range *appPermissions {
			AppUserPermission = append(AppUserPermission, general.AppUserPermission{
				UserId:       ad.UserId,
				AppId:        appperm.AppId,
				PermissionId: appperm.PermissionId,
			})

		}

		err = appreg.CreateAppUserPermissionsDBEntry(&AppUserPermission)

		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Something went wrong")
			return
		}

		// if urlParams.ResponseType == "code" && IsActiveu_sid {

		// 	tp, err := token.CreateTokenPair(RedisClient, ad.UserId, ad.DeviceId)
		// 	if err != nil {
		// 		log.Println(err)
		// 		general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Failed to generate token")
		// 		return
		// 	}

		// 	expStr := general.Envs["ACCESS_TOKEN_EXPIRY"]

		// 	log.Println("expStr:", expStr)

		// 	if len(expStr) == 0 {
		// 		log.Println("using default expiry for access token")
		// 		expStr = "900"
		// 	}

		// 	expInt, err := strconv.Atoi(expStr)
		// 	if err != nil {
		// 		log.Fatalf("Error loading .env file: %v", err)
		// 		return
		// 	}

		// 	rtexpStr := general.Envs["REFRESH_TOKEN_EXPIRY"]

		// 	if len(rtexpStr) == 0 {
		// 		log.Println("using default expiry for refresh token")
		// 		rtexpStr = "604800"
		// 	}

		// 	rtexpInt, err := strconv.Atoi(rtexpStr)
		// 	if err != nil {
		// 		log.Fatalf("Error loading .env file: %v", err)
		// 		return
		// 	}

		// 	c.SetCookie("u_at", tp.AccessToken, expInt, "/", ".appblocks.com", false, false)
		// 	c.SetCookie("u_rt", tp.RefreshToken, rtexpInt, "/", ".appblocks.com", false, false)

		// }

		err = GenerateAndRedirectAuthCode(c, RedisClient, &urlParams, ad, appds.AppId, urlParams.ResponseType)
		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Something went wrong")
			return
		}

		if IsActiveu_sid {
			codeuuid, err := base64.StdEncoding.DecodeString(Sessionstring)

			if err != nil {
				log.Println(err)
				general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Something went wrong")
				return
			}

			_, err = token.DeleteToken(string(codeuuid)+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
			if err != nil {
				log.Println(err)
				general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Something went wrong")
				return
			}

			// invalidate cookie
			c.SetCookie("u_sid", "", -1, "/", "", true, true)
		}

		return

	}

	general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-permission-error.html", "Bad request")

}

// GetAllAppPermissionsForAllow is used to fetch Permission Details of an app from its client id

// swagger:route GET /get-app-permissions-for-allow GetAllAppPermissionsForAllow
// End point to get app permissions for aouthorize-appname page
//
// security:
// - Key: []
// parameters:
//   - name: client_id
//     in: query
//     description: client id
//     type: string
//
// responses:
//
//	200: Response
//	400: Response
func GetAllAppPermissionsForAllow(c *gin.Context) {
	clientid := c.Query("client_id")
	if len(clientid) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "client_id missing in request")
		log.Println("client_id missing in request")
		return
	}

	// Fetch app details from DB
	appds, err := appreg.GetAppFromClientId(clientid)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
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

	var permissions []general.AppPermissionsForAllow

	db.Model(&general.AppPermission{}).Select("app_permissions.permission_id, permissions.permission_name, permissions.description, app_permissions.mandatory").Joins("inner join permissions on app_permissions.permission_id = permissions.permission_id").Where("app_permissions.app_id = ?", appds.AppId).Scan(&permissions)

	//c.JSON(http.StatusOK, &permissions)
	general.RespondHandler(c.Writer, true, http.StatusOK, &permissions)

}

// OpenApp  initiate the app open/install flow,
// this will check that the user already accepted the permission for this app, if not, it will redirect to allow page
// else, it will generate an app token pair

// swagger:route GET /auth/app/open OpenApp
// End point to start app open/install flow
//
// security:
// - Key: []
// parameters:
//
//	AppUrlParams
//
// responses:
//
//	200: Response
//	400: Response
func OpenApp(c *gin.Context) {
	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis client handling error")
		return
	}

	log.Println("inside app open")
	var urlParams general.AppUrlParams
	urlParams.ClientId = c.Query("client_id")
	urlParams.GrantType = c.Query("grant_type")

	log.Println(urlParams)

	if len(urlParams.ClientId) != 0 && urlParams.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {

		ad, _, err := token.GetDetailsFromIDToken(c)
		if ad == nil {
			ad = &general.AccessDetails{
				AccessUuid:  "",
				RefreshUuid: "",
				UserId:      "",
				DeviceId:    "",
			}
		}

		if err != nil || len(ad.UserId) == 0 {
			// u, err := url.Parse("/auth/app/open" + "?grant_type=" + urlParams.GrantType + "&client_id=" + urlParams.ClientId)
			// if err != nil {
			// 	log.Println("error parse:", err)
			// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed parse redirect url:"+err.Error())
			// 	return
			// }

			u, err := url.Parse("/login" + "?grant_type=" + urlParams.GrantType + "&client_id=" + urlParams.ClientId)
			if err != nil {
				log.Println("error parse:", err)
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed parse redirect url")
				return
			}

			c.Writer.Header().Set("Location", u.String())
			c.Writer.WriteHeader(http.StatusFound)

			log.Println("before open app login redirect")
			return

			// log.Println("open login.html")
			// general.OutputHTML(c.Writer, c.Request, "static/src/pages/login.html")
			// return
		}

		// Fetch app details from DB
		appds, err := appreg.GetAppFromClientId(urlParams.ClientId)
		if err != nil {
			if err.Error() == "record not found" {
				general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
				log.Println(err)
				return
			}
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
			log.Println(err)
			return
		}

		// application validations can be done here
		log.Println("appdetails:", appds)

		// Ckeck for permissions already set for this user
		PermissionAlreadySet, err := appreg.IsPermissionSetforUser(ad.UserId, appds.AppId)

		if PermissionAlreadySet {

			err = GenerateAndRedirectAppToken(c, RedisClient, appds, ad)

			if err != nil {
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
				log.Println(err)
				return
			}

			return
		}

		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
			return
		}

		log.Println("open permission")

		general.OutputHTML(c.Writer, c.Request, "static/src/pages/authorize-appname.html")
		return

	}

	general.RespondHandler(c.Writer, false, http.StatusBadRequest, "invalid parameters")

}

// CreateDeviceAccessToken  is to handle device app access token get request,
// this will validates grant type and device code
// it will creates a device access token using token.CreateDeviceAccessToken

// swagger:route GET /auth/device/get-token CreateDeviceAccessToken
// End point to create device access token
//
// security:
// - Key: []
// parameters:
//   - name: grant_type
//     in: query
//     description: grant type
//     type: string
//   - name: device_code
//     in: query
//     description: device code
//     type: string
//
// responses:
//
//	200: Response
//	400: Response
func CreateDeviceAccessToken(c *gin.Context) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis client handling error")
		return
	}

	GrantType := c.Query("grant_type")
	DeviceCode := c.Query("device_code")

	if GrantType != "urn:ietf:params:oauth:grant-type:device_code" || len(DeviceCode) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "failed")
		log.Println("failed")
		return
	}

	valid, authcodevalues, err := ValidateKeyAndFetchDetails(RedisClient, DeviceCode, "device_code")

	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if !valid {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed")
		log.Println("failed")
		return
	}

	expires := time.Now().Add(time.Hour * 24 * 30).Unix()
	td, err := token.CreateToken(general.DEVICEACCESS, expires)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	err = token.SaveToken(td.Uuid+":"+authcodevalues.DeviceId+":"+authcodevalues.UserId, td.Expires, RedisClient, authcodevalues)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	// invalidating code from redis
	codeuuid, err := base64.StdEncoding.DecodeString(DeviceCode)

	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed to generate token:"+err.Error())
		return
	}

	_, err = token.DeleteToken(string(codeuuid)+":"+authcodevalues.DeviceId+":"+authcodevalues.UserId, RedisClient)
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed to generate token:"+err.Error())
		return
	}

	general.RespondHandler(c.Writer, true, http.StatusOK, map[string]string{"access_token": td.Token, "token_type": "Bearer", "expires_in": strconv.FormatInt(td.Expires, 10)})
	//c.JSON(http.StatusOK, map[string]string{"access_token": td.Token, "token_type": "Bearer", "expires_in": strconv.FormatInt(td.Expires, 10)})

}

func CreateAppblockTokens(c *gin.Context) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis client handling error")
		return
	}

	GrantType := c.Query("grant_type")
	Code := c.Query("code")

	if GrantType != "authorization_code" || len(Code) == 0 /*|| len(RedirectUri) == 0 */ {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "failed")
		log.Println("failed in url values")
		return
	}

	valid, authcodevalues, err := ValidateKeyAndFetchDetails(RedisClient, Code, "code")

	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if !valid {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed")
		log.Println("failed in valid")
		return
	}

	// check if app type is 2 or 3; if 2: internal app, if 3: external app
	// Fetch app details from DB
	appds, err := appreg.GetAppFromClientId(authcodevalues.ClientId)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	var isAppblock bool

	switch appds.AppType {
	case 2:
		isAppblock = true
	case 3, 4:
		isAppblock = false
	default:

		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "app has no permission to access this")
		log.Println("app has no permission to access this")
		return
	}

	// create and set id token
	if err = SetIDToken(c, RedisClient, authcodevalues.UserId, authcodevalues.DeviceId); err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	td, err := token.GenerateAppTokenPair(RedisClient, authcodevalues, isAppblock)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	// invalidating code from redis
	codeuuid, err := base64.StdEncoding.DecodeString(Code)

	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed to generate token:"+err.Error())
		return
	}

	_, err = token.DeleteToken(string(codeuuid)+":"+authcodevalues.DeviceId+":"+authcodevalues.UserId, RedisClient)
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed to generate token:"+err.Error())
		return
	}

	general.RespondHandler(c.Writer, true, http.StatusOK,
		map[string]string{"ab_at": td.AccessToken, "token_type": "Bearer", "expires_in": strconv.FormatInt(td.AtExpires, 10), "ab_rt": td.RefreshToken})
	//c.JSON(http.StatusOK, map[string]string{"access_token": td.Token, "token_type": "Bearer", "expires_in": strconv.FormatInt(td.Expires, 10)})

}

func ValidateKeyAndFetchDetails(RedisClient *redis.Client, accesscode, codename string) (valid bool, authcodevalues *general.AuthCodeValues, err error) {

	codeuuid, err := base64.StdEncoding.DecodeString(accesscode)

	//log.Println(codeuuid)

	if err != nil {
		log.Println("error decoding base64:", err)
		return false, nil, err
	}

	authcodevalues, err = token.FetchDeviceAuthCode(string(codeuuid), RedisClient)
	if err != nil {
		return false, nil, err
	}

	//validate CodeName
	if authcodevalues.CodeName != codename {
		log.Println("code name not matched")
		err := errors.New("invalid code")
		return false, nil, err
	}
	return true, authcodevalues, nil
}

// CreateAppUserPermissions assign app permissions to user
// swagger:route POST /allow-app-permissions AllowAppPermissions
// To Allow permissions of app
//
// security:
// - Key: []
// parameters:
//
//	AppUrlParams
//	PermissionIds
//
// responses:
//
//	200: Response
//	400: Response
func CreateAppUserPermissions(c *gin.Context) {

	var urlParams general.AppUrlParams
	// err := c.BindJSON(&urlParams)
	// if err != nil {
	// 	c.IndentedJSON(http.StatusBadRequest, "Failed")
	// 	panic(err)
	// }

	urlParams.ClientId = c.Query("client_id")
	urlParams.ResponseType = c.Query("response_type")
	urlParams.State = c.Query("state")
	urlParams.RedirectUri = c.Query("redirect_uri")

	//for app token exchange
	urlParams.GrantType = c.Query("grant_type")

	log.Println("Url params:", urlParams)

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

	var appds *general.AppFromClientId
	if len(urlParams.ClientId) != 0 || len(urlParams.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(urlParams.ClientId, urlParams.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Not a valid client id")
			return
		}

		if err != nil {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Failed to fetch app details from db")
			return
		}

		if !Valid {
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Given redirect uri not registered for this app")
			return
		}

	} else {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Access blocked: authorisation error: invalid_request")
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Something went wrong")
		return
	}

	if len(urlParams.ClientId) != 0 && (urlParams.ResponseType == "code" /*|| urlParams.ResponseType == "device_code"*/) && len(urlParams.State) != 0 && len(urlParams.RedirectUri) != 0 {
		ad, _, err := token.GetDetailsFromIDToken(c)

		if ad == nil {
			ad = &general.AccessDetails{
				AccessUuid:  "",
				RefreshUuid: "",
				UserId:      "",
				DeviceId:    "",
			}
		}

		IsActiveu_sid := false
		Sessionstring := ""

		if err != nil || len(ad.UserId) == 0 {
			//check for session and fetch user id from it
			Sessionstring, err := c.Cookie("u_sid")
			if err != nil {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			valid, authcodevalues, err := ValidateKeyAndFetchDetails(RedisClient, Sessionstring, "u_sid")

			if err != nil {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			if !valid {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			ad.UserId = authcodevalues.UserId
			ad.DeviceId = authcodevalues.DeviceId

			IsActiveu_sid = true

		}

		//Fetch app permissions from DB
		appPermissions, err := appreg.FetchAppPermission(appds.AppId)
		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Something went wrong")
			return
		}

		var AppUserPermission []general.AppUserPermission

		for _, appperm := range *appPermissions {
			if appperm.Mandatory {
				AppUserPermission = append(AppUserPermission, general.AppUserPermission{
					UserId:       ad.UserId,
					AppId:        appperm.AppId,
					PermissionId: appperm.PermissionId,
				})

			} else {
				oppperm := c.PostForm(appperm.PermissionId)
				if oppperm == "on" {
					AppUserPermission = append(AppUserPermission, general.AppUserPermission{
						UserId:       ad.UserId,
						AppId:        appperm.AppId,
						PermissionId: appperm.PermissionId,
					})

				}

			}
		}

		err = appreg.CreateAppUserPermissionsDBEntry(&AppUserPermission)

		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Something went wrong")
			return
		}

		err = GenerateAndRedirectAuthCode(c, RedisClient, &urlParams, ad, appds.AppId, urlParams.ResponseType)
		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Something went wrong")
			return
		}

		if IsActiveu_sid {
			codeuuid, err := base64.StdEncoding.DecodeString(Sessionstring)

			if err != nil {
				log.Println(err)
				general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Something went wrong")
				return
			}

			_, err = token.DeleteToken(string(codeuuid)+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
			if err != nil {
				log.Println(err)
				general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Something went wrong")
				return
			}

			// invalidate cookie
			c.SetCookie("u_sid", "", -1, "/", "", true, true)
		}

		return
	}

	// if urlParams.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" && len(urlParams.ClientId) != 0 {
	// 	// app auth flow
	// 	log.Println("app auth")

	// 	ad, _, err := token.GetDetailsFromUserAccessToken(c)

	// 	if ad == nil {
	// 		ad = &general.AccessDetails{
	// 			AccessUuid:  "",
	// 			RefreshUuid: "",
	// 			UserId:      "",
	// 			DeviceId:    "",
	// 		}
	// 	}

	// 	IsActiveu_sid := false
	// 	Sessionstring := ""

	// 	if err != nil || len(ad.UserId) == 0 {
	// 		//check for session and fetch user id from it
	// 		Sessionstring, err := c.Cookie("u_sid")
	// 		if err != nil {
	// 			log.Println(err)
	// 			log.Println("open login.html")
	// 			general.OutputHTML(c.Writer, c.Request, "static/src/pages/login.html")
	// 			return
	// 		}

	// 		valid, authcodevalues, err := ValidateKeyAndFetchDetails(RedisClient, Sessionstring, "u_sid")

	// 		if err != nil {
	// 			log.Println(err)
	// 			log.Println("open login.html")
	// 			general.OutputHTML(c.Writer, c.Request, "static/src/pages/login.html")
	// 			return
	// 		}

	// 		if !valid {
	// 			log.Println("session expired")
	// 			log.Println("open login.html")
	// 			general.OutputHTML(c.Writer, c.Request, "static/src/pages/login.html")
	// 			return
	// 		}

	// 		ad.UserId = authcodevalues.UserId
	// 		ad.DeviceId = authcodevalues.DeviceId

	// 		IsActiveu_sid = true

	// 		log.Println("Authcode values b4 in CreateDeviceAppUserPermissions 1:", authcodevalues)
	// 	}

	// 	//Fetch app details from DB
	// 	appds, err := appreg.GetAppFromClientId(urlParams.ClientId)
	// 	if err != nil {
	// 		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 		log.Println(err)
	// 		return
	// 	}

	// 	log.Println("app details after GetAppFromClientId:", appds)

	// 	//Fetch app permissions from DB
	// 	appPermissions, err := appreg.FetchAppPermission(appds.AppId)
	// 	if err != nil {
	// 		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 		return
	// 	}

	// 	log.Println("app permissions:", appPermissions)

	// 	// var optionalPermissions []general.PermissionIds
	// 	// err = c.BindJSON(&optionalPermissions)
	// 	// if err != nil {
	// 	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	// 	log.Println(err)
	// 	// 	return
	// 	// }

	// 	// log.Println("optional permission ids from json:", &optionalPermissions)

	// 	var AppUserPermission []general.AppUserPermission

	// 	// for _, appperm := range *appPermissions {
	// 	// 	if appperm.Mandatory {
	// 	// 		AppUserPermission = append(AppUserPermission, general.AppUserPermission{
	// 	// 			UserId:       userid,
	// 	// 			AppId:        appperm.AppId,
	// 	// 			PermissionId: appperm.PermissionId,
	// 	// 		})

	// 	// 	} else {
	// 	// 		for _, opperm := range optionalPermissions {

	// 	// 			if opperm.PermissionId == appperm.PermissionId.String() {
	// 	// 				AppUserPermission = append(AppUserPermission, general.AppUserPermission{
	// 	// 					UserId:       userid,
	// 	// 					AppId:        appperm.AppId,
	// 	// 					PermissionId: appperm.PermissionId,
	// 	// 				})

	// 	// 			}

	// 	// 		}

	// 	// 	}
	// 	// }

	// 	for _, appperm := range *appPermissions {
	// 		if appperm.Mandatory {
	// 			AppUserPermission = append(AppUserPermission, general.AppUserPermission{
	// 				UserId:       ad.UserId,
	// 				AppId:        appperm.AppId,
	// 				PermissionId: appperm.PermissionId,
	// 			})

	// 		} else {
	// 			oppperm := c.PostForm(appperm.PermissionId)
	// 			if oppperm == "on" {
	// 				AppUserPermission = append(AppUserPermission, general.AppUserPermission{
	// 					UserId:       ad.UserId,
	// 					AppId:        appperm.AppId,
	// 					PermissionId: appperm.PermissionId,
	// 				})

	// 			}

	// 		}
	// 	}

	// 	//log.Println(AppUserPermission)

	// 	err = appreg.CreateAppUserPermissionsDBEntry(&AppUserPermission)

	// 	if err != nil {
	// 		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 		log.Println(err)
	// 		return
	// 	}

	// 	err = GenerateAndRedirectAppToken(c, RedisClient, appds, ad)
	// 	if err != nil {
	// 		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 		log.Println(err)
	// 		return
	// 	}

	// 	if IsActiveu_sid {
	// 		codeuuid, err := base64.StdEncoding.DecodeString(Sessionstring)

	// 		if err != nil {
	// 			log.Println(err)
	// 			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed to generate token:"+err.Error())
	// 			return
	// 		}

	// 		_, err = token.DeleteToken(string(codeuuid)+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
	// 		if err != nil {
	// 			log.Println(err)
	// 			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Failed to generate token:"+err.Error())
	// 			return
	// 		}

	// 		// invalidate cookie
	// 		c.SetCookie("u_sid", "", -1, "/", "", false, true)
	// 	}

	// 	return

	// }

	general.OutputHTMLTemplate(c.Writer, "static/src/pages/authorize-appname-error.html", "Bad request")

}

func GenerateAndRedirectAppToken(c *gin.Context, RedisClient *redis.Client, appDs *general.AppFromClientId, ad *general.AccessDetails) (err error) {

	//Fetching AppUserPermissions
	appUserPermissions, err := appreg.FetchAppUserPermissions(ad.UserId, appDs.AppId)
	if err != nil {
		log.Println(err)
		return err
	}

	// appUserPermissionString := ""
	// for _, appperm := range *appUserPermissions {
	// 	appUserPermissionString += appperm.PermissionId.String() + ","
	// }

	// if last := len(appUserPermissionString) - 1; last >= 0 {
	// 	appUserPermissionString = appUserPermissionString[:last]
	// }

	var appUserPermissionSlice []string
	for _, appperm := range *appUserPermissions {
		appUserPermissionSlice = append(appUserPermissionSlice, appperm.PermissionId)
	}

	authcodevalues := general.AuthCodeValues{
		UserId:            ad.UserId,
		ClientId:          appDs.ClientId,
		AppSname:          appDs.AppSname,
		AppUserPermission: appUserPermissionSlice,
		DeviceId:          ad.DeviceId,
	}

	td, err := token.GenerateAppTokenPair(RedisClient, &authcodevalues, false)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

	// expStr := general.Envs["APPACCESS_TOKEN_EXPIRY"]

	// log.Println("expStr:", expStr)

	// if len(expStr) == 0 {
	// 	log.Println("using default expiry for app access token")
	// 	expStr = "900"
	// }

	// expInt, err := strconv.Atoi(expStr)
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// 	return
	// }

	// rtexpStr := general.Envs["APPREFRESH_TOKEN_EXPIRY"]

	// if len(rtexpStr) == 0 {
	// 	log.Println("using default expiry for app refresh token")
	// 	rtexpStr = "604800"
	// }

	// rtexpInt, err := strconv.Atoi(rtexpStr)
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// 	return
	// }
	// c.SetCookie("a_at_"+appDs.AppSname, td.AccessToken, expInt, "/"+appDs.AppSname, ".appblocks.com", false, false)
	// c.SetCookie("a_rt_"+appDs.AppSname, td.RefreshToken, rtexpInt, "/"+appDs.AppSname, ".appblocks.com", false, false)
	general.RespondHandler(c.Writer, true, http.StatusOK, td)
	// general.RespondHandler(c.Writer, true, http.StatusOK, "success")

	// log.Println("before code redirect")
	return
}

// GetSignup page
func GetSignup(c *gin.Context) {

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

	if len(c.Query("error")) != 0 {
		error_string := general.FetchErrorFormCode(c.Query("error"))
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/signup-error.html", error_string)
		return
	}

	general.OutputHTML(c.Writer, c.Request, "static/src/pages/signup.html")
}

// GetVerifyEmail page
func GetVerifyEmail(c *gin.Context) {
	general.OutputHTML(c.Writer, c.Request, "static/src/pages/verify-email.html")
}

func GoogleSignup(c *gin.Context) {
	var sld general.SocialLoginDetails
	sld.ClientId = c.Query("client_id")
	sld.ResponseType = c.Query("response_type")
	sld.State = c.Query("state")
	sld.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	sld.GrantType = c.Query("grant_type")

	//set IsLogin false
	sld.IsLogin = false

	q := url.Values{}
	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, state, err := general.RandToken(32)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	Expires := time.Now().Add(time.Minute * 10)

	v, err := json.Marshal(sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(key, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	url := general.Oauth2Config.AuthCodeURL(state)

	http.Redirect(c.Writer, c.Request, url, http.StatusTemporaryRedirect)

}

func GoogleLogin(c *gin.Context) {
	var sld general.SocialLoginDetails
	sld.ClientId = c.Query("client_id")
	sld.ResponseType = c.Query("response_type")
	sld.State = c.Query("state")
	sld.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	sld.GrantType = c.Query("grant_type")

	//set IsLogin true
	sld.IsLogin = true

	q := url.Values{}
	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, state, err := general.RandToken(32)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	Expires := time.Now().Add(time.Minute * 10)

	v, err := json.Marshal(sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(key, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	url := general.Oauth2Config.AuthCodeURL(state)

	http.Redirect(c.Writer, c.Request, url, http.StatusTemporaryRedirect)

}

func GoogleAuth(c *gin.Context) {

	queryState := c.Request.URL.Query().Get("state")
	log.Println("queryState:", queryState)

	q := url.Values{}

	if len(queryState) == 0 {
		log.Println("Invalid state")
		q.Set("error", "50002")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, err := base64.StdEncoding.DecodeString(queryState)
	if err != nil {
		log.Println("error decoding base64:", err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	objStr, err := RedisClient.Get(string(key)).Result()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	b := []byte(objStr)

	sld := &general.SocialLoginDetails{}

	err = json.Unmarshal(b, sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	var appds *general.AppFromClientId
	if len(sld.ClientId) != 0 || len(sld.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(sld.ClientId, sld.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			q.Set("error", "40004")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if err != nil {
			q.Set("error", "50003")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if !Valid {
			q.Set("error", "40005")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

	}

	code := c.Request.URL.Query().Get("code")

	if len(code) == 0 {
		log.Println("code not found")
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	tok, err := general.Oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	HTTPclient := general.Oauth2Config.Client(context.Background(), tok)
	userinfo, err := HTTPclient.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}
	defer userinfo.Body.Close()

	data, _ := ioutil.ReadAll(userinfo.Body)
	u := general.GoogleUser{}
	if err = json.Unmarshal(data, &u); err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	u.Email = strings.ToLower(u.Email)

	log.Println("Google User:", u)

	//deleting sld data from redis
	_, err = token.DeleteToken(string(key), RedisClient)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	defer sqlDB.Close()

	var Users general.User

	result := db.Limit(1).Find(&Users, "email = ?", u.Email)
	if result.Error != nil {
		log.Println(result.Error)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	var userid string
	var UserProviders general.UserProvider

	if result.RowsAffected > 0 {
		// user already exists in user table

		userid = Users.UserID

		result := db.Limit(1).Find(&UserProviders, "user_id = ? AND provider = ?", userid, general.PROVIDER_GOOGLE)
		if result.Error != nil {
			log.Println(result.Error)
			q.Set("error", "50001")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if result.RowsAffected > 0 {
			// user already exists in both table
			if !(sld.IsLogin) {
				log.Println("user already exists")
				q.Set("error", "40001")
				location := url.URL{Path: "/signup", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

		} else {
			if sld.IsLogin {
				log.Println("user does not exists")
				q.Set("error", "40002")
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			if !u.EmailVerified {
				log.Println("google account not verified")
				q.Set("error", "40003")
				if sld.IsLogin {
					location := url.URL{Path: "/login", RawQuery: q.Encode()}
					c.Redirect(http.StatusFound, location.RequestURI())
					return
				}
				location := url.URL{Path: "/signup", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			// create entry to UserProvider table
			UserProviders.UserId = userid
			UserProviders.Provider = general.PROVIDER_GOOGLE

			result := db.Create(&UserProviders)
			if result.Error != nil {
				log.Println(result.Error)
				q.Set("error", "50001")
				location := url.URL{Path: "/signup", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

		}

	} else {
		// user not exists, go to signup

		if sld.IsLogin {
			log.Println("user does not exists")
			q.Set("error", "40002")
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if !u.EmailVerified {
			log.Println("google account not verified")
			q.Set("error", "40003")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		Password := general.RandPassword(8)

		hashPwd, err := general.HashPassword(Password)
		if err != nil {
			log.Println(err)
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		userid = nanoid.New()

		Users.UserID = userid
		Users.UserName = fmt.Sprintf("%s-%s", strings.Split(u.Email, "@")[0], nanoid.New())
		Users.FullName = u.Name
		Users.Email = u.Email
		Users.Password = hashPwd
		Users.Address1 = ""
		Users.Address2 = ""
		Users.Phone = ""
		Users.EmailVerificationCode = ""
		Users.EmailVerified = true

		UserProviders.UserId = userid
		UserProviders.Provider = general.PROVIDER_GOOGLE

		var tx = db.Begin()

		err = CreateUserSpace(tx, Users)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		result := tx.Create(&UserProviders)

		if result.Error != nil {
			log.Println(result.Error)
			tx.Rollback()
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
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
		// 	q.Set("error", "50001")
		// 	location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		// 	c.Redirect(http.StatusFound, location.RequestURI())
		// 	return
		// }

	}

	// generate device id
	DeviceId := general.RandPassword(8)

	urlParams := &general.AppUrlParams{
		ClientId:     sld.ClientId,
		ResponseType: sld.ResponseType,
		State:        sld.State,
		RedirectUri:  sld.RedirectUri,
		GrantType:    sld.GrantType,
	}

	//Complete auth
	if len(sld.ClientId) != 0 || len(sld.RedirectUri) != 0 {

		ad := &general.AccessDetails{
			AccessUuid:  "",
			RefreshUuid: "",
			UserId:      userid,
			DeviceId:    DeviceId,
		}
		session := &general.ActiveSession{
			IsActiveu_sid: false,
			Sessionstring: "",
		}

		err = CompleteAuth(c, RedisClient, urlParams, ad, appds, session)
		if err != nil {
			q.Set("error", "50004")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
	}
	// else {
	// 	err := SetAppblockToken(c, RedisClient, userid, DeviceId)
	// 	if err != nil {
	// 		q.Set("error", "50004")
	// 		if sld.IsLogin {
	// 			location := url.URL{Path: "/login", RawQuery: q.Encode()}
	// 			c.Redirect(http.StatusFound, location.RequestURI())
	// 			return
	// 		}
	// 		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
	// 		c.Redirect(http.StatusFound, location.RequestURI())
	// 		return
	// 	}

	// 	general.OutputHTML(c.Writer, c.Request, "static/src/pages/success-msg-general.html")
	// 	return
	// }

}

func TwitterSignup(c *gin.Context) {
	var sld general.SocialLoginDetails
	sld.ClientId = c.Query("client_id")
	sld.ResponseType = c.Query("response_type")
	sld.State = c.Query("state")
	sld.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	sld.GrantType = c.Query("grant_type")

	//set IsLogin false
	sld.IsLogin = false

	q := url.Values{}
	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, sessionid, err := general.RandToken(32)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	Expires := time.Now().Add(time.Minute * 10)

	v, err := json.Marshal(sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "Something went wrong")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(key, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpClient := new(http.Client)
	userConfig := &oauth1a.UserConfig{}

	if err = userConfig.GetRequestToken(c, general.Oauth1aService, httpClient); err != nil {
		log.Println("Could not get request token:", err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	var authurl string
	if authurl, err = userConfig.GetAuthorizeURL(general.Oauth1aService); err != nil {
		log.Println("Could not get authorization URL:", err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	v, err = json.Marshal(userConfig)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(sessionid, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	c.SetCookie("tw_session", sessionid, 600, "/", "", true, true)

	http.Redirect(c.Writer, c.Request, authurl, http.StatusFound)

}

func TwitterLogin(c *gin.Context) {
	var sld general.SocialLoginDetails
	sld.ClientId = c.Query("client_id")
	sld.ResponseType = c.Query("response_type")
	sld.State = c.Query("state")
	sld.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	sld.GrantType = c.Query("grant_type")

	//set IsLogin true
	sld.IsLogin = true

	q := url.Values{}
	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, sessionid, err := general.RandToken(32)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	Expires := time.Now().Add(time.Minute * 10)

	v, err := json.Marshal(sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(key, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpClient := new(http.Client)
	userConfig := &oauth1a.UserConfig{}

	if err = userConfig.GetRequestToken(c, general.Oauth1aService, httpClient); err != nil {
		log.Println("Could not get request token:", err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	var authurl string
	if authurl, err = userConfig.GetAuthorizeURL(general.Oauth1aService); err != nil {
		log.Println("Could not get authorization URL:", err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	v, err = json.Marshal(userConfig)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(sessionid, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	c.SetCookie("tw_session", sessionid, 600, "/", "", true, true)

	http.Redirect(c.Writer, c.Request, authurl, http.StatusFound)

}

func TwitterAuth(c *gin.Context) {

	q := url.Values{}

	sessionid, err := c.Cookie("tw_session")
	if err != nil {
		log.Println(err)
		q.Set("error", "50005")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		q.Set("error", "50005")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	objStr, err := RedisClient.Get(sessionid).Result()
	if err != nil {
		log.Println(err)
		q.Set("error", "50005")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	b := []byte(objStr)

	userConfig := &oauth1a.UserConfig{}

	err = json.Unmarshal(b, userConfig)
	if err != nil {
		log.Println(err)
		q.Set("error", "50005")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, err := base64.StdEncoding.DecodeString(sessionid)
	if err != nil {
		log.Println("error decoding base64:", err)
		q.Set("error", "50005")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	objStr, err = RedisClient.Get(string(key)).Result()
	if err != nil {
		log.Println(err)
		q.Set("error", "50005")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	b = []byte(objStr)

	sld := &general.SocialLoginDetails{}

	err = json.Unmarshal(b, sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50005")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	var appds *general.AppFromClientId
	if len(sld.ClientId) != 0 || len(sld.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(sld.ClientId, sld.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			q.Set("error", "40004")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if err != nil {
			q.Set("error", "50003")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if !Valid {
			q.Set("error", "40005")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

	}

	accesstoken, verifier, err := userConfig.ParseAuthorize(c.Request, general.Oauth1aService)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpClient := new(http.Client)
	if err = userConfig.GetAccessToken(c, accesstoken, verifier, general.Oauth1aService, httpClient); err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpRequest, err := http.NewRequest("GET", "https://api.twitter.com/1.1/account/verify_credentials.json?include_email=true", nil)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	if err = general.Oauth1aService.Sign(httpRequest, userConfig); err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	defer httpResponse.Body.Close()

	data, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	u := make(map[string]interface{})

	if err = json.Unmarshal(data, &u); err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	email := strings.ToLower(fmt.Sprintf("%v", u["email"]))
	name := fmt.Sprintf("%v", u["name"])

	if len(email) == 0 || email == "<nil>" {
		log.Println("Failed to fetch email from Twitter")
		q.Set("error", "50006")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	if name == "<nil>" {
		name = ""
	}

	//deleting sld data from redis
	_, err = token.DeleteToken(string(key), RedisClient)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	//deleting userConfig data from redis
	_, err = token.DeleteToken(sessionid, RedisClient)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	// invalidate tw_session cookie
	c.SetCookie("tw_session", "", -1, "/", "", true, true)

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	defer sqlDB.Close()

	var Users general.User

	result := db.Limit(1).Find(&Users, "email = ?", email)
	if result.Error != nil {
		log.Println(result.Error)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	var userid string
	var UserProviders general.UserProvider

	if result.RowsAffected > 0 {
		// user already exists in user table

		userid = Users.UserID

		result := db.Limit(1).Find(&UserProviders, "user_id = ? AND provider = ?", userid, general.PROVIDER_TWITTER)
		if result.Error != nil {
			log.Println(result.Error)
			q.Set("error", "Something went wrong")
			if sld.IsLogin {
				location := url.URL{Path: "/50001", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if result.RowsAffected > 0 {
			// user already exists in both table
			if !(sld.IsLogin) {
				log.Println("user already exists")
				q.Set("error", "40001")
				location := url.URL{Path: "/signup", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

		} else {
			if sld.IsLogin {
				log.Println("user does not exists")
				q.Set("error", "40002")
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			// create entry to UserProvider table
			UserProviders.UserId = userid
			UserProviders.Provider = general.PROVIDER_TWITTER

			result := db.Create(&UserProviders)
			if result.Error != nil {
				log.Println(result.Error)
				q.Set("error", "50001")
				location := url.URL{Path: "/signup", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

		}

	} else {
		// user not exists, go to signup

		if sld.IsLogin {
			log.Println("user does not exists")
			q.Set("error", "40002")
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		Password := general.RandPassword(8)

		hashPwd, err := general.HashPassword(Password)
		if err != nil {
			log.Println(err)
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		userid = nanoid.New()

		Users.UserID = userid
		Users.UserName = fmt.Sprintf("%s-%s", strings.Split(email, "@")[0], nanoid.New())
		Users.FullName = name
		Users.Email = email
		Users.Password = hashPwd
		Users.Address1 = ""
		Users.Address2 = ""
		Users.Phone = ""
		Users.EmailVerificationCode = ""
		Users.EmailVerified = true

		UserProviders.UserId = userid
		UserProviders.Provider = general.PROVIDER_TWITTER

		var tx = db.Begin()

		err = CreateUserSpace(tx, Users)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		result := tx.Create(&UserProviders)

		if result.Error != nil {
			log.Println(result.Error)
			tx.Rollback()
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
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
		// 	q.Set("error", "50001")
		// 	location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		// 	c.Redirect(http.StatusFound, location.RequestURI())
		// 	return
		// }

	}

	// generate device id
	DeviceId := general.RandPassword(8)

	urlParams := &general.AppUrlParams{
		ClientId:     sld.ClientId,
		ResponseType: sld.ResponseType,
		State:        sld.State,
		RedirectUri:  sld.RedirectUri,
		GrantType:    sld.GrantType,
	}

	//Complete auth
	if len(sld.ClientId) != 0 || len(sld.RedirectUri) != 0 {

		ad := &general.AccessDetails{
			AccessUuid:  "",
			RefreshUuid: "",
			UserId:      userid,
			DeviceId:    DeviceId,
		}
		session := &general.ActiveSession{
			IsActiveu_sid: false,
			Sessionstring: "",
		}

		err = CompleteAuth(c, RedisClient, urlParams, ad, appds, session)
		if err != nil {
			q.Set("error", "50004")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
	}
	//  else {
	// 	err := SetAppblockToken(c, RedisClient, userid, DeviceId)
	// 	if err != nil {
	// 		q.Set("error", "50004")
	// 		if sld.IsLogin {
	// 			location := url.URL{Path: "/login", RawQuery: q.Encode()}
	// 			c.Redirect(http.StatusFound, location.RequestURI())
	// 			return
	// 		}
	// 		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
	// 		c.Redirect(http.StatusFound, location.RequestURI())
	// 		return
	// 	}

	// 	general.OutputHTML(c.Writer, c.Request, "static/src/pages/success-msg-general.html")
	// 	return
	// }

}

func LinkedinSignup(c *gin.Context) {
	var sld general.SocialLoginDetails
	sld.ClientId = c.Query("client_id")
	sld.ResponseType = c.Query("response_type")
	sld.State = c.Query("state")
	sld.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	sld.GrantType = c.Query("grant_type")

	//set IsLogin false
	sld.IsLogin = false

	q := url.Values{}
	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Println("redis client handling error")
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, state, err := general.RandToken(32)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	Expires := time.Now().Add(time.Minute * 10)

	v, err := json.Marshal(sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(key, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	rq := url.Values{}
	rq.Set("response_type", "code")
	rq.Set("client_id", general.Envs["SHIELD_LINKEDIN_CLIENT_ID"])
	rq.Set("redirect_uri", general.Envs["SHIELD_LINKEDIN_REDIRECT"])
	rq.Set("state", state)
	rq.Set("scope", "r_liteprofile r_emailaddress")

	u := url.URL{}
	u.Scheme = "https"
	u.Host = "www.linkedin.com"
	u.Path = "oauth/v2/authorization"
	u.RawQuery = rq.Encode()

	authurl := u.String()

	http.Redirect(c.Writer, c.Request, authurl, http.StatusFound)
}

func LinkedinLogin(c *gin.Context) {
	var sld general.SocialLoginDetails
	sld.ClientId = c.Query("client_id")
	sld.ResponseType = c.Query("response_type")
	sld.State = c.Query("state")
	sld.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	sld.GrantType = c.Query("grant_type")

	//set IsLogin true
	sld.IsLogin = true

	q := url.Values{}
	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Println("redis client handling error")
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, state, err := general.RandToken(32)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	Expires := time.Now().Add(time.Minute * 10)

	v, err := json.Marshal(sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	err = RedisClient.Set(key, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	rq := url.Values{}
	rq.Set("response_type", "code")
	rq.Set("client_id", general.Envs["SHIELD_LINKEDIN_CLIENT_ID"])
	rq.Set("redirect_uri", general.Envs["SHIELD_LINKEDIN_REDIRECT"])
	rq.Set("state", state)
	rq.Set("scope", "r_liteprofile r_emailaddress")

	u := url.URL{}
	u.Scheme = "https"
	u.Host = "www.linkedin.com"
	u.Path = "oauth/v2/authorization"
	u.RawQuery = rq.Encode()

	authurl := u.String()

	http.Redirect(c.Writer, c.Request, authurl, http.StatusFound)
}

func LinkedinAuth(c *gin.Context) {

	queryState := c.Request.URL.Query().Get("state")
	q := url.Values{}

	if len(queryState) == 0 {
		log.Println("Invalid state")
		q.Set("error", "50007")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Println("redis client handling error")
		q.Set("error", "50007")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	key, err := base64.StdEncoding.DecodeString(queryState)
	if err != nil {
		log.Println("error decoding base64:", err)
		q.Set("error", "50007")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	objStr, err := RedisClient.Get(string(key)).Result()
	if err != nil {
		log.Println(err)
		q.Set("error", "50007")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	b := []byte(objStr)

	sld := &general.SocialLoginDetails{}

	err = json.Unmarshal(b, sld)
	if err != nil {
		log.Println(err)
		q.Set("error", "50007")
		location := url.URL{Path: "/login", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	if len(sld.ClientId) != 0 && ((sld.ResponseType == "code" || sld.ResponseType == "device_code") && len(sld.State) != 0 && len(sld.RedirectUri) != 0) {
		q.Set("client_id", sld.ClientId)
		q.Set("response_type", sld.ResponseType)
		q.Set("state", sld.State)
		q.Set("redirect_uri", sld.RedirectUri)
	}
	if len(sld.ClientId) != 0 && sld.GrantType == "urn:ietf:params:oauth:grant-type:token-exchange" {
		q.Set("grant_type", sld.GrantType)
		q.Set("client_id", sld.ClientId)
	}

	var appds *general.AppFromClientId
	if len(sld.ClientId) != 0 || len(sld.RedirectUri) != 0 {
		var Valid bool
		var err error
		Valid, appds, err = appreg.ValidateAppClientIdandRedirectUrl(sld.ClientId, sld.RedirectUri)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			q.Set("error", "40004")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if err != nil {
			q.Set("error", "50003")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if !Valid {
			q.Set("error", "40005")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

	}

	code := c.Request.URL.Query().Get("code")

	if len(code) == 0 {
		log.Println("code not found")
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpClient := &http.Client{}

	rq := url.Values{}
	rq.Set("grant_type", "authorization_code")
	rq.Set("code", code)
	rq.Set("redirect_uri", general.Envs["SHIELD_LINKEDIN_REDIRECT"])
	rq.Set("client_id", general.Envs["SHIELD_LINKEDIN_CLIENT_ID"])
	rq.Set("client_secret", general.Envs["SHIELD_LINKEDIN_CLIENT_SECRET"])

	u := url.URL{}
	u.Scheme = "https"
	u.Host = "www.linkedin.com"
	u.Path = "oauth/v2/accessToken"
	u.RawQuery = rq.Encode()

	authurl := u.String()

	req, err := http.NewRequest("POST", authurl, nil)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	var AccessTokenResp general.LinkedinATResponse

	err = json.Unmarshal(bodyBytes, &AccessTokenResp)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpRequest, err := http.NewRequest("GET", "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))", nil)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpRequest.Header.Add("Authorization", "Bearer "+AccessTokenResp.AccessToken)

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	defer httpResponse.Body.Close()

	data, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	emailbody := make(map[string]interface{})

	if err = json.Unmarshal(data, &emailbody); err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	email := strings.ToLower(emailbody["elements"].([]interface{})[0].(map[string]interface{})["handle~"].(map[string]interface{})["emailAddress"].(string))

	if len(email) == 0 {
		log.Println("Failed to fetch email from Linkedin")
		q.Set("error", "50008")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	// fetchig profile details
	httpRequest, err = http.NewRequest("GET", "https://api.linkedin.com/v2/me", nil)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	httpRequest.Header.Add("Authorization", "Bearer "+AccessTokenResp.AccessToken)

	httpResponse, err = httpClient.Do(httpRequest)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	defer httpResponse.Body.Close()

	data, err = ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	profile := make(map[string]interface{})

	if err = json.Unmarshal(data, &profile); err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	FirstName := profile["localizedFirstName"]
	LastName := profile["localizedLastName"]

	//deleting sld data from redis
	_, err = token.DeleteToken(string(key), RedisClient)
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	defer sqlDB.Close()

	var Users general.User

	result := db.Limit(1).Find(&Users, "email = ?", email)
	if result.Error != nil {
		log.Println(result.Error)
		q.Set("error", "50001")
		if sld.IsLogin {
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		c.Redirect(http.StatusFound, location.RequestURI())
		return
	}

	var userid string
	var UserProviders general.UserProvider

	if result.RowsAffected > 0 {
		// user already exists in user table

		userid = Users.UserID

		result := db.Limit(1).Find(&UserProviders, "user_id = ? AND provider = ?", userid, general.PROVIDER_LINKEDIN)
		if result.Error != nil {
			log.Println(result.Error)
			q.Set("error", "50001")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		if result.RowsAffected > 0 {
			// user already exists in both table
			if !(sld.IsLogin) {
				log.Println("user already exists")
				q.Set("error", "40001")
				location := url.URL{Path: "/signup", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

		} else {
			if sld.IsLogin {
				log.Println("user does not exists")
				q.Set("error", "40002")
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

			// create entry to UserProvider table
			UserProviders.UserId = userid
			UserProviders.Provider = general.PROVIDER_LINKEDIN

			result := db.Create(&UserProviders)
			if result.Error != nil {
				log.Println(result.Error)
				q.Set("error", "50001")
				location := url.URL{Path: "/signup", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}

		}

	} else {
		// user not exists, go to signup

		if sld.IsLogin {
			log.Println("user does not exists")
			q.Set("error", "40002")
			location := url.URL{Path: "/login", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		Password := general.RandPassword(8)

		hashPwd, err := general.HashPassword(Password)
		if err != nil {
			log.Println(err)
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		userid = nanoid.New()

		fullname := FirstName.(string) + " " + LastName.(string)
		fullname = strings.TrimSpace(fullname)

		Users.UserID = userid
		Users.UserName = fmt.Sprintf("%s-%s", strings.Split(email, "@")[0], nanoid.New())
		Users.FullName = fullname
		Users.Email = email
		Users.Password = hashPwd
		Users.Address1 = ""
		Users.Address2 = ""
		Users.Phone = ""
		Users.EmailVerificationCode = ""
		Users.EmailVerified = true

		UserProviders.UserId = userid
		UserProviders.Provider = general.PROVIDER_LINKEDIN

		var tx = db.Begin()

		err = CreateUserSpace(tx, Users)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}

		result := tx.Create(&UserProviders)

		if result.Error != nil {
			log.Println(result.Error)
			tx.Rollback()
			q.Set("error", "50001")
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
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
		// 	q.Set("error", "50001")
		// 	location := url.URL{Path: "/signup", RawQuery: q.Encode()}
		// 	c.Redirect(http.StatusFound, location.RequestURI())
		// 	return
		// }

	}

	// generate device id
	DeviceId := general.RandPassword(8)

	urlParams := &general.AppUrlParams{
		ClientId:     sld.ClientId,
		ResponseType: sld.ResponseType,
		State:        sld.State,
		RedirectUri:  sld.RedirectUri,
		GrantType:    sld.GrantType,
	}

	//Complete auth
	if len(sld.ClientId) != 0 || len(sld.RedirectUri) != 0 {

		ad := &general.AccessDetails{
			AccessUuid:  "",
			RefreshUuid: "",
			UserId:      userid,
			DeviceId:    DeviceId,
		}
		session := &general.ActiveSession{
			IsActiveu_sid: false,
			Sessionstring: "",
		}

		err = CompleteAuth(c, RedisClient, urlParams, ad, appds, session)
		if err != nil {
			q.Set("error", "50004")
			if sld.IsLogin {
				location := url.URL{Path: "/login", RawQuery: q.Encode()}
				c.Redirect(http.StatusFound, location.RequestURI())
				return
			}
			location := url.URL{Path: "/signup", RawQuery: q.Encode()}
			c.Redirect(http.StatusFound, location.RequestURI())
			return
		}
	}
	// else {
	// 	err := SetAppblockToken(c, RedisClient, userid, DeviceId)
	// 	if err != nil {
	// 		q.Set("error", "50004")
	// 		if sld.IsLogin {
	// 			location := url.URL{Path: "/login", RawQuery: q.Encode()}
	// 			c.Redirect(http.StatusFound, location.RequestURI())
	// 			return
	// 		}
	// 		location := url.URL{Path: "/signup", RawQuery: q.Encode()}
	// 		c.Redirect(http.StatusFound, location.RequestURI())
	// 		return
	// 	}

	// 	general.OutputHTML(c.Writer, c.Request, "static/src/pages/success-msg-general.html")
	// 	return
	// }

}

// DisplayErrorMessage
func DisplayErrorMessage(c *gin.Context) {
	general.OutputHTML(c.Writer, c.Request, "static/src/pages/error-msg-login.html")
}

// GetPasswordRecovery page
func GetPasswordRecovery(c *gin.Context) {
	general.OutputHTML(c.Writer, c.Request, "static/src/pages/password-recovery.html")
}

func PasswordRecovery(c *gin.Context) {

	var urlParams general.ResetPasswordDetails
	urlParams.ClientId = c.Query("client_id")
	urlParams.ResponseType = c.Query("response_type")
	urlParams.State = c.Query("state")
	urlParams.RedirectUri = c.Query("redirect_uri")
	//for app token exchange
	urlParams.GrantType = c.Query("grant_type")

	var RequestBody general.ResetPasswordRequestEvent

	err := c.BindJSON(&RequestBody)
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "invalid request body")
		return
	}

	err = general.ValidateStruct(&RequestBody)
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "invalid request body")
		return
	}

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "db error")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "db error")
		return
	}

	defer sqlDB.Close()

	var userdetails general.ResetPasswordUser

	// result := db.Model(&general.User{}).Select("users.user_id, users.user_name, users.email, users.password, users.email_verified, user_providers.provider").Joins("inner join user_providers on users.user_id = user_providers.user_id").Where("users.email = ?", email).Where("user_providers.provider = ?", general.PROVIDER_SHIELD).Limit(1).Scan(&userwprovider)
	result := db.Model(&general.User{}).Select("users.user_id, users.user_name, users.email").Joins("inner join user_providers on users.user_id = user_providers.user_id").Where("users.email = ?", strings.ToLower(RequestBody.Email)).Where("user_providers.provider = ?", general.PROVIDER_SHIELD).Limit(1).Scan(&userdetails)
	if result.Error != nil {
		log.Println(result.Error)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "db error")
		return
	}

	if result.RowsAffected <= 0 {
		log.Println("invalid email id:", RequestBody.Email)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "invalid email id")
		return
	}

	urlParams.UserId = userdetails.UserID

	v, err := json.Marshal(&urlParams)

	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to marshal url parameters")
		return
	}

	key := nanoid.New()
	resetcode := base64.URLEncoding.EncodeToString([]byte(key))
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to generate token")
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Println("redis client handling error")
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis client handling error")
		return
	}

	Expires := time.Now().Add(time.Minute * 10)

	err = RedisClient.Set(key, v, time.Until(Expires)).Err()
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis save error")
		return
	}

	// Send password recovery mail =>
	go func() {

		// TODO optimize

		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}

		rq := url.Values{}
		rq.Set("reset_code", resetcode)

		u := url.URL{}
		log.Println("u", u)
		u.Scheme = scheme
		log.Println("u", u)
		u.Host = c.Request.Host
		u.Path = "change-password"
		u.RawQuery = rq.Encode()

		resetDetails := general.ResetPasswordMail{
			Email:     userdetails.Email,
			UserName:  userdetails.UserName,
			ResetLink: u.String(),
		}

		mailer.PasswordRecoveryMailer(resetDetails)
	}()
	//  <= Send password recovery mail

	general.RespondHandler(c.Writer, true, http.StatusOK, "success")
}

// GetChangePassword page
func GetChangePassword(c *gin.Context) {
	general.OutputHTML(c.Writer, c.Request, "static/src/pages/change-password.html")
}

func ChangePassword(c *gin.Context) {

	reset_code := c.Query("reset_code")

	q := url.Values{}

	if len(reset_code) == 0 {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Reset code missing in request")
		return
	}

	q.Set("reset_code", reset_code)

	var RequestBody general.ChangePasswordRequestEvent

	RequestBody.Password = c.PostForm("password")
	RequestBody.PasswordRe = c.PostForm("passwordRe")

	err := general.ValidateStruct(&RequestBody)
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Invalid request body")
		return
	}

	if RequestBody.Password != RequestBody.PasswordRe {
		log.Println("password mismatch")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Password mismatch")
		return
	}

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Println("redis client handling error")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	key, err := base64.URLEncoding.DecodeString(reset_code)
	if err != nil {
		log.Println("error decoding base64:", err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	res, err := RedisClient.Get(string(key)).Result()

	if err != nil {
		log.Println("invalid or expired reset code")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Invalid or expired reset code")
		return
	}

	b := []byte(res)

	rpd := &general.ResetPasswordDetails{}

	err = json.Unmarshal(b, rpd)
	if err != nil {
		log.Println(err)
		q.Set("error", "Something went wrong")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	log.Println("rpd:", rpd)

	db, err := pghandling.SetupDB()
	if err != nil {
		log.Println(err)
		q.Set("error", "Something went wrong")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		q.Set("error", "Something went wrong")
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	defer sqlDB.Close()

	var userdetails general.ResetPasswordUser

	result := db.Model(&general.User{}).Select("users.user_id, users.user_name, users.email").Joins("inner join user_providers on users.user_id = user_providers.user_id").Where("users.user_id = ?", rpd.UserId).Where("user_providers.provider = ?", general.PROVIDER_SHIELD).Limit(1).Scan(&userdetails)
	if result.Error != nil {
		log.Println(result.Error)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	if result.RowsAffected <= 0 {
		log.Println("invalid user id:", rpd.UserId)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "User details does not exists")
		return
	}

	hashPwd, err := general.HashPassword(RequestBody.Password)
	if err != nil {
		log.Println(err)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	result = db.Model(&general.User{}).Where("user_Id = ?", rpd.UserId).Update("password", hashPwd)

	if result.Error != nil {
		log.Println(result.Error)
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
		return
	}

	if result.RowsAffected <= 0 {
		general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "User does not exists")
		return
	}

	// delete all tokens for the user
	var cursor uint64
	var n int
	var allkeys []string
	for {
		var keys []string
		var err error
		keys, cursor, err = RedisClient.Scan(cursor, "*:"+rpd.UserId, 10).Result()
		if err != nil {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
			return
		}
		allkeys = append(allkeys, keys...)
		n += len(keys)
		if cursor == 0 {
			break
		}
	}

	for _, v := range allkeys {
		deleted, err := token.DeleteToken(v, RedisClient)
		if err != nil || deleted == 0 {
			log.Println(err)
			general.OutputHTMLTemplate(c.Writer, "static/src/pages/change-password-error.html", "Something went wrong")
			return
		}
	}

	// invalidate cookie
	// c.SetCookie("u_at", "", -1, "/", ".appblocks.com", false, false)
	// c.SetCookie("u_rt", "", -1, "/", ".appblocks.com", false, false)

	if len(rpd.ClientId) != 0 {
		q.Set("client_id", rpd.ClientId)
	}
	if len(rpd.ResponseType) != 0 {
		q.Set("response_type", rpd.ResponseType)
	}
	if len(rpd.State) != 0 {
		q.Set("state", rpd.State)
	}
	if len(rpd.RedirectUri) != 0 {
		q.Set("redirect_uri", rpd.RedirectUri)
	}
	if len(rpd.GrantType) != 0 {
		q.Set("grant_type", rpd.GrantType)
	}

	q.Del("reset_code")

	location := url.URL{Path: "/login", RawQuery: q.Encode()}
	c.Redirect(http.StatusFound, location.RequestURI())

}

func ChangeUserPassword(c *gin.Context) {

	var RequestBody general.ChangeUserPasswordRequestEvent

	RequestBody.CurrentPassword = c.PostForm("current_password")
	RequestBody.NewPassword = c.PostForm("new_password")

	err := general.ValidateStruct(&RequestBody)
	if err != nil {
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "Invalid request body")
		return
	}

	if RequestBody.CurrentPassword == RequestBody.NewPassword {
		log.Println("new password cannot be the same as the current password")
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "New password cannot be the same as the current password")
		return
	}

	tokenString := token.ExtractToken(c)
	if len(tokenString) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "access token not found")
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
		// case general.IDTOKEN:
		// 	ad, Status, err = token.ProcessIDToken(c)

		// 	if err != nil {
		// 		general.RespondHandler(c.Writer, false, Status, err.Error())
		// 		log.Println(err)
		// 		return
		// 	}
		// case general.ACCESS:
		// 	ad, Status, err = token.ProcessUserAccessToken(c)

		// 	if err != nil {
		// 		general.RespondHandler(c.Writer, false, Status, err.Error())
		// 		log.Println(err)
		// 		return
		// 	}
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

		var userdetails general.ResetUserPasswordDetails

		result := db.Model(&general.User{}).Select("users.user_id, users.user_name, users.email, users.password").Joins("inner join user_providers on users.user_id = user_providers.user_id").Where("users.user_id = ?", ad.UserId).Where("user_providers.provider = ?", general.PROVIDER_SHIELD).Limit(1).Scan(&userdetails)
		if result.Error != nil {
			log.Println(result.Error)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
			return
		}

		if result.RowsAffected <= 0 {
			log.Println("invalid user id:", ad.UserId)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "User details does not exists")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(userdetails.Password), []byte(RequestBody.CurrentPassword))

		if err != nil {
			log.Println("hash password error/mismatch")
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, "Incorrect Password")
			return
		}

		hashPwd, err := general.HashPassword(RequestBody.NewPassword)
		if err != nil {
			log.Println(err)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
			return
		}

		result = db.Model(&general.User{}).Where("user_Id = ?", ad.UserId).Update("password", hashPwd)

		if result.Error != nil {
			log.Println(result.Error)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "Something went wrong")
			return
		}

		if result.RowsAffected <= 0 {
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "User does not exists")
			return
		}

		log.Println("password changed")

		general.RespondHandler(c.Writer, true, http.StatusOK, "Password successfully changed")

		return
	}

	err = errors.New("failed to fetch claims from token")

	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())

}

func DisplayAuthorizeAppnamePermission(c *gin.Context) {
	general.OutputHTML(c.Writer, c.Request, "static/src/pages/authorize-appname-permission.html")
}

func DisplayAuthorizeAppname(c *gin.Context) {
	general.OutputHTML(c.Writer, c.Request, "static/src/pages/authorize-appname.html")
}

func CreateUserSpace(tx *gorm.DB, usr general.User) error {

	member := general.Member{
		ID:   usr.UserID,
		Type: "U",
	}

	user := general.User{
		UserID:                  usr.UserID,
		UserName:                usr.UserName,
		FullName:                usr.FullName,
		Email:                   usr.Email,
		Password:                usr.Password,
		Address1:                usr.Address1,
		Address2:                usr.Address2,
		Phone:                   usr.Phone,
		EmailVerificationCode:   usr.EmailVerificationCode,
		EmailVerified:           usr.EmailVerified,
		EmailVerificationExpiry: usr.EmailVerificationExpiry,
		CreatedAt:               usr.CreatedAt,
		UpdatedAt:               usr.UpdatedAt,
	}

	SpaceID := nanoid.New()

	RoleID := nanoid.New()

	// Create role
	roles := general.Role{
		ID:           RoleID,
		IsOwner:      true,
		Name:         "owner",
		OwnerSpaceID: SpaceID,
		CreatedBy:    usr.UserID,
		UpdatedBy:    usr.UserID,
	}

	memberSpace := general.Member{
		ID:   SpaceID,
		Type: "S",
	}

	MemberRoleID := nanoid.New()
	// Create member role
	memberRole := general.MemberRole{
		ID:           MemberRoleID,
		OwnerUserID:  usr.UserID,
		RoleID:       RoleID,
		OwnerSpaceID: SpaceID,
	}

	spaceMember := general.SpaceMember{
		ID:           nanoid.New(),
		OwnerUserID:  usr.UserID,
		OwnerSpaceID: SpaceID,
	}

	space := general.Space{
		SpaceID: SpaceID,
		Type:    "P",
		Name:    usr.UserName,
		Email:   usr.Email,
	}

	// Set space as default user space
	defaultSpace := general.DefaultUserSpace{
		ID:           nanoid.New(),
		OwnerUserID:  user.UserID,
		OwnerSpaceID: SpaceID,
	}

	result := tx.Create(&member)

	if result.Error != nil {
		return result.Error
	}

	log.Println("member created")

	log.Println("user:", &user)
	result = tx.Create(&user)
	if result.Error != nil {
		return result.Error
	}

	log.Println("user created")

	result = tx.Create(&memberSpace)
	if result.Error != nil {
		return result.Error
	}

	log.Println("memberSpace created")

	result = tx.Create(&space)
	if result.Error != nil {
		return result.Error
	}

	log.Println("space created")

	result = tx.Create(&spaceMember)
	if result.Error != nil {
		return result.Error
	}

	log.Println("space member created")

	result = tx.Create(&defaultSpace)
	if result.Error != nil {
		return result.Error
	}

	log.Println("default user space created")

	result = tx.Create(&roles)
	if result.Error != nil {
		return result.Error
	}

	log.Println("role created")

	result = tx.Create(&memberRole)
	if result.Error != nil {
		return result.Error
	}

	log.Println("memberRole created")

	//deprecated space manage purchase access

	// valuesMap := make(map[string]interface{})
	// valuesMap["userID"] = usr.UserID
	// valuesMap["spaceID"] = SpaceID

	// type subExists struct {
	// 	PolGrpSubsExists bool `json:"pol_grp_subs_exists"`
	// }
	// var Exists subExists

	// result = tx.Raw(`select exists(select subs.id from ac_pol_grp_subs subs
	// 	inner join ac_pol_grps agp on agp.id=subs.ac_pol_grp_id
	// 	where  agp.name in ('Blocks-Manage-Purchase','Blocks Publish') and agp.is_predefined and agp.type=1
	// 	 and subs.owner_space_id=@spaceID and subs.owner_user_id=@userID
	// 	) as pol_grp_subs_exists`, valuesMap).Scan(&Exists)

	// if result.Error != nil {
	// 	return result.Error
	// }

	// if !Exists.PolGrpSubsExists {
	// 	result := tx.Exec(`INSERT INTO public.ac_pol_grp_subs(
	// 		id, created_at, updated_at, owner_space_id, role_id, owner_team_id, owner_user_id, ac_pol_grp_id)
	// 		select nanoid(),now(),now(),@spaceID,null,null,@userID,polgrp.id
	// 		from ac_pol_grps polgrp where
	// 		 polgrp.name in ('Blocks-Manage-Purchase','Blocks Publish') and polgrp.is_predefined and polgrp.type=1`, valuesMap)

	// 	if result.Error != nil {
	// 		return result.Error
	// 	}

	// }

	// rolePayload := general.RolePayload{
	// 	SpaceID: SpaceID,
	// 	RoleID:  RoleID,
	// }

	// err := AssignPredefinedPolicyToRole(tx, rolePayload)
	// return err

	return nil
}

func AssignPredefinedPolicyToRole(tx *gorm.DB, rolePayload general.RolePayload) error {

	var policyData []general.PolicyData
	var acPolGrpData general.AcPolGrpData
	var polGpPolicyData []general.PolGpPolicyData
	var polGrpPolicyPayload []general.PolGpPolicy
	var acPolGrpSubData general.AcPolGrpSubData

	result := tx.Raw(`select id from ac_policies where is_predefined`).Scan(&policyData)
	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("POlicy data is %v\n", policyData)

	result = tx.Create(general.AcPolGrp{
		ID:          nanoid.New(),
		MemberID:    rolePayload.SpaceID,
		Description: "Predefined policy group for personal space",
		OptCounter:  0,
	}).Scan(&acPolGrpData)

	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("ac pol grp data is %v\n", acPolGrpData)

	for _, v := range policyData {
		polGrpPolicyPayload = append(polGrpPolicyPayload, general.PolGpPolicy{ID: nanoid.New(), AcPolicyID: v.ID, AcPolGrpID: acPolGrpData.ID, OptCounter: 0})

	}

	fmt.Printf("acpolicy grp policy payload is %v", polGrpPolicyPayload)

	result = tx.Create(polGrpPolicyPayload).Scan(&polGpPolicyData)
	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("policy group create response is %v", polGpPolicyData)

	result = tx.Raw(`insert into ac_pol_grp_subs (id,role_id,ac_pol_grp_id,opt_counter) values($1,$2,$3,$4)`, nanoid.New(), rolePayload.RoleID, acPolGrpData.ID, 0).Scan(&acPolGrpSubData)

	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("pol group sub data is %v", acPolGrpSubData)

	return nil

}
