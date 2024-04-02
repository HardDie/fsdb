package fsentry

func (s *Service) CreateBinary(name string, data []byte, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.binary.Create(s.buildPath(path...), name, data)
}
func (s *Service) GetBinary(name string, path ...string) ([]byte, error) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.binary.Get(s.buildPath(path...), name)
}
func (s *Service) MoveBinary(oldName, newName string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.binary.Move(s.buildPath(path...), oldName, newName)
}
func (s *Service) UpdateBinary(name string, data []byte, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.binary.Update(s.buildPath(path...), name, data)
}
func (s *Service) RemoveBinary(name string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.binary.Remove(s.buildPath(path...), name)
}
