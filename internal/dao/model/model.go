package model

type Response[T any] struct {
	Data   *T             `json:"data,omitempty"`
	Paging *Paging        `json:"paging,omitempty"`
	Error  *ErrorResponse `json:"error,omitempty"`
}

type Paging struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPage  int64 `json:"total_page"`
	TotalCount int64 `json:"total_count"`
}

type ErrorResponse struct {
	RequestID string              `json:"request_id"`
	Errors    map[string][]string `json:"errors"`
}

func NewSuccessResponse[T any](data *T, paging *Paging) *Response[T] {
	return &Response[T]{
		Data:   data,
		Paging: paging,
	}
}

func NewErrorResponse[T any](err *ErrorResponse) *Response[T] {
	return &Response[T]{
		Error: err,
	}
}
