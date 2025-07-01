package dto

// ErrorResponse represents a standard error response
// @Description Standard error response format
// @Description Contains error details and status code
// @Description Example: {"error": "user not found", "code": 404}
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
	Code  int    `json:"code" example:"400"`
}

// PaginatedResponse represents a paginated response
// @Description Standard paginated response format
// @Description Contains pagination metadata and the data items
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalItems int64       `json:"total_items" example:"100"`
	Page       int         `json:"page" example:"1"`
	PageSize   int         `json:"page_size" example:"20"`
	TotalPages int         `json:"total_pages" example:"5"`
}
