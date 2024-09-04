package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"io"
	"log"
	"os"
	"path/filepath"
)

func (k *KeeperServiceClient) GetBytes(ctx context.Context, id string) (string, error) {
	// Initiate the GetBytes stream request using the provided ID
	stream, err := k.grpcClient.GetBytes(ctx, &pb.GetBytesRequest{ID: id})
	if err != nil {
		return "", fmt.Errorf("error initiating GetBytes stream: %w", err)
	}
	var filename string
	var hashSum string
	hasher := sha256.New()

	req, err := stream.Recv()
	if err != nil {
		return "", err
	}
	filename = req.GetFilename()
	if filename == "" {
		return "", fmt.Errorf("first message did not contain a filename")
	}

	saveDir := "downloaded_files"
	filePath := filepath.Join(saveDir, filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer outFile.Close()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		chunk := req.GetChunk()
		if chunk != nil {
			if _, err := outFile.Write(chunk); err != nil {
				return "", fmt.Errorf("failed to write chunk to file: %v", err)
			}
			hasher.Write(chunk) // Update hash with the chunk
			continue
		}

		hashSum = req.GetHashSum()
		if hashSum != "" {
			break // Received the hash sum
		}
	}

	// Calculate and compare hash sum
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))
	if calculatedHash != hashSum {
		return "", fmt.Errorf("hash mismatch: received %s, calculated %s", hashSum, calculatedHash)
	}

	return filePath, nil
}

// UploadFile uploads a file to the server using gRPC streaming.
func (k *KeeperServiceClient) SaveBytes(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	stream, err := k.grpcClient.SaveBytes(ctx)
	if err != nil {
		return err
	}

	// First, send the filename
	filename := filepath.Base(filePath)
	if err := stream.Send(&pb.SaveBytesRequest{
		Data: &pb.SaveBytesRequest_Filename{Filename: filename},
	}); err != nil {
		return err
	}

	buf := make([]byte, 1024) // Adjust buffer size as needed
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			// End of file, break the loop
			break
		}
		if err != nil {
			return err
		}

		// Update hasher with the chunk
		print("\nstart:\n" + string(buf[:n]) + "\nend\n")
		hasher.Write(buf[:n])

		// Send the chunk to the server
		if err := stream.Send(&pb.SaveBytesRequest{
			Data: &pb.SaveBytesRequest_Chunk{Chunk: buf[:n]},
		}); err != nil {
			return err
		}
	}

	// Calculate the file hash
	hashSum := hex.EncodeToString(hasher.Sum(nil))

	// Send the hash sum to the server
	if err := stream.Send(&pb.SaveBytesRequest{
		Data: &pb.SaveBytesRequest_HashSum{HashSum: hashSum},
	}); err != nil {
		return err
	}

	// Close the stream and get the server's response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	log.Printf("Upload response: %s", resp.ID)

	return nil
}

func (k *KeeperServiceClient) DeleteBytes(ctx context.Context, id string) (*pb.DeleteBytesResponse, error) {
	req := &pb.DeleteBytesRequest{
		ID: id,
	}

	resp, err := k.grpcClient.DeleteBytes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error deleting bytes: %w", err)
	}

	return resp, nil
}
