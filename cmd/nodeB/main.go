package main

import (
	"fmt"
	"time"
	rt "actorsys/runtime"
)

func WorkerActor(name string, dur time.Duration) rt.Receive {
	return func(ctx rt.Context, msg map[string]any) {
		kind, _ := msg["kind"].(string)
		switch kind {
		case ":started":
			fmt.Println("[", name, "] started on", ctx.Self())
		case "Task":
			id := int(msg["id"].(float64))
			payload, _ := msg["payload"].(string)
			time.Sleep(dur)
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
