package grpc

import (
	"context"
	"github.com/google/uuid"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedKeeperServiceServer
	//Shortener URLShortener
	//Worker    *workers.URLDeletionWorker
}

type ElementData struct {
	ID     uuid.UUID
	Name   string
	UserID uuid.UUID
	Type   string
}

var elementsDataSlice []ElementData

// Convert a slice of ElementData to a protobuf Elements message.
func convertToProto(elementsDataSlice []ElementData) *pb.Elements {
	protoElements := make([]*pb.Element, len(elementsDataSlice))

	for i, ed := range elementsDataSlice {
		protoElements[i] = &pb.Element{
			Id:     ed.ID.String(),
			Name:   ed.Name,
			UserId: ed.UserID.String(),
			Type:   ed.Type,
		}
	}

	return &pb.Elements{
		Elements: protoElements,
	}
}

// GetElements method implementation.
func (s *Server) GetElements(ctx context.Context, req *emptypb.Empty) (*pb.Elements, error) {
	return convertToProto(elementsDataSlice), nil
}

/*
func (s *Server) GetURLByID(ctx context.Context, in *pb.GetURLByIDRequest) (*pb.GetURLByIDResponse, error) {
	urlRow, ok := s.Shortener.GetURL(in.ShortURL)
	if urlRow.DeletedFlag {
		return &pb.GetURLByIDResponse{}, status.Error(codes.NotFound, "URL is deleted")
	}
	if ok {
		return &pb.GetURLByIDResponse{OriginalURL: urlRow.OriginalURL}, nil
	} else {
		return &pb.GetURLByIDResponse{}, status.Error(codes.InvalidArgument, "Bad request")
	}
}
*/
