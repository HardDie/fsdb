package binary

type Service interface {
	Create(path, name string, data []byte) error
	Get(path, name string) ([]byte, error)
	Move(path, oldName, newName string) error
	Update(path, name string, data []byte) error
	Remove(path, name string) error
	Duplicate(path, oldName, newName string) ([]byte, error)
}
