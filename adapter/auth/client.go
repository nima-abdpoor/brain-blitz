package auth_adapter

import (
	"BrainBlitz.com/game/contract/auth/golang"
	"context"
	"google.golang.org/grpc"
)

type Client struct {
	Conn *grpc.ClientConn
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		Conn: conn,
	}
}

func (c Client) GetAccessToken(ctx context.Context, request CreateAccessTokenRequest) (CreateAccessTokenResponse, error) {
	client := golang.NewTokenServiceClient(c.Conn)

	requestData := make([]*golang.CreateTokenRequest, 0)
	for _, data := range request.Data {
		requestData = append(requestData, &golang.CreateTokenRequest{
			Key:   data.Key,
			Value: data.Value,
		})
	}
	req := &golang.CreateAccessTokenRequest{
		Data: requestData,
	}

	res, err := client.GetAccessToken(ctx, req)
	if err != nil || res == nil {
		return CreateAccessTokenResponse{}, err
	}

	return CreateAccessTokenResponse{
		AccessToken: res.AccessToken,
		ExpireTime:  res.ExpireTime,
	}, nil
}

func (c Client) GetRefreshToken(ctx context.Context, request CreateRefreshTokenRequest) (CreateRefreshTokenResponse, error) {
	client := golang.NewTokenServiceClient(c.Conn)

	requestData := make([]*golang.CreateTokenRequest, 0)
	for _, data := range request.Data {
		requestData = append(requestData, &golang.CreateTokenRequest{
			Key:   data.Key,
			Value: data.Value,
		})
	}
	req := &golang.CreateAccessTokenRequest{
		Data: requestData,
	}

	res, err := client.GetAccessToken(ctx, req)
	if err != nil || res == nil {
		return CreateRefreshTokenResponse{}, err
	}

	return CreateRefreshTokenResponse{
		RefreshToken: res.AccessToken,
		ExpireTime:   res.ExpireTime,
	}, nil
}
