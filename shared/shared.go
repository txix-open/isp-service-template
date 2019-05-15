package shared

type ObjectDomain struct {
	Id   int32
	Name string
}

type ObjectController interface {
	GetAll() ([]ObjectDomain, error)
}
