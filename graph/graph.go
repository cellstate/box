package graph

//keys refer to a single node
type Key []byte

//nodes in the graph can link to other nodes
//to form a DAG (directed) acyclic graph, nodes
//can have arbritary metadata to allow optmizations
type Node interface {
	Link(Node) error //link from this node to another node
	Key() (Key, error)
	Metadata() map[string]string
	Links() ([]Key, error)
	Data() ([]byte, error)
}

//a graph that can be compared and serialized
//into a linear list of nodes for storage into
//any key/value store
type Graph interface {
	Index() error                    //(re)index all nodes in the graph
	Compare(b Graph) ([]Node, error) //return nodes from graph A that are missing in B
	List() ([]Node, error)           //return all nodes of the graph in serial fashion
	Put(Node) (Key, error)
	Get(Key) (Node, error)
}

//@todo implement a "lazy" graph that only loads part of the tree into memory
//when it requires comparison?
