package fsentry

import (
	"github.com/HardDie/fsentry/dto"
)

func (s *Service) CreateFolder(name string, data interface{}, path ...string) (*dto.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.folder.Create(s.buildPath(path...), name, data)
}
func (s *Service) GetFolder(name string, path ...string) (*dto.FolderInfo, error) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.folder.Get(s.buildPath(path...), name)
}
func (s *Service) MoveFolder(oldName, newName string, path ...string) (*dto.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.folder.Move(s.buildPath(path...), oldName, newName)
}
func (s *Service) UpdateFolder(name string, data interface{}, path ...string) (*dto.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.folder.Update(s.buildPath(path...), name, data)
}
func (s *Service) RemoveFolder(name string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.folder.Remove(s.buildPath(path...), name)
}
func (s *Service) DuplicateFolder(srcName, dstName string, path ...string) (*dto.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.folder.Duplicate(s.buildPath(path...), srcName, dstName)
}
func (s *Service) UpdateFolderNameWithoutTimestamp(oldName, newName string, path ...string) (*dto.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.folder.MoveWithoutTimestamp(s.buildPath(path...), oldName, newName)
}
