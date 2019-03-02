package mapper

import (
	"fmt"
	"strconv"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/martinxsliu/protoc-gen-graphql/descriptor"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Graph struct {
	g        *simple.DirectedGraph
	nodes    map[graph.Node]*descriptor.Message
	revNodes map[string]graph.Node
	sorted   []*descriptor.Message
}

func NewGraph(messages []*descriptor.Message) *Graph {
	g := simple.NewDirectedGraph()

	nodes := make(map[graph.Node]*descriptor.Message)
	revNodes := make(map[string]graph.Node)
	for _, message := range messages {
		node := g.NewNode()
		nodes[node] = message
		revNodes[message.FullName] = node
		g.AddNode(node)
	}

	for _, message := range messages {
		for _, field := range message.Proto.GetField() {
			if field.GetType() == pb.FieldDescriptorProto_TYPE_MESSAGE {
				// A message may have fields with types not included within the
				// messages slice. If this happens then create a proxy node to
				// represent this type.
				from, ok := revNodes[field.GetTypeName()]
				if !ok {
					from = g.NewNode()
				}
				to := revNodes[message.FullName]
				g.SetEdge(g.NewEdge(from, to))
			}
		}
	}

	return &Graph{
		g:        g,
		nodes:    nodes,
		revNodes: revNodes,
	}
}

// Sort returns topologically sorted messages such that earlier messages
// will not have any fields of a later message's type.
func (g *Graph) Sort() ([]*descriptor.Message, error) {
	if g.sorted != nil {
		return g.sorted, nil
	}

	sortedNodes, err := topo.Sort(g.g)
	if err != nil {
		if unorderable, ok := err.(topo.Unorderable); ok {
			var nameSets [][]string
			for _, set := range unorderable {
				var names []string
				for _, node := range set {
					message, ok := g.nodes[node]
					if ok {
						names = append(names, message.FullName)
					} else {
						names = append(names, strconv.Itoa(int(node.ID())))
					}
				}
				nameSets = append(nameSets, names)
			}
			return nil, fmt.Errorf("%s: %v", err, nameSets)
		}
		return nil, err
	}

	var sorted []*descriptor.Message
	for _, node := range sortedNodes {
		// Note, not all nodes will map back to messages. This is due to potential
		// proxy nodes created to represent types not included within the input.
		message, ok := g.nodes[node]
		if ok {
			sorted = append(sorted, message)
		}
	}

	g.sorted = sorted
	return sorted, nil
}

func (g *Graph) SortTo(toMessages []*descriptor.Message) ([]*descriptor.Message, error) {
	sorted, err := g.Sort()
	if err != nil {
		return nil, err
	}

	seen := make(map[*descriptor.Message]bool)
	stack := toMessages[:]
	for len(stack) > 0 {
		toMessage := stack[0]
		stack = stack[1:]

		if seen[toMessage] {
			continue
		}
		seen[toMessage] = true

		to, ok := g.revNodes[toMessage.FullName]
		if !ok {
			continue
		}

		nodes := g.g.To(to.ID())
		for {
			if !nodes.Next() {
				break
			}

			node := nodes.Node()
			if node == nil {
				break
			}

			// Note, not all nodes will map back to messages. This is due to potential
			// proxy nodes created to represent types not included within the input.
			message, ok := g.nodes[node]
			if ok {
				stack = append(stack, message)
			}
		}
	}

	var filtered []*descriptor.Message
	for _, message := range sorted {
		if seen[message] {
			filtered = append(filtered, message)
		}
	}

	return filtered, nil
}
