package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/appblocks-hub/SHIELD/functions/general"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

// FetchSecret fetches Access or Refresh Secret from environment variables
// For security, remove the if blocks before moving to production
func FetchSecret(which string) string {

	Secret := ""
	switch which {
	case general.IDTOKEN:
		Secret = general.Envs["ID_SECRET"]
		if len(Secret) == 0 {
			Secret = "abcd"
		}
	case general.ACCESS:
		Secret = general.Envs["ACCESS_SECRET"]
		if len(Secret) == 0 {
			Secret = "abcd"
		}
	case general.REFRESH:
		Secret = general.Envs["REFRESH_SECRET"]
		if len(Secret) == 0 {
			Secret = "efgh"
		}
	case general.APPBACCESS:
		Secret = general.Envs["APPBACCESS_SECRET"]
		if len(Secret) == 0 {
			Secret = "uvwx"
		}

	case general.APPBREFRESH:
		Secret = general.Envs["APPBREFRESH_SECRET"]
		if len(Secret) == 0 {
			Secret = "yzab"
		}
	case general.DEVICEACCESS:
		Secret = general.Envs["DEVICEACCESS_SECRET"]
		if len(Secret) == 0 {
			Secret = "ijkl"
		}

	case general.APPACCESS:
		Secret = general.Envs["APPACCESS_SECRET"]
		if len(Secret) == 0 {
			Secret = "mnop"
		}

	case general.APPREFRESH:
		Secret = general.Envs["APPREFRESH_SECRET"]
		if len(Secret) == 0 {
			Secret = "qrst"
		}

	case general.VERIFY_EMAIL_SECRET:
		Secret = general.Envs["VERIFY_EMAIL_SECRET"]
		if len(Secret) == 0 {
			Secret = "VERIFY_EMAIL_SECRET"
		}
	}

	return Secret
}

// CreateToken creates Access token for userid and secret key string.
//
// Expiry time is set to 15 minutes.
//
// Token creates using jwt.NewWithClaims().
func CreateToken(tokentype string, expires int64) (*general.TokenDetails, error) {
	td := &general.TokenDetails{}
	key := FetchSecret(tokentype)
	td.Expires = expires //time.Now().Add(time.Minute * 15).Unix()
	// td.Expires = time.Now().Add(time.Second * 30).Unix()
	td.Uuid = uuid.New().String()

	var err error
	Claims := jwt.MapClaims{}
	Claims["authorized"] = true
	Claims["token_uuid"] = td.Uuid
	Claims["token_type"] = tokentype
	Claims["exp"] = td.Expires
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	td.Token, err = token.SignedString([]byte(key))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return td, nil

}

// SaveToken save the token uuid as key and claims as data to redis database.
//
// Expiry time in Unix time converted to local time and calculates the remaining time to expiry
// and entry will be automatically removed from redis storage after expiry
func SaveToken(key string, expires int64, client *redis.Client, claims interface{}) error {
	et := time.Unix(expires, 0)
	now := time.Now()

	v, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	err = client.Set(key, v, et.Sub(now)).Err()
	if err != nil {
		return err
	}

	return nil
}

// GenerateTokenPair create user token pairs and save using SaveTokens()
func GenerateTokenPair(RedisClient *redis.Client, userid, deviceid string) (general.TokenPairs, error) {
	tp := &general.TokenPairs{}

	//  atexpires := time.Now().Add(time.Minute * 15).Unix()
	// atexpires := time.Now().Add(time.Second * 20).Unix()
	// atexpires := time.Now().Add(time.Hour * 4).Unix()

	expStr := general.Envs["ACCESS_TOKEN_EXPIRY"]

	log.Println("expStr:", expStr)

	if len(expStr) == 0 {
		log.Println("using default expiry for access token")
		expStr = "900"
	}

	expInt, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return *tp, err
	}

	atexpires := time.Now().Add(time.Second * time.Duration(expInt)).Unix()

	at, err := CreateToken(general.ACCESS, atexpires)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	rtexpStr := general.Envs["REFRESH_TOKEN_EXPIRY"]

	log.Println("rtexpStr:", rtexpStr)

	if len(rtexpStr) == 0 {
		log.Println("using default expiry for refresh token")
		rtexpStr = "604800"
	}

	rtexpInt, err := strconv.ParseInt(rtexpStr, 10, 64)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return *tp, err
	}

	rtexpires := time.Now().Add(time.Second * time.Duration(rtexpInt)).Unix()

	// rtexpires := time.Now().Add(time.Hour * 24 * 7).Unix()

	rt, err := CreateToken(general.REFRESH, rtexpires)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	atd := &general.UserAccessDetails{}
	atd.PairTokenUuid = rt.Uuid
	atd.UserId = userid
	atd.DeviceId = deviceid

	err = SaveToken(at.Uuid+":"+deviceid+":"+userid, at.Expires, RedisClient, atd)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	rtd := &general.UserAccessDetails{}
	rtd.PairTokenUuid = at.Uuid
	rtd.UserId = userid
	rtd.DeviceId = deviceid
	err = SaveToken(rt.Uuid+":"+deviceid+":"+userid, rt.Expires, RedisClient, rtd)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	tp.AccessToken = at.Token
	tp.RefreshToken = rt.Token

	return *tp, nil

}

// GenerateAppTokenPair create app token pairs and save using SaveTokens()
func GenerateAppTokenPair(RedisClient *redis.Client, appauthclaims *general.AuthCodeValues, isAppBlock bool) (general.TokenPairDetails, error) {
	tp := &general.TokenPairDetails{}

	atokentype, rtokentype := general.APPACCESS, general.APPREFRESH

	if isAppBlock {
		atokentype, rtokentype = general.APPBACCESS, general.APPBREFRESH
	}

	// atexpires := time.Now().Add(time.Minute * 15).Unix()
	// atexpires := time.Now().Add(time.Second * 20).Unix()

	expStr := general.Envs["APPACCESS_TOKEN_EXPIRY"]

	if len(expStr) == 0 {
		log.Println("using default expiry for app access token")
		expStr = "900"
	}

	expInt, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return *tp, err
	}

	atexpires := time.Now().Add(time.Second * time.Duration(expInt)).Unix()

	// atexpires := time.Now().Add(time.Hour * 4).Unix()

	at, err := CreateToken(atokentype, atexpires)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	rtexpStr := general.Envs["APPREFRESH_TOKEN_EXPIRY"]

	if len(rtexpStr) == 0 {
		log.Println("using default expiry for app refresh token")
		rtexpStr = "604800"
	}

	rtexpInt, err := strconv.ParseInt(rtexpStr, 10, 64)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return *tp, err
	}

	rtexpires := time.Now().Add(time.Second * time.Duration(rtexpInt)).Unix()

	// rtexpires := time.Now().Add(time.Hour * 24 * 7).Unix()

	rt, err := CreateToken(rtokentype, rtexpires)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	atd := &general.AppAccessDetails{}
	atd.PairTokenUuid = rt.Uuid
	atd.UserId = appauthclaims.UserId
	atd.ClientId = appauthclaims.ClientId
	atd.AppSname = appauthclaims.AppSname
	atd.AppUserPermission = appauthclaims.AppUserPermission
	atd.DeviceId = appauthclaims.DeviceId

	// saving access token to redis
	err = SaveToken(at.Uuid+":"+appauthclaims.DeviceId+":"+appauthclaims.UserId, at.Expires, RedisClient, atd)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	rtd := &general.AppAccessDetails{}
	rtd.PairTokenUuid = at.Uuid
	rtd.UserId = appauthclaims.UserId
	rtd.ClientId = appauthclaims.ClientId
	rtd.AppSname = appauthclaims.AppSname
	rtd.AppUserPermission = appauthclaims.AppUserPermission
	rtd.DeviceId = appauthclaims.DeviceId

	// saving refresh token to redis

	err = SaveToken(rt.Uuid+":"+appauthclaims.DeviceId+":"+appauthclaims.UserId, rt.Expires, RedisClient, rtd)
	if err != nil {
		log.Println(err)
		return *tp, err
	}

	tp.AccessToken = at.Token
	tp.RefreshToken = rt.Token
	tp.AccessUuid = at.Uuid
	tp.RefreshUuid = rt.Uuid
	tp.AtExpires = at.Expires
	tp.RtExpires = rt.Expires

	return *tp, nil

}

// CreateTokenPair craete and return access and refresh tokens.
func CreateTokenPair(RedisClient *redis.Client, UserId, DeviceId string) (general.TokenPairs, error) {
	tokens, err := GenerateTokenPair(RedisClient, UserId, DeviceId)

	if err != nil {
		tokens.AccessToken = ""
		tokens.RefreshToken = ""
		return tokens, err
	}

	return tokens, nil

}

// ExtractCookieToken extracts token from Authorization Header if not it will check for cookie
func ExtractCookieToken(c *gin.Context, tokenName string) string {
	tokenString, err := c.Cookie(tokenName)
	if err != nil {
		log.Println(err)
		return ""
	}
	// log.Println("token string:", tokenString)
	return tokenString

}

// ExtractToken extracts token from Authorization Header if not it will check for cookie
func ExtractToken(c *gin.Context) string {
	bearToken := c.Request.Header.Get("Authorization")
	if len(bearToken) != 0 {
		strArr := strings.Split(bearToken, " ")
		if strArr[0] != "Bearer" {
			log.Println("not a Bearer token")
			return ""
		}
		if len(strArr) == 2 || len(strArr) == 3 {
			return strArr[1]
		}
	}

	return ""

}

// ExtractTokenPairs extracts tokens including refresh token from Authorization Header
func ExtractTokenPairs(c *gin.Context) (string, string) {
	bearToken := c.Request.Header.Get("Authorization")
	if len(bearToken) != 0 {
		strArr := strings.Split(bearToken, " ")

		if strArr[0] != "Bearer" {
			log.Println("not Bearer tokens")
			return "", ""
		}

		if len(strArr) == 3 {
			return strArr[1], strArr[2]
		}
	}

	return "", ""

}

func FetchFromRedisUsingKey(client *redis.Client, key string) ([]byte, error) {
	var cursor uint64
	var n int
	var allkeys []string
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(cursor, key+":*", 10).Result()
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allkeys = append(allkeys, keys...)
		n += len(keys)
		if cursor == 0 {
			break
		}
	}

	fmt.Printf("found %d keys\n", n)
	log.Println("keys:", allkeys)

	if n == 0 {
		log.Println("user logged out")
		err := errors.New("user logged out")
		return nil, err
	}

	res, err := client.Get(allkeys[0]).Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return []byte(res), nil

}

// ParseAndVerifyToken parrse and validate token string and returns jwt token.
func ParseAndVerifyToken(tokenString, tokenType string) (*jwt.Token, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		Secret := FetchSecret(tokenType)

		return []byte(Secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// ExtractIDTokenMetadata extracts claims from redis for given id token
func ExtractIDTokenMetadata(RedisClient *redis.Client, token *jwt.Token) (*general.AccessDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		idUuid, ok := claims["token_uuid"].(string)
		if !ok {
			err := errors.New("failed to fetch access uuid from token")
			log.Println(err)
			return nil, err
		}

		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token type from token")
			log.Println(err)
			return nil, err
		}

		if tokenType != general.IDTOKEN {
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return nil, err
		}

		b, err := FetchFromRedisUsingKey(RedisClient, idUuid)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		values := &general.UserAccessDetails{}

		err = json.Unmarshal(b, values)
		if err != nil {
			log.Println("json.Unmarshal error:", err)
			return nil, err
		}

		ad := &general.AccessDetails{
			AccessUuid:  idUuid,
			RefreshUuid: values.PairTokenUuid, // blank
			UserId:      values.UserId,
			DeviceId:    values.DeviceId,
		}

		log.Println("Access details:", ad)
		return ad, nil
	}
	err := errors.New("failed to fetch claims from token")
	return nil, err
}

// ExtractUserAccessTokenMetadata extracts claims from redis for given user access token
func ExtractUserAccessTokenMetadata(RedisClient *redis.Client, token *jwt.Token) (*general.AccessDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["token_uuid"].(string)
		if !ok {
			err := errors.New("failed to fetch access uuid from token")
			log.Println(err)
			return nil, err
		}

		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token type from token")
			log.Println(err)
			return nil, err
		}

		if tokenType != general.ACCESS {
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return nil, err
		}

		b, err := FetchFromRedisUsingKey(RedisClient, accessUuid)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		values := &general.UserAccessDetails{}

		err = json.Unmarshal(b, values)
		if err != nil {
			log.Println("json.Unmarshal error:", err)
			return nil, err
		}

		ad := &general.AccessDetails{
			AccessUuid:  accessUuid,
			RefreshUuid: values.PairTokenUuid,
			UserId:      values.UserId,
			DeviceId:    values.DeviceId,
		}

		log.Println("Access details:", ad)
		return ad, nil
	}
	err := errors.New("failed to fetch claims from token")
	return nil, err
}

// ExtractDeviceTokenMetadata extracts claims from redis for given device access token
func ExtractDeviceTokenMetadata(RedisClient *redis.Client, token *jwt.Token) (*general.AuthCodeValues, string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		accessUuid, ok := claims["token_uuid"].(string)
		if !ok {
			err := errors.New("failed to fetch access uuid from token")
			log.Println(err)
			return nil, "", err
		}

		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token type from token")
			log.Println(err)
			return nil, "", err
		}

		if tokenType != general.DEVICEACCESS {
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return nil, "", err
		}

		b, err := FetchFromRedisUsingKey(RedisClient, accessUuid)
		if err != nil {
			return nil, "", err
		}
		values := &general.AuthCodeValues{}

		err = json.Unmarshal(b, values)
		if err != nil {
			log.Println("json.Unmarshal error:", err)
			return nil, "", err
		}

		log.Println("Device Access details:", values)
		return values, accessUuid, nil

		// clientId, appShortName := "", ""
		// if tokenType == general.DEVICEACCESS {
		// 	clientId, ok = claims["client_id"].(string)
		// 	if !ok {
		// 		err := errors.New("failed to fetch client id from token")
		// 		log.Println(err)
		// 		return nil, "", err
		// 	}
		// }

		// if tokenType == general.APPACCESS || tokenType == general.APPBACCESS {
		// 	clientId, ok = claims["client_id"].(string)
		// 	if !ok {
		// 		err := errors.New("failed to fetch client id from token")
		// 		log.Println(err)
		// 		return nil, "", err
		// 	}

		// 	appShortName, ok = claims["app_shortname"].(string)
		// 	if !ok {
		// 		err := errors.New("failed to fetch app short name from token")
		// 		log.Println(err)
		// 		return nil, "", err
		// 	}
		// }

		// permissionIds, ok := claims["permission"]
		// if !ok {
		// 	err := errors.New("failed to fetch permissions from token")
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// p, err := json.Marshal(permissionIds)
		// if err != nil {
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// var permissions []string
		// err = json.Unmarshal(p, &permissions)
		// if err != nil {
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// userId, ok := claims["user_id"].(string)
		// if !ok {
		// 	err := errors.New("failed to fetch user id from token")
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// acv := &general.AuthCodeValues{
		// 	UserId:            userId,
		// 	ClientId:          clientId,
		// 	AppSname:          appShortName,
		// 	AppUserPermission: permissions,
		// }

		// return acv, accessUuid, nil
	}
	err := errors.New("failed to fetch claims from token")
	log.Println(err)
	return nil, "", err
}

// ExtractAppTokenMetadata extracts claims from redis for given app access token
func ExtractAppTokenMetadata(RedisClient *redis.Client, token *jwt.Token) (*general.AppAccessDetails, string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		accessUuid, ok := claims["token_uuid"].(string)
		if !ok {
			err := errors.New("failed to fetch access uuid from token")
			log.Println(err)
			return nil, "", err
		}

		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token type from token")
			log.Println(err)
			return nil, "", err
		}

		if tokenType != general.APPACCESS && tokenType != general.APPBACCESS {
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return nil, "", err
		}

		b, err := FetchFromRedisUsingKey(RedisClient, accessUuid)
		if err != nil {
			return nil, "", err
		}

		values := &general.AppAccessDetails{}

		err = json.Unmarshal(b, values)
		if err != nil {
			log.Println("json.Unmarshal error:", err)
			return nil, "", err
		}

		log.Println("Device Access details:", values)
		return values, accessUuid, nil

		// clientId, appShortName := "", ""
		// if tokenType == general.DEVICEACCESS {
		// 	clientId, ok = claims["client_id"].(string)
		// 	if !ok {
		// 		err := errors.New("failed to fetch client id from token")
		// 		log.Println(err)
		// 		return nil, "", err
		// 	}
		// }

		// if tokenType == general.APPACCESS || tokenType == general.APPBACCESS {
		// 	clientId, ok = claims["client_id"].(string)
		// 	if !ok {
		// 		err := errors.New("failed to fetch client id from token")
		// 		log.Println(err)
		// 		return nil, "", err
		// 	}

		// 	appShortName, ok = claims["app_shortname"].(string)
		// 	if !ok {
		// 		err := errors.New("failed to fetch app short name from token")
		// 		log.Println(err)
		// 		return nil, "", err
		// 	}
		// }

		// permissionIds, ok := claims["permission"]
		// if !ok {
		// 	err := errors.New("failed to fetch permissions from token")
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// p, err := json.Marshal(permissionIds)
		// if err != nil {
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// var permissions []string
		// err = json.Unmarshal(p, &permissions)
		// if err != nil {
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// userId, ok := claims["user_id"].(string)
		// if !ok {
		// 	err := errors.New("failed to fetch user id from token")
		// 	log.Println(err)
		// 	return nil, "", err
		// }

		// acv := &general.AuthCodeValues{
		// 	UserId:            userId,
		// 	ClientId:          clientId,
		// 	AppSname:          appShortName,
		// 	AppUserPermission: permissions,
		// }

		// return acv, accessUuid, nil
	}
	err := errors.New("failed to fetch claims from token")
	log.Println(err)
	return nil, "", err
}

// FetchAccessAuth fetches userid and refresh uuuid from redis.
// This can be used for validating token
// func FetchAccessAuth(authD *general.AccessDetails, tokenType string, client *redis.Client) error {

// 	var cursor uint64
// 	var n int
// 	var allkeys []string
// 	for {
// 		var keys []string
// 		var err error
// 		keys, cursor, err = client.Scan(cursor, authD.AccessUuid+":*", 10).Result()
// 		if err != nil {
// 			log.Println(err)
// 			return err
// 		}
// 		allkeys = append(allkeys, keys...)
// 		n += len(keys)
// 		if cursor == 0 {
// 			break
// 		}
// 	}

// 	fmt.Printf("found %d keys\n", n)
// 	log.Println("keys:", allkeys)

// 	if n == 0 {
// 		log.Println("user logged out")
// 		err := errors.New("user logged out")
// 		return err
// 	}

// 	res, err := client.Get(allkeys[0]).Result()

// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}

// 	log.Println("redis get string:", res)

// 	if tokenType == general.DEVICEACCESS {
// 		return nil
// 	}

// 	b := []byte(res)

// 	values := &general.AccessDetails{}

// 	err = json.Unmarshal(b, values)
// 	if err != nil {
// 		log.Println("json.Unmarshal error:", err)
// 		return err
// 	}

// 	log.Println("refresh values: ", values)
// 	authD.RefreshUuid = values.RefreshUuid

// 	log.Println("authD values: ", authD)

// 	return nil
// }

// FetchDeviceAuthCode fetches claims from redis for validate device token.
func FetchDeviceAuthCode(devicecode string, client *redis.Client) (*general.AuthCodeValues, error) {

	b, err := FetchFromRedisUsingKey(client, devicecode)
	if err != nil {
		return nil, err
	}
	authcodevalues := &general.AuthCodeValues{}

	err = json.Unmarshal(b, authcodevalues)
	if err != nil {
		return nil, err
	}

	return authcodevalues, nil
}

// ValidateAccessToken just returns a success response,
// since UserTokenAuthMiddleware and DeviceTokenAuthMiddleware alredy validates token.

// swagger:route GET /validate-user-acess-token ValidateToken
// End point for validate access token
//
// security:
// - Bearer: []
// responses:
//
//	200: Response
//	400: Response
func ValidateAccessToken(c *gin.Context) {
	general.RespondHandler(c.Writer, true, http.StatusOK, "valid")
}

func VerifyAccessToken(c *gin.Context) {

	ClientId := c.Request.Header.Get("Client-Id")

	if len(ClientId) == 0 {

		err := errors.New("client id missing in request")
		log.Println(err)
		general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
		return
	}

	//extract token string from request
	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		err := errors.New("access token not found")
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}

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
			authCodeValues, _, Status, err = ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPACCESS:
			authCodeValues, _, Status, err = ProcessAppAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPBACCESS:
			authCodeValues, _, Status, err = ProcessAppblockAccessToken(c)
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

		general.RespondHandler(c.Writer, true, http.StatusOK, "valid")

		return

	}

	err = errors.New("failed to fetch claims from token")
	general.RespondHandler(c.Writer, false, http.StatusBadGateway, err.Error())
	log.Println(err)
}

func ValidateIDToken(c *gin.Context) {

	_, status, err := ProcessIDToken(c)
	if err != nil {
		general.RespondHandler(c.Writer, false, status, "invalid")
		return
	}
	general.RespondHandler(c.Writer, true, http.StatusOK, "valid")
}

// RefreshToken handles the refresh token request.

// swagger:route POST /refresh-token RefreshToken
// To get new token pair by refreshing existing token pair
//
// security:
// - Key: []
// parameters:
//   - name: refresh_token
//     in: body
//     description: refresh token
//     type: string
//
// responses:
//
//	200: Response
//	400: Response
func RefreshToken(c *gin.Context) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis client handling error")
		return
	}

	at, rt := ExtractTokenPairs(c)
	if len(at) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "access token not found")
		return
	}

	if len(rt) == 0 {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "refresh token not found")
		return
	}

	token, err := jwt.Parse(at, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			log.Println(err)
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to fetch token_type from token claim")
			return
		}

		accessUuid, ok := claims["token_uuid"].(string)
		if !ok {
			err := errors.New("failed to fetch access uuid from token")
			log.Println(err)
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to fetch token_type from token claim")
			return
		}

		//var authcodeValues *general.AuthCodeValues
		// var ad *general.AccessDetails
		var Status int
		var err error
		refreshTokenType := ""

		switch tokenType {
		case general.ACCESS:
			refreshTokenType = general.REFRESH
		case general.APPACCESS:
			refreshTokenType = general.APPREFRESH
		case general.APPBACCESS:
			refreshTokenType = general.APPBREFRESH
		default:
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token is not allowed to access this")
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return
		}

		_, err = jwt.Parse(at, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			Secret := FetchSecret(tokenType)

			return []byte(Secret), nil
		})

		if err != nil {
			if !strings.Contains(err.Error(), "Token is expired") {
				log.Println(err)
				general.RespondHandler(c.Writer, false, Status, err.Error())
				return
			}

		}

		//parse and verify refresh token
		token, err = ParseAndVerifyToken(rt, refreshTokenType)
		if err != nil {
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to parse refresh token")
			log.Println("failed to parse refresh token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			refreshUuid, ok := claims["token_uuid"].(string)
			if !ok {
				general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to fetch refresh_uuid from refresh token")
				log.Println("failed to fetch refresh_uuid from refresh token")
				return
			}

			refreshTokenType, ok := claims["token_type"].(string)
			if !ok {
				general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to fetch token type from refresh token")
				log.Println("failed to fetch token type from refresh token")
				return
			}

			switch tokenType {
			case general.ACCESS:

				//extract token data

				// ad, err = ExtractUserAccessTokenMetadata(token)
				// if err != nil {
				// 	log.Println(err)
				// 	general.RespondHandler(c.Writer, false, Status, err.Error())
				// 	return
				// }

				//jnn
				if refreshTokenType != general.REFRESH {
					general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "invalid refresh token type")
					log.Println("invalid refresh token type")
					return
				}

				b, err := FetchFromRedisUsingKey(RedisClient, refreshUuid)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
					log.Println(err)
					return
				}

				values := &general.UserAccessDetails{}

				err = json.Unmarshal(b, values)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
					log.Println(err)
					return
				}

				if values.PairTokenUuid != accessUuid {
					general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token pairs not matched")
					log.Println("token pairs not matched")
					return
				}

				// return success if already refreshed
				if values.Refreshed {
					general.RespondHandler(c.Writer, true, http.StatusOK, "")
					return
				}

				ad := general.AccessDetails{
					AccessUuid:  accessUuid,
					RefreshUuid: refreshUuid,
					UserId:      values.UserId,
					DeviceId:    values.DeviceId,
				}

				//Delete the Access Token
				// if !tokenExpired {
				// 	deleted, err := DeleteToken(ad.AccessUuid+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
				// 	if err != nil || deleted == 0 {
				// 		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete access token from redis")
				// 		log.Println(err)
				// 		return
				// 	}
				// }

				//Delete the Refresh Token
				// deleted, err := DeleteToken(refreshUuid+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
				// if err != nil || deleted == 0 {
				// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete refresh token")
				// 	log.Println(err)
				// 	return
				// }

				// change refreshed flag and expiry in redis for refresh token
				rtd := &general.UserAccessDetails{}
				rtd.PairTokenUuid = ad.AccessUuid
				rtd.UserId = ad.UserId
				rtd.DeviceId = ad.DeviceId
				rtd.Refreshed = true

				// fetch refresh token expiry extension
				rtexpextStr := general.Envs["RT_EXPIRY_EXT"]

				if len(rtexpextStr) == 0 {
					log.Println("using default expiry extension for refresh token")
					rtexpextStr = "10"
				}

				rtexpextInt, err := strconv.ParseInt(rtexpextStr, 10, 64)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "internal server error")
					log.Println(err)
					return
				}

				rtexpiresext := time.Now().Add(time.Second * time.Duration(rtexpextInt)).Unix()
				err = SaveToken(refreshUuid+":"+ad.DeviceId+":"+ad.UserId, rtexpiresext, RedisClient, rtd)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to update refresh token")
					log.Println(err)
					return
				}

				//Create new pairs of refresh and access tokens for user
				tokens, err := GenerateTokenPair(RedisClient, ad.UserId, ad.DeviceId)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to create token pair")
					log.Println(err)
					return
				}

				general.RespondHandler(c.Writer, true, http.StatusOK, tokens)
				return

			case general.APPACCESS, general.APPBACCESS:

				rtType := general.APPREFRESH

				if tokenType != general.APPACCESS {
					rtType = general.APPBREFRESH
				}

				if refreshTokenType != rtType {
					general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "invalid refresh token type")
					log.Println("invalid refresh token type")
					return
				}

				b, err := FetchFromRedisUsingKey(RedisClient, refreshUuid)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
					log.Println(err)
					return
				}

				values := &general.AppAccessDetails{}

				err = json.Unmarshal(b, values)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
					log.Println(err)
					return
				}

				if values.PairTokenUuid != accessUuid {
					general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token pairs not matched")
					log.Println("token pairs not matched")
					return
				}

				// return success if already refreshed
				if values.Refreshed {
					general.RespondHandler(c.Writer, true, http.StatusOK, "")
					return
				}

				authcodeValues := &general.AuthCodeValues{
					UserId:            values.UserId,
					ClientId:          values.ClientId,
					AppSname:          values.AppSname,
					AppUserPermission: values.AppUserPermission,
					CodeName:          "",
					DeviceId:          values.DeviceId,
				}

				//Delete the Access Token
				// if !tokenExpired {
				// 	deleted, err := DeleteToken(accessUuid+":"+authcodeValues.DeviceId+":"+authcodeValues.UserId, RedisClient)
				// 	if err != nil || deleted == 0 {
				// 		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete access token from redis")
				// 		log.Println(err)
				// 		return
				// 	}
				// }

				//Delete the Refresh Token
				// deleted, err := DeleteToken(refreshUuid+":"+authcodeValues.DeviceId+":"+authcodeValues.UserId, RedisClient)
				// if err != nil || deleted == 0 {
				// 	general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete refresh token")
				// 	log.Println(err)
				// 	return
				// }

				// change refreshed flag and expiry in redis for refresh token
				rtd := &general.AppAccessDetails{
					PairTokenUuid:     values.PairTokenUuid,
					UserId:            values.UserId,
					ClientId:          values.ClientId,
					AppSname:          values.AppSname,
					AppUserPermission: values.AppUserPermission,
					DeviceId:          values.DeviceId,
					Refreshed:         true,
				}

				// fetch refresh token expiry extension
				rtexpextStr := general.Envs["RT_EXPIRY_EXT"]

				if len(rtexpextStr) == 0 {
					log.Println("using default expiry extension for refresh token")
					rtexpextStr = "10"
				}

				rtexpextInt, err := strconv.ParseInt(rtexpextStr, 10, 64)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "internal server error")
					log.Println(err)
					return
				}

				rtexpiresext := time.Now().Add(time.Second * time.Duration(rtexpextInt)).Unix()

				// saving refresh token to redis
				err = SaveToken(refreshUuid+":"+rtd.DeviceId+":"+rtd.UserId, rtexpiresext, RedisClient, rtd)
				if err != nil {
					general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to update refresh token")
					log.Println(err)
					return
				}

				var tokens general.TokenPairDetails
				if tokenType == general.APPACCESS {
					//Create new pairs of refresh and access tokens for app
					tokens, err = GenerateAppTokenPair(RedisClient, authcodeValues, false)
					if err != nil {
						general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to create token pair")
						log.Println(err)
						return
					}

				} else {

					//Create new pairs of refresh and access tokens for appblocks
					tokens, err = GenerateAppTokenPair(RedisClient, authcodeValues, true)
					if err != nil {
						general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to create token pair")
						log.Println(err)
						return
					}

				}

				general.RespondHandler(c.Writer, true, http.StatusOK, tokens)
				return

			default:
				general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token is not allowed to access this")
				err := errors.New("token is not allowed to access this")
				log.Println(err)
				return

			}

		} else {
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "refresh token expired")
		}
	}
}

// DeleteToken deletes redis entry for passing uuid.
func DeleteToken(key string, client *redis.Client) (int64, error) {
	deleted, err := client.Del(key).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// Logout handles the logout end point.
// It will deletes access and refresh token entries from redis for user logout.
// Access token deleted from redis for device

// swagger:route POST /logout Logout
// To logout
//
// security:
// - Key: []
// responses:
//
//	200: Response
//	400: Response
func Logout(c *gin.Context) {
	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis client handling error")
		return
	}

	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		tokenString = ExtractCookieToken(c, general.Envs["ID_TOKEN_NAME"])
		if len(tokenString) == 0 {
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token not found")
			return
		}

	}

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to fetch token_type from token claim")
			return
		}

		var ad *general.AccessDetails
		// var authcodevalues *general.AuthCodeValues
		var Status int
		var err error
		switch tokenType {
		case general.IDTOKEN:
			ad, Status, err = ProcessIDToken(c)

			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.ACCESS:
			ad, Status, err = ProcessUserAccessToken(c)

			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.DEVICEACCESS:
			_, ad, Status, err = ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPACCESS:
			_, ad, Status, err = ProcessAppAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPBACCESS:
			_, ad, Status, err = ProcessAppblockAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		default:
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token is not allowed to access this")
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return

		}

		deleted, err := DeleteToken(ad.AccessUuid+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
		if err != nil || deleted == 0 {
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete access token from redis")
			log.Println(err)
			return
		}

		if tokenType != general.DEVICEACCESS && tokenType != general.IDTOKEN {
			deleted, err = DeleteToken(ad.RefreshUuid+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
			if err != nil || deleted == 0 {
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete refresh token from redis")
				log.Println(err)
				return
			}
		}

		var cursor uint64
		var n int
		var allkeys []string
		for {
			var keys []string
			var err error
			keys, cursor, err = RedisClient.Scan(cursor, "*:"+ad.DeviceId+":"+ad.UserId, 10).Result()
			if err != nil {
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
				log.Println(err)
				return
			}
			allkeys = append(allkeys, keys...)
			n += len(keys)
			if cursor == 0 {
				break
			}
		}

		for _, v := range allkeys {
			deleted, err = DeleteToken(v, RedisClient)
			if err != nil || deleted == 0 {
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete tokens from redis")
				log.Println(err)
				return
			}
		}

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

		general.RespondHandler(c.Writer, true, http.StatusOK, "successfully logged out")
		return

	}

	general.RespondHandler(c.Writer, false, http.StatusBadRequest, "failed to fetch token claims")

}

// LogoutFromAll handles the logout-from-all end point.
// It will deletes access and refresh token entries created for user from redis.

// swagger:route POST /logout-from-all LogoutFromAll
// To logout from all
//
// security:
// - Key: []
// responses:
//
//	200: Response
//	400: Response
func LogoutFromAll(c *gin.Context) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		log.Printf("redis client handling error")
		general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "redis client handling error")
		return
	}

	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		tokenString = ExtractCookieToken(c, general.Envs["ID_TOKEN_NAME"])
		if len(tokenString) == 0 {
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token not found")
			return
		}
	}

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		general.RespondHandler(c.Writer, false, http.StatusUnauthorized, err.Error())
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to fetch token_type from token claim")
			return
		}

		var ad *general.AccessDetails
		// var authcodevalues *general.AuthCodeValues
		var Status int
		var err error
		switch tokenType {
		case general.IDTOKEN:
			ad, Status, err = ProcessIDToken(c)

			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.ACCESS:
			ad, Status, err = ProcessUserAccessToken(c)

			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.DEVICEACCESS:
			_, ad, Status, err = ProcessDeviceAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPACCESS:
			_, ad, Status, err = ProcessAppAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		case general.APPBACCESS:
			_, ad, Status, err = ProcessAppblockAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				log.Println(err)
				return
			}
		default:
			general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "token is not allowed to access this")
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return

		}

		deleted, err := DeleteToken(ad.AccessUuid+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
		if err != nil || deleted == 0 {
			general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete access token from redis")
			log.Println(err)
			return
		}

		if tokenType != general.DEVICEACCESS && tokenType != general.IDTOKEN {
			deleted, err = DeleteToken(ad.RefreshUuid+":"+ad.DeviceId+":"+ad.UserId, RedisClient)
			if err != nil || deleted == 0 {
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete refresh token from redis")
				log.Println(err)
				return
			}
		}

		var cursor uint64
		var n int
		var allkeys []string
		for {
			var keys []string
			var err error
			keys, cursor, err = RedisClient.Scan(cursor, "*:"+ad.UserId, 10).Result()
			if err != nil {
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, err.Error())
				log.Println(err)
				return
			}
			allkeys = append(allkeys, keys...)
			n += len(keys)
			if cursor == 0 {
				break
			}
		}

		for _, v := range allkeys {
			deleted, err = DeleteToken(v, RedisClient)
			if err != nil || deleted == 0 {
				general.RespondHandler(c.Writer, false, http.StatusInternalServerError, "failed to delete tokens from redis")
				log.Println(err)
				return
			}
		}

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

		general.RespondHandler(c.Writer, true, http.StatusOK, "successfully logged out")
		return

	}

	general.RespondHandler(c.Writer, false, http.StatusBadRequest, "failed to fetch token claims")

}

// ProcessIDToken extract user id token string, parse and verify token string, extracts token data and return token details stored in redis
func ProcessIDToken(c *gin.Context) (*general.AccessDetails, int, error) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		err := errors.New("redis client handling error")
		return nil, http.StatusInternalServerError, err
	}

	//extract token string from request
	tokenString := ExtractCookieToken(c, general.Envs["ID_TOKEN_NAME"])
	if len(tokenString) == 0 {
		err := errors.New("id token not found")
		return nil, http.StatusUnauthorized, err
	}

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		return nil, http.StatusBadRequest, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {

		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token_type from token claim")
			return nil, http.StatusUnauthorized, err
		}

		if tokenType != general.IDTOKEN {
			err := errors.New("invalid token type")
			return nil, http.StatusUnauthorized, err
		}

		//parse and verify token
		token, err := ParseAndVerifyToken(tokenString, tokenType)
		if err != nil {
			//general.RespondHandler(c.Writer, false, http.StatusBadRequest, "failed to parse access token")
			log.Println("failed to parse access token after ParseAndVerifyToken")
			return nil, http.StatusUnauthorized, err
		}

		//extract token data
		ad, err := ExtractIDTokenMetadata(RedisClient, token)
		if err != nil {
			//general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
			log.Println("failed to parse access token after ExtractUserAccessTokenMetadata")
			return nil, http.StatusUnauthorized, err
		}

		return ad, http.StatusOK, nil
	}

	err = errors.New("failed to fetch claims from token")

	return nil, http.StatusBadGateway, err
}

// ProcessUserAccessToken extract user access token string, parse and verify token string, extracts token data and return access token details stored in redis
func ProcessUserAccessToken(c *gin.Context) (*general.AccessDetails, int, error) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		err := errors.New("redis client handling error")
		return nil, http.StatusInternalServerError, err
	}

	//extract token string from request
	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		//general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "access token not found")
		err := errors.New("access token not found")
		return nil, http.StatusUnauthorized, err
	}

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		return nil, http.StatusBadRequest, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {

		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token_type from token claim")
			return nil, http.StatusUnauthorized, err
		}

		if tokenType != general.ACCESS {
			err := errors.New("invalid token type")
			return nil, http.StatusUnauthorized, err
		}

		//parse and verify token
		token, err := ParseAndVerifyToken(tokenString, tokenType)
		if err != nil {
			//general.RespondHandler(c.Writer, false, http.StatusBadRequest, "failed to parse access token")
			log.Println("failed to parse access token after ParseAndVerifyToken")
			return nil, http.StatusUnauthorized, err
		}

		//extract token data
		ad, err := ExtractUserAccessTokenMetadata(RedisClient, token)
		if err != nil {
			//general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
			log.Println("failed to parse access token after ExtractUserAccessTokenMetadata")
			return nil, http.StatusUnauthorized, err
		}

		return ad, http.StatusOK, nil
	}

	err = errors.New("failed to fetch claims from token")

	return nil, http.StatusBadGateway, err
}

// ProcessDeviceAccessToken extract device access token string, parse and verify token string, extracts token data and return access token details
func ProcessDeviceAccessToken(c *gin.Context) (*general.AuthCodeValues, *general.AccessDetails, int, error) {
	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		err := errors.New("redis client handling error")
		return nil, nil, http.StatusInternalServerError, err
	}

	//extract token string from request
	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		err := errors.New("access token not found")
		return nil, nil, http.StatusUnauthorized, err
	}

	// log.Println("token string:", tokenString)

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		return nil, nil, http.StatusBadRequest, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token_type from token claim")
			return nil, nil, http.StatusUnauthorized, err
		}

		if tokenType != general.DEVICEACCESS {
			err := errors.New("invalid token type")
			return nil, nil, http.StatusUnauthorized, err
		}

		log.Println("token type b4 ParseAndVerifyToken:", tokenType)
		//parse and verify token
		token, err := ParseAndVerifyToken(tokenString, tokenType)
		if err != nil {
			log.Println("failed to parse access token after ParseAndVerifyToken")
			return nil, nil, http.StatusUnauthorized, err
		}

		//extract token data
		authCodeValues, accessUuid, err := ExtractDeviceTokenMetadata(RedisClient, token)
		if err != nil {
			log.Println("failed to parse access token after ExtractDeviceTokenMetadata")
			return nil, nil, http.StatusBadRequest, err
		}

		accessdetails := general.AccessDetails{
			AccessUuid:  accessUuid,
			RefreshUuid: "",
			UserId:      authCodeValues.UserId,
			DeviceId:    authCodeValues.DeviceId,
		}

		return authCodeValues, &accessdetails, http.StatusOK, nil
	}

	err = errors.New("failed to fetch claims from token")

	return nil, nil, http.StatusBadGateway, err
}

// ProcessAppAccessToken extract device access token string, parse and verify token string, extracts token data and return access token details
func ProcessAppAccessToken(c *gin.Context) (*general.AuthCodeValues, *general.AccessDetails, int, error) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		err := errors.New("redis client handling error")
		return nil, nil, http.StatusInternalServerError, err
	}

	//extract token string from request
	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		//general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "access token not found")
		err := errors.New("access token not found")
		return nil, nil, http.StatusUnauthorized, err
	}

	// log.Println("token string:", tokenString)

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		return nil, nil, http.StatusBadRequest, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			// general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to fetch token_type from token claim")
			err := errors.New("failed to fetch token_type from token claim")
			return nil, nil, http.StatusUnauthorized, err
		}

		if tokenType != general.APPACCESS {
			//general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "invalid token type")
			err := errors.New("invalid token type")
			return nil, nil, http.StatusUnauthorized, err
		}

		log.Println("token type b4 ParseAndVerifyToken:", tokenType)
		//parse and verify token
		token, err := ParseAndVerifyToken(tokenString, tokenType)
		if err != nil {
			//general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "failed to parse access token")
			log.Println("failed to parse access token after ParseAndVerifyToken")
			return nil, nil, http.StatusUnauthorized, err
		}

		//extract token data
		appaccessdetails, accessUuid, err := ExtractAppTokenMetadata(RedisClient, token)
		if err != nil {
			//general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
			log.Println("failed to parse access token after ExtractDeviceTokenMetadata")
			return nil, nil, http.StatusBadRequest, err
		}

		accessdetails := general.AccessDetails{
			AccessUuid:  accessUuid,
			RefreshUuid: appaccessdetails.PairTokenUuid,
			UserId:      appaccessdetails.UserId,
			DeviceId:    appaccessdetails.DeviceId,
		}

		authCodeValues := general.AuthCodeValues{
			UserId:            appaccessdetails.UserId,
			ClientId:          appaccessdetails.ClientId,
			AppSname:          appaccessdetails.AppSname,
			AppUserPermission: appaccessdetails.AppUserPermission,
			CodeName:          "",
			DeviceId:          appaccessdetails.DeviceId,
		}

		return &authCodeValues, &accessdetails, http.StatusOK, nil
	}

	err = errors.New("failed to fetch claims from token")

	return nil, nil, http.StatusBadGateway, err
}

// ProcessAppblockAccessToken extract appblocks access token string, parse and verify token string, extracts token data and return access token details
func ProcessAppblockAccessToken(c *gin.Context) (*general.AuthCodeValues, *general.AccessDetails, int, error) {

	RedisClient, ok := c.MustGet("RedisClient").(*redis.Client)
	if !ok {
		err := errors.New("redis client handling error")
		return nil, nil, http.StatusInternalServerError, err
	}

	//extract token string from request
	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		//general.RespondHandler(c.Writer, false, http.StatusUnauthorized, "access token not found")
		err := errors.New("access token not found")
		return nil, nil, http.StatusUnauthorized, err
	}

	// log.Println("token string:", tokenString)

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		return nil, nil, http.StatusBadRequest, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token_type from token claim")
			return nil, nil, http.StatusUnauthorized, err
		}

		if tokenType != general.APPBACCESS {
			err := errors.New("invalid token type")
			return nil, nil, http.StatusUnauthorized, err
		}

		log.Println("token type b4 ParseAndVerifyToken:", tokenType)
		//parse and verify token
		token, err := ParseAndVerifyToken(tokenString, tokenType)
		if err != nil {
			log.Println("failed to parse access token after ParseAndVerifyToken")
			return nil, nil, http.StatusUnauthorized, err
		}

		//extract token data
		appaccessdetails, accessUuid, err := ExtractAppTokenMetadata(RedisClient, token)
		if err != nil {
			//general.RespondHandler(c.Writer, false, http.StatusBadRequest, err.Error())
			log.Println("failed to parse access token after ExtractDeviceTokenMetadata")
			return nil, nil, http.StatusBadRequest, err
		}

		accessdetails := general.AccessDetails{
			AccessUuid:  accessUuid,
			RefreshUuid: appaccessdetails.PairTokenUuid,
			UserId:      appaccessdetails.UserId,
			DeviceId:    appaccessdetails.DeviceId,
		}

		authCodeValues := general.AuthCodeValues{
			UserId:            appaccessdetails.UserId,
			ClientId:          appaccessdetails.ClientId,
			AppSname:          appaccessdetails.AppSname,
			AppUserPermission: appaccessdetails.AppUserPermission,
			CodeName:          "",
			DeviceId:          appaccessdetails.DeviceId,
		}

		return &authCodeValues, &accessdetails, http.StatusOK, nil
	}

	err = errors.New("failed to fetch claims from token")

	return nil, nil, http.StatusBadGateway, err
}

// UserTokenAuthMiddleware used to verify the user token, to secure the routs.
func UserTokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, Status, err := ProcessUserAccessToken(c)
		if err != nil {
			general.RespondHandler(c.Writer, false, Status, err.Error())
			c.Abort()
			log.Println("failed to parse access token")
			return
		}
		c.Next()
	}
}

// DeviceTokenAuthMiddleware used to verify the device or app token, to secure the routs.
func DeviceTokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, Status, err := ProcessDeviceAccessToken(c)
		if err != nil {
			general.RespondHandler(c.Writer, false, Status, err.Error())
			c.Abort()
			log.Println("failed to parse access token")
			return
		}
		c.Next()
	}
}

// AppblockTokenAuthMiddleware used to verify the appblocks token, to secure the routs.
func AppblockTokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, Status, err := ProcessAppblockAccessToken(c)
		if err != nil {
			general.RespondHandler(c.Writer, false, Status, err.Error())
			c.Abort()
			log.Println("failed to parse access token")
			return
		}
		c.Next()
	}
}

// AppblockTokenAuthMiddleware used to verify the appblocks token, to secure the routs.
func AppAndAppblockTokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, Status, err := ProcessAppblockAccessToken(c)
		if err != nil && strings.Contains(err.Error(), "invalid token type") {
			_, _, Status, err := ProcessAppAccessToken(c)
			if err != nil {
				general.RespondHandler(c.Writer, false, Status, err.Error())
				c.Abort()
				log.Println("failed to parse access token")
				return
			}
		} else if err != nil {
			general.RespondHandler(c.Writer, false, Status, err.Error())
			c.Abort()
			log.Println("failed to parse access token")
			return
		}
		c.Next()
	}
}

// AppTokenAuthMiddleware used to verify the appblocks token, to secure the routs.
func AppTokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, Status, err := ProcessAppAccessToken(c)
		if err != nil {
			general.RespondHandler(c.Writer, false, Status, err.Error())
			c.Abort()
			log.Println("failed to parse access token")
			return
		}
		c.Next()
	}
}

// GetDetailsFromIDToken returns Access details from token
func GetDetailsFromIDToken(c *gin.Context) (*general.AccessDetails, int, error) {
	// extracts UserId from token
	ad, Status, err := ProcessIDToken(c)
	if err != nil {
		log.Println(err)
		return nil, Status, err
	}

	return ad, Status, nil
}

// GetUserIdFromAccessToken returns userid from token
func GetUserIdFromAccessToken(c *gin.Context) (string, int, error) {
	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		err := errors.New("access token not found")
		return "", http.StatusUnauthorized, err
	}

	token, err := jwt.Parse(tokenString, nil)
	if err != nil && !(strings.Contains(err.Error(), "no Keyfunc was provided")) {
		return "", http.StatusBadRequest, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		tokenType, ok := claims["token_type"].(string)
		if !ok {
			err := errors.New("failed to fetch token_type from token claim")
			return "", http.StatusUnauthorized, err
		}

		var ad *general.AccessDetails
		var Status int
		var err error
		switch tokenType {
		// case general.ACCESS:
		// 	ad, Status, err = ProcessUserAccessToken(c)

		// 	if err != nil {
		// 		log.Println(err)
		// 		return "", Status, err
		// 	}
		case general.DEVICEACCESS:
			_, ad, Status, err = ProcessDeviceAccessToken(c)
			if err != nil {
				log.Println(err)
				return "", Status, err
			}
		case general.APPACCESS:
			_, ad, Status, err = ProcessAppAccessToken(c)
			if err != nil {
				log.Println(err)
				return "", Status, err
			}
		case general.APPBACCESS:
			_, ad, Status, err = ProcessAppblockAccessToken(c)
			if err != nil {
				log.Println(err)
				return "", Status, err
			}
		default:
			err := errors.New("token is not allowed to access this")
			log.Println(err)
			return "", http.StatusUnauthorized, err

		}

		return ad.UserId, http.StatusOK, nil

	}

	err = errors.New("failed to fetch claims from token")

	return "", http.StatusBadGateway, err
}
