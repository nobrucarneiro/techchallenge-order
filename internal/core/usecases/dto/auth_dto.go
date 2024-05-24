package dto

type AuthorizerResponse struct {
	UserId       int    `json:"userId"`
	IsAuthorized bool   `json:"isAuthorized"`
	Message      string `json:"message"`
}