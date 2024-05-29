package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/appblocks-hub/SHIELD/common_services"
	"github.com/appblocks-hub/SHIELD/functions/appreg"
	"github.com/appblocks-hub/SHIELD/functions/auth"
	_ "github.com/appblocks-hub/SHIELD/functions/docs"
	"github.com/appblocks-hub/SHIELD/functions/general"
	"github.com/appblocks-hub/SHIELD/functions/token"
	"github.com/appblocks-hub/SHIELD/functions/user"
	gen "github.com/appblocks-hub/SHIELD/shield_gen/go/proxy"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"google.golang.org/grpc/metadata"
)

type Service struct {
	gen.UnimplementedShieldProxyServer
}

func (g *Service) ShieldCallService(ctx context.Context, request *gen.ShieldRequest) (*gen.ShieldReply, error) {

	// parsing body.
	var userIdValues []string
	var userID string
	var urlValues []string
	var url string
	var userNameValues []string
	var userName string
	var clientIdValues []string
	var clientId string
	var clientSecretValues []string
	var clientSecret string

	//decoding url and userID from request context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		userIdValues = md.Get("user-id")
		urlValues = md.Get("url")
		userNameValues = md.Get("user-name")
		clientIdValues = md.Get("client-id")
		clientSecretValues = md.Get("client-secret")
	}

	if len(userIdValues) > 0 {
		userID = userIdValues[0]
	}
	if len(urlValues) > 0 {
		url = urlValues[0]
	}
	if len(userNameValues) > 0 {
		userName = urlValues[0]
	}
	if len(clientIdValues) > 0 {
		clientId = clientIdValues[0]
	}
	if len(clientSecretValues) > 0 {
		clientSecret = clientSecretValues[0]
	}

	request.Queryparams = make(map[string]string)

	request.Queryparams["client-id"] = clientId
	request.Queryparams["client-secret"] = clientSecret

	invReply := InvokeShieldFunction(funcs, common_services.HandlerPayload{Url: url, RequestBody: request.Body, UserID: userID, UserName: userName, Queryparams: request.Queryparams})

	// result := timed(InvokeFunctionPayload{UserID: userID, Url: url, RequestBody: request.Body})

	fmt.Printf("invite reply is %v", invReply)

	// // converting pb struct to Map Interface.
	// body := request.Body.AsMap()

	// // convert map to json string
	// jsonString, _ := json.Marshal(body)
	// s := (string(jsonString))
	// log.Println(s)
	// log.Println(reflect.TypeOf(s))

	// // converting json string to struct
	// as_struct := Body{}
	// err := json.Unmarshal([]byte(s), &as_struct)
	// if err != nil {
	// 	log.Println(err.Error())
	// }

	// log.Println(as_struct)
	// log.Println("#####################################")

	if err := request.Validate(); err != nil {
		return nil, err
	}

	// strWrapper := wrapperspb.String("foo")
	// strAny, _ := anypb.New(strWrapper)

	return &gen.ShieldReply{
		Err:    invReply.Err,
		Data:   invReply.Data,
		Status: int32(invReply.Status),
	}, nil
}

func RedisMiddleware(RedisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("RedisClient", RedisClient)
		c.Next()
	}
}

func main() {

	redisDBValue, err := strconv.Atoi(general.Envs["SHIELD_REDIS_DB"])
	if err != nil {
		log.Fatalf("Error redisDBValue convert: %v", err)
	}
	RedisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", general.Envs["SHIELD_REDIS_HOST"], general.Envs["SHIELD_REDIS_PORT"]),
		Password: general.Envs["SHIELD_REDIS_PASSWORD"],
		DB:       redisDBValue,
	})

	defer RedisClient.Close()

	// Ping the Redis server to confirm the connection
	_, err = RedisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Error Redis ping: %v", err)
	}

	router := gin.Default()

	router.Use(RedisMiddleware(RedisClient))

	var allowedOrigins []string

	allowedOrigins = append(allowedOrigins, strings.Split(general.Envs["ALLOWED_DOMAINS"], ",")...)

	//router.Use(cors.Default())
	router.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Client-Id", "Client-Secret", "Access-Control-Allow-Origin"},
		//ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/assets/css", "./static/assets/css")
	router.Static("/assets/img", "./static/assets/img")
	router.Static("/js", "./static/js")

	router.GET("/signup", auth.GetSignup)
	router.POST("/signup", user.Signup)
	router.POST("/login", user.Login)
	router.GET("/login", auth.GetLogin)

	router.GET("/validate-idt", token.ValidateIDToken)

	router.GET("/validate-user-acess-token", token.UserTokenAuthMiddleware(), token.ValidateAccessToken)
	router.POST("/refresh-token", token.RefreshToken)
	router.POST("/logout", token.Logout)
	router.POST("/logout-from-all", token.LogoutFromAll)

	router.GET("/password-recovery", auth.GetPasswordRecovery)
	router.POST("/password-recovery", auth.PasswordRecovery)

	router.GET("/change-password", auth.GetChangePassword)
	router.POST("/change-password", auth.ChangePassword)
	router.POST("/change-user-password", auth.ChangeUserPassword)

	router.PATCH("/update-user-profile", user.UpdateUserProfile)

	router.POST("/block-app-registration", appreg.BlockAppRegistration)
	router.POST("/update-block-app-redirect-url", appreg.UpdateBlockAppRedirectUrl)
	router.POST("/get-block-app-redirect-url", appreg.GetBlockAppRedirectUrl)
	router.POST("/get-block-app-client-id", appreg.GetBlockAppClientId)
	router.POST("/get-block-app-client-secret", appreg.GetBlockAppClientSecret)
	router.POST("/get-block-app-scopes", appreg.GetBlockAppScopes)
	router.POST("/update-block-app-scopes", appreg.UpdateBlockAppScopes)

	router.GET("/get-all-permissions", token.UserTokenAuthMiddleware(), appreg.GetAllPermissions)
	router.POST("/create-app-permissions", token.AppblockTokenAuthMiddleware(), appreg.CreateAppPermissions)

	router.POST("/allow-permissions" /*token.UserTokenAuthMiddleware(),*/, auth.CreateShieldAppUserPermissions)

	router.GET("/get-app-permissions-for-allow" /* token.UserTokenAuthMiddleware(), */, auth.GetAllAppPermissionsForAllow)

	router.POST("/allow-app-permissions" /*token.UserTokenAuthMiddleware(),*/, auth.CreateAppUserPermissions)

	router.GET("/auth/app/open", auth.OpenApp)

	//router.GET("/auth/device-authorize", auth.AuthHandler)
	router.GET("/auth/get-token", auth.CreateAppblockTokens)
	router.GET("/auth/device/get-token", auth.CreateDeviceAccessToken)

	router.GET("/auth/google/signup", auth.GoogleSignup)
	router.GET("/auth/google/login", auth.GoogleLogin)
	router.GET("/auth/google/callback", auth.GoogleAuth)

	router.GET("/auth/twitter/signup", auth.TwitterSignup)
	router.GET("/auth/twitter/login", auth.TwitterLogin)
	router.GET("/auth/twitter/callback", auth.TwitterAuth)

	router.GET("/auth/linkedin/signup", auth.LinkedinSignup)
	router.GET("/auth/linkedin/login", auth.LinkedinLogin)
	router.GET("/auth/linkedin/callback", auth.LinkedinAuth)

	router.GET("/validate-device-app-acess-token", token.DeviceTokenAuthMiddleware(), token.ValidateAccessToken)

	router.GET("/validate-appblocks-acess-token", token.AppAndAppblockTokenAuthMiddleware(), token.ValidateAccessToken)

	// common for App/Appb/Device tokens
	router.POST("/verify-appblocks-acess-token", token.VerifyAccessToken)

	router.GET("/validate-app-acess-token", token.AppTokenAuthMiddleware(), token.ValidateAccessToken)

	router.POST("/device/verify-token", user.VerifyTokenAndGetEmailForDevice)
	router.GET("/device/get-email", user.GetEmailForDevice)

	router.GET("/get-user-id", user.GetUserId)
	router.GET("/get-user-details", user.GetUserDetails)

	router.POST("/get-uid", user.GetUid)
	router.POST("/get-user", user.GetUser)

	router.GET("/verify-user-email", auth.GetVerifyEmail)
	router.POST("/verify-user-email", user.VerfiyUserEmail)

	router.POST("/resend-user-email-otp", user.ResendUserEmailOTP)

	router.GET("/display-error-message", auth.DisplayErrorMessage)

	router.GET("/authorize-appname-permission", auth.DisplayAuthorizeAppnamePermission)
	router.GET("/authorize-appname", auth.DisplayAuthorizeAppname)

	log.Println("Running...")
	if general.Envs["GRPC_ENABLE"] == "1" {
		// Load env File ==>
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}

		DBInit()
		// defer CloseDbCOnn()

		startGRPCServer()
	}
	router.Run(general.Envs["SHIELD_HTTP_PORT"])

	//http.ListenAndServe(":8000", router)
}
