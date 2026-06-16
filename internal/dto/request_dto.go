package dto

type CreateRequestDto struct {
    UserID    string `json:"user_id" binding:"omitempty,uuid"`
    QueryText string `json:"query_text" binding:"required"`
}

type SearchRequestsDto struct {
    Q    string `form:"q"`
    Size int    `form:"size"`
}
