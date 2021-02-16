package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Remneva/anti-bruteforce/internal/server/pb"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"google.golang.org/grpc"
)

func TestServerGRPC(t *testing.T) {
	host := os.Getenv("INTEGRATION_TEST_SERVICE_HOST")
	if host == "" {
		host = "localhost:50051"
	} else {
		host = host + ":50051"
	}

	conn, err := grpc.Dial(host, grpc.WithInsecure())
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
		assert.Equal(t, true, response.Result.State)
	})

	t.Run("Authorization failed - Anti-Fraud Protection", func(t *testing.T) {
		response, err := client.Auth(ctx, request)
		require.NoError(t, err)
		response, err = client.Auth(ctx, request)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, false, response.Result.State)
	})

	t.Run("Authorization error", func(t *testing.T) {
		request = &pb.AuthorizationRequest{
			Authorization: &pb.Authorization{
				Password: "qwerty",
			},
		}
		_, err := client.Auth(ctx, request)
		require.Errorf(t, err, "Database query failed")
		assert.Equal(t, "rpc error: code = InvalidArgument desc = login, password or ip can`t be empty", err.Error())
	})

	t.Run("Add to black list success", func(t *testing.T) {
		requestBlackIp := &pb.AddToBlackListRequest{
			Ip: &pb.Ip{
				Ip: "193.3.3.0/25",
			},
		}
		_, err := client.AddToBlackList(ctx, requestBlackIp)
		require.NoError(t, err)
	})

	t.Run("Add to white list success", func(t *testing.T) {
		requestWhiteIp := &pb.AddToWhiteListRequest{
			Ip: &pb.Ip{
				Ip: "193.3.3.0/25",
			},
		}
		_, err := client.AddToWhiteList(ctx, requestWhiteIp)
		require.NoError(t, err)
	})

	t.Run("Delete from black list success", func(t *testing.T) {
		requestBlackIp := &pb.DeleteFromBlackListRequest{
			Ip: &pb.Ip{
				Ip: "193.3.3.0/25",
			},
		}
		_, err := client.DeleteFromBlackList(ctx, requestBlackIp)
		require.NoError(t, err)
	})

	t.Run("Delete from white list success", func(t *testing.T) {
		requestWhiteIp := &pb.DeleteFromWhiteListRequest{
			Ip: &pb.Ip{
				Ip: "193.3.3.0/25",
			},
		}
		_, err := client.DeleteFromWhiteList(ctx, requestWhiteIp)
		require.NoError(t, err)
	})

	t.Run("Delete from white list error", func(t *testing.T) {
		requestWhiteIp := &pb.DeleteFromWhiteListRequest{
			Ip: &pb.Ip{
				Ip: "194.4.4.0/25",
			},
		}
		_, err := client.DeleteFromWhiteList(ctx, requestWhiteIp)
		require.Error(t, err)
		assert.Equal(t, "rpc error: code = Internal desc = delete from white list error: ip does not exist in white list: 194.4.4.0/25", err.Error())
	})

	t.Run("Delete from black list error", func(t *testing.T) {
		requestBlackIp := &pb.DeleteFromBlackListRequest{
			Ip: &pb.Ip{
				Ip: "194.4.4.0/25",
			},
		}
		_, err := client.DeleteFromBlackList(ctx, requestBlackIp)
		require.Error(t, err)
		assert.Equal(t, "rpc error: code = Internal desc = delete from black list error: ip does not exist in black list: 194.4.4.0/25", err.Error())
	})
}
