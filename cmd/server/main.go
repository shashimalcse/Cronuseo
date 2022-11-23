package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shashimalcse/cronuseo/internal/config"
	"github.com/shashimalcse/cronuseo/internal/keto"
	"github.com/shashimalcse/cronuseo/internal/organization"
	"github.com/shashimalcse/cronuseo/internal/permission"
	"github.com/shashimalcse/cronuseo/internal/resource"
	"github.com/shashimalcse/cronuseo/internal/role"
	"github.com/shashimalcse/cronuseo/internal/user"
	"google.golang.org/grpc"

	_ "github.com/shashimalcse/cronuseo/docs"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	acl "github.com/ory/keto/proto/ory/keto/acl/v1alpha1"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var Version = "1.0.0"

// var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")
var flagConfig = flag.String("config", "/Users/thilinashashimal/Desktop/Cronuseo/config/local.yml", "path to the config file")

// @title          Cronuseo API
// @version        1.0
// @description    This is a sample server celler server.
// @termsOfService http://swagger.io/terms/

// @contact.name  API Support
// @contact.url   http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8080
// @BasePath /api/v1
func main() {
	flag.Parse()
	// load application configurations
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		log.Fatal("error loading config")
		os.Exit(-1)
	}

	//connect db
	db, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		log.Fatalln("error connecting databse")
		os.Exit(-1)
	}

	//keto clients

	conn, err := grpc.Dial("127.0.0.1:4467", grpc.WithInsecure())
	if err != nil {
		panic("Encountered error: " + err.Error())
	}

	writeClient := acl.NewWriteServiceClient(conn)

	conn, err = grpc.Dial("127.0.0.1:4466", grpc.WithInsecure())
	if err != nil {
		panic("Encountered error: " + err.Error())
	}
	readClient := acl.NewReadServiceClient(conn)

	conn, err = grpc.Dial("127.0.0.1:4466", grpc.WithInsecure())
	if err != nil {
		panic("Encountered error: " + err.Error())
	}
	checkClient := acl.NewCheckServiceClient(conn)

	clients := keto.KetoClients{WriteClient: writeClient, ReadClient: readClient, CheckClient: checkClient}

	e := buildHandler(db, cfg, clients)
	address := fmt.Sprintf(":%v", cfg.ServerPort)
	e.Logger.Fatal(e.Start(address))

}

func buildHandler(db *sqlx.DB, cfg *config.Config, clients keto.KetoClients) *echo.Echo {
	router := echo.New()
	router.Use(middleware.CORS())
	router.GET("/swagger/*", echoSwagger.WrapHandler)
	rg := router.Group("/api/v1")
	organization.RegisterHandlers(rg, organization.NewService(organization.NewRepository(db)))
	user.RegisterHandlers(rg, user.NewService(user.NewRepository(db)))
	resource.RegisterHandlers(rg, resource.NewService(resource.NewRepository(db)))
	role.RegisterHandlers(rg, role.NewService(role.NewRepository(db)))
	permission.RegisterHandlers(rg, permission.NewService(permission.NewRepository(db)))
	keto.RegisterHandlers(rg, keto.NewService(clients))
	return router
}
