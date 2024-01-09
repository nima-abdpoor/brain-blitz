package main

import (
	proto "BrainBlitz.com/game/api/api"
	"BrainBlitz.com/game/internal/controller/grpc"
	"BrainBlitz.com/game/internal/core/server"
	grpcServer "BrainBlitz.com/game/internal/core/server/grpc"
	"BrainBlitz.com/game/internal/core/service"
	infraConf "BrainBlitz.com/game/internal/infra/config"
	"BrainBlitz.com/game/internal/infra/repository"
	"go.uber.org/zap"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// Initialize the database connection
	db, err := repository.NewDB(
		infraConf.DatabaseConfig{
			Driver:                 "mysql",
			Url:                    "bbGame:root@tcp(127.0.0.1:3310)/brainBlitz_db?charset=utf8mb4&parseTime=true&loc=UTC&tls=false&readTimeout=3s&writeTimeout=3s&timeout=3s&clientFoundRows=true",
			ConnMaxLifeTimeMinutes: 3,
			MaxOpenCons:            10,
			MaxIdleCons:            1,
		})
	if err != nil {
		log.Fatalf("failed to new database err=%s\n", err.Error())
	}

	// Create the UserRepository
	userRepo := repository.NewUserRepository(db)

	// Create the UserService
	userService := service.NewUserService(userRepo)

	// Create the UserController
	userController := grpc.NewUserController(userService)
	gServer, err := grpcServer.NewGRPCServer(infraConf.GrpcServerConfig{
		Port: 9090,
		KeepaliveParams: keepalive.ServerParameters{
			MaxConnectionIdle:     100,
			MaxConnectionAge:      7200,
			MaxConnectionAgeGrace: 60,
			Time:                  10,
			Timeout:               3,
		},
		KeepalivePolicy: keepalive.EnforcementPolicy{
			MinTime:             10,
			PermitWithoutStream: true,
		},
	})
	if err != nil {
		log.Fatalf("failed to new grpc server err=%s\n", err.Error())
	}

	go gServer.Start(func(server *grpc2.Server) {
		proto.RegisterUserServiceServer(server, userController)
	})
	server.AddShutdownHook(gServer, db)
}
