package cluster

type Phase int

const (
	Election Phase = iota
	ClusterHead
	ClusterMember
)

type LocalState struct {
	NodeId    int
	ClusterId int
	Term      int
	Pos       []float64

	// TODO: change this
	AvailableNodes []int
	Positions      [][]float64
	Phase
}

func NewState(id int, pos []float64) *LocalState {
	return &LocalState{
		NodeId:         id,
		Phase:          Election,
		Pos:            pos,
		AvailableNodes: make([]int, 0),
		Positions:      make([][]float64, 0),
	}
}

func (s *LocalState) AddPeer(id int, pos []float64) {
	s.AvailableNodes = append(s.AvailableNodes, id)
	s.Positions = append(s.Positions, pos)
}
func (s *LocalState) Update(newPos []float64) {
	s.Pos = newPos
	s.AvailableNodes = make([]int, 0)
	s.Positions = make([][]float64, 0)
}
