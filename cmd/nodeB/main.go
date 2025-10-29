package main

import (
	"fmt"
	"time"
	rt "actorsys/runtime"
)

// Helper function to convert id field to int (handles both int and float64)
func getID(msg map[string]any) int {
	switch v := msg["id"].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func WorkerActor(name string, dur time.Duration) rt.Receive {
	return func(ctx rt.Context, msg map[string]any) {
		kind, _ := msg["kind"].(string)
		switch kind {
		case ":started":
			fmt.Println("[", name, "] started on", ctx.Self())
		case "Task":
			id := getID(msg)
			payload, _ := msg["payload"].(string)
			fmt.Printf("[%s] processing Task id=%d payload=%s\n", name, id, payload)
			time.Sleep(dur)
			fmt.Printf("[%s] completed Task id=%d, sending Result\n", name, id)
			ctx.System().Tell(ctx.Sender(), map[string]any{"kind":"Result", "id": id, "status": "ok:"+payload}, ctx.Self())
		default:
			fmt.Println("[", name, "] unknown:", msg)
		}
	}
}

func main(){
	sys := rt.NewActorSystem(":9002")
	_ = sys.Spawn("worker0", WorkerActor("worker0", 5*time.Millisecond))
	_ = sys.Spawn("worker1", WorkerActor("worker1", 8*time.Millisecond))
	_ = sys.Spawn("worker2", WorkerActor("worker2", 12*time.Millisecond))
	select{}
}
