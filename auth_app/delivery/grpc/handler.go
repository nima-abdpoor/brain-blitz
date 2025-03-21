package grpc

import (
	"BrainBlitz.com/game/auth_app/service"
	pb "BrainBlitz.com/game/contract/auth/golang"
	"BrainBlitz.com/game/pkg/logger"
	"context"
)

type Handler struct {
	pb.UnimplementedTokenServiceServer
	AuthService service.Service
	Logger      logger.SlogAdapter
}

func NewHandler(srv service.Service, logger logger.SlogAdapter) Handler {
	return Handler{
		UnimplementedTokenServiceServer: pb.UnimplementedTokenServiceServer{},
		AuthService:                     srv,
		Logger:                          logger,
	}
}

func (h Handler) GetAccessToken(ctx context.Context, req *pb.CreateAccessTokenRequest) (*pb.CreateAccessTokenResponse, error) {
	op := "grpc_GetAccessToken"
	h.Logger.Info(op, "userId", req.Data)

	requestData := make([]service.CreateTokenRequest, 0)
	for _, data := range req.GetData() {
		requestData = append(requestData, service.CreateTokenRequest{
			Key:   data.GetKey(),
			Value: data.GetValue(),
		})
	}
	res, err := h.AuthService.CreateAccessToken(ctx, service.CreateAccessTokenRequest{
		Data: requestData,
	})

	if err != nil {
		h.Logger.Error(op, "error", err.Error())
		return nil, err
	}

	return &pb.CreateAccessTokenResponse{
		AccessToken: res.AccessToken,
		ExpireTime:  res.ExpireTime,
	}, nil
}

func (h Handler) GetRefreshToken(ctx context.Context, req *pb.CreateRefreshTokenRequest) (*pb.CreateRefreshTokenResponse, error) {
	op := "grpc_GetRefreshToken"
	h.Logger.Info(op, "userId", req.Data)

	requestData := make([]service.CreateTokenRequest, 0)
	for _, data := range req.GetData() {
		requestData = append(requestData, service.CreateTokenRequest{
			Key:   data.GetKey(),
			Value: data.GetValue(),
		})
	}
	res, err := h.AuthService.CreateRefreshToken(ctx, service.CreateRefreshTokenRequest{
		Data: requestData,
	})

	if err != nil {
		h.Logger.Error(op, "error", err.Error())
		return nil, err
	}

	return &pb.CreateRefreshTokenResponse{
		RefreshToken: res.RefreshToken,
		ExpireTime:   res.ExpireTime,
	}, nil
}
