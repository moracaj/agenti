package runtime

import (
	"fmt"
	"sync"
)

type ActorSystem struct {
	addr    string // host:port
	mu      sync.RWMutex
	actors  map[string]*actorCell // Path -> cell
	remote  *Remote // TCP server/klijent
}

func NewActorSystem(addr string) *ActorSystem {
	s := &ActorSystem{
		addr:   addr,
		actors: make(map[string]*actorCell),
	}
	s.remote = NewRemote(s, addr)
	// start TCP server
	go s.remote.listen()
	return s
}

func (s *ActorSystem) Address() string { return s.addr }

func (s *ActorSystem) Spawn(name string, behavior Receive) PID {
	pid := PID{Node: s.addr, Path: "/user/" + name}
	cell := &actorCell{
		pid:      pid,
		behavior: behavior,
		mbox:     make(chan Envelope, 1024),
		sys:      s,
	}
	s.mu.Lock()
	s.actors[pid.Path] = cell
	s.mu.Unlock()
	go cell.run()
	fmt.Printf("[SPAWN] %s\n", pid)
	return pid
}

func (s *ActorSystem) stop(pid PID) {
	s.mu.Lock()
	cell := s.actors[pid.Path]
	delete(s.actors, pid.Path)
	s.mu.Unlock()
	if cell != nil {
		cell.close()
	}
}

func (s *ActorSystem) localDeliver(env Envelope) {
	s.mu.RLock()
	cell := s.actors[env.Receiver.Path]
	s.mu.RUnlock()
	if cell == nil {
		fmt.Printf("[WARN] unknown local actor %s\n", env.Receiver.Path)
		return
	}
	cell.enqueue(env)
}

// Tell šalje poruku lokalno ili udaljeno.
func (s *ActorSystem) Tell(to PID, msg map[string]any, from PID) {
	env := Envelope{Receiver: to, Sender: from, Msg: msg}
	if to.Node == s.addr || to.Node == "" {
		s.localDeliver(env)
		return
	}
	// remote
	s.remote.send(env)
}
