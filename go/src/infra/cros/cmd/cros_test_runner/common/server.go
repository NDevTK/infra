package common

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectWithService connects with the service at the provided server address.
func ConnectWithService(ctx context.Context, serverAddress string) (*grpc.ClientConn, error) {
	var err error
	step, ctx := build.StartStep(ctx, "Connect to server")
	defer func() { step.End(err) }()

	logging.Infof(ctx, "Trying to connect with address %q with %s timeout", serverAddress, ServiceConnectionTimeout.String())

	conn, err := grpc.Dial(serverAddress, getGrpcDialOpts(ctx, ServiceConnectionTimeout)...)
	if err != nil {
		return nil, errors.Annotate(err, "error during connecting to service address %s: ", serverAddress).Err()
	}

	return conn, nil
}

// getGrpcDialOpts provides the grpc dial options used to connect with a service.
func getGrpcDialOpts(ctx context.Context, timeout time.Duration) []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	return opts
}
