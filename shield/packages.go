package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/appblocks-hub/SHIELD/common_services"
	"github.com/appblocks-hub/SHIELD/functions/appreg"
	gen "github.com/appblocks-hub/SHIELD/shield_gen/go/proxy"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var funcs map[string]interface{}
var db *gorm.DB

func startGRPCServer() {
	//Load func map
	funcloadErr := loadFuncs()
	if funcloadErr != nil {
		log.Fatalf("Error loading function map %v", funcloadErr)
	}

	//Initialise common db object for grpc handlers invocation

	// create new gRPC servera
	server := grpc.NewServer()

	// register the GreeterServerImpl on the gRPC server
	gen.RegisterShieldProxyServer(server, &Service{})

	// start listening on port :8080 for a tcp connection

	listener, err := net.Listen("tcp", os.Getenv("SHIELD_GRPC_PORT"))
	if err != nil {
		log.Fatalf("could not attach listener to port: %v", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("could not start grpc server: %v", err)
		}
	}()
}

func InvokeShieldFunction(funcs map[string]interface{}, payload common_services.HandlerPayload) common_services.HandlerResponse {
	f, functionExists := funcs[payload.Url]
	fmt.Printf("function exists%v", payload.Url)
	if !functionExists {
		return common_services.HandlerResponse{Data: "", Status: 404, Err: true}
	}

	timed := InvokeGrpcFunction(f).(func(common_services.HandlerPayload) common_services.HandlerResponse)

	result := timed(common_services.HandlerPayload{Url: payload.Url, UserID: payload.UserID, RequestBody: payload.RequestBody, Db: db, Queryparams: payload.Queryparams})

	return result
}

func loadFuncs() error {
	funcs = map[string]interface{}{

		"/app-registration":        appreg.AppRegistration,
		"/update-app-redirect-url": appreg.UpdateAppRedirectUrl,
		"/get-app-redirect-url":    appreg.GetAppRedirectUrl,
		"/get-app-client-id":       appreg.GetAppClientId,
		"/get-app-client-secret":   appreg.GetAppClientSecret,
		"/get-app-scopes":          appreg.GetAppScopes,
		"/update-app-scopes":       appreg.UpdateAppScopes,
	}

	return nil
}

func InvokeGrpcFunction(f interface{}) interface{} {
	rf := reflect.TypeOf(f)
	if rf.Kind() != reflect.Func {
		panic("expects a function")
	}
	vf := reflect.ValueOf(f)
	wrapperF := reflect.MakeFunc(rf, func(in []reflect.Value) []reflect.Value {
		start := time.Now()
		out := vf.Call(in)
		end := time.Now()
		fmt.Printf("calling %s took %v\n", runtime.FuncForPC(vf.Pointer()).Name(), end.Sub(start))
		return out
	})
	return wrapperF.Interface()
}

func DBInit() {
	dbinf := &common_services.DBInfo{}
	var dbErr error

	dbinf.Host = os.Getenv("SHIELD_POSTGRES_HOST")
	dbinf.User = os.Getenv("SHIELD_POSTGRES_USER")
	dbinf.Password = os.Getenv("SHIELD_POSTGRES_PASSWORD")
	dbinf.Name = os.Getenv("SHIELD_POSTGRES_NAME")
	dbinf.Port = os.Getenv("SHIELD_POSTGRES_PORT")
	dbinf.Sslmode = os.Getenv("SHIELD_POSTGRES_SSLMODE")
	dbinf.Timezone = os.Getenv("SHIELD_POSTGRES_TIMEZONE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", dbinf.Host, dbinf.User, dbinf.Password, dbinf.Name, dbinf.Port, dbinf.Sslmode, dbinf.Timezone)
	db, dbErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if dbErr != nil {
		panic("DB connection err")
	}
}

func CloseDbCOnn() {
	//closing connection to db
	sqlDB, dberr := db.DB()
	if dberr != nil {
		log.Fatalln(dberr)
	}
	defer sqlDB.Close()

}
