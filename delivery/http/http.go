package http

//
//import (
//	_ "net/http/pprof"
//)

//
//func main() {
//	const op = "main.main"
//	// TODO - read config path from command line
//	//logger.Logger.Named(op).Info("cfg", zap.Any("config", cfg))
//
//	// Create a new instance of the Echo router
//	echoInstance := echo.New()
//	echoInstance.Use(middleware.RequestID())
//	echoInstance.Use(middleware.RequestLoggerWithConfig(echo2.RequestLoggerConfig))
//	echoInstance.Use(middleware.Recover())
//
//	// backoffice
//	//backofficeRepo := repository.New(db)
//	//backofficeHandler := backofficeUserHandler.New(backofficeRepo)
//
//	// presence
//	//todo resolve two instances of presence client
//	//redisDB := redis.New(cfg.Redis)
//	//presenceRepo := presence.New(redisDB, cfg.GetPresence)
//	//presenceS := presenceService.New(presenceRepo, cfg.Presence)
//
//	//controllerServices := service.Service{
//	//	BackofficeUserService: backofficeHandler,
//	//	Presence:              presenceS,
//	//}
//	//todo move this to somewhere better
//	httpController := controller.NewController(echoInstance, controllerServices)
//	httpController.InitRouter()
//
//	//create httpServer
//	httpServer := http.NewHTTPServer(echoInstance, cfg.HTTPServer)
//
//	if cfg.Feature.Infra {
//		if cfg.Feature.Metrics {
//			metrics.InitMetrics()
//			prometheusHttpHandler := promhttp.Handler()
//			infraHttpController := controller.NewInfraHttpController(prometheusHttpHandler)
//			infraHttpController.InitRouter()
//			InfraHttpServer := http.NewInfraHTTPServer(prometheusHttpHandler, cfg.HTTPServer)
//			defer InfraHttpServer.StopInfraServer()
//			InfraHttpServer.StartInfraServer()
//		}
//	}
//
//	httpServer.Start()
//	defer httpServer.Stop()
//
//	var wg sync.WaitGroup
//
//	// Listen for OS signals to perform a graceful shutdown
//	//logger.Logger.Named(op).Info("listening signals...", zap.Int("processId", os.Getpid()))
//	quite := make(chan os.Signal, 1)
//	signal.Notify(
//		quite,
//		os.Interrupt,
//		syscall.SIGHUP,
//		syscall.SIGINT,
//		syscall.SIGQUIT,
//		syscall.SIGTERM,
//	)
//	<-quite
//	//logger.Logger.Named(op).Info("graceful shutdown...")
//	time.Sleep(5 * time.Second)
//	wg.Wait()
//}
//
//func getMysqlDB(config repository.Config) (repository2.Database, error) {
//	const op = "main.getMysqlDB"
//	db, err := repository.NewDB(mysqlConfig.DatabaseConfig{
//		Driver:                 "mysql",
//		Url:                    fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=true&loc=UTC&tls=false&readTimeout=3s&writeTimeout=3s&timeout=3s&clientFoundRows=true", config.Username, config.Password, config.Host, config.Port, config.DBName),
//		ConnMaxLifeTimeMinutes: config.ConnMaxLifeTimeMinutes,
//		MaxOpenCons:            config.MaxOpenCons,
//		MaxIdleCons:            config.MaxIdleCons,
//	})
//	if err != nil {
//		//logger.Logger.Named(op).Error("failed to connect to mysql", zap.Error(err))
//		return nil, err
//	} else {
//		return db, nil
//	}
//}
