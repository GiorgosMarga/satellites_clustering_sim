package cluster

import (
	"fmt"
	"math"
	"slices"
	"sync"
)

const (
	a              float64 = 0.5
	b              float64 = 0.8
	PlaneStart             = 1
	PlaneEnd               = 71
	MiddlePlane            = (PlaneEnd + PlaneStart) / 2
	MaxPlaneOffset         = (PlaneEnd - PlaneStart) / 2
)

type ClusterAlgo interface {
	OnEvent(*LocalState, *Event) []*Event
	OnTick(*LocalState) []*Event
}

type LayeredClustering struct {
	mtx *sync.Mutex
}

func NewLayeredClustering() *LayeredClustering {
	return &LayeredClustering{
		mtx: &sync.Mutex{},
	}
}
func (lc *LayeredClustering) OnTick(state *LocalState) []*Event {
	lc.mtx.Lock()
	defer lc.mtx.Unlock()
	switch state.Phase {
	case ClusterHead:
		return handleCHPhase(state)
	case ClusterMember:
		return handleCMPhase(state)
	case Election:
		return handleElectionPhase(state)
	default:
		return nil
	}
}

func handleCHPhase(state *LocalState) []*Event {
	bestScore := math.MaxFloat32
	bestPeer := 0
	if state.ClusterId != state.NodeId {
		return nil
	}
	for idx, peer := range state.AvailableNodes {
		if !isPlaneCH(peer) || isSamePlane(state.NodeId, peer) {
			continue
		}
		peerPos := state.Positions[idx]

		score := calculateScore(state, peer, peerPos)
		if score < bestScore {
			bestPeer = peer
			bestScore = score
		}
	}

	peerPlaneCentrality := calculatePlaneCentrality(bestPeer / 21)
	myPlaneCentrality := calculatePlaneCentrality(state.NodeId / 21)

	if myPlaneCentrality < peerPlaneCentrality || (myPlaneCentrality == peerPlaneCentrality && state.NodeId < bestPeer) {
		// join best peer
		events := []*Event{{From: state.NodeId, To: bestPeer, Payload: &JoinEvent{SatId: state.NodeId, ClusterId: bestPeer}}, {From: state.NodeId, To: state.NodeId - 1, Payload: &ClusterHeadEvent{ClusterId: bestPeer, ClusterheadId: state.NodeId}}, {From: state.NodeId, To: state.NodeId + 1, Payload: &ClusterHeadEvent{ClusterId: bestPeer, ClusterheadId: state.NodeId}}}

		for _, member := range state.ClusterMembers {
			events = append(events, &Event{
				From:    state.NodeId,
				To:      member,
				Payload: &ClusterHeadEvent{ClusterId: bestPeer, ClusterheadId: state.NodeId},
			})
		}
		state.ClusterId = bestPeer
		state.ChId = bestPeer
		state.Phase = ClusterMember
		return events
	}
	return nil
}
func handleCMPhase(state *LocalState) []*Event {
	// if the cluster id is not in the available nodes, return to plane cluster. Current node is always
	// a plane cluster head. Plane cluster members are always in contact with plane cluster head.
	if !slices.Contains(state.AvailableNodes, state.ChId) {
		state.Phase = Election
		state.ClusterId = state.NodeId
		state.ChId = state.NodeId
		fmt.Printf("[%d] return to election state\n", state.NodeId)
	}
	return nil
}

func handleElectionPhase(state *LocalState) []*Event {
	// * 2 * * 5 * * 8 * * 11 * * 14 * * 17 * * 20 *
	state.Term++
	if isPlaneCH(state.NodeId) {
		// middle node in every 3-pairs of nodes
		state.Phase = ClusterHead
		state.ClusterId = state.NodeId
		payload := &ClusterHeadEvent{
			ClusterId:     state.NodeId,
			ClusterheadId: state.NodeId,
		}
		prevNodeMsg := &Event{From: state.NodeId, To: state.NodeId - 1, Payload: payload}
		nextNodeMsg := &Event{From: state.NodeId, To: state.NodeId + 1, Payload: payload}
		return []*Event{prevNodeMsg, nextNodeMsg}
	}

	state.Phase = ClusterMember

	// cluster member on same plane
	return nil
}

func (lc *LayeredClustering) OnEvent(state *LocalState, event *Event) []*Event {
	lc.mtx.Lock()
	defer lc.mtx.Unlock()
	switch msg := event.Payload.(type) {
	case *JoinEvent:
		return handleJoinEvent(state, msg)
	case *ClusterHeadEvent:
		return handleClusterHeadEvent(state, msg)
	case *ClusterMemberEvent:
		return handleClusterMemberEvent(state, msg)
	case *LeaveEvent:
		return handleLeaveEvent(state, msg)
	default:
		fmt.Println("Invalid Message")
		return nil
	}
}

func handleJoinEvent(state *LocalState, payload *JoinEvent) []*Event {
	if !slices.Contains(state.ClusterMembers, payload.SatId) {
		state.ClusterMembers = append(state.ClusterMembers, payload.SatId)
	}
	switch state.Phase {
	case ClusterHead:
	case ClusterMember:
		return []*Event{{From: state.NodeId, To: payload.SatId, Payload: &ClusterHeadEvent{
			ClusterId:     state.ClusterId,
			ClusterheadId: state.NodeId,
		}}}
	}
	return nil
}
func handleLeaveEvent(state *LocalState, payload *LeaveEvent) []*Event {
	return nil

}
func handleClusterHeadEvent(state *LocalState, payload *ClusterHeadEvent) []*Event {
	state.ClusterId = payload.ClusterId
	state.ChId = payload.ClusterheadId
	state.Term++
	state.Phase = ClusterMember
	if isPlaneCH(state.NodeId) {
		// propagate message to members
		events := make([]*Event, 0, len(state.ClusterMembers)+2)
		events = append(events, &Event{
			From: state.NodeId,
			To:   state.NodeId - 1,
			Payload: &ClusterHeadEvent{
				ClusterId:     payload.ClusterId,
				ClusterheadId: state.NodeId,
			},
		}, &Event{
			From: state.NodeId,
			To:   state.NodeId + 1,
			Payload: &ClusterHeadEvent{
				ClusterId:     payload.ClusterId,
				ClusterheadId: state.NodeId,
			},
		})
		for _, clusterMember := range state.ClusterMembers {
			if clusterMember == payload.ClusterheadId {
				continue
			}
			events = append(events, &Event{
				From: state.NodeId,
				To:   clusterMember,
				Payload: &ClusterHeadEvent{
					ClusterId:     payload.ClusterId,
					ClusterheadId: state.NodeId,
				},
			})
		}
		return events
	}
	return nil

}
func handleClusterMemberEvent(state *LocalState, payload *ClusterMemberEvent) []*Event {
	return nil

}

func calculateScore(state *LocalState, peerId int, peerPos []float64) float64 {
	dx := math.Abs(state.Pos[0] - peerPos[0])
	dy := math.Abs(state.Pos[1] - peerPos[1])
	dz := math.Abs(state.Pos[2] - peerPos[2])
	dist := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2) + math.Pow(dz, 2))

	myPlaneCentrality := calculatePlaneCentrality(state.NodeId / 21)
	otherPlaneCentrality := calculatePlaneCentrality(peerId / 21)
	dPlane := math.Abs(float64(myPlaneCentrality) - float64(otherPlaneCentrality))

	return a*dist + b*dPlane
}
func calculatePlaneCentrality(planeId int) float64 {
	return 1.0 - (float64(planeId-MiddlePlane) / MaxPlaneOffset)
}
func isPlaneCH(peerId int) bool {
	return peerId%3 == 2
}
func isSamePlane(id, peerId int) bool {
	return peerId/21 == id/21
}
