package game

type Character int

const (
	Empty Character = iota
	Black
	White
	Wall
	None
)

func CharacterToStr(c Character) string {
	switch c {
	case Black:
		return "○"
	case White:
		return "◉"
	case Empty:
		return " "
	}
	return ""
}

func OpponentCharacter(me Character) Character {
	switch me {
	case Black:
		return White
	case White:
		return Black
	}
	panic("invalid state")
}
