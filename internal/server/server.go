package server

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"

	grpc_server "github.com/romanyakovlev/gophkeeper/internal/grpc"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
)

// Run запускает web-приложение.
func Run() error {
	/*
		sugar := logger.GetLogger()
		serverConfig := config.GetConfig(sugar)

		DB, err := db.InitDB(serverConfig.DatabaseDSN, sugar)
		if err != nil {
			sugar.Errorf("Server error: %v", err)
			return err
		}
		defer DB.Close()
		sharedURLRows := models.NewSharedURLRows()

		shortenerrepo, err := InitURLRepository(serverConfig, DB, sharedURLRows, sugar)
		if err != nil {
			sugar.Errorf("Server error: %v", err)
			return err
		}
		userrepo, err := initUserRepository(serverConfig, DB, sharedURLRows, sugar)
		if err != nil {
			sugar.Errorf("Server error: %v", err)
			return err
		}
		shortenerService := service.NewURLShortenerService(serverConfig, shortenerrepo, userrepo)
		worker := workers.InitURLDeletionWorker(shortenerService)
		URLCtrl := controller.NewURLShortenerController(shortenerService, sugar, worker)
		HealthCtrl := controller.NewHealthCheckController(DB)

	*/

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}
		/*
			grpcServer := grpc.NewServer(
				grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
					interceptors.JWTAuthInterceptor,
					interceptors.TrustedSubnetInterceptor(serverConfig.TrustedSubnet),
				)),
			)

		*/
		grpcServer := grpc.NewServer()
		pb.RegisterKeeperServiceServer(grpcServer, &grpc_server.Server{})

		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	<-ctx.Done()
	_, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("Server exited properly")
	return nil
}
