package handler

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kazuki.matsumoto/reversi/build"
	"kazuki.matsumoto/reversi/game"
	"kazuki.matsumoto/reversi/gen/pb"
	"sync"
	"time"
)

type MatchingHandler struct {
	pb.UnimplementedMatchingServiceServer
	sync.RWMutex
	Rooms       map[int32]*game.Room
	maxPlayerID int32
}

func NewMatchingHandler() *MatchingHandler {
	return &MatchingHandler{
		Rooms: make(map[int32]*game.Room),
	}
}

func (h *MatchingHandler) JoinRoom(req *pb.JoinRoomRequest, stream pb.MatchingService_JoinRoomServer) error {
	ctx, cancel := context.WithTimeout(stream.Context(), 2*time.Minute)
	defer cancel()

	// h.roomsは複数のクライアントから同時にアクセスされるので、mutexで保護する。
	h.Lock()
	// Playerの新規作成
	me := &game.Player{
		ID: h.maxPlayerID,
	}

	// 空いている部屋を探す
	// 作成されているh.roomsのうち、guestがnilのやつを探す
	// roomsを全件探索するので、一つでもroomに空きがあれば必ずマッチする。
	for _, room := range h.Rooms {
		if room.Guest == nil {
			me.Character = game.White
			room.Guest = me
			err := stream.Send(&pb.JoinRoomResponse{
				Status: pb.JoinRoomResponse_MATCHED,
				Room:   build.PBRoom(room),
				Me:     build.PBPlayer(room.Guest),
			})
			if err != nil {
				return err
			}
			h.Unlock()
			fmt.Printf("matched room_id=%\n", room.ID)
			return nil
		}
	}

	// 部屋が空いてなかったら新規作成
	me.Character = game.Black
	room := &game.Room{
		ID:   int32(len(h.Rooms)) + 1,
		Host: me,
	}
	h.Rooms[room.ID] = room
	h.Unlock()

	err := stream.Send(&pb.JoinRoomResponse{
		Room:   build.PBRoom(room),
		Status: pb.JoinRoomResponse_WAITING,
	})
	if err != nil {
		return err
	}

	// このchはdeadlineのみを監視する
	// ここのgo routineの中で1秒おきにforを回すことで、room.Guestに値が入るまで待機することができる。
	// go routineを使っている理由は、非ブロッキングなので負荷が少ないこと、
	//select文でguestの参加(case <-ch)とcontextのdoneをトラッキングし、適切な処理ができること。
	// 並行処理なのでスレッドをまるまる使用しないことなどが挙げられる。(コルーチン)
	// 通常のforだとそのループが使われているスレッドがまるまる処理を待つことになってしまう。
	ch := make(chan int)
	go func(ch chan<- int) {
		for {
			// この前後でguestに値が入ったらstateの整合性が崩れるのでRLock
			h.RLock()
			guest := room.Guest
			h.RUnlock()

			if guest != nil {
				err := stream.Send(&pb.JoinRoomResponse{
					Status: pb.JoinRoomResponse_MATCHED,
					Room:   build.PBRoom(room),
					Me:     build.PBPlayer(room.Host),
				})
				if err != nil {
					return
				}
				ch <- 0
				break
			}
			time.Sleep(1 * time.Second)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}(ch)

	select {
	case <-ch:
	case <-ctx.Done():
		return status.Errorf(codes.DeadlineExceeded, "マッチングできませんでした。")
	}
	return nil
}
