package runtime

import "fmt"

// PID je jedinstvena adresa aktera u sistemu.
type PID struct {
	Node string // "host:port"
	Path string // npr. "/user/router"
}

func (p PID) String() string {
	return fmt.Sprintf("actor://%s%s", p.Node, p.Path)
}

// Envelope nosi sve što je potrebno za isporuku poruke.
type Envelope struct {
	Receiver PID            `json:"receiver"`
	Sender   PID            `json:"sender"`
	Msg      map[string]any `json:"msg"` // JSON-friendly, očekuje polje "kind"
}
