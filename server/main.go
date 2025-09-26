package main

import (
	"log"
	"net"
	"net/http"
	"os"

	chat "chat-grpc/chat/proto"
	echo "chat-grpc/echo/proto"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	_ "net/http/pprof"
)

func main() {
	godotenv.Load()

	if os.Getenv("ENABLE_PPROF") == "true" {
		go func() {
			addr := "localhost:6060"
			log.Printf("[PPROF] disponível em http://%s/debug/pprof/", addr)
			log.Println(http.ListenAndServe(addr, nil))
		}()
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Erro ao escutar: %v", err)
	}

	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, NewChatServer())
	echo.RegisterEchoServiceServer(grpcServer, NewEchoServer())

	log.Println("Servidor rodando na porta :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Erro ao rodar gRPC server: %v", err)
	}
}
