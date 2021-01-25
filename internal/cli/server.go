package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/Remneva/anti-bruteforce/internal/app"
	"github.com/Remneva/anti-bruteforce/internal/cli/pb"
	"github.com/Remneva/anti-bruteforce/internal/storage"
	"github.com/mitchellh/cli"
	"google.golang.org/grpc"
)

type grpcCommands struct {
	commands map[string]cli.CommandFactory
	cli      *CliC
}

type Clean struct {
}

type AddToWhiteList struct {
}

type AddToBlackList struct {
}

type DeleteFromWhiteList struct {
}

func (d DeleteFromWhiteList) Help() string {
	return "hello [arg0] [arg1] ... says hello to everyone"
}

func (d DeleteFromWhiteList) Run(args []string) int {
	panic("implement me")
}

func (d DeleteFromWhiteList) Synopsis() string {
	panic("implement me")
}

type DeleteFromBlackList struct {
}

func (d DeleteFromBlackList) Help() string {
	return "hello [arg0] [arg1] ... says hello to everyone"
}

func (d DeleteFromBlackList) Run(args []string) int {
	fmt.Println("Delete ip from bucket command", args)
	return 0
}

func (d DeleteFromBlackList) Synopsis() string {
	panic("implement me")
}

func (g *grpcCommands) Clean(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["clean"], arg.Args)
	var us storage.User
	us.Login = arg.Args[0]
	us.IP = arg.Args[1]
	g.cli.app.CleanBucket(us)
	fmt.Println(us.Login)
	fmt.Println(us.IP)
	fmt.Println("print CLEAN")

	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, err
}

func (g *grpcCommands) AddToWhiteList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["addToWhiteList"], arg.Args)
	var ip storage.IP
	ip.IP = arg.Args[0]
	ip.Mask = arg.Args[1]
	g.cli.app.AddToWhiteList(ctx, ip)
	fmt.Println("print AddToWhiteList")
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, err
}

func (g *grpcCommands) AddToBlackList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["addToBlackList"], arg.Args)
	var ip storage.IP
	ip.IP = arg.Args[0]
	ip.Mask = arg.Args[1]
	g.cli.app.AddToBlackList(ctx, ip)
	fmt.Println("print AddToBlackList")
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, err
}

func (g *grpcCommands) DeleteFromWhiteList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["deleteFromBlackList"], arg.Args)
	var ip storage.IP
	ip.IP = arg.Args[0]
	ip.Mask = arg.Args[1]
	g.cli.app.DeleteFromWhiteList(ctx, ip)
	fmt.Println("print DeleteFromWhiteList")
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, err
}

func (g *grpcCommands) DeleteFromBlackList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["deleteFromBlackList"], arg.Args)
	var ip storage.IP
	ip.IP = arg.Args[0]
	ip.Mask = arg.Args[1]
	g.cli.app.DeleteFromBlackList(ctx, ip)
	fmt.Println("print DeleteFromBlackList")
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, err
}

func (t *Clean) Help() string {
	return "hello [arg0] [arg1] ... says hello to everyone"
}

func (t *Clean) Run(args []string) int {
	fmt.Println("CleanBucket command", args)

	return 0
}

func (t *Clean) Synopsis() string {
	return "A sample command that says hello on stdout"
}

func (t *AddToWhiteList) Help() string {
	return "hello [arg0] [arg1] ... says hello to everyone"
}

func (t *AddToWhiteList) Run(args []string) int {
	fmt.Println("AddToWhiteList command", args)
	return 0
}

func (t *AddToWhiteList) Synopsis() string {
	return "A sample command that says hello on stdout"
}

func (t *AddToBlackList) Help() string {
	return "hello [arg0] [arg1] ... says hello to everyone"
}

func (t *AddToBlackList) Run(args []string) int {
	fmt.Println("AddToBlackList command", args)
	return 0
}

func (t *AddToBlackList) Synopsis() string {
	return "A sample command that says hello on stdout"
}

func wrapper(cf cli.CommandFactory, args []string) (int32, []byte, []byte, error) {
	var ret int32
	oldStdout := os.Stdout // keep backup of the real stdout
	oldStderr := os.Stderr

	// Backup the stdout
	r, w, err := os.Pipe()
	if err != nil {
		return ret, nil, nil, err
	}
	re, we, err := os.Pipe()
	if err != nil {
		return ret, nil, nil, err
	}
	os.Stdout = w
	os.Stderr = we

	runner, err := cf()
	if err != nil {
		return ret, nil, nil, err
	}
	ret = int32(runner.Run(args))

	outC := make(chan []byte)
	errC := make(chan []byte)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.Bytes()
	}()
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, re)
		errC <- buf.Bytes()
	}()

	// back to normal state
	w.Close()
	we.Close()
	os.Stdout = oldStdout // restoring the real stdout
	os.Stderr = oldStderr
	stdout := <-outC
	stderr := <-errC
	return ret, stdout, stderr, nil
}

func main() {
	c := cli.NewCLI("server", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"clean": func() (cli.Command, error) {
			return &Clean{}, nil
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
		listener, err := net.Listen("tcp", "127.0.0.1:1234")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		fmt.Println("1")
		grpcServer := grpc.NewServer()
		pb.RegisterAntifraudServiceServer(grpcServer, &grpcCommands{commands: c.Commands})
		fmt.Println("2")
		// determine whether to use TLS
		grpcServer.Serve(listener)
		fmt.Println("3")
	}
	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

func (s *CliC) RunCli() {
	c := cli.NewCLI("server", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"clean": func() (cli.Command, error) {
			return &Clean{}, nil
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
		listener, err := net.Listen("tcp", "127.0.0.1:1234")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		fmt.Println("1")
		grpcServer := grpc.NewServer()
		s.grpc = grpcServer

		pb.RegisterAntifraudServiceServer(grpcServer, &grpcCommands{commands: c.Commands, cli: s})
		fmt.Println("2")

		// determine whether to use TLS
		fmt.Println("3")
		grpcServer.Serve(listener)
	}
	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

type CliC struct {
	grpc *grpc.Server
	app  *app.App
}

func NewCli(app *app.App) *CliC {
	c := &CliC{
		app: app,
	}
	return c
}

func (s *CliC) Stop() {
	s.grpc.GracefulStop()
}
