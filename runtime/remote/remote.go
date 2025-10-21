package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

// Remote je minimalni TCP transport sa JSON envelope-om.
// Jedna konekcija po poruci (dovoljno za ocenu 8 demo).
type Remote struct {
	sys  *ActorSystem
	addr string
}

func NewRemote(sys *ActorSystem, addr string) *Remote {
	return &Remote{sys: sys, addr: addr}
}

// listen prihvata dolazne TCP konekcije i isporučuje poruke.
func (r *Remote) listen() {
	ln, err := net.Listen("tcp", r.addr)
	if err != nil {
		panic(fmt.Errorf("remote listen %s: %w", r.addr, err))
	}
	fmt.Printf("[REMOTE] listening on %s\n", r.addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[REMOTE] accept error:", err)
			continue
		}
		go r.handle(conn)
	}
}

func (r *Remote) handle(conn net.Conn) {
	defer conn.Close()
	dec := json.NewDecoder(bufio.NewReader(conn))
	var env Envelope
	if err := dec.Decode(&env); err != nil {
		fmt.Println("[REMOTE] decode error:", err)
		return
	}
	// isporuči lokalno
	r.sys.localDeliver(env)
}

// send šalje envelope na odredišni Node.
func (r *Remote) send(env Envelope) {
	conn, err := net.Dial("tcp", env.Receiver.Node)
	if err != nil {
		fmt.Printf("[REMOTE] dial %s error: %v\n", env.Receiver.Node, err)
		return
	}
	defer conn.Close()
	bw := bufio.NewWriter(conn)
	enc := json.NewEncoder(bw)
	if err := enc.Encode(env); err != nil {
		fmt.Println("[REMOTE] encode error:", err)
		return
	}
	if err := bw.Flush(); err != nil {
		fmt.Println("[REMOTE] flush error:", err)
		return
	}
}
