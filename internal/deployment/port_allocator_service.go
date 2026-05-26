package deployment

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type portAllocatorService struct {
	l *sync.Mutex
	cache map[int]bool
}

type PortAllocatorService interface {
	GetFreePort() (int, error)
	Clear()
}

func NewPortAllocatorService() PortAllocatorService {
	var l sync.Mutex
	cache := make(map[int]bool)
	return &portAllocatorService{&l, cache}
}

func (s *portAllocatorService) GetFreePort() (int, error) {
	s.l.Lock()
	defer s.l.Unlock()
	
	port := 1024

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	
	for ; (err == nil || s.cache[port]) && port <= 49151; {
		if conn != nil {
			conn.Close()
		}
		port++
		conn, err = net.Dial("tcp", fmt.Sprintf(":%d", port))		
	}

	if err != nil && strings.Contains(err.Error(), "connect: connection refused") {
		s.cache[port] = true
		return port, nil
	}

	return 0, err
}

func (s *portAllocatorService) Clear() {
	s.cache = make(map[int]bool)
}