package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

func HandleNotFoundRequestResponse(c *gin.Context, message string, err error) {
	notFoundRequestError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusNotFound, notFoundRequestError)
}

func HandleBadRequestResponse(c *gin.Context, message string, err error) {
	badRequestError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusBadRequest, badRequestError)
}

func HandleNotFoundResponse(c *gin.Context, message string, err error) {
	notFoundError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusNotFound, notFoundError)
}

func HandleInternalServerResponse(c *gin.Context, message string, err error) {
	internalServerError := ErrorResponse{
		Message: message,
		Err:     err.Error(),
	}
	c.JSON(http.StatusInternalServerError, internalServerError)
}
