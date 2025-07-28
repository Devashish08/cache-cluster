package cluster

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type Hash func(data []byte) uint32

type Ring struct {
	mu sync.RWMutex

	nodes      map[string]struct{}
	hashes     []uint32
	hash       Hash
	nodeHashes map[uint32]string

	virtualNodes int
}

func NewRing(virtualNodes int) *Ring {
	if virtualNodes <= 0 {
		virtualNodes = 100
	}
	return &Ring{
		nodes:        make(map[string]struct{}),
		hashes:       []uint32{},
		hash:         crc32.ChecksumIEEE,
		nodeHashes:   make(map[uint32]string),
		virtualNodes: virtualNodes,
	}
}

func (r *Ring) AddNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.nodes[node]; ok {
		return
	}

	r.nodes[node] = struct{}{}
	for i := 0; i < r.virtualNodes; i++ {
		virtualNodeKey := node + "#" + strconv.Itoa(i)
		hash := r.hash([]byte(virtualNodeKey))

		r.hashes = append(r.hashes, hash)
		r.nodeHashes[hash] = node
	}

	sort.Slice(r.hashes, func(i, j int) bool {
		return r.hashes[i] < r.hashes[j]
	})
}

func (r *Ring) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.hashes) == 0 {
		return ""
	}

	keyhash := r.hash([]byte(key))

	idx := sort.Search(len(r.hashes), func(i int) bool {
		return r.hashes[i] >= keyhash
	})

	if idx == len(r.hashes) {
		idx = 0
	}

	return r.nodeHashes[r.hashes[idx]]
}
