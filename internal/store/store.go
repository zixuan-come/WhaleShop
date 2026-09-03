package store

import (
	"sync"
	"time"

	"github.com/zixuan-come/whaleshop/internal/model"
)

// Store 内存多租户存储:每个 project_id 一个独立订单空间和自增 id。
type Store struct {
	mu     sync.RWMutex
	tables map[int]map[int]*model.Order
	nextID map[int]int
}

func New() *Store {
	s := &Store{
		tables: make(map[int]map[int]*model.Order),
		nextID: make(map[int]int),
	}
	// 默认项目(pid=1)塞点种子,方便 demo
	s.tables[1] = map[int]*model.Order{}
	s.nextID[1] = 1
	seeds := []model.Order{
		{Item: "iPhone 15", Quantity: 1, Price: 6999.00, Status: "paid"},
		{Item: "AirPods Pro", Quantity: 2, Price: 1899.00, Status: "shipped"},
		{Item: "MacBook Air", Quantity: 1, Price: 8999.00, Status: "pending"},
	}
	for i := range seeds {
		id := s.nextID[1]
		seeds[i].ID = id
		seeds[i].CreatedAt = time.Now()
		s.tables[1][id] = &seeds[i]
		s.nextID[1]++
	}
	return s
}

func (s *Store) List(pid int) []*model.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Order, 0, len(s.tables[pid]))
	for _, o := range s.tables[pid] {
		out = append(out, o)
	}
	return out
}

func (s *Store) Get(pid, id int) (*model.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.tables[pid][id]
	return o, ok
}

func (s *Store) Create(pid int, o *model.Order) *model.Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tables[pid] == nil {
		s.tables[pid] = map[int]*model.Order{}
		s.nextID[pid] = 1
	}
	id := s.nextID[pid]
	o.ID = id
	o.CreatedAt = time.Now()
	if o.Status == "" {
		o.Status = "pending"
	}
	s.tables[pid][id] = o
	s.nextID[pid]++
	return o
}

func (s *Store) Update(pid, id int, patch *model.Order) (*model.Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.tables[pid][id]
	if !ok {
		return nil, false
	}
	patch.ID = id
	patch.CreatedAt = existing.CreatedAt
	s.tables[pid][id] = patch
	return patch, true
}

func (s *Store) Delete(pid, id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tables[pid][id]; !ok {
		return false
	}
	delete(s.tables[pid], id)
	return true
}
