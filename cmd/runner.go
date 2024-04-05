package main

import (
	"BrainBlitz.com/game/config"
	"BrainBlitz.com/game/internal/controller"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/internal/core/server/http"
	coreService "BrainBlitz.com/game/internal/core/service"
	"BrainBlitz.com/game/internal/core/service/backofficeUserHandler"
	mysqlConfig "BrainBlitz.com/game/internal/infra/config"
	"BrainBlitz.com/game/internal/infra/repository"
	repository3 "BrainBlitz.com/game/internal/infra/repository/authorization"
	"BrainBlitz.com/game/internal/infra/repository/mongo"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// TODO - read config path from command line
	cfg := config.Load("config.yml")
	fmt.Printf("cfg: %+v\n", cfg)

	// Create a new instance of the Gin router
	ginInstance := gin.New()
	ginInstance.Use(gin.Recovery())

	db, err := repository.NewDB(mysqlConfig.DatabaseConfig{
		Driver:                 "mysql",
		Url:                    "bbGame:root@tcp(127.0.0.1:3310)/brainBlitz_db?charset=utf8mb4&parseTime=true&loc=UTC&tls=false&readTimeout=3s&writeTimeout=3s&timeout=3s&clientFoundRows=true",
		ConnMaxLifeTimeMinutes: 3,
		MaxOpenCons:            10,
		MaxIdleCons:            1,
	})
	if err != nil {
		log.Fatalf("failed to new database err=%s\n", err.Error())
	}

	//create the UserRepository
	userRepo := repository.NewUserRepository(db)
	backofficeRepo := repository.New(db)

	//create the user service
	authService := coreService.NewJWTAuthService("salam", "exp", time.Now().Add(time.Hour*24).Unix(), time.Now().Add(time.Hour*24*7).Unix())
	uService := coreService.NewUserService(userRepo, authService)
	mongoDB, err := mongo.NewMongoDB()
	if err != nil {
		log.Fatal("cant connect to mongo", err)
	}
	authorizationRepo := repository3.NewAuthorizationRepo(mongoDB)
	authorizationService := coreService.NewAuthorizationService(authorizationRepo)
	backofficeHandler := backofficeUserHandler.New(backofficeRepo)
	controllerServices := service.Service{
		UserService:           uService,
		BackofficeUserService: backofficeHandler,
		AuthService:           authService,
		AuthorizationService:  authorizationService,
	}
	httpController := controller.NewController(ginInstance, controllerServices)
	httpController.InitRouter()

	//create httpServer
	httpServer := http.NewHTTPServer(ginInstance, mysqlConfig.HttpServerConfig{
		Port: 8000,
	})

	httpServer.Start()
	defer httpServer.Stop()

	// Listen for OS signals to perform a graceful shutdown
	log.Printf("listening signals on %d ...", os.Getpid())
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	<-c
	log.Println("graceful shutdown...")
}
