package main

import (
	"fmt"
	"net/http"

	"github.com/akhileshh/state-server/state"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var corsConfig = middleware.CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPut,
		http.MethodPatch,
		http.MethodPost,
		http.MethodDelete,
	},
	AllowHeaders: []string{echo.HeaderAuthorization},
	ExposeHeaders: []string{
		echo.HeaderContentType,
		echo.HeaderContentLength,
		echo.HeaderContentEncoding,
		echo.HeaderAccept,
	},
}

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(corsConfig))

	e.GET(state.JSONStateEP, func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET(fmt.Sprintf("%v/:id", state.JSONStateEP), state.GetJSON)
	e.POST(state.JSONStatePostEP, state.SaveJSON)

	e.Logger.Fatal(e.Start(":8001"))
}
