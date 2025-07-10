package dtousergo

type SearchUsersRequest struct {
	Query    string `json:"query" binding:"required,min=2"`
	Role     string `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}
