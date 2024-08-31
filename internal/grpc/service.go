package grpc

import (
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
)

type Server struct {
	pb.UnimplementedKeeperServiceServer
	//Shortener URLShortener
	//Worker    *workers.URLDeletionWorker
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
