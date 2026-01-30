package api

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse is a standard success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// PaginatedResponse is a paginated response wrapper
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int64       `json:"total_count"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}
