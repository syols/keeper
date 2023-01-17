package pkg

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/syols/keeper/config"
	"github.com/syols/keeper/internal/models"
	pb "github.com/syols/keeper/internal/rpc/proto"
)

type GrpcService struct {
	pb.UnimplementedKeeperServer
	ctx        context.Context
	authorizer Authorizer
	database   Database
	grpcServer *grpc.Server
	settings   config.Config
}

func NewGrpcService(ctx context.Context, settings config.Config) (GrpcService, error) {
	conn := NewDatabaseURLConnection(*settings.Server.DatabaseConnectionString)
	database, err := NewDatabase(conn)
	if err != nil {
		return GrpcService{}, err
	}

	var opts []grpc.ServerOption
	service := GrpcService{
		ctx:        ctx,
		authorizer: NewAuthorizer(settings),
		database:   database,
		grpcServer: grpc.NewServer(opts...),
		settings:   settings,
	}
	pb.RegisterKeeperServer(service.grpcServer, service)
	return service, nil
}

func (g *GrpcService) Run(port uint16) error {
	uri := fmt.Sprintf(":%d", port)
	listen, err := net.Listen("tcp", uri)
	if err != nil {
		return err
	}

	if err := g.grpcServer.Serve(listen); err != nil {
		return err
	}
	return nil
}

func (g *GrpcService) Shutdown() {
	if g.grpcServer != nil {
		g.grpcServer.GracefulStop()
	}
}

func (g GrpcService) Registration(ctx context.Context, request *pb.SignInRequest) (*pb.SignInResponse, error) {
	user := models.NewUser(request)
	if err := g.database.Register(ctx, &user); err != nil {
		return nil, status.Error(codes.InvalidArgument, "database error")
	}

	return g.createSignInResponse(request)
}

func (g GrpcService) Authorization(ctx context.Context, request *pb.SignInRequest) (*pb.SignInResponse, error) {
	user := models.NewUser(request)
	value, err := g.database.Login(ctx, &user)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect user")
	}
	if value == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	return g.createSignInResponse(request)
}

func (g GrpcService) AccessTokenRequest(_ context.Context, request *pb.Token) (*pb.Token, error) {
	username, err := g.authorizer.VerifyToken(request.String())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	token, err := g.authorizer.CreateToken(username, time.Hour)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "create access token error")
	}

	return createToken(token), nil
}

func (g GrpcService) AddRecord(ctx context.Context, in *pb.Record) (*pb.Record, error) {
	username, err := g.authorizer.VerifyToken(in.AccessToken.Value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := g.database.SaveRecord(ctx, username, *models.NewRecord(in)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return in, nil
}

func (g GrpcService) SyncRecord(token *pb.Token, srv pb.Keeper_SyncRecordServer) error {
	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()

	username, err := g.authorizer.VerifyToken(token.GetValue())
	if err != nil {
		return status.Error(codes.InvalidArgument, "token error")
	}

	wg := sync.WaitGroup{}
	errs := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		userRecords, err := g.database.UserRecords(ctx, username)
		if err != nil {
			errs <- status.Error(codes.InvalidArgument, "database error")
			return
		}
		if userRecords == nil {
			return
		}

		for _, userRecord := range userRecords {
			record := pb.Record{
				AccessToken: token,
				Metadata:    userRecord.Metadata,
				DetailType:  userRecord.DetailType,
			}

			userRecord.PrivateData.SetPrivateData(&record)
			if err := srv.Send(&record); err != nil {
				errs <- status.Errorf(codes.Internal, err.Error())
			}
		}
		close(errs)
	}()
	wg.Wait()

	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func createToken(token string) *pb.Token {
	result := pb.Token{
		Value: token,
	}
	return &result
}

func (g GrpcService) createSignInResponse(request *pb.SignInRequest) (*pb.SignInResponse, error) {
	accessToken, err := g.authorizer.CreateToken(request.Login, time.Hour)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "create access token error")
	}

	refreshToken, err := g.authorizer.CreateToken(request.Login, time.Hour*24)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "create refresh token error")
	}

	return &pb.SignInResponse{
		Access:  createToken(accessToken),
		Refresh: createToken(refreshToken),
	}, nil
}
