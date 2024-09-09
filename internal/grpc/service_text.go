package grpc

import (
	"context"
	"github.com/google/uuid"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
)

type TextData struct {
	ID     uuid.UUID
	Text   string
	Name   string
	UserID uuid.UUID
}

var textDataSlice []TextData

func (s *Server) GetText(ctx context.Context, in *pb.GetTextRequest) (*pb.GetTextResponse, error) {

	objID, _ := uuid.Parse(in.ID)

	for _, element := range textDataSlice {
		if element.ID == objID {
			return &pb.GetTextResponse{Text: element.Text, Name: element.Name}, nil
		}
	}

	return &pb.GetTextResponse{}, nil

}

func (s *Server) SaveText(ctx context.Context, in *pb.SaveTextRequest) (*pb.SaveTextResponse, error) {

	objID := uuid.New()
	textDataSlice = append(textDataSlice, TextData{
		ID:     objID,
		Text:   in.Text,
		Name:   in.Name,
		UserID: uuid.New(),
	})

	return &pb.SaveTextResponse{ID: objID.String()}, nil

}
