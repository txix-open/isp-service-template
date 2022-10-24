package domain

type ByIdRequest struct {
	Id int `valid:"required"`
}
