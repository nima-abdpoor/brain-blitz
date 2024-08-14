package main

import (
	"BrainBlitz.com/game/adapter/broker/kafka"
	"BrainBlitz.com/game/config"
	"BrainBlitz.com/game/internal/controller"
	"BrainBlitz.com/game/internal/core/model/request"
	repository2 "BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/internal/core/server/http"
	coreService "BrainBlitz.com/game/internal/core/service"
	"BrainBlitz.com/game/internal/core/service/backofficeUserHandler"
	"BrainBlitz.com/game/internal/core/service/match"
	matchMakingHandler "BrainBlitz.com/game/internal/core/service/matchMaking"
	presenceService "BrainBlitz.com/game/internal/core/service/presence"
	mysqlConfig "BrainBlitz.com/game/internal/infra/config"
	"BrainBlitz.com/game/internal/infra/repository"
	repository3 "BrainBlitz.com/game/internal/infra/repository/authorization"
	"BrainBlitz.com/game/internal/infra/repository/matchmaking"
	"BrainBlitz.com/game/internal/infra/repository/matchmanager"
	"BrainBlitz.com/game/internal/infra/repository/mongo"
	"BrainBlitz.com/game/internal/infra/repository/presence"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"BrainBlitz.com/game/logger"
	echo2 "BrainBlitz.com/game/pkg/echo"
	"BrainBlitz.com/game/scheduler"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	http2 "net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	const op = "main.main"
	// TODO - read config path from command line
	cfg := config.Load("config.yml")
	logger.Logger.Named(op).Info("cfg", zap.Any("config", cfg))

	if cfg.Infra.PPROF {
		go listenPprofService()
	}

	// Create a new instance of the Echo router
	echoInstance := echo.New()
	echoInstance.Use(middleware.RequestID())
	echoInstance.Use(middleware.RequestLoggerWithConfig(echo2.RequestLoggerConfig))
	echoInstance.Use(middleware.Recover())

	db, err := getMysqlDB(cfg.Mysql)
	for err != nil {
		db, err = getMysqlDB(cfg.Mysql)
		time.Sleep(cfg.Mysql.RetryConnection)
	}

	//create the UserRepository
	userRepo := repository.NewUserRepository(db)

	// backoffice
	backofficeRepo := repository.New(db)
	backofficeHandler := backofficeUserHandler.New(backofficeRepo)

	//create the user service
	//todo mv secret key into env files
	authService := coreService.NewJWTAuthService("salam", "exp", time.Now().Add(time.Hour*24).Unix(), time.Now().Add(time.Hour*24*7).Unix())
	uService := coreService.NewUserService(userRepo, authService)

	// authorization
	mongoDB, err := mongo.NewMongoDB(cfg.Mongo)
	if err != nil {
		//todo add to metrics
		logger.Logger.Named(op).Error("cant connect to mongo", zap.Error(err))
	}
	authorizationRepo := repository3.NewAuthorizationRepo(mongoDB)
	authorizationService := coreService.NewAuthorizationService(authorizationRepo)

	// presence
	//todo resolve two instances of presence client
	redisDB := redis.New(cfg.Redis)
	presenceRepo := presence.New(redisDB, cfg.GetPresence)
	presenceS := presenceService.New(presenceRepo, cfg.Presence)
	presenceClientRepo := presence.NewPresenceClient(redisDB, cfg.GetPresence)

	// matchMaking
	kafkaPublisher := kafka.NewKafkaPublisher(cfg.Kafka)
	matchMakingRepo := matchmaking.NewMatchMakingRepo(redisDB, cfg.MatchMakingPrefix)
	matchMakingService := matchMakingHandler.NewMatchMakingService(matchMakingRepo, presenceClientRepo, kafkaPublisher, cfg.MatchMakingTimeOut)

	// matchManagement
	kafkaConsumer := kafka.NewKafkaConsumer(cfg.Kafka)
	matchManagerRepo := matchmanager.New(mongoDB)
	matchManagerSvc := match.New(matchManagerRepo, kafkaConsumer)

	controllerServices := service.Service{
		UserService:            uService,
		BackofficeUserService:  backofficeHandler,
		AuthService:            authService,
		AuthorizationService:   authorizationService,
		MatchMakingService:     matchMakingService,
		MatchManagementService: matchManagerSvc,
		Presence:               presenceS,
	}
	//todo move this to somewhere better
	go controllerServices.MatchManagementService.StartMatchCreator(request.StartMatchCreatorRequest{})
	httpController := controller.NewController(echoInstance, controllerServices)
	httpController.InitRouter()

	//create httpServer
	httpServer := http.NewHTTPServer(echoInstance, cfg.HTTPServer)

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
	logger.Logger.Named(op).Info("listening signals...", zap.Int("processId", os.Getpid()))
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
	logger.Logger.Named(op).Info("graceful shutdown...")
	time.Sleep(5 * time.Second)
	wg.Wait()
}

func getMysqlDB(config repository.Config) (repository2.Database, error) {
	const op = "main.getMysqlDB"
	db, err := repository.NewDB(mysqlConfig.DatabaseConfig{
		Driver:                 "mysql",
		Url:                    fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=true&loc=UTC&tls=false&readTimeout=3s&writeTimeout=3s&timeout=3s&clientFoundRows=true", config.Username, config.Password, config.Host, config.Port, config.DBName),
		ConnMaxLifeTimeMinutes: config.ConnMaxLifeTimeMinutes,
		MaxOpenCons:            config.MaxOpenCons,
		MaxIdleCons:            config.MaxIdleCons,
	})
	if err != nil {
		logger.Logger.Named(op).Error("failed to connect to mysql", zap.Error(err))
		return nil, err
	} else {
		return db, nil
	}
}

func listenPprofService() {
	if err := http2.ListenAndServe(":8099", nil); err != nil {
		fmt.Printf("error in serving PProf %v\n", err)
	}
}
