package history

type Manager interface {
	Add(name, namespace string)
	GetLastName(current string) *string
	GetLastNamespace(name, current string) *string

	DeleteByName(name string)
	DeleteAll()

	List() []*Record

	Save() error
}

type Record struct {
	Timestamp int64
	Name      string
	Namespace string
}
