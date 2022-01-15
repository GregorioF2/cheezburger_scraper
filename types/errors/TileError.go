package errors

import "fmt"

type TileError struct {
	Title string
	Err   string
}

func (m *TileError) Error() string {
	return fmt.Sprintf("Tile operation error %s:: %s", m.Title, m.Err)
}
