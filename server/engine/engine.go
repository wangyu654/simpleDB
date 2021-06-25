package engine

type Engine interface {
	Find(uint64) (string, error)
	Insert(uint64, string) error
	Update(uint64, string) error
	Delete(uint64) error
}
 