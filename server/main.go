package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	chat "chat-grpc/chat/proto"
	echo "chat-grpc/echo/proto"

	"google.golang.org/grpc"

	_ "net/http/pprof"
)

var enablePprof bool

func init() {
	// Flag de terminal: --pprof=true
	flag.BoolVar(&enablePprof, "pprof", false, "Habilita pprof HTTP server")
	flag.Parse()
}

func main() {
	if enablePprof {
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
