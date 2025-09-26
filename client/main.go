package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	chat "chat-grpc/chat/proto"
	echo "chat-grpc/echo/proto"

	"google.golang.org/grpc"
)

func simulateUser(id int, wg *sync.WaitGroup, done <-chan struct{}) {
	defer wg.Done()

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Printf("[User %d] não conectou: %v", id, err)
		return
	}
	defer conn.Close()

	echoClient := echo.NewEchoServiceClient(conn)
	_, err = echoClient.Echo(context.Background(), &echo.EchoRequest{Message: fmt.Sprintf("Hello from User %d", id)})
	if err != nil {
		log.Printf("[User %d] erro no echo: %v", id, err)
	}

	client := chat.NewChatServiceClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Printf("[User %d] erro ao iniciar chat: %v", id, err)
		return
	}

	// Recebendo mensagens do servidor
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_, err := stream.Recv()
				if err != nil {
					log.Printf("[User %d] conexão encerrada: %v", id, err)
					return
				}
			}
		}
	}()

	username := fmt.Sprintf("User%d", id)
	log.Printf("[%s] conectado", username)

	// Envia mensagens para sempre
	for {
		select {
		case <-done:
			err := stream.CloseSend()
			if err != nil {
				log.Printf("[%s] erro ao fechar stream: %v", username, err)
			} else {
				log.Printf("[%s] stream encerrado com sucesso", username)
			}
			return
		default:
			time.Sleep(1 * time.Second)
			msg := &chat.ChatMessage{
				User:      username,
				Message:   fmt.Sprintf("Oi, eu sou %s! %d", username, rand.Intn(1000)),
				Timestamp: time.Now().Unix(),
			}
			if err := stream.Send(msg); err != nil {
				log.Printf("[%s] erro ao enviar: %v", username, err)
				return
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run client.go <num_usuarios>")
		return
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n <= 0 {
		fmt.Println("Número inválido de usuários")
		return
	}

	var wg sync.WaitGroup
	wg.Add(n)

	done := make(chan struct{})

	// Captura sinais de encerramento
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Encerrando todos os usuários...")
		close(done)
	}()

	for i := 1; i <= n; i++ {
		go simulateUser(i, &wg, done)
	}

	wg.Wait()
	log.Println("Todos os usuários desconectados")
}
