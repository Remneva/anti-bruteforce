package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/Remneva/anti-bruteforce/internal/app"
	"github.com/Remneva/anti-bruteforce/internal/server"
	"github.com/Remneva/anti-bruteforce/internal/server/pb"
	"github.com/Remneva/anti-bruteforce/internal/storage"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ server.Stopper = (*Server)(nil)

type Server struct {
	pb.UnimplementedAntiBruteForceServiceServer
	l      *zap.Logger
	server *grpc.Server
	lsn    net.Listener
	app    *app.App
}

func LogRequest(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {
	fmt.Printf("Request for : %s\n", info.FullMethod)
	// Last but super important, execute the handler so that the actually gRPC request is also performed
	return handler(ctx, req)
}

func NewServer(app *app.App, l *zap.Logger, address string) (*Server, error) {
	l.Info("grpc is running...")
	lsn, err := net.Listen("tcp", address)
	if err != nil {
		l.Error("Listening Error", zap.Error(err))
		return &Server{}, fmt.Errorf("database query failed: %w", err)
	}
	server := grpc.NewServer(
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_zap.StreamServerInterceptor(l))),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_zap.UnaryServerInterceptor(l))))

	srv := &Server{ //nolint:exhaustivestruct
		app:    app,
		server: server,
		lsn:    lsn,
		l:      l,
	}
	pb.RegisterAntiBruteForceServiceServer(server, srv)
	return srv, nil
}

func (s *Server) Start() error {
	if err := s.server.Serve(s.lsn); err != nil {
		s.l.Error("Error", zap.Error(err))
		return fmt.Errorf("creating a new ServerTransport failed: %w", err)
	}
	s.l.Info("starting grpc server", zap.String("Addr", s.lsn.Addr().String()))
	return nil
}

func (s *Server) Stop() {
	s.l.Info("grpc stopping...")
	s.server.GracefulStop()
}

func (s *Server) Auth(ctx context.Context, req *pb.AuthorizationRequest) (*pb.AuthorizationResponse, error) {
	s.l.Info("Auth grpc method")
	if req.Authorization.Login == "" || req.Authorization.Password == "" || req.Authorization.Ip == "" {
		s.l.Error("login, password, ip can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "login, password or ip can`t be empty")
	}
	var auth storage.Auth
	auth.IP = req.Authorization.Ip
	auth.Login = req.Authorization.Login
	auth.Password = req.Authorization.Password
	success, err := s.app.Validate(ctx, auth)
	r := &pb.Result{State: success}
	if err != nil {
		s.l.Error("Validation error", zap.Error(err))
		return &pb.AuthorizationResponse{Result: r}, status.Error(codes.Internal, err.Error())
	}

	return &pb.AuthorizationResponse{Result: r}, nil
}

func (s *Server) CleanBucket(ctx context.Context, req *pb.CleanBucketRequest) (*pb.CleanBucketResponse, error) {
	s.l.Info("clean grpc method")
	if req.User.Login == "" || req.User.Ip == "" {
		s.l.Error("login or ip can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "login or ip can`t be empty")
	}
	var us storage.User
	us.IP = req.User.Ip
	us.Login = req.User.Login
	if err := s.app.CleanBucket(ctx, us); err != nil {
		s.l.Error("server error", zap.Error(err))
		return &pb.CleanBucketResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.CleanBucketResponse{}, nil
}

func (s *Server) AddToWhiteList(ctx context.Context, req *pb.AddToWhiteListRequest) (*pb.AddToWhiteListResponse, error) {
	s.l.Info("add to white list grpc method")
	if req.Ip.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	ip := parseToStorageIP(req.Ip)
	if err := s.app.AddToWhiteList(ctx, ip); err != nil {
		return &pb.AddToWhiteListResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.AddToWhiteListResponse{}, nil
}

func (s *Server) AddToBlackList(ctx context.Context, req *pb.AddToBlackListRequest) (*pb.AddToBlackListResponse, error) {
	s.l.Info("add to black list grpc method")
	if req.Ip.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	ip := parseToStorageIP(req.Ip)
	if err := s.app.AddToBlackList(ctx, ip); err != nil {
		return &pb.AddToBlackListResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.AddToBlackListResponse{}, nil
}

func (s *Server) DeleteFromWhiteList(ctx context.Context, req *pb.DeleteFromWhiteListRequest) (*pb.DeleteFromWhiteListResponse, error) {
	s.l.Info("delete from white list grpc method")
	if req.Ip.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	ip := parseToStorageIP(req.Ip)
	if err := s.app.DeleteFromWhiteList(ctx, ip); err != nil {
		return &pb.DeleteFromWhiteListResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteFromWhiteListResponse{}, nil
}

func (s *Server) DeleteFromBlackList(ctx context.Context, req *pb.DeleteFromBlackListRequest) (*pb.DeleteFromBlackListResponse, error) {
	s.l.Info("delete from black list grpc method")
	if req.Ip.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	ip := parseToStorageIP(req.Ip)
	if err := s.app.DeleteFromBlackList(ctx, ip); err != nil {
		return &pb.DeleteFromBlackListResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteFromBlackListResponse{}, nil
}

func parseToStorageIP(req *pb.Ip) storage.IP {
	var ip storage.IP
	ip.IP = req.Ip
	ip.Mask = req.Mask
	return ip
}
