package client

import (
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"google.golang.org/grpc"
)

// FileUploaderClient is a wrapper around the pb.FileServiceClient to provide simplified file upload functionality.
type KeeperServiceClient struct {
	grpcClient pb.KeeperServiceClient
}

// NewFileUploaderClient creates a new FileUploaderClient.
func NewKeeperServiceClient(grpcConn *grpc.ClientConn) *KeeperServiceClient {
	return &KeeperServiceClient{
		grpcClient: pb.NewKeeperServiceClient(grpcConn),
	}
}
