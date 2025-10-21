package main

import (
	"fmt"
	"time"
	rt "actorsys/runtime"
)

const MAX_INFLIGHT = 20
const RECOVER_LEVEL = 10

// RouterActor: active <-> throttled
func RouterActor(workers []rt.PID) rt.Receive {
	inflight := 0
	idx := 0
	pending := map[int]rt.PID{} // id -> original sender

	active := func(ctx rt.Context, msg map[string]any) {
		kind, _ := msg["kind"].(string)
		switch kind {
		case ":started":
			fmt.Println("[Router] started on", ctx.Self())
		case "Route":
			id := int(msg["id"].(float64)) // JSON numbers are float64
			payload, _ := msg["payload"].(string)
			target := workers[idx%len(workers)]
			idx++
			pending[id] = ctx.Sender()
			ctx.System().Tell(target, map[string]any{"kind":"Task", "id": id, "payload": payload}, ctx.Self())
			inflight++
			if inflight >= MAX_INFLIGHT {
				fmt.Println("[Router] >>> THROTTLED (inflight=", inflight, ")")
				ctx.Become(throttled)
			}
		case "Result":
			id := int(msg["id"].(float64))
			if orig, ok := pending[id]; ok {
				delete(pending, id)
				inflight--
				ctx.System().Tell(orig, msg, ctx.Self())
			}
		default:
			fmt.Println("[Router] unknown:", msg)
		}
	}

	throttled := func(ctx rt.Context, msg map[string]any) {
		kind, _ := msg["kind"].(string)
		switch kind {
		case "Route":
			id := int(msg["id"].(float64))
			fmt.Println("[Router] drop Route id", id, "(backpressure)")
		case "Result":
			id := int(msg["id"].(float64))
			if orig, ok := pending[id]; ok {
				delete(pending, id)
				inflight--
				ctx.System().Tell(orig, msg, ctx.Self())
			}
			if inflight <= RECOVER_LEVEL {
				fmt.Println("[Router] <<< RECOVER to active (inflight=", inflight, ")")
				ctx.Become(active)
			}
		}
	}

	return active
}

// Load generator that sends N Route messages then waits Result
func LoadGen(router rt.PID, N int) rt.Receive {
	completed := 0
	var start time.Time
	return func(ctx rt.Context, msg map[string]any) {
		kind, _ := msg["kind"].(string)
		switch kind {
		case ":started":
			fmt.Println("[LoadGen] started; sending", N, "jobs")
			start = time.Now()
			for i:=0;i<N;i++{
				ctx.System().Tell(router, map[string]any{"kind":"Route", "id": i, "payload": fmt.Sprintf("work-%d", i)}, ctx.Self())
			}
		case "Result":
			completed++
			if completed == N {
				elapsed := time.Since(start)
				pps := float64(N) / elapsed.Seconds()
				fmt.Printf("[LoadGen] done: %d results in %s (%.1f msg/s)\n", N, elapsed, pps)
			}
		}
	}
}

func main(){
	sys := rt.NewActorSystem(":9001")
	workers := []rt.PID{
		{Node: "127.0.0.1:9002", Path: "/user/worker0"},
		{Node: "127.0.0.1:9002", Path: "/user/worker1"},
		{Node: "127.0.0.1:9002", Path: "/user/worker2"},
	}
	router := sys.Spawn("router", RouterActor(workers))
	_ = sys.Spawn("loadgen", LoadGen(router, 200))
	select{}
}
