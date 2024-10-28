package build

import (
	"fmt"

	"kazuki.matsumoto/reversi/game"
	"kazuki.matsumoto/reversi/gen/pb"
)

func Room(r *pb.Room) *game.Room {
	return &game.Room{
		ID:    r.GetId(),
		Host:  Player(r.GetHost()),
		Guest: Player(r.GetGuest()),
	}
}

func Player(p *pb.Player) *game.Player {
	return &game.Player{
		ID:        p.GetId(),
		Character: Character(p.GetCharacter()),
	}
}

func Character(c pb.Character) game.Character {
	switch c {
	case pb.Character_BLACK:
		return game.Black
	case pb.Character_WHITE:
		return game.White
	case pb.Character_EMPTY:
		return game.Empty
	case pb.Character_WALL:
		return game.Wall
	}

	panic(fmt.Sprintf("unknwon color=%v", c))
}
