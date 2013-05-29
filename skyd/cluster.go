package skyd

import (
	"errors"
	"sort"
	"sync"
)

//------------------------------------------------------------------------------
//
// Globals
//
//------------------------------------------------------------------------------

var NodeGroupRequiredError = errors.New("Node group required")
var NodeGroupNotFoundError = errors.New("Node group not found")
var DuplicateNodeError = errors.New("Duplicate node already exists")

//------------------------------------------------------------------------------
//
// Typedefs
//
//------------------------------------------------------------------------------

// The cluster manages the topology of servers for a distributed Sky database.
// Clusters are made up of cluster groups which are sets of servers that
// manage a subset of the total dataset.
type Cluster struct {
	groups []*NodeGroup
	mutex  sync.Mutex
}

//------------------------------------------------------------------------------
//
// Constructor
//
//------------------------------------------------------------------------------

// Creates a new cluster.
func NewCluster() *Cluster {
	return &Cluster{
		groups: []*NodeGroup{},
	}
}

//------------------------------------------------------------------------------
//
// Methods
//
//------------------------------------------------------------------------------

//--------------------------------------
// Node Groups
//--------------------------------------

// Finds a group in the cluster by id.
func (c *Cluster) GetNodeGroup(id string) *NodeGroup {
	c.mutex.Lock()
	c.mutex.Unlock()
	return c.getNodeGroup(id)
}

func (c *Cluster) getNodeGroup(id string) *NodeGroup {
	for _, group := range c.groups {
		if group.id == id {
			return group
		}
	}
	return nil
}

// Adds a group to the cluster.
func (c *Cluster) AddNodeGroup(group *NodeGroup) {
	c.mutex.Lock()
	c.mutex.Unlock()
	c.groups = append(c.groups, group)
	sort.Sort(NodeGroups(c.groups))
}

// Removes a group from the cluster.
func (c *Cluster) RemoveNodeGroup(group *NodeGroup) error {
	c.mutex.Lock()
	c.mutex.Unlock()

	if group == nil {
		return NodeGroupRequiredError
	}
	for index, g := range c.groups {
		if g == group {
			c.groups = append(c.groups[:index], c.groups[index+1:]...)
			return nil
		}
	}

	return NodeGroupNotFoundError
}

//--------------------------------------
// Nodes
//--------------------------------------

// Retrieves a node and its group from the cluster by id.
func (c *Cluster) GetNode(id string) (*Node, *NodeGroup) {
	c.mutex.Lock()
	c.mutex.Unlock()
	return c.getNode(id)
}

func (c *Cluster) getNode(id string) (*Node, *NodeGroup) {
	for _, group := range c.groups {
		if node := group.getNode(id); node != nil {
			return node, group
		}
	}
	return nil, nil
}

// Adds a node to an existing group in the cluster.
func (c *Cluster) AddNode(node *Node, group *NodeGroup) error {
	c.mutex.Lock()
	c.mutex.Unlock()

	// Validate node.
	if node == nil {
		return NodeRequiredError
	}

	// Check if the node id exists in the cluster already.
	if n, _ := c.getNode(node.id); n != nil {
		return DuplicateNodeError
	}

	// Find the group.
	if group == nil {
		return NodeGroupRequiredError
	}
	if group = c.getNodeGroup(group.id); group == nil {
		return NodeGroupNotFoundError
	}

	return group.addNode(node)
}

// Removes a node from a group in the cluster.
func (c *Cluster) RemoveNode(node *Node) error {
	c.mutex.Lock()
	c.mutex.Unlock()

	if node == nil {
		return NodeRequiredError
	}

	var group *NodeGroup
	if node, group = c.getNode(node.id); node == nil {
		return NodeNotFoundError
	}

	return group.removeNode(node)
}

//--------------------------------------
// Serialization
//--------------------------------------

// Converts the cluster topology to an object that can be easily serialized
// to JSON outside the cluster lock.
func (c *Cluster) serialize() map[string]interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Serialize groups.
	groups := []interface{}{}
	for _, group := range c.groups {
		groups = append(groups, group.Serialize())
	}

	return map[string]interface{}{"groups": groups}
}
