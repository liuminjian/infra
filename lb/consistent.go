package lb

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type uints []uint32

func (u uints) Len() int {
	return len(u)
}

func (u uints) Less(i, j int) bool {
	return u[i] < u[j]
}

func (u uints) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

type Consistent struct {
	circle      map[uint32]string
	virtualNode int
	sortHashes  uints
	mutex       sync.RWMutex
}

func (c *Consistent) Next(key string) *ServerInstance {
	name, err := c.Get(key)
	if err != nil {
		log.Error(err)
		return nil
	}
	return &ServerInstance{
		Address: name,
		Status:  Enable,
	}
}

func NewConsistent() *Consistent {
	return &Consistent{
		circle:      make(map[uint32]string),
		virtualNode: 20,
	}
}

func (c *Consistent) Add(element string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.add(element)
}

func (c *Consistent) add(element string) {
	for i := 0; i < c.virtualNode; i++ {
		c.circle[c.hashKey(c.generateKey(element, i))] = element
	}
	c.updateSortHashes()
}

func (c *Consistent) generateKey(element string, index int) string {
	return element + strconv.Itoa(index)
}

func (c *Consistent) hashKey(key string) uint32 {
	if len(key) < 64 {
		var barray [64]byte
		copy(barray[:], key)
		return crc32.ChecksumIEEE(barray[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) updateSortHashes() {
	hashes := c.sortHashes[:0]
	if cap(c.sortHashes)/(c.virtualNode*4) > len(c.circle) {
		hashes = nil
	}

	for k := range c.circle {
		hashes = append(hashes, k)
	}

	sort.Sort(hashes)
	c.sortHashes = hashes
}

func (c *Consistent) Remove(element string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.remove(element)
}

func (c *Consistent) remove(element string) {
	for i := 0; i < c.virtualNode; i++ {
		delete(c.circle, c.hashKey(c.generateKey(element, i)))
	}
	c.updateSortHashes()
}

func (c *Consistent) search(key uint32) int {
	f := func(x int) bool {
		return c.sortHashes[x] > key
	}

	i := sort.Search(len(c.sortHashes), f)
	if i >= len(c.sortHashes) {
		i = 0
	}
	return i
}

func (c *Consistent) Get(name string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if len(c.circle) == 0 {
		return "", errors.New("no hash data")
	}
	key := c.hashKey(name)
	i := c.search(key)
	return c.circle[c.sortHashes[i]], nil
}
