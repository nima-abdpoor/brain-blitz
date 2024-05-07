package main

import (
	"BrainBlitz.com/game/config"
	"BrainBlitz.com/game/internal/controller"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/internal/core/server/http"
	coreService "BrainBlitz.com/game/internal/core/service"
	"BrainBlitz.com/game/internal/core/service/backofficeUserHandler"
	matchMakingHandler "BrainBlitz.com/game/internal/core/service/matchMaking"
	mysqlConfig "BrainBlitz.com/game/internal/infra/config"
	"BrainBlitz.com/game/internal/infra/repository"
	repository3 "BrainBlitz.com/game/internal/infra/repository/authorization"
	"BrainBlitz.com/game/internal/infra/repository/matchmaking"
	"BrainBlitz.com/game/internal/infra/repository/mongo"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"BrainBlitz.com/game/scheduler"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// TODO - read config path from command line
	cfg := config.Load("config.yml")
	fmt.Printf("cfg: %+v\n", cfg)

	// Create a new instance of the Echo router
	echoInstance := echo.New()
	echoInstance.Use(middleware.Logger())
	echoInstance.Use(middleware.Recover())

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

	// backoffice
	backofficeRepo := repository.New(db)
	backofficeHandler := backofficeUserHandler.New(backofficeRepo)

	//create the user service
	authService := coreService.NewJWTAuthService("salam", "exp", time.Now().Add(time.Hour*24).Unix(), time.Now().Add(time.Hour*24*7).Unix())
	uService := coreService.NewUserService(userRepo, authService)

	// authorization
	mongoDB, err := mongo.NewMongoDB()
	if err != nil {
		log.Fatal("cant connect to mongo", err)
	}
	authorizationRepo := repository3.NewAuthorizationRepo(mongoDB)
	authorizationService := coreService.NewAuthorizationService(authorizationRepo)

	// matchMaking
	redisDB := redis.New(cfg.Redis)
	matchMakingRepo := matchmaking.NewMatchMakingRepo(redisDB, cfg.MatchMakingPrefix)
	matchMakingService := matchMakingHandler.NewMatchMakingService(matchMakingRepo, cfg.MatchMakingTimeOut)

	controllerServices := service.Service{
		UserService:           uService,
		BackofficeUserService: backofficeHandler,
		AuthService:           authService,
		AuthorizationService:  authorizationService,
		MatchMakingService:    matchMakingService,
	}
	httpController := controller.NewController(echoInstance, controllerServices)
	httpController.InitRouter()

	//create httpServer
	httpServer := http.NewHTTPServer(echoInstance, mysqlConfig.HttpServerConfig{
		Port: 8000,
	})

	httpServer.Start()
	defer httpServer.Stop()

	done := make(chan bool)
	var wg sync.WaitGroup
	go func() {
		sch := scheduler.New(matchMakingService, cfg.Scheduler)
		wg.Add(1)
		sch.Start(done, &wg)
	}()

	// Listen for OS signals to perform a graceful shutdown
	log.Printf("listening signals on %d ...", os.Getpid())
	quite := make(chan os.Signal, 1)
	signal.Notify(
		quite,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	<-quite
	done <- true
	log.Println("graceful shutdown...")
	time.Sleep(5 * time.Second)
	wg.Wait()
}
