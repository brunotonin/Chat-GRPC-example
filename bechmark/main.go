package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	chat "chat-grpc/chat/proto"

	"google.golang.org/grpc"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type Result struct {
	mu        sync.Mutex
	latencies []time.Duration
	totalMsgs int
}

func (r *Result) Add(lat time.Duration, msgs int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.latencies = append(r.latencies, lat)
	r.totalMsgs += msgs
}

func (r *Result) AvgLatency() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.latencies) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range r.latencies {
		total += d
	}
	return total / time.Duration(len(r.latencies))
}

func simulateUser(id int, totalUsers int, connectWg *sync.WaitGroup, startCh chan struct{}, result *Result) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Printf("[User %d] erro ao conectar: %v", id, err)
		connectWg.Done()
		return
	}
	defer conn.Close()

	client := chat.NewChatServiceClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Printf("[User %d] erro ao abrir stream: %v", id, err)
		connectWg.Done()
		return
	}

	userName := fmt.Sprintf("User%d", id)
	expected := totalUsers - 1
	recvCount := 0
	recvDone := make(chan struct{})

	connectWg.Done()

	go func() {
		for {
			_, err := stream.Recv()
			if err != nil {
				break
			}

			recvCount++
			if recvCount >= expected {
				break
			}
		}
		close(recvDone)
	}()

	time.Sleep(2 * time.Second)
	<-startCh

	err = stream.Send(&chat.ChatMessage{
		User:    userName,
		Message: fmt.Sprintf("Oi, eu sou %s", userName),
	})
	if err != nil {
		return
	}
	start := time.Now()

	<-recvDone
	result.Add(time.Since(start), recvCount)
}

func runSimulation(users int) (time.Duration, int) {
	var connectWg sync.WaitGroup
	var finishWg sync.WaitGroup
	result := &Result{}
	startCh := make(chan struct{})

	connectWg.Add(users)
	finishWg.Add(users)
	for i := 1; i <= users; i++ {
		go func(id int) {
			defer finishWg.Done()
			simulateUser(id, users, &connectWg, startCh, result)
		}(i)
	}

	connectWg.Wait()
	close(startCh)
	finishWg.Wait()

	return result.AvgLatency(), result.totalMsgs
}

func exportGraph(xVals []float64, yVals []float64, filename string) error {
	p := plot.New()
	p.Title.Text = "Latência média x Número de mensagens"
	p.X.Label.Text = "Número de mensagens"
	p.Y.Label.Text = "Latência média (ms)"

	pts := make(plotter.XYs, len(xVals))
	for i := range xVals {
		pts[i].X = xVals[i]
		pts[i].Y = yVals[i]
	}

	// Adiciona linha conectando os pontos
	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	line.Color = plotter.DefaultLineStyle.Color
	p.Add(line)

	// Adiciona todos os pontos como scatter
	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		return err
	}
	scatter.Shape = draw.CircleGlyph{}
	scatter.Color = plotter.DefaultLineStyle.Color
	p.Add(scatter)

	return p.Save(6*vg.Inch, 4*vg.Inch, filename)
}

func main() {
	users := 50
	var xVals, yVals []float64

	for {
		fmt.Printf("\n▶ Rodando simulação com %d usuários...\n", users)
		avg, msgs := runSimulation(users)

		fmt.Printf("   Latência média: %v\n", avg)
		fmt.Printf("   Mensagens processadas: %d\n", msgs)

		xVals = append(xVals, float64(msgs))
		yVals = append(yVals, float64(avg.Milliseconds()))

		nextUsers := users + users/3
		if nextUsers > 1000 {
			fmt.Println("⚠ Limite de 1000 usuários atingido, encerrando teste.")
			break
		}
		users = nextUsers
		time.Sleep(10 * time.Second)
	}

	// Exporta gráfico no final
	if err := exportGraph(xVals, yVals, "latencia_vs_mensagens.png"); err != nil {
		log.Fatalf("Erro ao exportar gráfico: %v", err)
	}
	fmt.Println("✅ Gráfico exportado para latencia_vs_mensagens.png")
}
