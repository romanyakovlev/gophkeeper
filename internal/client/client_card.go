package client

import (
	"context"
	"fmt"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
)

func (k *KeeperServiceClient) SaveCreditCard(ctx context.Context, number string, exp string, cvv string) (string, error) {
	req := &pb.SaveCreditCardRequest{
		CardNumber: number,
		Exp:        exp,
		Cvv:        cvv,
	}

	resp, err := k.grpcClient.SaveCreditCard(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error saving credit card: %w", err)
	}

	return resp.ID, nil
}

func (k *KeeperServiceClient) GetCreditCard(ctx context.Context, id string) (*pb.GetCreditCardResponse, error) {
	req := &pb.GetCreditCardRequest{
		ID: id,
	}

	resp, err := k.grpcClient.GetCreditCard(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error getting credit card: %w", err)
	}

	return resp, nil
}

func (k *KeeperServiceClient) DeleteCreditCard(ctx context.Context, id string) (*pb.DeleteCreditCardResponse, error) {
	req := &pb.DeleteCreditCardRequest{
		ID: id,
	}

	resp, err := k.grpcClient.DeleteCreditCard(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error deleting credit card: %w", err)
	}

	return resp, nil
}
