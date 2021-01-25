package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/Remneva/anti-bruteforce/internal/app"
	"github.com/Remneva/anti-bruteforce/internal/server"
	"github.com/Remneva/anti-bruteforce/internal/server/pb"
	"github.com/Remneva/anti-bruteforce/internal/storage"
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

func NewServer(app *app.App, l *zap.Logger, address string) (*Server, error) {
	l.Info("grpc is running...")
	lsn, err := net.Listen("tcp", address)
	if err != nil {
		l.Error("Listening Error", zap.Error(err))
		return &Server{}, fmt.Errorf("database query failed: %w", err)
	}
	server := grpc.NewServer()
	srv := &Server{
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

func (s *Server) Auth(ctx context.Context, req *pb.Authorization) (*pb.Result, error) {
	s.l.Info("Auth grpc method")
	if req.Login == "" || req.Password == "" || req.Ip == "" {
		s.l.Error("login, password, ip can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "login, password, ip can`t be empty")
	}
	var auth storage.Auth
	auth.IP = req.Ip
	auth.Login = req.Login
	auth.Password = req.Password
	success, err := s.app.Validate(ctx, auth)
	if err != nil {
		s.l.Error("Validation error", zap.Error(err))
	}

	return &pb.Result{Result: success}, nil
}

func (s *Server) CleanBucket(ctx context.Context, req *pb.User) (*pb.Empty, error) {
	s.l.Info("clean grpc method")
	if req.Login == "" || req.Ip == "" {
		s.l.Error("login, ip can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "login, ip can`t be empty")
	}
	var us storage.User
	us.IP = req.Ip
	us.Login = req.Login
	if err := s.app.CleanBucket(us); err != nil {
		s.l.Error("servr error", zap.Error(err))
	}
	return &pb.Empty{}, nil
}

func (s *Server) AddToWhiteList(ctx context.Context, req *pb.Ip) (*pb.Empty, error) {
	s.l.Info("add to white list grpc method")
	if req.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	var ip storage.IP
	ip.IP = req.Ip
	ip.Mask = req.Mask
	if err := s.app.AddToWhiteList(ctx, ip); err != nil {
		return &pb.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

func (s *Server) AddToBlackList(ctx context.Context, req *pb.Ip) (*pb.Empty, error) {
	s.l.Info("add to black list grpc method")
	if req.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	var ip storage.IP
	ip.IP = req.Ip
	ip.Mask = req.Mask
	if err := s.app.AddToBlackList(ctx, ip); err != nil {
		return &pb.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

func (s *Server) DeleteFromWhiteList(ctx context.Context, req *pb.Ip) (*pb.Empty, error) {
	s.l.Info("delete from white list grpc method")
	if req.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	var ip storage.IP
	ip.IP = req.Ip
	ip.Mask = req.Mask
	if err := s.app.DeleteFromWhiteList(ctx, ip); err != nil {
		return &pb.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

func (s *Server) DeleteFromBlackList(ctx context.Context, req *pb.Ip) (*pb.Empty, error) {
	s.l.Info("delete from black list grpc method")
	if req.Ip == "" {
		s.l.Error("IP can`t be empty")
		return nil, status.Error(codes.InvalidArgument, "IP can`t be empty")
	}
	var ip storage.IP
	ip.IP = req.Ip
	ip.Mask = req.Mask
	if err := s.app.DeleteFromBlackList(ctx, ip); err != nil {
		return &pb.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}
