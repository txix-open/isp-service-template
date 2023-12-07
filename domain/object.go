package domain

const (
	ErrCodeObjectNotFound = 800
)

type Object struct {
	Name string `validate:"required,max=32"`
}
