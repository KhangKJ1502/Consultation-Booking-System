package dtousergo

type UserListResponse struct {
	Users      []UserProfileResponse `json:"users"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int64                 `json:"total_pages"`
}
