package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/Remneva/anti-bruteforce/internal/cli/pb"
	"google.golang.org/grpc"
)

func main() {
	ipaddr := "localhost"
	port := "1234"
	//	fmt.Printf("Addresses returned by LookupHost(%s): %v\n", "127.0.0.1")
	fmt.Printf("ipaddr: %s, port %s\n", ipaddr, port)
	addr := net.JoinHostPort(ipaddr, port)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewAntifraudServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	args := os.Args[1:]
	var retcode int32
	if args[0] == "clean" {
		fmt.Println("CLeAN")
		r, _ := c.Clean(ctx, &pb.Arg{Args: os.Args[1:]})
		retcode = r.Retcode
	} else if args[0] == "addToWhiteList" {
		fmt.Println("addToWhiteList")
		r, _ := c.AddToWhiteList(ctx, &pb.Arg{Args: os.Args[1:]})
		retcode = r.Retcode
	} else if args[0] == "deleteFromWhiteList" {
		fmt.Println("deleteFromWhiteList")
		r, _ := c.DeleteFromWhiteList(ctx, &pb.Arg{Args: os.Args[1:]})
		retcode = r.Retcode
	} else if args[0] == "addToBlackList" {
		fmt.Println("addToBlackList")
		r, _ := c.AddToBlackList(ctx, &pb.Arg{Args: os.Args[1:]})
		retcode = r.Retcode
	} else if args[0] == "deleteFromBlackList" {
		fmt.Println("deleteFromBlackList")
		r, _ := c.DeleteFromBlackList(ctx, &pb.Arg{Args: os.Args[1:]})
		retcode = r.Retcode
	}

	log.Println(">msg: ", retcode)
}
