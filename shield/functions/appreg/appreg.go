package appreg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aidarkhanov/nanoid"
	"github.com/appblocks-hub/SHIELD/common_services"
	"github.com/appblocks-hub/SHIELD/functions/general"
	"github.com/appblocks-hub/SHIELD/functions/pghandling"
	"github.com/appblocks-hub/SHIELD/functions/token"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func BlockAppRegistration(c *gin.Context) {
	var newApp general.BlockApprRegEvent
	err := c.BindJSON(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = general.ValidateStruct(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

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

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		clientId := newApp.AppId + "-" + strconv.Itoa(1000+r.Intn(9999-1000))

		if len(general.Envs["CLIENT_SECRET_KEY"]) == 0 {
			general.Envs["CLIENT_SECRET_KEY"] = "uvwxyz"
		}
		key := []byte(general.Envs["CLIENT_SECRET_KEY"])
		message := uuid.New().String()

		sig := hmac.New(sha256.New, key)
		sig.Write([]byte(message))

		clientSecret := hex.EncodeToString(sig.Sum(nil))

		redirecturl := []string{}

		Apps = general.ShieldApp{
			AppId:        newApp.AppId,
			ClientId:     clientId,
			ClientSecret: clientSecret,
			UserId:       ad.UserId,
			AppName:      newApp.AppId,
			AppSname:     newApp.AppId,
			Description:  "",
			LogoUrl:      "",
			AppUrl:       "",
			RedirectUrl:  redirecturl,
			AppType:      4,
		}

		AppUrls := general.ShieldAppDomainMapping{
			ID:         nanoid.New(),
			OwnerAppID: newApp.AppId,
			Url:        "",
		}

		// Fetch mandatory permissions
		var Permissions []general.Permission

		result = db.Find(&Permissions, "Mandatory = ?", true)

		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, result.Error.Error())
			log.Println(result.Error)
			return
		}

		var AppPermission []general.AppPermission

		for _, perm := range Permissions {
			AppPermission = append(AppPermission, general.AppPermission{
				AppId:        newApp.AppId,
				PermissionId: perm.PermissionId,
				Mandatory:    true,
			})

		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&Apps).Error; err != nil {
				// return any error - will rollback
				return err
			}
			if err := tx.Create(&AppPermission).Error; err != nil {
				// return any error - will rollback
				return err
			}
			if err := tx.Create(&AppUrls).Error; err != nil {
				// return any error - will rollback
				return err
			}
			// return nil will commit the whole transaction
			return nil
		})

		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err)
			log.Println(err)
			return
		}

		general.RespondHandler(c.Writer, true, http.StatusCreated, "success")
		return
	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

func GetBlockAppClientId(c *gin.Context) {
	var newApp general.BlockApprRegEvent
	err := c.BindJSON(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = general.ValidateStruct(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if len(newApp.AppId) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		log.Println("app_id missing in request")
		return
	}

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

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

		var authCodeValues *general.AuthCodeValues
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, _, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}

		case general.APPBACCESS:
			authCodeValues, _, Status, err = token.ProcessAppblockAccessToken(c)
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

		var AppCreds general.BlockAppClientIdResponse
		result = db.Model(&Apps).First(&AppCreds, "app_id = ?", newApp.AppId)
		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
			log.Println(result.Error)
			return
		}

		AppCreds.CreatedAt = time.Now()

		general.RespondHandler(c.Writer, true, http.StatusOK, &AppCreds)
		return
	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

func GetBlockAppClientSecret(c *gin.Context) {
	var newApp general.BlockApprRegEvent
	err := c.BindJSON(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = general.ValidateStruct(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if len(newApp.AppId) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		log.Println("app_id missing in request")
		return
	}

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

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
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, _, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}

		case general.APPBACCESS:
			authCodeValues, _, Status, err = token.ProcessAppblockAccessToken(c)
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

		var AppCreds general.BlockAppClientSecretResponse
		result = db.Model(&Apps).First(&AppCreds, "app_id = ?", newApp.AppId)
		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
			log.Println(result.Error)
			return
		}

		AppCreds.CreatedAt = time.Now()

		general.RespondHandler(c.Writer, true, http.StatusOK, &AppCreds)
		return
	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

func UpdateBlockAppRedirectUrl(c *gin.Context) {
	var newRedirects general.BlockApprRedirectUrl
	err := c.BindJSON(&newRedirects)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = general.ValidateStruct(&newRedirects)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if len(newRedirects.AppId) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		log.Println("app_id missing in request")
		return
	}

	// validate redirect urls
	for _, redirect := range newRedirects.RedirectUrl {

		if !(general.IsValidRedirectUrl(redirect)) {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, "contains invalid redirect_url")
			log.Println("contains invalid redirect_url:", redirect)
			return
		}

	}

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

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
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, _, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}

		case general.APPBACCESS:
			authCodeValues, _, Status, err = token.ProcessAppblockAccessToken(c)
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

		result = db.Model(&general.ShieldApp{}).Where("app_id = ?", newRedirects.AppId).Update("redirect_url", pq.Array(newRedirects.RedirectUrl))
		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
			log.Println(result.Error)
			return
		}

		if result.RowsAffected < 1 {
			general.RespondHandler(c.Writer, false, http.StatusNoContent, "NO RECORD FOUND!")
			log.Println("NO RECORD FOUND!")
			return
		}

		general.RespondHandler(c.Writer, true, http.StatusOK, "success")
		return
	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

func GetBlockAppRedirectUrl(c *gin.Context) {
	var newApp general.BlockApprRegEvent
	err := c.BindJSON(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = general.ValidateStruct(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if len(newApp.AppId) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		log.Println("app_id missing in request")
		return
	}

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

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

		var authCodeValues *general.AuthCodeValues
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, _, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}

		case general.APPBACCESS:
			authCodeValues, _, Status, err = token.ProcessAppblockAccessToken(c)
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

		Apps.AppId = ""
		Apps.RedirectUrl = []string{}

		result = db.First(&Apps, "app_id = ?", newApp.AppId)
		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
			log.Println(result.Error)
			return
		}

		general.RespondHandler(c.Writer, true, http.StatusOK, &Apps.RedirectUrl)
		return
	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

func AppRegistration(payload common_services.HandlerPayload) common_services.HandlerResponse {
	var newApp general.ApprRegEvent
	// err := c.BindJSON(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// err = general.ValidateStruct(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	var handlerResp common_services.HandlerResponse

	if err := json.Unmarshal([]byte(payload.RequestBody), &newApp); err != nil {
		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	db := payload.Db

	ClientId := payload.Queryparams["client-id"]
	ClientSecret := payload.Queryparams["client-secret"]

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		// err := errors.New("client id or secret missing in request")
		// log.Println(err)
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client id or secret missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// db, err := pghandling.SetupDB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	return
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", ClientId)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if Apps.AppType != 2 {
		// err := errors.New("app is not allowed to access this")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app is not allowed to access this", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	//if client secret passing not matched with client secret from db
	if ClientSecret != Apps.ClientSecret {
		// err := errors.New("client secret not matched")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client secret not matched", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	clientId := newApp.AppId + "-" + strconv.Itoa(1000+r.Intn(9999-1000))

	if len(general.Envs["CLIENT_SECRET_KEY"]) == 0 {
		general.Envs["CLIENT_SECRET_KEY"] = "uvwxyz"
	}
	key := []byte(general.Envs["CLIENT_SECRET_KEY"])
	message := uuid.New().String()

	sig := hmac.New(sha256.New, key)
	sig.Write([]byte(message))

	clientSecret := hex.EncodeToString(sig.Sum(nil))

	redirecturl := []string{}

	Apps = general.ShieldApp{
		AppId:        newApp.AppId,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		UserId:       payload.UserID,
		AppName:      newApp.AppId,
		AppSname:     newApp.AppId,
		Description:  "",
		LogoUrl:      "",
		AppUrl:       "",
		RedirectUrl:  redirecturl,
		AppType:      4,
	}

	// Fetch mandatory permissions
	var Permissions []general.Permission

	result = db.Find(&Permissions, "Mandatory = ?", true)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusInternalServerError, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Internal Server Error", general.ResponseTemplate{}, http.StatusInternalServerError)
		return handlerResp
	}

	var AppPermission []general.AppPermission

	for _, perm := range Permissions {
		AppPermission = append(AppPermission, general.AppPermission{
			AppId:        newApp.AppId,
			PermissionId: perm.PermissionId,
			Mandatory:    true,
		})

	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&Apps).Error; err != nil {
			// return any error - will rollback
			return err
		}
		if err := tx.Create(&AppPermission).Error; err != nil {
			// return any error - will rollback
			return err
		}
		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		// general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err)
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Internal Server Error", general.ResponseTemplate{}, http.StatusInternalServerError)
		return handlerResp
	}

	// general.RespondHandler(c.Writer, true, http.StatusCreated, "success")

	handlerResp = common_services.BuildResponse(false, "success", general.ResponseTemplate{Data: "success", Success: true, Message: "success", Status: http.StatusOK}, http.StatusOK)
	return handlerResp

}

func GetAppClientId(payload common_services.HandlerPayload) common_services.HandlerResponse {
	var newApp general.BlockApprRegEvent
	// err := c.BindJSON(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// err = general.ValidateStruct(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	var handlerResp common_services.HandlerResponse

	if err := json.Unmarshal([]byte(payload.RequestBody), &newApp); err != nil {
		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	db := payload.Db

	if len(newApp.AppId) == 0 {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		// log.Println("app_id missing in request")
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app_id missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	ClientId := payload.Queryparams["client-id"]
	ClientSecret := payload.Queryparams["client-secret"]

	log.Println("client-id secret:", ClientId, ClientSecret)

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		// err := errors.New("client id or secret missing in request")
		// log.Println(err)
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client id or secret missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// db, err := pghandling.SetupDB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	return
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", ClientId)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if Apps.AppType != 2 {
		// err := errors.New("app is not allowed to access this")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app is not allowed to access this", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	//if client secret passing not matched with client secret from db
	if ClientSecret != Apps.ClientSecret {
		// err := errors.New("client secret not matched")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client secret not matched", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	var AppCreds general.BlockAppClientIdResponse
	result = db.Model(&Apps).First(&AppCreds, "app_id = ?", newApp.AppId)
	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, result.Error.Error(), general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	AppCreds.CreatedAt = time.Now()

	// general.RespondHandler(c.Writer, true, http.StatusOK, &AppCreds)

	handlerResp = common_services.BuildResponse(false, "success", general.ResponseTemplate{Data: &AppCreds, Success: true, Message: "success", Status: http.StatusOK}, http.StatusOK)
	return handlerResp

}

func GetAppClientSecret(payload common_services.HandlerPayload) common_services.HandlerResponse {
	var newApp general.BlockApprRegEvent
	// err := c.BindJSON(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// err = general.ValidateStruct(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	var handlerResp common_services.HandlerResponse

	if err := json.Unmarshal([]byte(payload.RequestBody), &newApp); err != nil {
		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	db := payload.Db

	if len(newApp.AppId) == 0 {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		// log.Println("app_id missing in request")
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app_id missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	ClientId := payload.Queryparams["client-id"]
	ClientSecret := payload.Queryparams["client-secret"]

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		// err := errors.New("client id or secret missing in request")
		// log.Println(err)
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client id or secret missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// db, err := pghandling.SetupDB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	return
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", ClientId)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if Apps.AppType != 2 {
		// err := errors.New("app is not allowed to access this")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app is not allowed to access this", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	//if client secret passing not matched with client secret from db
	if ClientSecret != Apps.ClientSecret {
		// err := errors.New("client secret not matched")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client secret not matched", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	var AppCreds general.BlockAppClientSecretResponse
	result = db.Model(&Apps).First(&AppCreds, "app_id = ?", newApp.AppId)
	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, result.Error.Error(), general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	AppCreds.CreatedAt = time.Now()

	// general.RespondHandler(c.Writer, true, http.StatusOK, &AppCreds)

	handlerResp = common_services.BuildResponse(false, "success", general.ResponseTemplate{Data: &AppCreds, Success: true, Message: "success", Status: http.StatusOK}, http.StatusOK)
	return handlerResp
}

func UpdateAppRedirectUrl(payload common_services.HandlerPayload) common_services.HandlerResponse {
	var newRedirects general.BlockApprRedirectUrl
	// err := c.BindJSON(&newRedirects)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// err = general.ValidateStruct(&newRedirects)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	var handlerResp common_services.HandlerResponse

	if err := json.Unmarshal([]byte(payload.RequestBody), &newRedirects); err != nil {
		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	db := payload.Db

	if len(newRedirects.AppId) == 0 {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		// log.Println("app_id missing in request")
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app_id missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// validate redirect urls
	for _, redirect := range newRedirects.RedirectUrl {

		if !(general.IsValidRedirectUrl(redirect)) {
			// general.RespondHandler(c.Writer, false, http.StatusBadRequest, "contains invalid redirect_url")
			// log.Println("contains invalid redirect_url:", redirect)
			// return

			handlerResp = common_services.BuildErrorResponse(true, "contains invalid redirect_url", general.ResponseTemplate{}, http.StatusBadRequest)
			return handlerResp
		}

	}

	ClientId := payload.Queryparams["client-id"]
	ClientSecret := payload.Queryparams["client-secret"]

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		// err := errors.New("client id or secret missing in request")
		// log.Println(err)
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client id or secret missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// db, err := pghandling.SetupDB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	return
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", ClientId)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if Apps.AppType != 2 {
		// err := errors.New("app is not allowed to access this")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app is not allowed to access this", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	//if client secret passing not matched with client secret from db
	if ClientSecret != Apps.ClientSecret {
		// err := errors.New("client secret not matched")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client secret not matched", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	result = db.Model(&general.ShieldApp{}).Where("app_id = ?", newRedirects.AppId).Update("redirect_url", pq.Array(newRedirects.RedirectUrl))
	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, result.Error.Error(), general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if result.RowsAffected < 1 {
		// general.RespondHandler(c.Writer, false, http.StatusNoContent, "NO RECORD FOUND!")
		// log.Println("NO RECORD FOUND!")
		// return

		handlerResp = common_services.BuildErrorResponse(true, "NO RECORD FOUND!", general.ResponseTemplate{}, http.StatusNoContent)
		return handlerResp
	}

	// general.RespondHandler(c.Writer, true, http.StatusOK, "success")

	handlerResp = common_services.BuildResponse(false, "success", general.ResponseTemplate{Data: "success", Success: true, Message: "success", Status: http.StatusOK}, http.StatusOK)
	return handlerResp

}

func GetAppRedirectUrl(payload common_services.HandlerPayload) common_services.HandlerResponse {
	var newApp general.BlockApprRegEvent
	// err := c.BindJSON(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// err = general.ValidateStruct(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	var handlerResp common_services.HandlerResponse

	if err := json.Unmarshal([]byte(payload.RequestBody), &newApp); err != nil {
		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	db := payload.Db

	if len(newApp.AppId) == 0 {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		// log.Println("app_id missing in request")
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app_id missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	ClientId := payload.Queryparams["client-id"]
	ClientSecret := payload.Queryparams["client-secret"]

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		// err := errors.New("client id or secret missing in request")
		// log.Println(err)
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client id or secret missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}
	// db, err := pghandling.SetupDB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	return
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", ClientId)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if Apps.AppType != 2 {
		// err := errors.New("app is not allowed to access this")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app is not allowed to access this", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	//if client secret passing not matched with client secret from db
	if ClientSecret != Apps.ClientSecret {
		// err := errors.New("client secret not matched")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client secret not matched", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	Apps.AppId = ""
	Apps.RedirectUrl = []string{}

	result = db.First(&Apps, "app_id = ?", newApp.AppId)
	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, result.Error.Error(), general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// general.RespondHandler(c.Writer, true, http.StatusOK, &Apps.RedirectUrl)

	handlerResp = common_services.BuildResponse(false, "success", general.ResponseTemplate{Data: &Apps.RedirectUrl, Success: true, Message: "success", Status: http.StatusOK}, http.StatusOK)
	return handlerResp
}

func GetAllPermissions(c *gin.Context) {
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

	db.Find(&Permissions)

	log.Println(&Permissions)
	general.RespondHandler(c.Writer, true, http.StatusOK, &Permissions)

}

func GetBlockAppScopes(c *gin.Context) {
	var newApp general.BlockApprRegEvent
	err := c.BindJSON(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = general.ValidateStruct(&newApp)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if len(newApp.AppId) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		log.Println("app_id missing in request")
		return
	}

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

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

		var authCodeValues *general.AuthCodeValues
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, _, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}

		case general.APPBACCESS:
			authCodeValues, _, Status, err = token.ProcessAppblockAccessToken(c)
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

		//if client id passing not matched with client id in token
		if ClientId != authCodeValues.ClientId {
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

		// var Permissions []general.Permission

		var Permissions []general.GetScopes
		appPermissionsQuery := `select permissions.permission_id, permissions.permission_name, 
	case when permissions.mandatory is true then false else true end as editable,
	case when app_permissions.mandatory is null then false else app_permissions.mandatory end as mandatory,
	case when app_permissions.app_id is null then false else true end as selected
	from permissions
	left join app_permissions on app_permissions.permission_id = permissions.permission_id and app_id = ?`

		res := db.Raw(appPermissionsQuery, newApp.AppId).Scan(&Permissions)
		if res.Error != nil {
			log.Println(res.Error)
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, res.Error)
			return
		}

		// db.Model(&general.Permission{}).Select("permissions.permission_id, permissions.permission_name, permissions.mandatory as editable, app_permissions.")(&Permissions)

		// db.Find(&Permissions).Scan(&Permissions)

		// db.Find(&Permissions)

		log.Println(&Permissions)
		general.RespondHandler(c.Writer, true, http.StatusOK, &Permissions)
		return

	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

func UpdateBlockAppScopes(c *gin.Context) {
	var reqPayload general.CreateBlockAppScopes
	err := c.BindJSON(&reqPayload)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = general.ValidateStruct(&reqPayload)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	if len(reqPayload.AppId) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		log.Println("app_id missing in request")
		return
	}

	ClientId := c.Request.Header.Get("Client-Id")
	ClientSecret := c.Request.Header.Get("Client-Secret")

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		err := errors.New("client id or secret missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

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

		var authCodeValues *general.AuthCodeValues
		var Status int
		var err error
		switch tokenType {
		case general.DEVICEACCESS:
			authCodeValues, _, Status, err = token.ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}

		case general.APPBACCESS:
			authCodeValues, _, Status, err = token.ProcessAppblockAccessToken(c)
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

		//if client id passing not matched with client id in token
		if ClientId != authCodeValues.ClientId {
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

		// fetch mandatory scopes
		var ManPermissions []general.BlockAppScopes

		db.Model(&general.Permission{}).Where("mandatory = ?", true).Scan(&ManPermissions)

		ManPermissions = append(ManPermissions, reqPayload.Scopes...)

		// fetch existing app scopes
		var ExAppPermissions []general.BlockAppScopes
		db.Model(&general.AppPermission{}).Where("app_id = ?", reqPayload.AppId).Scan(&ExAppPermissions)

		var exPermList []string
		for _, a := range ExAppPermissions {
			exPermList = append(exPermList, a.PermissionId)
		}

		var NewAppPermissions []general.AppPermission

		var curPermList []string
		var UpdatePermissions []general.BlockAppScopes
		for _, a := range ManPermissions {

			exists := false

			curPermList = append(curPermList, a.PermissionId)

			for _, b := range ExAppPermissions {
				if a.PermissionId == b.PermissionId {
					exists = true
					if a.Mandatory != b.Mandatory {
						UpdatePermissions = append(UpdatePermissions, a)
					}
				}

			}
			if !(exists) {
				NewAppPermissions = append(NewAppPermissions, general.AppPermission{
					AppId:        reqPayload.AppId,
					PermissionId: a.PermissionId,
					Mandatory:    a.Mandatory,
				})
			}

		}

		// appUserPermissionsDeleteQuery := `delete from app_user_permissions where app_id = ? and permission_id IN (?) and permission_id NOT IN (?)`

		err = db.Transaction(func(tx *gorm.DB) error {
			if len(exPermList) > 0 && len(curPermList) > 0 {
				if err := tx.Where("app_id = ? and permission_id IN (?) and permission_id NOT IN (?)", reqPayload.AppId, exPermList, curPermList).Delete(&general.AppUserPermission{}).Error; err != nil {
					// return any error - will rollback
					return err
				}

				if err := tx.Where("app_id = ? and permission_id IN (?) and permission_id NOT IN (?)", reqPayload.AppId, exPermList, curPermList).Delete(&general.AppPermission{}).Error; err != nil {
					// return any error - will rollback
					return err
				}

			}

			if len(exPermList) > 0 && len(curPermList) <= 0 {
				if err := tx.Where("app_id = ? and permission_id IN (?)", reqPayload.AppId, exPermList).Delete(&general.AppUserPermission{}).Error; err != nil {
					// return any error - will rollback
					return err
				}

				if err := tx.Where("app_id = ? and permission_id IN (?)", reqPayload.AppId, exPermList).Delete(&general.AppPermission{}).Error; err != nil {
					// return any error - will rollback
					return err
				}

			}

			for _, a := range UpdatePermissions {
				log.Println(a)

				if err := tx.Model(&general.AppPermission{AppId: reqPayload.AppId, PermissionId: a.PermissionId}).Select("mandatory", "updated_at").Updates(general.AppPermission{Mandatory: a.Mandatory, UpdatedAt: time.Now()}).Error; err != nil {
					// return any error - will rollback
					return err
				}

			}

			if len(NewAppPermissions) > 0 {
				if err := tx.Create(&NewAppPermissions).Error; err != nil {
					// return any error - will rollback
					return err
				}
			}

			// return nil will commit the whole transaction
			return nil
		})

		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
			log.Println(err)
			return
		}

		general.RespondHandler(c.Writer, true, http.StatusOK, "success")
		return

	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)

}

func GetAppScopes(payload common_services.HandlerPayload) common_services.HandlerResponse {
	var newApp general.BlockApprRegEvent
	// err := c.BindJSON(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// err = general.ValidateStruct(&newApp)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	var handlerResp common_services.HandlerResponse

	if err := json.Unmarshal([]byte(payload.RequestBody), &newApp); err != nil {
		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	db := payload.Db

	if len(newApp.AppId) == 0 {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		// log.Println("app_id missing in request")
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app_id missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	ClientId := payload.Queryparams["client-id"]
	ClientSecret := payload.Queryparams["client-secret"]

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		// err := errors.New("client id or secret missing in request")
		// log.Println(err)
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client id or secret missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// db, err := pghandling.SetupDB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", ClientId)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if Apps.AppType != 2 {
		// err := errors.New("app is not allowed to access this")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app is not allowed to access this", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	//if client secret passing not matched with client secret from db
	if ClientSecret != Apps.ClientSecret {
		// err := errors.New("client secret not matched")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client secret not matched", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// var Permissions []general.Permission

	var Permissions []general.GetScopes
	appPermissionsQuery := `select permissions.permission_id, permissions.permission_name, 
	case when permissions.mandatory is true then false else true end as editable,
	case when app_permissions.mandatory is null then false else app_permissions.mandatory end as mandatory,
	case when app_permissions.app_id is null then false else true end as selected
	from permissions
	left join app_permissions on app_permissions.permission_id = permissions.permission_id and app_id = ?`

	res := db.Raw(appPermissionsQuery, newApp.AppId).Scan(&Permissions)
	if res.Error != nil {
		// log.Println(res.Error)
		// general.RespondHandler(c.Writer, false, http.StatusInternalServerError, res.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Internal Server Error", general.ResponseTemplate{}, http.StatusInternalServerError)
		return handlerResp
	}

	// db.Model(&general.Permission{}).Select("permissions.permission_id, permissions.permission_name, permissions.mandatory as editable, app_permissions.")(&Permissions)

	// db.Find(&Permissions).Scan(&Permissions)

	// db.Find(&Permissions)

	// general.RespondHandler(c.Writer, true, http.StatusOK, &Permissions)

	handlerResp = common_services.BuildResponse(false, "success", general.ResponseTemplate{Data: &Permissions, Success: true, Message: "success", Status: http.StatusOK}, http.StatusOK)
	return handlerResp

}

func UpdateAppScopes(payload common_services.HandlerPayload) common_services.HandlerResponse {
	var reqPayload general.CreateBlockAppScopes
	// err := c.BindJSON(&reqPayload)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// err = general.ValidateStruct(&reqPayload)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	var handlerResp common_services.HandlerResponse

	if err := json.Unmarshal([]byte(payload.RequestBody), &reqPayload); err != nil {
		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp

	}

	db := payload.Db

	if len(reqPayload.AppId) == 0 {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, "app_id missing in request")
		// log.Println("app_id missing in request")
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app_id missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	ClientId := payload.Queryparams["client-id"]
	ClientSecret := payload.Queryparams["client-secret"]

	if len(ClientId) == 0 || len(ClientSecret) == 0 {

		// err := errors.New("client id or secret missing in request")
		// log.Println(err)
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client id or secret missing in request", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// db, err := pghandling.SetupDB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", ClientId)

	if result.Error != nil {
		// general.RespondHandler(c.Writer, false, http.StatusBadRequest, result.Error.Error())
		// log.Println(result.Error)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Invalid Request Payload", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	if Apps.AppType != 2 {
		// err := errors.New("app is not allowed to access this")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "app is not allowed to access this", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	//if client secret passing not matched with client secret from db
	if ClientSecret != Apps.ClientSecret {
		// err := errors.New("client secret not matched")
		// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "client secret not matched", general.ResponseTemplate{}, http.StatusBadRequest)
		return handlerResp
	}

	// fetch mandatory scopes
	var ManPermissions []general.BlockAppScopes

	db.Model(&general.Permission{}).Where("mandatory = ?", true).Scan(&ManPermissions)

	ManPermissions = append(ManPermissions, reqPayload.Scopes...)

	// fetch existing app scopes
	var ExAppPermissions []general.BlockAppScopes
	db.Model(&general.AppPermission{}).Where("app_id = ?", reqPayload.AppId).Scan(&ExAppPermissions)

	var exPermList []string
	for _, a := range ExAppPermissions {
		exPermList = append(exPermList, a.PermissionId)
	}

	var NewAppPermissions []general.AppPermission

	var curPermList []string
	var UpdatePermissions []general.BlockAppScopes
	for _, a := range ManPermissions {

		exists := false

		curPermList = append(curPermList, a.PermissionId)

		for _, b := range ExAppPermissions {
			if a.PermissionId == b.PermissionId {
				exists = true
				if a.Mandatory != b.Mandatory {
					UpdatePermissions = append(UpdatePermissions, a)
				}
			}

		}
		if !(exists) {
			NewAppPermissions = append(NewAppPermissions, general.AppPermission{
				AppId:        reqPayload.AppId,
				PermissionId: a.PermissionId,
				Mandatory:    a.Mandatory,
			})
		}

	}

	// appUserPermissionsDeleteQuery := `delete from app_user_permissions where app_id = ? and permission_id IN (?) and permission_id NOT IN (?)`

	err := db.Transaction(func(tx *gorm.DB) error {
		if len(exPermList) > 0 && len(curPermList) > 0 {
			if err := tx.Where("app_id = ? and permission_id IN (?) and permission_id NOT IN (?)", reqPayload.AppId, exPermList, curPermList).Delete(&general.AppUserPermission{}).Error; err != nil {
				// return any error - will rollback
				return err
			}

			if err := tx.Where("app_id = ? and permission_id IN (?) and permission_id NOT IN (?)", reqPayload.AppId, exPermList, curPermList).Delete(&general.AppPermission{}).Error; err != nil {
				// return any error - will rollback
				return err
			}

		}

		if len(exPermList) > 0 && len(curPermList) <= 0 {
			if err := tx.Where("app_id = ? and permission_id IN (?)", reqPayload.AppId, exPermList).Delete(&general.AppUserPermission{}).Error; err != nil {
				// return any error - will rollback
				return err
			}

			if err := tx.Where("app_id = ? and permission_id IN (?)", reqPayload.AppId, exPermList).Delete(&general.AppPermission{}).Error; err != nil {
				// return any error - will rollback
				return err
			}

		}

		for _, a := range UpdatePermissions {

			if err := tx.Model(&general.AppPermission{AppId: reqPayload.AppId, PermissionId: a.PermissionId}).Select("mandatory", "updated_at").Updates(general.AppPermission{Mandatory: a.Mandatory, UpdatedAt: time.Now()}).Error; err != nil {
				// return any error - will rollback
				return err
			}

		}

		if len(NewAppPermissions) > 0 {
			if err := tx.Create(&NewAppPermissions).Error; err != nil {
				// return any error - will rollback
				return err
			}
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		// general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
		// log.Println(err)
		// return

		handlerResp = common_services.BuildErrorResponse(true, "Internal Server Error", general.ResponseTemplate{}, http.StatusInternalServerError)
		return handlerResp
	}

	// general.RespondHandler(c.Writer, true, http.StatusOK, "success")

	handlerResp = common_services.BuildResponse(false, "success", general.ResponseTemplate{Data: "success", Success: true, Message: "success", Status: http.StatusOK}, http.StatusOK)
	return handlerResp
}

// swagger:route POST /create-app-permissions AppPermissionsRequest
// Create App Permissions
//
// security:
// - Key: []
// parameters:
//
//	AppPermissionsRequest
//
// responses:
//
//	201: Response
//	400: Response
func CreateAppPermissions(c *gin.Context) {
	var newAppPermission general.AppPermissionEvent
	err := c.BindJSON(&newAppPermission)
	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	log.Println(newAppPermission)

	err = general.ValidateStruct(&newAppPermission)

	if err != nil {
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
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

	// userid, status, err := token.GetUserIdFromUserAccessToken(c)
	// if err != nil {
	// 	general.RespondHandler(c.Writer, false, status, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	// isowner, err := IsOwnerOfApp(db, userid.String(), newAppPermission.AppId)

	// if !isowner {
	// 	if err.Error() == "invalid user" {
	// 		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
	// 		log.Println(err)
	// 		return
	// 	}

	// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
	// 	log.Println(err)
	// 	return
	// }

	AppPermissions := general.AppPermission{}

	for i, p := range newAppPermission.PermissionId {

		AppPermissions.AppId = newAppPermission.AppId
		AppPermissions.PermissionId = p

		AppPermissions.Mandatory = newAppPermission.Mandatory[i]
		result := db.Create(&AppPermissions)
		if result.Error != nil {
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, result.Error.Error())
			log.Println(result.Error)
			return
		}

	}

	general.RespondHandler(c.Writer, true, http.StatusCreated, "success")

}

func FetchAppPermission(appid string) (AppPermissions *[]general.AppPermission, err error) {
	db, err := pghandling.SetupDB()
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	defer sqlDB.Close()

	db.Find(&AppPermissions, "app_id = ?", appid)

	return AppPermissions, nil

}

func CreateAppUserPermissionsDBEntry(AppUserPermission *[]general.AppUserPermission) (err error) {
	db, err := pghandling.SetupDB()
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		return err
	}

	defer sqlDB.Close()

	result := db.Create(AppUserPermission)

	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}

	return nil

}

func GetAppFromClientId(clientid string) (AppDetails *general.AppFromClientId, err error) {

	db, err := pghandling.SetupDB()
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", clientid)

	if result.Error != nil {
		return nil, result.Error
	}

	return &general.AppFromClientId{
		AppId:       Apps.AppId,
		ClientId:    Apps.ClientId,
		RedirectUrl: Apps.RedirectUrl,
		AppSname:    Apps.AppSname,
		AppType:     Apps.AppType,
	}, nil

}

func IsOwnerOfApp(db *gorm.DB, UserId, AppId string) (bool, error) {

	//validate userid from token is owner of app
	var Apps general.ShieldApp

	db.First(&Apps, "app_id = ?", AppId)

	log.Println(Apps)

	if Apps.UserId != UserId {
		err := errors.New("invalid user")
		return false, err
	}

	return true, nil
}

// IsPermissionSetforUser checks app permissions assigned for user
func IsPermissionSetforUser(userid, appid string) (bool, error) {

	db, err := pghandling.SetupDB()
	if err != nil {
		return false, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return false, err
	}

	defer sqlDB.Close()

	//check for valid app id

	var Apps general.ShieldApp

	var exists bool
	err = db.Model(Apps).
		Select("count(*) > 0").
		Where("app_id = ?", appid).
		Find(&exists).
		Error

	if err != nil {
		log.Println("failed to fetch app id from db:", appid)
		return false, err
	}

	if !exists {
		log.Println("failed to fetch app id from db:", appid)
		return false, err
	}

	var AppUserPermissions general.AppUserPermission

	result := db.First(&AppUserPermissions, "user_id = ? AND app_id = ?", userid, appid)
	log.Println("rows affected:", result.RowsAffected)
	if result.RowsAffected > 0 {
		return true, nil
	}

	return false, nil

}

func FetchAppUserPermissions(userid, appid string) (AppUserPermissions *[]general.AppUserPermission, err error) {
	db, err := pghandling.SetupDB()
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	defer sqlDB.Close()

	db.Find(&AppUserPermissions, "user_id = ? AND app_id = ?", userid, appid)

	return AppUserPermissions, nil
}

func ValidateAppClientIdandRedirectUrl(clientid string, redirecturl string) (Valid bool, AppDetails *general.AppFromClientId, err error) {

	u, err := url.Parse(redirecturl)
	if err != nil {
		return false, nil, err
	}

	reidrect_url_chk := u.Scheme + "://" + u.Host

	db, err := pghandling.SetupDB()
	if err != nil {
		return false, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return false, nil, err
	}

	defer sqlDB.Close()

	var Apps general.ShieldApp

	result := db.First(&Apps, "client_id = ?", clientid)

	if result.Error != nil {
		return false, nil, result.Error
	}

	validRedirect := false
	for _, a := range Apps.RedirectUrl {
		if strings.Trim(a, " ") == reidrect_url_chk {
			validRedirect = true
		}
	}

	if !validRedirect {
		return false, nil, nil
	}

	return true, &general.AppFromClientId{
		AppId:       Apps.AppId,
		ClientId:    Apps.ClientId,
		RedirectUrl: Apps.RedirectUrl,
		AppSname:    Apps.AppSname,
		AppType:     Apps.AppType,
	}, nil

}
