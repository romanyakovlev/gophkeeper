package grpc

import (
	"context"
	"github.com/google/uuid"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
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
