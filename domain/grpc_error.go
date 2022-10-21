package domain

type GrpcError struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    string `json:"errorCode"`
	Details      []any  `json:"details"`
}
