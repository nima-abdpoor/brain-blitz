package auth_adapter

import (
	"BrainBlitz.com/game/contract/auth/golang"
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// mockTokenServer records which methods were called and returns distinct token values.
type mockTokenServer struct {
	golang.UnimplementedTokenServiceServer
	getAccessTokenCalled  bool
	getRefreshTokenCalled bool
}

func (s *mockTokenServer) GetAccessToken(_ context.Context, _ *golang.CreateAccessTokenRequest) (*golang.CreateAccessTokenResponse, error) {
	s.getAccessTokenCalled = true
	return &golang.CreateAccessTokenResponse{AccessToken: "access-token-value", ExpireTime: 100}, nil
}

func (s *mockTokenServer) GetRefreshToken(_ context.Context, _ *golang.CreateRefreshTokenRequest) (*golang.CreateRefreshTokenResponse, error) {
	s.getRefreshTokenCalled = true
	return &golang.CreateRefreshTokenResponse{RefreshToken: "refresh-token-value", ExpireTime: 200}, nil
}

func newTestClient(t *testing.T, srv *mockTokenServer) *Client {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	golang.RegisterTokenServiceServer(s, srv)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return New(conn)
}

func TestGetAccessToken_CallsCorrectGRPCMethod(t *testing.T) {
	srv := &mockTokenServer{}
	client := newTestClient(t, srv)

	res, err := client.GetAccessToken(context.Background(), CreateAccessTokenRequest{
		Data: []CreateTokenRequest{{Key: "id", Value: "1"}},
	})

	require.NoError(t, err)
	assert.True(t, srv.getAccessTokenCalled, "GetAccessToken should call the GetAccessToken gRPC method")
	assert.False(t, srv.getRefreshTokenCalled, "GetAccessToken must not call GetRefreshToken")
	assert.Equal(t, "access-token-value", res.AccessToken)
	assert.Equal(t, int64(100), res.ExpireTime)
}

func TestGetRefreshToken_CallsCorrectGRPCMethod(t *testing.T) {
	srv := &mockTokenServer{}
	client := newTestClient(t, srv)

	res, err := client.GetRefreshToken(context.Background(), CreateRefreshTokenRequest{
		Data: []CreateTokenRequest{{Key: "id", Value: "1"}},
	})

	require.NoError(t, err)
	assert.True(t, srv.getRefreshTokenCalled, "GetRefreshToken should call the GetRefreshToken gRPC method")
	assert.False(t, srv.getAccessTokenCalled, "GetRefreshToken must not call GetAccessToken")
	assert.Equal(t, "refresh-token-value", res.RefreshToken)
	assert.Equal(t, int64(200), res.ExpireTime)
}

func TestGetRefreshToken_ReturnsRefreshTokenNotAccessToken(t *testing.T) {
	srv := &mockTokenServer{}
	client := newTestClient(t, srv)

	res, err := client.GetRefreshToken(context.Background(), CreateRefreshTokenRequest{
		Data: []CreateTokenRequest{{Key: "role", Value: "user"}},
	})

	require.NoError(t, err)
	// The returned value must be the server's refresh token, not the access token.
	assert.Equal(t, "refresh-token-value", res.RefreshToken)
	assert.NotEqual(t, "access-token-value", res.RefreshToken,
		"GetRefreshToken must not return an access token value")
}
