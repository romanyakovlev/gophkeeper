package grpc

import (
	"context"
	"github.com/google/uuid"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreditCardData struct {
	ID         uuid.UUID
	CardNumber string
	Exp        string
	CVV        string
	UserID     uuid.UUID
}

var creditCardDataSlice []CreditCardData

func (s *Server) GetCreditCard(ctx context.Context, in *pb.GetCreditCardRequest) (*pb.GetCreditCardResponse, error) {
	objID, _ := uuid.Parse(in.ID)

	for _, element := range creditCardDataSlice {
		if element.ID == objID {
			return &pb.GetCreditCardResponse{
				CardNumber: element.CardNumber,
				Exp:        element.Exp,
				Cvv:        element.CVV,
			}, nil
		}
	}

	return &pb.GetCreditCardResponse{}, nil
}

func (s *Server) SaveCreditCard(ctx context.Context, in *pb.SaveCreditCardRequest) (*pb.SaveCreditCardResponse, error) {
	objID := uuid.New()
	creditCardDataSlice = append(creditCardDataSlice, CreditCardData{
		ID:         objID,
		CardNumber: in.CardNumber,
		Exp:        in.Exp,
		CVV:        in.Cvv,
		UserID:     uuid.New(),
	})

	elementsDataSlice = append(elementsDataSlice, ElementData{
		ID:     objID,
		Name:   in.CardNumber,
		UserID: uuid.New(),
		Type:   "card",
	})
	return &pb.SaveCreditCardResponse{ID: objID.String()}, nil
}

// DeleteCreditCard deletes creditCard from the server's data slice based on the provided ID.
func (s *Server) DeleteCreditCard(ctx context.Context, in *pb.DeleteCreditCardRequest) (*pb.DeleteCreditCardResponse, error) {
	objID, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, err
	}

	var found bool
	var indexToDelete int
	for index, element := range creditCardDataSlice {
		if element.ID == objID {
			found = true
			indexToDelete = index
			break
		}
	}

	if !found {
		return nil, status.Errorf(codes.NotFound, "Credit card with ID %s not found", in.ID)
	}

	// Remove the BytesData from the slice
	creditCardDataSlice = append(creditCardDataSlice[:indexToDelete], creditCardDataSlice[indexToDelete+1:]...)

	for index, element := range elementsDataSlice {
		if element.ID == objID {
			indexToDelete = index
			break
		}
	}

	if !found {
		return nil, status.Errorf(codes.NotFound, "Credit card with ID %s not found", in.ID)
	}

	// Remove the BytesData from the slice
	elementsDataSlice = append(elementsDataSlice[:indexToDelete], elementsDataSlice[indexToDelete+1:]...)

	if found {
		return &pb.DeleteCreditCardResponse{Success: true}, nil
	} else {
		return &pb.DeleteCreditCardResponse{Success: false}, nil
	}
}
