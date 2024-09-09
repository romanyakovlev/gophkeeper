package grpc

import (
	"context"
	"github.com/google/uuid"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CredentialsData struct {
	ID       uuid.UUID
	Login    string
	Password string
	UserID   uuid.UUID
}

var credentialsDataSlice []CredentialsData

func (s *Server) GetCredentials(ctx context.Context, in *pb.GetCredentialsRequest) (*pb.GetCredentialsResponse, error) {
	objID, _ := uuid.Parse(in.ID)

	for _, element := range credentialsDataSlice {
		if element.ID == objID {
			return &pb.GetCredentialsResponse{Login: element.Login, Password: element.Password}, nil
		}
	}

	return &pb.GetCredentialsResponse{}, nil
}

func (s *Server) SaveCredentials(ctx context.Context, in *pb.SaveCredentialsRequest) (*pb.SaveCredentialsResponse, error) {
	objID := uuid.New()
	credentialsDataSlice = append(credentialsDataSlice, CredentialsData{
		ID:       objID,
		Login:    in.Login,
		Password: in.Password,
		UserID:   uuid.New(),
	})

	elementsDataSlice = append(elementsDataSlice, ElementData{
		ID:     objID,
		Name:   in.Login,
		UserID: uuid.New(),
		Type:   "credentials",
	})

	return &pb.SaveCredentialsResponse{ID: objID.String()}, nil
}

// DeleteCredentials deletes credentials from the server's data slice based on the provided ID.
func (s *Server) DeleteCredentials(ctx context.Context, in *pb.DeleteCredentialsRequest) (*pb.DeleteCredentialsResponse, error) {
	objID, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, err
	}

	var found bool
	var indexToDelete int
	for index, element := range credentialsDataSlice {
		if element.ID == objID {
			found = true
			indexToDelete = index
			break
		}
	}

	if !found {
		return nil, status.Errorf(codes.NotFound, "credentials with ID %s not found", in.ID)
	}

	// Remove the BytesData from the slice
	credentialsDataSlice = append(credentialsDataSlice[:indexToDelete], credentialsDataSlice[indexToDelete+1:]...)

	for index, element := range elementsDataSlice {
		if element.ID == objID {
			indexToDelete = index
			break
		}
	}

	if !found {
		return nil, status.Errorf(codes.NotFound, "credentials with ID %s not found", in.ID)
	}

	// Remove the BytesData from the slice
	elementsDataSlice = append(elementsDataSlice[:indexToDelete], elementsDataSlice[indexToDelete+1:]...)

	if found {
		return &pb.DeleteCredentialsResponse{Success: true}, nil
	} else {
		return &pb.DeleteCredentialsResponse{Success: false}, nil
	}
}
