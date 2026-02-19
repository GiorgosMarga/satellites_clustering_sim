package node

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/GiorgosMarga/satellites/cluster"
	"github.com/GiorgosMarga/satellites/transport"
)

type Position struct {
	X float64
	Y float64
	Z float64
}

type Node struct {
	ID        int
	Neighbors map[int]*Node
	Transport transport.Transport
	Position
	ctx    context.Context
	cancel context.CancelFunc
	// For Clustering
	State       *cluster.LocalState
	clusterAlgo cluster.ClusterAlgo
}

func New(id int, pos []float64) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		ID:        id,
		Neighbors: make(map[int]*Node),
		Transport: transport.NewDefault(),
		Position: Position{
			X: pos[0],
			Y: pos[1],
			Z: pos[2],
		},
		ctx:         ctx,
		cancel:      cancel,
		State:       cluster.NewState(id, pos),
		clusterAlgo: cluster.NewLayeredClustering(),
	}
}

func (n *Node) Update(newPos []float64) {
	n.Position = Position{
		X: newPos[0],
		Y: newPos[1],
		Z: newPos[2],
	}
	n.State.Update(newPos)
}
func (n *Node) AddPeer(p *Node) {
	n.Neighbors[p.ID] = p
	n.Transport.AddPeer(p.ID, p.Transport.Chan())
	// TODO: fix this
	n.State.AddPeer(p.ID, []float64{p.X, p.Y, p.Z})
}

func (n *Node) Stop() {
	n.cancel()
}
func (n *Node) Start() {
	go n.Transport.Start(n.ctx)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case transMsg := <-n.Transport.Consume():
			switch msg := transMsg.(type) {
			case *cluster.Event:
				n.handleEvent(msg)
			}
		case <-ticker.C:
			n.handleTick()
		case <-n.ctx.Done():
			return
		}
	}
}
func (n *Node) handleTick() {
	outgoingEvents := n.clusterAlgo.OnTick(n.State)
	for _, event := range outgoingEvents {
		if err := n.Transport.Send(&transport.Message{
			From:    n.ID,
			To:      event.To,
			Payload: event,
		}); err != nil {
			fmt.Println(err)
		}
	}
}
func (n *Node) handleEvent(event *cluster.Event) {
	outgoingEvents := n.clusterAlgo.OnEvent(n.State, event)
	for _, event := range outgoingEvents {
		if err := n.Transport.Send(&transport.Message{
			From:    n.ID,
			To:      event.To,
			Payload: event,
		}); err != nil {
			fmt.Println(err)
		}
	}
}
func GetEuclidianDistance(pos1, pos2 Position) float64 {
	dx := math.Pow(pos1.X-pos2.X, 2)
	dy := math.Pow(pos1.Y-pos2.Y, 2)
	dz := math.Pow(pos1.Z-pos2.Z, 2)
	return math.Sqrt(dx + dy + dz)
}
