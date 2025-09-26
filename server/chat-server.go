package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	chat "chat-grpc/chat/proto"

	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
)

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	exchangeName = "chat_exchange"
)

type ChatServer struct {
	chat.UnimplementedChatServiceServer
	conn    *amqp.Connection
	pubCh   *amqp.Channel
	mu      sync.Mutex
	clients map[string]chat.ChatService_ChatServer
}

func NewChatServer() *ChatServer {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Erro ao conectar RabbitMQ: %v", err)
	}

	pubCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("Erro ao criar canal global de publicação: %v", err)
	}

	err = pubCh.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,  // durable
		false, // autoDelete
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Erro ao declarar exchange: %v", err)
	}

	return &ChatServer{
		conn:    conn,
		pubCh:   pubCh,
		clients: make(map[string]chat.ChatService_ChatServer),
	}
}

func (s *ChatServer) Chat(stream chat.ChatService_ChatServer) error {
	clientID := fmt.Sprintf("%p", stream)
	log.Printf("Cliente %s conectado", clientID)

	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		"",    // nome gerado
		false, // não durable
		true,  // autoDelete
		true,  // exclusive
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return err
	}

	err = ch.QueueBind(q.Name, "", exchangeName, false, nil)
	if err != nil {
		ch.Close()
		return err
	}

	s.mu.Lock()
	s.clients[clientID] = stream
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		msgs, _ := ch.Consume(q.Name, "", true, true, false, true, nil)
		for d := range msgs {
			var msg chat.ChatMessage
			if err := proto.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Erro ao desserializar: %v", err)
				continue
			}
			if err := stream.Send(&msg); err != nil {
				log.Printf("[%s] erro ao enviar: %v", clientID, err)
				break
			}
		}
		// close(done)
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("[%s] erro ao receber: %v", clientID, err)
			close(done)
			break
		}

		data, err := proto.Marshal(msg)
		if err != nil {
			log.Printf("Erro ao serializar: %v", err)
			continue
		}

		err = s.pubCh.Publish(
			exchangeName,
			"",
			false,
			false,
			amqp.Publishing{
				ContentType: "application/protobuf",
				Body:        data,
				Timestamp:   time.Now(),
			},
		)
		if err != nil {
			log.Printf("Erro ao publicar: %v", err)
		}
	}

	<-done
	s.cleanupClient(clientID, ch, q.Name)
	return nil
}

func (s *ChatServer) cleanupClient(clientID string, ch *amqp.Channel, queueName string) {
	// Remove explicitamente a fila
	_, err := ch.QueueDelete(queueName, false, false, false)
	if err != nil {
		log.Printf("[%s] erro ao deletar fila: %v", clientID, err)
	}

	// Fecha o canal do cliente
	err = ch.Close()
	if err != nil {
		log.Printf("[%s] erro ao fechar canal: %v", clientID, err)
	}

	// Remove o cliente do mapa
	s.mu.Lock()
	delete(s.clients, clientID)
	s.mu.Unlock()

	log.Printf("Cliente %s desconectado e fila removida", clientID)
}
