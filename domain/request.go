package domain

type ByIdRequest struct {
	Id int `validate:"required"`
}
