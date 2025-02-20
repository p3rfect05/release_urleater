package handlers

import (
	"html/template"
	"io"
	_ "urleater/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type ServerInterface interface {
	GetMainPage(echo.Context) error
	PostLogin(c echo.Context) error
	PostRegister(c echo.Context) error
	GetLogout(c echo.Context) error
	CreateShortLink(c echo.Context) error
	UpdateUserShortLinks(c echo.Context) error
	GetRegisterPage(c echo.Context) error
	GetLoginPage(c echo.Context) error
	GetUserShortLinks(c echo.Context) error
	GetCreateShortLink(c echo.Context) error
	GetShortLink(c echo.Context) error
	GetSubscriptions(c echo.Context) error
	GetSubscriptionsPage(c echo.Context) error
	GetUser(c echo.Context) error
	DeleteShortLink(c echo.Context) error
	GetUserShortLinksNumber(c echo.Context) error
	GetLinksPage(c echo.Context) error
	GetShortLinksMatchingPattern(c echo.Context) error
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// @title			URLEater Swagger API
// @version		1.0
// @description	Это описание API для работы с сайтом по сокращению ссылок
// @host			localhost:8080
// @BasePath		/
func GetRoutes(si ServerInterface) *echo.Echo {
	e := echo.New()

	e.Use(middleware.CORS())

	e.Use(middleware.Static("/static"))

	t := &Template{
		templates: template.Must(template.ParseGlob("./templates/*.html")),
	}

	e.Renderer = t

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/", si.GetMainPage)
	e.GET("/login", si.GetLoginPage)
	e.GET("/register", si.GetRegisterPage)
	e.POST("/login", si.PostLogin)
	e.POST("/register", si.PostRegister)
	e.GET("/logout", si.GetLogout)
	e.POST("/create_link", si.CreateShortLink)
	e.GET("/create_link", si.GetCreateShortLink)
	e.GET("/:short_link", si.GetShortLink)
	e.GET("/subscriptions", si.GetSubscriptionsPage)
	e.GET("/get_subscriptions", si.GetSubscriptions)
	e.GET("/user", si.GetUser)
	e.GET("/get_links", si.GetUserShortLinks)
	e.GET("/get_total_links_number", si.GetUserShortLinksNumber)
	e.GET("/my_links", si.GetLinksPage)
	e.GET("/short_links", si.GetShortLinksMatchingPattern)
	e.DELETE("/delete_link", si.DeleteShortLink)

	return e

}
