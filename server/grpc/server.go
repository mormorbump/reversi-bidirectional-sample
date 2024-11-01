package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"kazuki.matsumoto/reversi/gen/pb"
	"kazuki.matsumoto/reversi/server/handler"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	port := 50052
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	pb.RegisterMatchingServiceServer(server, handler.NewMatchingHandler())
	pb.RegisterGameServiceServer(server, handler.NewGameHandler())

	reflection.Register(server)

	go func() {
		log.Printf("start gRPC server port: %v", port)
		server.Serve(lis)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	server.GracefulStop()
}
