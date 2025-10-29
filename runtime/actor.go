package runtime

import (
	"context"
	"fmt"
)

// Receive je funkcija koja obrađuje poruke.
type Receive func(ctx Context, msg map[string]any)

// Context interfejs iz perspektive aktera.
type Context interface {
	Self() PID
	Sender() PID
	System() *ActorSystem
	Become(Receive)
	Stop()
	Spawn(name string, behavior Receive) PID
}

// internalna implementacija konteksta
type actorContext struct {
	sys   *ActorSystem
	self  PID
	sender PID
	cell  *actorCell
}

func (c *actorContext) Self() PID               { return c.self }
func (c *actorContext) Sender() PID             { return c.sender }
func (c *actorContext) System() *ActorSystem    { return c.sys }
func (c *actorContext) Become(r Receive)        { c.cell.behavior = r }
func (c *actorContext) Stop()                   { c.sys.stop(c.self) }
func (c *actorContext) Spawn(name string, b Receive) PID { return c.sys.Spawn(name, b) }

// actorCell predstavlja "proces" aktera.
type actorCell struct {
	pid      PID
	behavior Receive
	mbox     chan Envelope
	sys      *ActorSystem
	stopped  bool
}

func (c *actorCell) run() {
	// lifecycle: :started
	c.mbox <- Envelope{Receiver: c.pid, Sender: PID{}, Msg: map[string]any{"kind":":started"}}

	for env := range c.mbox {
		ctx := &actorContext{sys: c.sys, self: c.pid, sender: env.Sender, cell: c}
		// recover od panike da ne srušimo ceo runtime
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[ERROR] actor %s panic: %v\n", c.pid, r)
				}
			}()
			c.behavior(ctx, env.Msg)
		}()
	}
	// lifecycle: :stopped
	fmt.Printf("[LIFECYCLE] %s stopped\n", c.pid)
}

// util: blokirajuće slanje u mailbox (back-pressure)
func (c *actorCell) enqueue(env Envelope) {
	if c.stopped { return }
	c.mbox <- env
}

// util: graceful stop
func (c *actorCell) close() {
	if !c.stopped {
		c.stopped = true
		close(c.mbox)
	}
	// :stopped će se ispisati u run()
	_ = context.Background()
}
