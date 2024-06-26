package authgrpc

import (
	"context"
	"errors"

	"github.com/kochnevns/finanses-auth/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authGrpcSrv "github.com/kochnevns/finances-protos/auth"
	"github.com/kochnevns/finanses-auth/internal/services/auth"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	authGrpcSrv.UnimplementedAuthServer
	auth Auth
}

func Register(gRPCServer *grpc.Server, auth Auth) {
	authGrpcSrv.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *authGrpcSrv.LoginRequest,
) (*authGrpcSrv.LoginResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &authGrpcSrv.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *authGrpcSrv.RegisterRequest,
) (*authGrpcSrv.RegisterResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &authGrpcSrv.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	in *authGrpcSrv.IsAdminRequest,
) (*authGrpcSrv.IsAdminResponse, error) {
	if in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, in.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to check admin status")
	}

	return &authGrpcSrv.IsAdminResponse{IsAdmin: isAdmin}, nil
}
