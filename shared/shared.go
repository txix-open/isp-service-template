package shared

type ObjectDomain struct {
	Id   int64
	Name string
}

type ObjectController interface {
	GetAll() ([]ObjectDomain, error)
}
