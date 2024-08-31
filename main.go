package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	gamepb "gabaithon-grpc-server/pkg/grpc"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type MatchServer struct {
	gamepb.UnimplementedMatchServiceServer
	rooms         map[string]chan *gamepb.MatchRequest
	lock          sync.Mutex
	currentRoomId int
}

func NewMatchServer() *MatchServer {
	return &MatchServer{
		rooms:         make(map[string]chan *gamepb.MatchRequest),
		currentRoomId: 1,
	}
}

func main() {
	port := 8081
	grpcServer := grpc.NewServer()
	gamepb.RegisterMatchServiceServer(grpcServer, NewMatchServer())
	reflection.Register(grpcServer)

	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
	)

	handler := func(resp http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(req) || wrappedGrpc.IsGrpcWebSocketRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)
			return
		}
		http.NotFound(resp, req)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}

	log.Printf("Server started on port %v", port)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Stopping server...")
	grpcServer.GracefulStop()
}

func (s *MatchServer) FindMatch(req *gamepb.MatchRequest, stream gamepb.MatchService_FindMatchServer) error {
	s.lock.Lock()
	roomID := strconv.Itoa(s.currentRoomId)
	queue, exists := s.rooms[roomID]
	if !exists {
		queue = make(chan *gamepb.MatchRequest, 2)
		s.rooms[roomID] = queue
	}
	queue <- req
	if len(queue) == 2 {
		s.currentRoomId++
	}
	s.lock.Unlock()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case peer := <-queue:
			if peer.PlayerId != req.PlayerId {
				response := &gamepb.MatchResponse{
					Message:  "Match found with " + peer.PlayerId + " in room " + roomID,
					RoomId:   roomID,
					PlayerId: peer.PlayerId,
				}
				if err := stream.Send(response); err != nil {
					return err
				}
				s.lock.Lock()
				s.currentRoomId++
				s.lock.Unlock()
				return nil
			} else {
				s.lock.Lock()
				queue <- peer
				s.lock.Unlock()
			}
		case <-timeout:
			s.lock.Lock()
			newQueue := make(chan *gamepb.MatchRequest, 2)
			for len(queue) > 0 {
				p := <-queue
				if p.PlayerId != req.PlayerId {
					newQueue <- p
				}
			}
			s.rooms[roomID] = newQueue
			s.lock.Unlock()
			return fmt.Errorf("timeout: no match found for Player ID: %s in room %s", req.PlayerId, roomID)
		}
	}
}
