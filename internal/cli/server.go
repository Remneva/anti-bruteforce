package cli

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Remneva/anti-bruteforce/internal/app"
	"github.com/Remneva/anti-bruteforce/internal/server/pb"
	"github.com/Remneva/anti-bruteforce/internal/storage"
	"github.com/mitchellh/cli"
	"google.golang.org/grpc"
)

const text = "something strange happens"

type grpcCommands struct {
	commands map[string]cli.CommandFactory
	cli      *Servercli
}

type Servercli struct {
	grpc *grpc.Server
	app  *app.App
}

func New(app *app.App) *Servercli {
	c := &Servercli{
		app: app,
	}
	return c
}

func (s *Servercli) Stop() {
	s.grpc.GracefulStop()
}

func (s *Servercli) RunCli() {
	c := cli.NewCLI("server", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"auth": func() (cli.Command, error) {
			return &Auth{}, nil
		},
		"cleanBucket": func() (cli.Command, error) {
			return &CleanBucket{}, nil
		},
		"addToWhiteList": func() (cli.Command, error) {
			return &AddToWhiteList{}, nil
		},
		"addToBlackList": func() (cli.Command, error) {
			return &AddToBlackList{}, nil
		},
		"deleteFromWhiteList": func() (cli.Command, error) {
			return &DeleteFromWhiteList{}, nil
		},
		"deleteFromBlackList": func() (cli.Command, error) {
			return &DeleteFromBlackList{}, nil
		},
	}

	if len(c.Args) == 0 {
		listener, err := net.Listen("tcp", "localhost:12344")
		if err != nil {
			fmt.Println(err)
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		s.grpc = grpcServer

		pb.RegisterAntiBruteForceServiceServer(grpcServer, &grpcCommands{commands: c.Commands, cli: s})

		err = grpcServer.Serve(listener)
		if err != nil {
			log.Fatalf("failed to start: %v", err)
		}
	}
	_, err := c.Run()
	if err != nil {
		log.Println(err)
	}
}

func (g *grpcCommands) Auth(ctx context.Context, request *pb.AuthorizationRequest) (*pb.AuthorizationResponse, error) {
	panic("implement me")
}

func (g *grpcCommands) CleanBucket(ctx context.Context, request *pb.CleanBucketRequest) (*pb.CleanBucketResponse, error) {
	var us storage.User
	us.Login = request.User.Ip
	us.IP = request.User.Login
	if err := g.cli.app.CleanBucket(ctx, us); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.CleanBucketResponse{}, nil
}

func (g *grpcCommands) AddToWhiteList(ctx context.Context, request *pb.AddToWhiteListRequest) (*pb.AddToWhiteListResponse, error) {
	ip := parseToStorage(request.Ip.Ip, request.Ip.Mask)
	if err := g.cli.app.AddToWhiteList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.AddToWhiteListResponse{}, nil
}

func (g *grpcCommands) DeleteFromWhiteList(ctx context.Context, request *pb.DeleteFromWhiteListRequest) (*pb.DeleteFromWhiteListResponse, error) {
	ip := parseToStorage(request.Ip.Ip, request.Ip.Mask)
	if err := g.cli.app.DeleteFromWhiteList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.DeleteFromWhiteListResponse{}, nil
}

func (g *grpcCommands) AddToBlackList(ctx context.Context, request *pb.AddToBlackListRequest) (*pb.AddToBlackListResponse, error) {
	ip := parseToStorage(request.Ip.Ip, request.Ip.Mask)
	if err := g.cli.app.AddToBlackList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.AddToBlackListResponse{}, nil
}

func (g *grpcCommands) DeleteFromBlackList(ctx context.Context, request *pb.DeleteFromBlackListRequest) (*pb.DeleteFromBlackListResponse, error) {
	ip := parseToStorage(request.Ip.Ip, request.Ip.Mask)
	if err := g.cli.app.DeleteFromBlackList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.DeleteFromBlackListResponse{}, nil
}

func parseToStorage(arg ...string) storage.IP {
	var ip storage.IP
	ip.IP = arg[0]
	ip.Mask = arg[1]
	return ip
}

func (t *CleanBucket) Run(args []string) int {
	return 0
}

func (t *CleanBucket) Synopsis() string {
	return text
}

func (t *AddToWhiteList) Run(args []string) int {
	return 0
}

func (t *AddToWhiteList) Synopsis() string {
	return text
}

func (t *AddToBlackList) Run(args []string) int {
	return 0
}

func (t *AddToBlackList) Synopsis() string {
	return text
}

type CleanBucket struct {
}

func (t *CleanBucket) Help() string {
	return text
}

type AddToWhiteList struct {
}

func (t *AddToWhiteList) Help() string {
	return text
}

type AddToBlackList struct {
}

func (t *AddToBlackList) Help() string {
	return text
}

type DeleteFromWhiteList struct {
}

func (d DeleteFromWhiteList) Help() string {
	return text
}

func (d DeleteFromWhiteList) Run(args []string) int {
	return 0
}

func (d DeleteFromWhiteList) Synopsis() string {
	return text
}

type DeleteFromBlackList struct {
}

func (d DeleteFromBlackList) Help() string {
	return text
}

func (d DeleteFromBlackList) Run(args []string) int {
	return 0
}

func (d DeleteFromBlackList) Synopsis() string {
	return text
}

type Auth struct {
}

func (d *Auth) Help() string {
	return text
}

func (d *Auth) Run(args []string) int {
	return 0
}

func (d *Auth) Synopsis() string {
	return text
}
