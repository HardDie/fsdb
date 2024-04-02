package fsentry

import (
	"github.com/HardDie/fsentry/dto"
)

func (s *Service) CreateEntry(name string, data interface{}, path ...string) (*dto.Entry, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.entry.Create(s.buildPath(path...), name, data)
}
func (s *Service) GetEntry(name string, path ...string) (*dto.Entry, error) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.entry.Get(s.buildPath(path...), name)
}
func (s *Service) MoveEntry(oldName, newName string, path ...string) (*dto.Entry, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.entry.Move(s.buildPath(path...), oldName, newName)
}
func (s *Service) UpdateEntry(name string, data interface{}, path ...string) (*dto.Entry, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.entry.Update(s.buildPath(path...), name, data)
}
func (s *Service) RemoveEntry(name string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.entry.Remove(s.buildPath(path...), name)
}
func (s *Service) DuplicateEntry(srcName, dstName string, path ...string) (*dto.Entry, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.entry.Duplicate(s.buildPath(path...), srcName, dstName)
}
