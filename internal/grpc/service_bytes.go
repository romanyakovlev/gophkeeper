package grpc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
	"path/filepath"
)

type BytesData struct {
	ID       uuid.UUID
	FileName string
	UserID   uuid.UUID
}

var bytesDataSlice []BytesData

func (s *Server) GetBytes(req *pb.GetBytesRequest, stream pb.KeeperService_GetBytesServer) error {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to parse ID: %v", err)
	}

	var fileName string
	var objID uuid.UUID

	// Find the file metadata in bytesDataSlice
	var bytesData *BytesData
	for _, bd := range bytesDataSlice {
		if bd.ID == id {
			fileName = bd.FileName
			objID = bd.ID
			break
		}
	}
	/*
		if bytesData == nil {
			return status.Errorf(codes.NotFound, "file with ID %s not found", req.ID)
		}

	*/
	if fileName == "" {
		return status.Errorf(codes.NotFound, "filename %s is empty", bytesData.FileName)
	}

	// Send the filename first
	if err := stream.Send(&pb.GetBytesResponse{
		Data: &pb.GetBytesResponse_Filename{Filename: fileName},
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to send filename: %v", err)
	}

	// Open file and stream it
	filePath := filepath.Join("uploaded_files", objID.String())
	file, err := os.Open(filePath)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to open file: %v", err)
	}
	defer file.Close()

	hasher := sha256.New()
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(codes.Internal, "failed to read file: %v", err)
		}

		chunk := buf[:n]
		hasher.Write(chunk)

		if err := stream.Send(&pb.GetBytesResponse{
			Data: &pb.GetBytesResponse_Chunk{Chunk: chunk},
		}); err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk: %v", err)
		}
	}

	// After all chunks have been sent, send the hash sum
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))
	if err := stream.Send(&pb.GetBytesResponse{
		Data: &pb.GetBytesResponse_HashSum{HashSum: calculatedHash},
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to send hash sum: %v", err)
	}

	return nil
}

/*
func (s *Server) SaveBytes(ctx context.Context, in *pb.SaveBytesRequest) (*pb.SaveBytesResponse, error) {

	objID := uuid.New()
	bytesDataSlice = append(bytesDataSlice, BytesData{
		ID:     objID,
		Bytes:  in.Bytes,
		Name:   in.Name,
		UserID: uuid.New(),
	})

	return &pb.SaveBytesResponse{ID: objID.String()}, nil

}

*/

func (s *Server) SaveBytes(stream pb.KeeperService_SaveBytesServer) error {
	var filename string
	var hashSum string
	hasher := sha256.New()

	req, err := stream.Recv()
	if err != nil {
		return err
	}
	filename = req.GetFilename()
	if filename == "" {
		return fmt.Errorf("first message did not contain a filename")
	}

	objID := uuid.New()
	saveDir := "uploaded_files"
	filePath := filepath.Join(saveDir, objID.String())

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer outFile.Close()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		chunk := req.GetChunk()
		if chunk != nil {
			if _, err := outFile.Write(chunk); err != nil {
				return fmt.Errorf("failed to write chunk to file: %v", err)
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
		return fmt.Errorf("hash mismatch: received %s, calculated %s", hashSum, calculatedHash)
	}

	bytesDataSlice = append(bytesDataSlice, BytesData{
		ID:       objID,
		FileName: filename,
		UserID:   uuid.New(),
	})

	if err := stream.SendAndClose(&pb.SaveBytesResponse{
		ID: objID.String(),
	}); err != nil {
		return fmt.Errorf("failed to send status response: %v", err)
	}

	return nil
}

/*
	func (s *Server) SaveBytes(ctx context.Context, in *BytesData) (*string, error) {
		objID := uuid.New()

		// User provided bucket name
		bucketName := "your-bucket-name"

		// Generate an object name based on the UUID and/or other unique data
		objectName := fmt.Sprintf("%s-%s", objID, in.Name)

		// Using PutObject to upload to MinIO
		info, err := s.MinioClient.PutObject(ctx, bucketName, objectName, in.Bytes, int64(len(in.Bytes)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}

		fmt.Printf("Successfully uploaded bytes to %s of size %dB\n", info.Key, info.Size)

		objIDStr := objID.String()
		return &objIDStr, nil
	}
*/

/*

Sending large files using gRPC involves utilizing gRPC streaming, which can break down a large file into smaller chunks and send these chunks sequentially over a single open gRPC stream. This process helps to manage memory usage efficiently and avoids the need to load the entire file into memory before sending. gRPC supports different types of streaming: client streaming, server streaming, and bidirectional streaming. For sending large files from a client to a server, client streaming is typically used.

Here’s a step-by-step guide on how to implement client streaming for sending large files using gRPC in Go:

### 1. Define the gRPC Service in a `.proto` File

First, define the service and messages in a protobuf file. This specifies the structure of the data you'll be sending over the stream and the service methods.

```proto
syntax = "proto3";

package filetransfer;

service FileService {
  rpc Upload(stream Chunk) returns (UploadStatus) {}
}

message Chunk {
  bytes content = 1;
}

message UploadStatus {
  bool success = 1;
  string message = 2;
}
```

This defines an `Upload` RPC method where the client streams `Chunk` messages to the server, which contain portions of the file as byte arrays. After uploading, the server responds with an `UploadStatus` indicating the success or failure of the operation.

### 2. Generate Go Code from the `.proto` File

Use the protoc compiler with the Go plugins to generate Go code from your `.proto` file. You need the `protoc-gen-go` and `protoc-gen-go-grpc` plugins installed.

```sh
protoc --go_out=. --go-grpc_out=. path/to/your/filetransfer.proto
```

### 3. Implement the Server

Implement the gRPC server, including the `Upload` function to receive the file chunks.

```go
package main

import (
    "context"
    "io"
    "log"
    "net"

    "google.golang.org/grpc"
    pb "path/to/your/generated/protos/filetransfer"
)

type server struct {
    pb.UnimplementedFileServiceServer
}

func (s *server) Upload(stream pb.FileService_UploadServer) error {
    // Example: store the file's contents in a buffer
    var content []byte

    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            // End of file stream
            break
        }
        if err != nil {
            return err
        }

        // Append the chunk's content to the buffer
        content = append(content, chunk.Content...)
    }

    // Process the received content, e.g., save to a file
    // ...

    return stream.SendAndClose(&pb.UploadStatus{
        Success: true,
        Message: "File uploaded successfully.",
    })
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()
    pb.RegisterFileServiceServer(s, &server{})
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

### 4. Implement the Client

Implement the client to read the file in chunks and send each chunk to the server using the client-streaming API.

```go
package main

import (
    "context"
    "log"
    "os"
    "path/filepath"

    "google.golang.org/grpc"
    pb "path/to/your/generated/protos/filetransfer"
)

func uploadFile(client pb.FileServiceClient, filename string) {
    stream, err := client.Upload(context.Background())
    if err != nil {
        log.Fatalf("could not upload file: %v", err)
    }

    file, err := os.Open(filename)
    if err != nil {
        log.Fatalf("could not open file %s: %v", filename, err)
    }
    defer file.Close()

    buf := make([]byte, 1024) // Buffer for reading file chunks
    for {
        n, err := file.Read(buf)
        if err == io.EOF {
            break // End of file
        }
        if err != nil {
            log.Fatalf("could not read chunk from file: %v", err)
        }

        // Send the chunk to the server
        if err := stream.Send(&pb.Chunk{Content: buf[:n]}); err != nil {
            log.Fatalf("could not send file chunk to server: %v", err)
        }
    }

    status, err := stream.CloseAndRecv()
    if err != nil {
        log.Fatalf("could not receive upload status: %v", err)
    }

    log.Printf("Upload finished with message: %s", status.GetMessage())
}

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewFileServiceClient(conn)

    // Assuming you have a file named "large_file.txt" in the same directory
    filepath := filepath.Join(".", "large_file.txt")
    uploadFile(client, filepath)
}
```

### 5. Run the Server and Client

First, run your server:

```sh
go run server.go
```

Then, run your client by passing the path to the large file you want to upload:

```sh
go run client.go
```

This demonstrates a basic implementation of uploading a file using client streaming in gRPC with Go. Depending on your requirements, you may want to add additional features like error handling, metadata exchange (e.g., filename, MIME type), and security measures (e.g., authentication, encryption).


*/
//
//
//
//
//
//
//
//
//
//
//
//

/*

Yes, you can use MinIO to save a file each time a chunk arrives, especially when you are implementing a gRPC streaming service for uploading large files in chunks. Instead of accumulating the entire file content in server memory or a temporary location before saving it to MinIO, you can write chunks directly to MinIO as they arrive. This approach is efficient for memory usage and allows for handling very large files.

To achieve this, you would use MinIO's ability to create an object by streaming chunks of data into it. This process typically involves using the MinIO client's `PutObject` method in a way that it supports streaming.

Here's an adjustment to the server-side implementation of the `Upload` function from the earlier streaming example. This version writes each received chunk directly to MinIO:

### Adjusted Server Implementation for MinIO

First, ensure you have MinIO client imported and configured:

```go
import (
    "context"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    // Other imports...
)

// Initialize a global MinIO client
var minioClient *minio.Client

func init() {
    var err error
    minioClient, err = minio.New("minio-server:9000", &minio.Options{
        Creds:  credentials.NewStaticV4("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
        Secure: false,
    })
    if err != nil {
        log.Fatalln(err)
    }
}
```

Modify your `Upload` function like so:

```go
func (s *server) Upload(stream pb.FileService_UploadServer) error {
    objID := uuid.New().String() // Unique object ID
    bucketName := "your-bucket-name"
    objectName := fmt.Sprintf("%s-file", objID) // Create a unique object name

    // Create a pipe: writer side will be used for PutObject, reader for receiving chunks
    reader, writer := io.Pipe()

    // Use a channel to signal when the upload is complete
    doneCh := make(chan struct{})
    errCh := make(chan error)

    // Start a goroutine to deal with MinIO upload
    go func() {
        // Uploading data to MinIO as it arrives
        _, err := minioClient.PutObject(context.Background(), bucketName, objectName, reader, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
        if err != nil {
            errCh <- err
            return
        }
        doneCh <- struct{}{}
    }()

    // Receive chunks from gRPC stream and write to MinIO through the pipe
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            // Close the writer when done reading to signal EOF to the PutObject
            writer.Close()
            break
        }
        if err != nil {
            writer.CloseWithError(err)
            return err
        }

        // Writing received chunk to the pipe
        if _, err := writer.Write(chunk.Content); err != nil {
            return err
        }
    }

    // Wait for the upload to finish or error
    select {
    case <-doneCh:
        return stream.SendAndClose(&pb.UploadStatus{Success: true, Message: "File uploaded successfully."})
    case err := <-errCh:
        return err
    }
}
```

### Explanation

- When a chunk arrives from the client, it is immediately written to a pipe (`io.Pipe()`). The writer end of the pipe is bound to the gRPC service, where chunks are written as they are received. The reader end is passed to `minioClient.PutObject`, which reads from this pipe as if it was a file.
- Using `io.Pipe()` allows you to start uploading parts of the file to MinIO as soon as they arrive without waiting for the entire file to be received. This effectively decreases memory usage and can support very large files.
- The call to `minioClient.PutObject` is done in a separate goroutine to allow for simultaneous receiving of chunks via gRPC and uploading these chunks to MinIO.
- Proper error handling and cleanup are essential when dealing with IO operations and goroutines like this to avoid memory leaks or goroutine leaks.

This approach effectively makes your gRPC server act as a proxy, receiving file chunks from the client and streaming those chunks directly to MinIO. It's efficient for handling large files and can be adapted for various streaming data scenarios.

*/
