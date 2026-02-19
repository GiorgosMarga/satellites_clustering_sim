package cluster

type Event struct {
	From    int
	To      int
	Payload any
}

// JoinEvent is a type of event that is sent between 2 plane clusterheads to merge their plane clusters.
type JoinEvent struct {
	SatId int
	ClusterId int
}
// LeaveEvent is a type of event that is sent between 2 plane clusterheads to unmerge their plane clusters.
type LeaveEvent struct {}

// ClusterMemeberEvent is a type of event that is sent from the members to the clusterhead of the same cluster.
type ClusterMemberEvent struct {}

// ClusterHeadEvent is a type of event that is sent from the clusterhead to the clustermembers of the same cluster.
type ClusterHeadEvent struct {
	ClusterId int
	ClusterheadId int
}
