package models

// StandardResponse represents a standard API response format
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	RequestID string      `json:"request_id"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      *ErrorInfo  `json:"error,omitempty"`
	RequestID  string      `json:"request_id"`
	Pagination *Pagination `json:"pagination"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPages int  `json:"total_pages"`
}

// OrderWithProduct combines order and product information
type OrderWithProduct struct {
	Order   *Order   `json:"order"`
	Product *Product `json:"product,omitempty"`
}

// CreateSuccessResponse creates a standardized success response
func CreateSuccessResponse(data interface{}, requestID string) StandardResponse {
	return StandardResponse{
		Success:   true,
		Data:      data,
		RequestID: requestID,
	}
}

// CreateErrorResponse creates a standardized error response
func CreateErrorResponse(message, code, requestID string) StandardResponse {
	return StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Message: message,
			Code:    code,
		},
		RequestID: requestID,
	}
}

// CreatePaginatedResponse creates a standardized paginated response
func CreatePaginatedResponse(data interface{}, pagination *Pagination, requestID string) PaginatedResponse {
	return PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		RequestID:  requestID,
	}
}