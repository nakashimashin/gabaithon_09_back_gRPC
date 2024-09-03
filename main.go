package main

import (
	"fmt"
	"io"
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

type KeyGameSession struct {
	playerKeys map[string]int32
	gameOver   bool
	winnerID   string
	listeners  map[string]chan *gamepb.KeyCollectGameStatus
}

type MatchServer struct {
	gamepb.UnimplementedMatchServiceServer
	rooms         map[string]*KeyGameSession
	lock          sync.Mutex
	currentRoomId int
}

func NewMatchServer() *MatchServer {
	return &MatchServer{
		rooms:         make(map[string]*KeyGameSession),
		currentRoomId: 1,
	}
}

func NewKeyGameSession() *KeyGameSession {
	return &KeyGameSession{
		playerKeys: make(map[string]int32),
		gameOver:   false,
		winnerID:   "",
		listeners:  make(map[string]chan *gamepb.KeyCollectGameStatus),
	}
}

func (s *MatchServer) FindMatch(req *gamepb.MatchRequest, stream gamepb.MatchService_FindMatchServer) error {
	s.lock.Lock()
	gameTypeStr := strconv.Itoa(int(req.GameType))
	roomID := "room_" + gameTypeStr + "_" + strconv.Itoa(s.currentRoomId)

	session, exists := s.rooms[roomID]
	if !exists {
		session = &KeyGameSession{
			playerKeys: make(map[string]int32),
			gameOver:   false,
			winnerID:   "",
		}
		s.rooms[roomID] = session
	}
	session.playerKeys[req.PlayerId] = 0
	playerCount := len(session.playerKeys)
	s.lock.Unlock()

	if playerCount > 2 {
		return fmt.Errorf("too many players in room %s", roomID)
	}

	if playerCount == 2 {
		s.lock.Lock()
		s.currentRoomId++
		s.lock.Unlock()
	}

	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			s.lock.Lock()
			delete(s.rooms[roomID].playerKeys, req.PlayerId)
			if len(s.rooms[roomID].playerKeys) < 2 {
				delete(s.rooms, roomID)
			}
			s.lock.Unlock()
			return fmt.Errorf("timeout: no match found for Player ID: %s in room %s", req.PlayerId, roomID)
		case <-ticker.C:
			s.lock.Lock()
			if len(s.rooms[roomID].playerKeys) == 2 {
				for id := range s.rooms[roomID].playerKeys {
					if id != req.PlayerId {
						response := &gamepb.MatchResponse{
							Message:  "Match found with " + id + " in room " + roomID,
							RoomId:   roomID,
							PlayerId: id,
							GameType: req.GameType,
						}
						s.lock.Unlock()
						if err := stream.Send(response); err != nil {
							return err
						}
						return nil
					}
				}
			}
			s.lock.Unlock()
		}
	}
}

func (s *MatchServer) StartGame(req *gamepb.GameRequest, stream gamepb.MatchService_StartGameServer) error {
	s.lock.Lock()
	session, exists := s.rooms[req.RoomId]
	if !exists {
		s.lock.Unlock()
		return fmt.Errorf("room not found")
	}
	s.lock.Unlock()

	status := &gamepb.KeyCollectGameStatus{
		Message:    "Game started",
		RoomId:     req.RoomId,
		PlayerKeys: session.playerKeys,
		GameOver:   session.gameOver,
		WinnerId:   session.winnerID,
		GameType:   req.GameType,
	}
	if err := stream.Send(status); err != nil {
		return err
	}
	return nil
}

func (s *MatchServer) CollectKey(stream gamepb.MatchService_CollectKeyServer) error {
	for {
		req, err := stream.Recv() // クライアントからのリクエストを受け取る
		if err == io.EOF {
			return nil // クライアントからのストリーミング終了
		}
		if err != nil {
			return err // 受信エラー
		}

		s.lock.Lock()
		session, exists := s.rooms[req.RoomId]
		if !exists {
			session = NewKeyGameSession() // 新しいセッションを作成
			s.rooms[req.RoomId] = session
		}

		// キーの総数を更新
		session.playerKeys[req.PlayerId] = req.TotalKeys
		if req.TotalKeys >= 5 {
			session.gameOver = true
			session.winnerID = req.PlayerId
		}

		status := &gamepb.KeyCollectGameStatus{
			Message:    fmt.Sprintf("Player %s has collected %d keys", req.PlayerId, req.TotalKeys),
			RoomId:     req.RoomId,
			PlayerKeys: session.playerKeys,
			GameOver:   session.gameOver,
			WinnerId:   session.winnerID,
			GameType:   req.GameType,
		}

		s.lock.Unlock()

		if err := stream.Send(status); err != nil { // クライアントに状態を送信
			return err
		}
	}
}

func main() {
	port := 8081
	grpcServer := grpc.NewServer()
	gamepb.RegisterMatchServiceServer(grpcServer, NewMatchServer())
	reflection.Register(grpcServer)

	wrappedGrpc := grpcweb.WrapServer(grpcServer, grpcweb.WithCorsForRegisteredEndpointsOnly(false), grpcweb.WithOriginFunc(func(origin string) bool { return true }))

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
