package response

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

type Success struct {
	Result interface{} `json:"result"`
}

type Error struct {
	Message string `json:"error"`
}

// JSON writes any JSON response with a given status code
func JSON(c *ginext.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// OK sends a 200 OK response
func OK(c *ginext.Context, result interface{}) {
	JSON(c, http.StatusOK, Success{Result: result})
}

// Created sends a 201 Created response
func Created(c *ginext.Context, result interface{}) {
	JSON(c, http.StatusCreated, Success{Result: result})
}

// Fail sends an error response with a given status code
func Fail(c *ginext.Context, status int, err error) {
	JSON(c, status, Error{Message: err.Error()})
}
