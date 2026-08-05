package deployment

type StubPortAllocatorService struct {
	Get_free_port_return int
	Get_free_port_error error
	Get_free_port_n_calls int
}

func (s *StubPortAllocatorService) Clear() {
	s.Get_free_port_return = 0
	s.Get_free_port_error = nil
	s.Get_free_port_n_calls = 0
}

func (s *StubPortAllocatorService) GetFreePort() (int, error) {
	s.Get_free_port_n_calls += 1
	return s.Get_free_port_return, s.Get_free_port_error
}
