package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

func handleBadRequestResponse(c *gin.Context, message string, err error) {
	badRequestError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusBadRequest, badRequestError)
}

func handleNotFoundResponse(c *gin.Context, message string, err error) {
	notFoundError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusNotFound, notFoundError)
}

func handleUnauthorizedResponse(c *gin.Context, message string, err error) {
	unauthorizedError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusForbidden, unauthorizedError)
}

func handleInternalServerResponse(c *gin.Context, message string, err error) {
	internalServerError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusInternalServerError, internalServerError)
}
