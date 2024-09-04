package client

import (
	"context"
	"fmt"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
)

func (k *KeeperServiceClient) SaveCredentials(ctx context.Context, login, password string) (string, error) {
	req := &pb.SaveCredentialsRequest{
		Login:    login,
		Password: password,
	}

	resp, err := k.grpcClient.SaveCredentials(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error saving credentials: %w", err)
	}

	return resp.ID, nil
}

func (k *KeeperServiceClient) GetCredentials(ctx context.Context, id string) (*pb.GetCredentialsResponse, error) {
	req := &pb.GetCredentialsRequest{
		ID: id,
	}

	resp, err := k.grpcClient.GetCredentials(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error getting credentials: %w", err)
	}

	return resp, nil
}
