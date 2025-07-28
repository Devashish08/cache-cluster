// internal/cluster/consistent_hash_test.go
package cluster

import (
	"strconv"
	"testing"
)

func TestRing_GetNode(t *testing.T) {
	ring := NewRing(100)

	nodes := []string{"node1:8080", "node2:8080", "node3:8080"}
	for _, node := range nodes {
		ring.AddNode(node)
	}

	testCases := map[string]string{
		"my-key-1":         "node2:8080",
		"another-key":      "node3:8080",
		"user:123:profile": "node3:8080",
		"a-very-long-key-that-hashes-differently": "node2:8080",
		"wrap-around-key":                         "node2:8080",
	}

	for key, expectedNode := range testCases {
		t.Run(key, func(t *testing.T) {
			node := ring.GetNode(key)
			if node != expectedNode {
				t.Errorf("for key '%s', got node '%s', expected '%s'", key, node, expectedNode)
			}
		})
	}
}

func TestRing_AddNodeIdempotency(t *testing.T) {
	ring := NewRing(10)
	ring.AddNode("node1:8080")
	ring.AddNode("node1:8080") // Add the same node again to test idempotency

	if len(ring.hashes) != 10 {
		t.Errorf("expected 10 hashes in the ring, got %d", len(ring.hashes))
	}
	if len(ring.nodes) != 1 {
		t.Errorf("expected 1 node in the node set, got %d", len(ring.nodes))
	}
}

func TestRing_Distribution(t *testing.T) {
	ring := NewRing(100)
	nodes := []string{"node1", "node2", "node3"}
	for _, node := range nodes {
		ring.AddNode(node)
	}

	distribution := make(map[string]int)
	for i := 0; i < 10000; i++ {
		key := "some-key-" + strconv.Itoa(i)
		node := ring.GetNode(key)
		distribution[node]++
	}

	ideal := 10000 / 3.0
	for node, count := range distribution {
		// Check if the distribution is within a reasonable range (e.g., 50%) of the ideal.
		if float64(count) < ideal*0.5 || float64(count) > ideal*1.5 {
			t.Errorf("Node %s has poor distribution: got %d keys, ideal is ~%.0f", node, count, ideal)
		}
	}
	t.Logf("Distribution results: %v", distribution)
}
