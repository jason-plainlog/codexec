package main

import (
	"codexec/internal/isolate"
	"codexec/internal/submission"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	isolate.Init()

	e := echo.New()
	e.Use(middleware.Gzip())

	e.POST("/submission", submission.Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
