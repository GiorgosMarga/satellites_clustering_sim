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
	ChId      int
	Term      int
	Pos       []float64

	// TODO: change this
	AvailableNodes []int
	Positions      [][]float64
	ClusterMembers []int
	Phase
}

func NewState(id int, pos []float64) *LocalState {
	return &LocalState{
		NodeId:         id,
		Phase:          Election,
		Pos:            pos,
		AvailableNodes: make([]int, 0),
		Positions:      make([][]float64, 0),
		ClusterMembers: make([]int, 0),
	}
}

func (s *LocalState) AddPeer(id int, pos []float64) {
	s.AvailableNodes = append(s.AvailableNodes, id)
	s.Positions = append(s.Positions, pos)
}
func (s *LocalState) Reset() {
	s.Phase = Election
	s.AvailableNodes = make([]int, 0)
	s.Positions = make([][]float64, 0)
	s.ClusterMembers = make([]int, 0)
}
func (s *LocalState) Update(newPos []float64) {
	s.Pos = newPos
}
