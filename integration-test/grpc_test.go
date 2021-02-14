package test

import (
	"context"
	"fmt"
	"github.com/Remneva/anti-bruteforce/internal/server/pb"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"google.golang.org/grpc"
	"testing"
)

func TestServerGRPC(t *testing.T) {
	conn, err := grpc.Dial("service:50051", grpc.WithInsecure())
	ctx := context.Background()
	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()

	client := pb.NewAntiBruteForceServiceClient(conn)
	request := &pb.AuthorizationRequest{
		Authorization: &pb.Authorization{
			Login:    "login",
			Password: "qwerty",
			Ip:       "192.1.1.0/25",
		},
	}

	t.Run("Authorization success", func(t *testing.T) {
		response, err := client.Auth(ctx, request)
		if err != nil {
			fmt.Printf("fail to dial: %v\n", err)
		}
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, true, response.Result)
		fmt.Println("result", response.Result)
	})

	t.Run("Authorization failed", func(t *testing.T) {
		response, err := client.Auth(ctx, request)
		if err != nil {
			fmt.Printf("fail to dial: %v\n", err)
		}
		require.NoError(t, err)
		response, err = client.Auth(ctx, request)
		if err != nil {
			fmt.Printf("fail to dial: %v\n", err)
		}
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, false, response.Result)
		fmt.Println("result", response.Result)
	})

	t.Run("Authorization success", func(t *testing.T) {
		request = &pb.AuthorizationRequest{
			Authorization: &pb.Authorization{
				Password: "qwerty",
			},
		}
		response, err := client.Auth(ctx, request)
		if err != nil {
			fmt.Printf("fail to dial: %v\n", err)
		}
		require.Error(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "", response.String())
	})

}