package examples

import (
	"context"

	api "github.com/dcaf-labs/go-retryable-http-client"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/sirupsen/logrus"
)

func GetSolanaGoRPCClient(retryClientProvider api.RetryableHTTPClientProvider) (*rpc.Client, error) {
	url, callsPerSecond := rpc.MainNetBeta_RPC, 1.0
	options := api.GetDefaultRateLimitHTTPClientOptions()
	options.CallsPerSecond = callsPerSecond
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: retryClientProvider(options),
	}
	rpcClient := rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(url, opts))

	_, err := rpcClient.GetVersion(context.Background())
	if err != nil {
		logrus.WithError(err).Fatalf("failed to get clients version info")
		return nil, err
	}
	return rpcClient, nil
}
