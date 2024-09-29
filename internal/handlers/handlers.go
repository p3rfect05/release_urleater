package handlers

import (
	"context"
	"fmt"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
	_ "urleater/docs"
	"urleater/internal/repository/postgresDB"

	"github.com/antonlindstrom/pgstore"
	"github.com/labstack/echo/v4"
)

type Service interface {
	LoginUser(ctx context.Context, email string, password string) error
	RegisterUser(ctx context.Context, email string, password string) error
	CreateShortLink(ctx context.Context, shortLink string, longLink string, userEmail string) (*postgresDB.Link, error)
	UpdateUserShortLinks(ctx context.Context, email string, deltaLinks int) (*postgresDB.User, error)
	GetUserShortLinksWithOffsetAndLimit(ctx context.Context, email string, offset int, limit int) ([]postgresDB.Link, *postgresDB.User, error)
	GetSubscriptions(ctx context.Context) ([]postgresDB.Subscription, error)
	GetShortLink(ctx context.Context, shortLink string) (*postgresDB.Link, error)
	GetUser(ctx context.Context, email string) (*postgresDB.User, error)
	DeleteShortLink(ctx context.Context, shortLink string, email string) error
}

type SessionStore interface {
	RetrieveEmailFromSession(c echo.Context) (string, error)
	Get(r *http.Request, key string) (*sessions.Session, error)
	Save(c echo.Context, email string, session *sessions.Session) error
}

// TODO populate
var domain string = "http://localhost:8080"

type Handlers struct {
	Service Service
	Store   SessionStore
}

type PostgresSessionStore struct {
	store *pgstore.PGStore
}

func NewPostgresSessionStore(store *pgstore.PGStore) SessionStore {
	return &PostgresSessionStore{store}
}

func (pg *PostgresSessionStore) RetrieveEmailFromSession(c echo.Context) (string, error) {
	session, err := pg.store.Get(c.Request(), "session_key")

	if err != nil {
		return "", fmt.Errorf("error getting session: %w", err)
	}

	if _, ok := session.Values["email"]; !ok {
		return "", nil
	}
	res := session.Values["email"].(string)
	return res, nil
}

func (pg *PostgresSessionStore) Get(r *http.Request, key string) (*sessions.Session, error) {
	session, err := pg.store.Get(r, key)

	return session, err
}

func (db *PostgresSessionStore) Save(c echo.Context, email string, session *sessions.Session) error {
	session.Values["email"] = email

	err := session.Save(c.Request(), c.Response())

	if err != nil {
		return err
	}

	return nil
}

type redirectResponse struct {
	redirectTo string
}

// GetMainPage godoc
//
// @Summary Gets main page HTML
// @Produce	html
// @Success 200
// @Failure 500 {} nil
// @Failure 307 {} nil
// @Router /	[get]
func (h *Handlers) GetMainPage(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	return c.Render(http.StatusOK, "main_page.html", nil)
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// PostLogin godoc
//
//	@Summary		Logins a user
//	@Accept			json
//	@Param			username	body		string	true	"Username"
//	@Param			password	body		string	true	"Password"
//	@Success		200			{object}	redirectResponse
//	@Failure		400			{} nil
//	@Failure		500			{} nil
//	@Router			/login      [post]
func (h *Handlers) PostLogin(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email != "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"redirectTo": "/",
		})
	}

	ctx := c.Request().Context()

	requestData := new(LoginRequest)

	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Echo().Validator != nil {
		if err := c.Validate(requestData); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}

	err = h.Service.LoginUser(ctx, requestData.Email, requestData.Password)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	session, err := h.Store.Get(c.Request(), "session_key")

	if err != nil {
		log.Printf("Error getting session: %v\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if err = h.Store.Save(c, requestData.Email, session); err != nil {
		log.Printf("Error saving session: %v\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())

	}

	return c.JSON(http.StatusOK, echo.Map{
		"redirectTo": "/",
	})
}

// GetLogout godoc
//
//	@Summary		Logs out a user
//	@Success		307			{object}	redirectResponse
//	@Failure		500			{} nil
//	@Router			/logout      [get]
func (h *Handlers) GetLogout(c echo.Context) error {
	session, err := h.Store.Get(c.Request(), "session_key")

	if err != nil {
		log.Printf("Error getting session: %v\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	session.Options.MaxAge = -1
	if err = session.Save(c.Request(), c.Response()); err != nil {
		log.Printf("Error saving session: %v\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// PostRegister godoc
//
//	@Summary		Registers a user
//	@Accept			json
//	@Param			username	body		string	true	"Username"
//	@Param			password	body		string	true	"Password"
//	@Success		200			{object}	redirectResponse
//	@Failure		400			{} nil
//	@Failure		500			{} nil
//	@Router			/register      [post]
func (h *Handlers) PostRegister(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email != "" {
		return c.JSON(http.StatusOK, echo.Map{
			"redirectTo": "/",
		})
	}

	ctx := c.Request().Context()

	requestData := new(RegisterRequest)

	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Echo().Validator != nil {
		if err := c.Validate(requestData); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}

	err = h.Service.RegisterUser(ctx, requestData.Email, requestData.Password)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	session, err := h.Store.Get(c.Request(), "session_key")

	if err != nil {
		log.Printf("Error getting session: %v\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if err = h.Store.Save(c, requestData.Email, session); err != nil {
		log.Printf("Error saving session: %v\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())

	}

	return c.JSON(http.StatusOK, echo.Map{
		"redirectTo": "/",
	})
}

type CreateShortLinkRequest struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url" validate:"required"`
}

type CreateShortLinkResponse struct {
	Link postgresDB.Link `json:"link"`
}

// CreateShortLink godoc
//
//	@Summary		Creates a link
//	@Accept			json
//	@Param			short_url	body		string	true	"Short URL"
//	@Param			long_url	body		string	true	"Long URL"
//	@Success		200			{object}	CreateShortLinkResponse
//	@Failure		400			{} nil
//	@Failure		500			{} nil
//	@Router			/create_link      [post]
func (h *Handlers) CreateShortLink(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"redirectTo": "/login",
		})
	}

	ctx := c.Request().Context()

	requestData := new(CreateShortLinkRequest)

	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Echo().Validator != nil {
		if err := c.Validate(requestData); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}

	link, err := h.Service.CreateShortLink(ctx, requestData.ShortURL, requestData.LongURL, email)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, CreateShortLinkResponse{
		Link: *link,
	})
}

type GetUserShortLinksResponse struct {
	Links []postgresDB.Link `json:"links"`
	User  postgresDB.User   `json:"user"`
}

// GetUserShortLinks godoc
//
//	@Summary		Gets user's short links
//	@Accept			json
//	@Param			limit	query		int	true	"Limit of a number of user's short links"
//	@Param			offset	query		int	true	"Maximum amount of links to show"
//	@Success		200			{object}	GetUserShortLinksResponse
//	@Failure		400			{} nil
//	@Failure		500			{} nil
//	@Router			/get_links      [get]
func (h *Handlers) GetUserShortLinks(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"redirectTo": "/login",
		})
	}

	limitParam, offsetParam := c.QueryParam("limit"), c.QueryParam("offset")

	limit, err := strconv.Atoi(limitParam)

	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	offset, err := strconv.Atoi(offsetParam)

	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	links, user, err := h.Service.GetUserShortLinksWithOffsetAndLimit(ctx, email, offset, limit)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, GetUserShortLinksResponse{
		Links: links,
		User:  *user,
	})
}

// GetLoginPage godoc
//
// @Summary Gets login page HTML
// @Produce	html
// @Success 200
// @Failure 500
// @Failure 307
// @Router /login	[get]
func (h *Handlers) GetLoginPage(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email != "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	return c.Render(http.StatusOK, "login_page.html", nil)
}

// GetRegisterPage godoc
//
// @Summary Gets register page HTML
// @Produce	html
// @Success 200
// @Failure 500
// @Failure 307
// @Router /register	[get]
func (h *Handlers) GetRegisterPage(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email != "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	return c.Render(http.StatusOK, "register_page.html", nil)
}

type UpdateUserShortLinksRequest struct {
	Email      string `json:"email" validate:"required"`
	DeltaLinks int    `json:"delta_links" validate:"required"`
}

type UpdateUserShortLinksResponse struct {
	User postgresDB.User `json:"user"`
}

func (h *Handlers) UpdateUserShortLinks(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email != "admin@admin.com" {
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("user %s is not authorized to change links number", email))
	}

	ctx := c.Request().Context()

	requestData := new(UpdateUserShortLinksRequest)

	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Echo().Validator != nil {
		if err := c.Validate(requestData); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}

	user, err := h.Service.UpdateUserShortLinks(ctx, requestData.Email, requestData.DeltaLinks)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, UpdateUserShortLinksResponse{
		User: *user,
	})

}

// GetCreateShortLink godoc
//
// @Summary Gets create link page HTML
// @Produce	html
// @Success 200
// @Failure 500
// @Failure 307
// @Router /create_link	[get]
func (h *Handlers) GetCreateShortLink(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	return c.Render(http.StatusOK, "create_link_page.html", nil)
}

// GetShortLink godoc
//
//	@Summary		Gets short link
//	@Param			ShortLink	path		string	true	"Short link to get"
//	@Success		307			{object}	DeleteShortLinkRequest
//	@Failure		400			{} nil
//	@Router			/      [get]
func (h *Handlers) GetShortLink(c echo.Context) error {
	ctx := c.Request().Context()

	shortLink := c.Param("short_link")

	link, err := h.Service.GetShortLink(ctx, shortLink)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusMovedPermanently, link.LongUrl)
}

type GetSubscriptionsResponse struct {
	Subscriptions []postgresDB.Subscription `json:"subscriptions"`
}

// GetSubscriptions godoc
//
//	@Summary		Gets all subscriptions
//	@Success		200			{object}	GetSubscriptionsResponse
//	@Failure		400			{} nil
//	@Failure		500			{} nil
//	@Router			/get_subscriptions      [get]
func (h *Handlers) GetSubscriptions(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"redirectTo": "/login",
		})
	}

	ctx := c.Request().Context()

	subscriptions, err := h.Service.GetSubscriptions(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"subscriptions": subscriptions,
	})
}

// GetSubscriptionsPage godoc
//
// @Summary Gets subscription page HTML
// @Produce	html
// @Success 200
// @Failure 500
// @Failure 307
// @Router /subscriptions	[get]
func (h *Handlers) GetSubscriptionsPage(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	return c.Render(http.StatusOK, "subscriptions.html", nil)

}

type GetUserResponse struct {
	User postgresDB.User `json:"user"`
}

// GetUser godoc
//
//	@Summary		Gets user from session
//	@Success		200			{object}	GetUserResponse
//	@Failure		400			{} nil
//	@Failure		500			{} nil
//	@Router			/user      [get]
func (h *Handlers) GetUser(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"redirectTo": "/login",
		})
	}

	ctx := c.Request().Context()

	user, err := h.Service.GetUser(ctx, email)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, GetUserResponse{
		User: *user,
	})

}

type DeleteShortLinkRequest struct {
	ShortLink string `json:"short_link"`
}

// DeleteShortLink godoc
//
//	@Summary		Tries to delete the short link
//	@Param			ShortLink	body		string	true	"Short link to delete"
//	@Success		200			{object}	DeleteShortLinkRequest
//	@Failure		400			{} nil
//	@Failure		500			{} nil
//	@Router			/delete_link      [delete]
func (h *Handlers) DeleteShortLink(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if email == "" {
		return c.JSON(http.StatusBadRequest, redirectResponse{
			redirectTo: "/login",
		})
	}

	ctx := c.Request().Context()

	requestData := new(DeleteShortLinkRequest)

	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Echo().Validator != nil {
		if err := c.Validate(requestData); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}

	err = h.Service.DeleteShortLink(ctx, requestData.ShortLink, email)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, nil)
}
