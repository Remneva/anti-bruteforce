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

const text = "something strange happens"

type grpcCommands struct {
	commands map[string]cli.CommandFactory
	cli      *Servercli
}

func (s *Servercli) RunCli() {
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
		listener, err := net.Listen("tcp", "antifrod:1234")
		if err != nil {
			fmt.Println(err)
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		s.grpc = grpcServer

		pb.RegisterAntifraudServiceServer(grpcServer, &grpcCommands{commands: c.Commands, cli: s})

		err = grpcServer.Serve(listener)
		if err != nil {
			log.Fatalf("failed to start: %v", err)
		}
	}
	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}

func (g *grpcCommands) Clean(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["clean"], arg.Args)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	var us storage.User
	us.Login = arg.Args[0]
	us.IP = arg.Args[1]
	if err = g.cli.app.CleanBucket(ctx, us); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, nil
}

func (g *grpcCommands) AddToWhiteList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["addToWhiteList"], arg.Args)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	ip := parseToStorage(arg.Args[0], arg.Args[1])
	if err = g.cli.app.AddToWhiteList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, nil
}

func (g *grpcCommands) AddToBlackList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["addToBlackList"], arg.Args)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	ip := parseToStorage(arg.Args[0], arg.Args[1])
	if err = g.cli.app.AddToBlackList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, nil
}

func (g *grpcCommands) DeleteFromWhiteList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["deleteFromBlackList"], arg.Args)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	ip := parseToStorage(arg.Args[0], arg.Args[1])
	if err = g.cli.app.DeleteFromWhiteList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, nil
}

func (g *grpcCommands) DeleteFromBlackList(ctx context.Context, arg *pb.Arg) (*pb.Output, error) {
	ret, stdout, stderr, err := wrapper(g.commands["deleteFromBlackList"], arg.Args)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	ip := parseToStorage(arg.Args[0], arg.Args[1])
	if err = g.cli.app.DeleteFromBlackList(ctx, ip); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &pb.Output{Retcode: ret, Stdout: stdout, Stderr: stderr}, nil
}

func (t *Clean) Run(args []string) int {
	return 0
}

func (t *Clean) Synopsis() string {
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

func wrapper(cf cli.CommandFactory, args []string) (int32, []byte, []byte, error) {
	var ret int32
	oldStdout := os.Stdout // keep backup of the real stdout
	oldStderr := os.Stderr

	// Backup the stdout
	read, write, err := os.Pipe()
	if err != nil {
		return ret, nil, nil, fmt.Errorf("error: %w", err)
	}
	// Backup the stderr
	reade, writee, err := os.Pipe()
	if err != nil {
		return ret, nil, nil, fmt.Errorf("error: %w", err)
	}
	// assigne stdout and stderr to the input of the pipe
	os.Stdout = write
	os.Stderr = writee

	runner, err := cf()
	if err != nil {
		return ret, nil, nil, fmt.Errorf("error: %w", err)
	}
	ret = int32(runner.Run(args))

	outC := make(chan []byte)
	errC := make(chan []byte)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, err = io.Copy(&buf, read)
		outC <- buf.Bytes()
	}()
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, err = io.Copy(&buf, reade)

		errC <- buf.Bytes()
	}()

	// back to normal state
	write.Close()
	writee.Close()
	os.Stdout = oldStdout // restoring the real stdout
	os.Stderr = oldStderr
	stdout := <-outC
	stderr := <-errC
	return ret, stdout, stderr, nil
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

func parseToStorage(arg ...string) storage.IP {
	var ip storage.IP
	ip.IP = arg[0]
	ip.Mask = arg[1]
	return ip
}

type Clean struct {
}

func (t *Clean) Help() string {
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
