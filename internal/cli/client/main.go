package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/Remneva/anti-bruteforce/internal/server/pb"
	"google.golang.org/grpc"
)

func main() {
	ipaddr := "localhost"
	port := "12344"
	fmt.Printf("ipaddr: %s, port %s\n", ipaddr, port) // nolint:forbidigo
	addr := net.JoinHostPort(ipaddr, port)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewAntiBruteForceServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	args := os.Args[0:]
	var resp interface{}
	switch {
	case args[0] == "clean":
		resp, _ = c.CleanBucket(ctx, &pb.CleanBucketRequest{
			User: &pb.User{
				Login: os.Args[1],
				Ip:    os.Args[2],
			},
		})
	case args[0] == "addToWhiteList":
		resp, _ = c.AddToWhiteList(ctx, &pb.AddToWhiteListRequest{
			Ip: &pb.Ip{
				Ip:   os.Args[1],
				Mask: os.Args[2],
			},
		})
	case args[0] == "deleteFromWhiteList":
		resp, _ = c.DeleteFromWhiteList(ctx, &pb.DeleteFromWhiteListRequest{
			Ip: &pb.Ip{
				Ip:   os.Args[1],
				Mask: os.Args[2],
			},
		})
	case args[0] == "addToBlackList":
		resp, _ = c.AddToBlackList(ctx, &pb.AddToBlackListRequest{
			Ip: &pb.Ip{
				Ip:   os.Args[1],
				Mask: os.Args[2],
			},
		})
	case args[0] == "deleteFromBlackList":
		resp, _ = c.DeleteFromBlackList(ctx, &pb.DeleteFromBlackListRequest{
			Ip: &pb.Ip{
				Ip:   os.Args[1],
				Mask: os.Args[2],
			},
		})
	}
	log.Println(resp)
}
