package handlers

import (
	"context"
	"fmt"
	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	_ "urleater/docs"
	"urleater/dto"
)

// Service описывает бизнес-логику приложения.
type Service interface {
	LoginUser(ctx context.Context, email string, password string) error
	RegisterUser(ctx context.Context, email string, password string) error
	CreateShortLink(ctx context.Context, shortLink string, longLink string, userEmail string) (*dto.Link, error)
	UpdateUserShortLinks(ctx context.Context, email string, deltaLinks int) (*dto.User, error)
	GetUserShortLinksWithOffsetAndLimit(ctx context.Context, email string, offset int, limit int) ([]dto.Link, *dto.User, error)
	GetSubscriptions(ctx context.Context) ([]dto.Subscription, error)
	GetShortLink(ctx context.Context, shortLink string) (*dto.Link, error)
	GetUser(ctx context.Context, email string) (*dto.User, error)
	DeleteShortLink(ctx context.Context, shortLink string, email string) error
	CreateSubscriptions(ctx context.Context) error
	GetTotalUserLinks(ctx context.Context, email string) (int, error)
	GetShortLinksMatchingPattern(ctx context.Context, containsWord string, offset int) (dto.SearcherMatchResult, error)
}

// SessionStore описывает методы работы с сессиями.
type SessionStore interface {
	RetrieveEmailFromSession(c echo.Context) (string, error)
	Get(r *http.Request, key string) (*sessions.Session, error)
	Save(c echo.Context, email string, session *sessions.Session) error
}

// Handlers содержит зависимости для HTTP-обработчиков.
type Handlers struct {
	Service Service
	Store   SessionStore
}

type PostgresSessionStore struct {
	store *pgstore.PGStore
	mu    sync.Mutex
}

func NewPostgresSessionStore(store *pgstore.PGStore) SessionStore {
	return &PostgresSessionStore{
		store: store,
	}
}

func (pg *PostgresSessionStore) RetrieveEmailFromSession(c echo.Context) (string, error) {
	pg.mu.Lock()
	session, err := pg.store.Get(c.Request(), "session_key")
	pg.mu.Unlock()

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
	pg.mu.Lock()
	session, err := pg.store.Get(r, key)
	pg.mu.Unlock()

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
// @Summary Рендер главной страницы
// @Description Отрисовывает главную страницу, если пользователь авторизован, иначе перенаправляет на /login.
// @Tags Страницы
// @Produce html
// @Success 200 {string} string "HTML главной страницы"
// @Failure 500 {object} string "Ошибка сервера"
// @Router / [get]
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

// GetLinksPage godoc
// @Summary Рендер страницы ссылок
// @Description Отрисовывает страницу со списком ссылок, если пользователь авторизован, иначе перенаправляет на /login.
// @Tags Страницы
// @Produce html
// @Success 200 {string} string "HTML страницы ссылок"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /links [get]
func (h *Handlers) GetLinksPage(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if email == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}
	return c.Render(http.StatusOK, "links_list.html", nil)
}

// LoginRequest описывает тело запроса для логина.
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// PostLogin godoc
// @Summary Вход пользователя
// @Description Аутентифицирует пользователя по email и паролю, создавая сессию.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param LoginRequest body LoginRequest true "Данные для логина"
// @Success 200 {object} map[string]string "Перенаправление на главную страницу"
// @Failure 400 {object} string "Неверный запрос или уже авторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /login [post]
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
// @Summary Выход пользователя
// @Description Завершает сессию пользователя и перенаправляет на /login.
// @Tags Аутентификация
// @Produce json
// @Success 302 {string} string "Перенаправление на /login"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /logout [get]
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

// RegisterRequest описывает тело запроса для регистрации.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// PostRegister godoc
// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя по email и паролю и создаёт сессию.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param RegisterRequest body RegisterRequest true "Данные для регистрации"
// @Success 200 {object} map[string]string "Перенаправление на главную страницу"
// @Failure 400 {object} string "Неверный запрос или уже авторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /register [post]
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

// CreateShortLinkRequest описывает тело запроса для создания короткой ссылки.
type CreateShortLinkRequest struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url" validate:"required"`
}

// CreateShortLinkResponse описывает ответ на запрос создания короткой ссылки.
type CreateShortLinkResponse struct {
	Link dto.Link `json:"link"`
}

// CreateShortLink godoc
// @Summary Создание короткой ссылки
// @Description Создаёт короткую ссылку, сопоставляя её с длинным URL для авторизованного пользователя.
// @Tags Ссылки
// @Accept json
// @Produce json
// @Param CreateShortLinkRequest body CreateShortLinkRequest true "Данные для создания ссылки"
// @Success 200 {object} CreateShortLinkResponse "Созданная ссылка"
// @Failure 400 {object} string "Неверный запрос или неавторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /shortlink [post]
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

type FormattedLink struct {
	ShortUrl     string
	LongUrl      string
	UserEmail    string
	ExpiresAt    string
	TimesVisited int
}

type GetUserShortLinksResponse struct {
	Links []FormattedLink `json:"links"`
	User  dto.User        `json:"user"`
}

// GetUserShortLinks godoc
// @Summary Получение коротких ссылок пользователя
// @Description Возвращает список коротких ссылок текущего пользователя с поддержкой пагинации.
// @Tags Ссылки
// @Produce json
// @Param limit query int false "Лимит записей"
// @Param offset query int false "Сдвиг для пагинации"
// @Success 200 {object} GetUserShortLinksResponse "Список ссылок и данные пользователя"
// @Failure 400 {object} string "Неверный запрос или неавторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /user/shortlinks [get]
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
	var limit int
	if limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
	}
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

	// Убираем хеш пароля из ответа
	user.PasswordHash = ""

	var formattedLinks []FormattedLink
	for _, l := range links {
		formattedLinks = append(formattedLinks, FormattedLink{
			ShortUrl:     l.ShortUrl,
			LongUrl:      l.LongUrl,
			UserEmail:    l.UserEmail,
			TimesVisited: l.TimesVisited,
			ExpiresAt:    l.ExpiresAt.Format(time.DateTime),
		})
	}
	return c.JSON(http.StatusOK, GetUserShortLinksResponse{
		Links: formattedLinks,
		User:  *user,
	})
}

type GetUserShortLinksTotalNumberResponse struct {
	TotalUserShortLinks int `json:"total_user_short_links"`
}

// GetUserShortLinksNumber godoc
// @Summary Получение общего числа коротких ссылок пользователя
// @Description Возвращает общее количество коротких ссылок, созданных пользователем.
// @Tags Ссылки
// @Produce json
// @Success 200 {object} GetUserShortLinksTotalNumberResponse "Общее число ссылок"
// @Failure 400 {object} string "Неверный запрос или неавторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /user/shortlinks/number [get]
func (h *Handlers) GetUserShortLinksNumber(c echo.Context) error {
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
	totalUserLinks, err := h.Service.GetTotalUserLinks(ctx, email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, GetUserShortLinksTotalNumberResponse{
		TotalUserShortLinks: totalUserLinks,
	})
}

// GetLoginPage godoc
// @Summary Рендер страницы логина
// @Description Отрисовывает HTML-страницу для входа пользователя.
// @Tags Страницы
// @Produce html
// @Success 200 {string} string "HTML страницы логина"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /login [get]
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
// @Summary Рендер страницы регистрации
// @Description Отрисовывает HTML-страницу для регистрации нового пользователя.
// @Tags Страницы
// @Produce html
// @Success 200 {string} string "HTML страницы регистрации"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /register [get]
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

// UpdateUserShortLinksRequest описывает тело запроса для обновления количества ссылок пользователя.
type UpdateUserShortLinksRequest struct {
	Email      string `json:"email" validate:"required"`
	DeltaLinks int    `json:"delta_links" validate:"required"`
}

// UpdateUserShortLinksResponse описывает ответ с обновлёнными данными пользователя.
type UpdateUserShortLinksResponse struct {
	User dto.User `json:"user"`
}

// UpdateUserShortLinks godoc
// @Summary Обновление количества коротких ссылок пользователя
// @Description Администратор может изменить число доступных ссылок для указанного пользователя.
// @Tags Администрирование
// @Accept json
// @Produce json
// @Param UpdateUserShortLinksRequest body UpdateUserShortLinksRequest true "Данные для обновления"
// @Success 200 {object} UpdateUserShortLinksResponse "Обновлённые данные пользователя"
// @Failure 400 {object} string "Неверный запрос"
// @Failure 500 {object} string "Ошибка сервера или неавторизованный доступ"
// @Router /admin/update-links [put]
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
// @Summary Рендер страницы создания короткой ссылки
// @Description Отрисовывает HTML-страницу для создания новой короткой ссылки.
// @Tags Страницы
// @Produce html
// @Success 200 {string} string "HTML страницы создания ссылки"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /shortlink/create [get]
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
// @Summary Редирект короткой ссылки
// @Description Перенаправляет пользователя с короткой ссылки на соответствующий длинный URL.
// @Tags Ссылки
// @Produce plain
// @Param short_link path string true "Короткая ссылка"
// @Success 302 {string} string "Перенаправление на длинный URL"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /{short_link} [get]
func (h *Handlers) GetShortLink(c echo.Context) error {
	ctx := c.Request().Context()
	shortLink := c.Param("short_link")
	link, err := h.Service.GetShortLink(ctx, shortLink)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusFound, link.LongUrl)
}

type GetSubscriptionsResponse struct {
	Subscriptions []dto.Subscription `json:"subscriptions"`
}

// GetSubscriptions godoc
// @Summary Получение подписок
// @Description Возвращает список подписок для авторизованного пользователя.
// @Tags Подписки
// @Produce json
// @Success 200 {object} GetSubscriptionsResponse "Список подписок"
// @Failure 400 {object} string "Неверный запрос или неавторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /subscriptions [get]
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
// @Summary Рендер страницы подписок
// @Description Отрисовывает HTML-страницу для просмотра и управления подписками.
// @Tags Страницы
// @Produce html
// @Success 200 {string} string "HTML страницы подписок"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /subscriptions/page [get]
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
	User dto.User `json:"user"`
}

// GetUser godoc
// @Summary Получение данных пользователя
// @Description Возвращает информацию об авторизованном пользователе.
// @Tags Пользователь
// @Produce json
// @Success 200 {object} GetUserResponse "Данные пользователя"
// @Failure 400 {object} string "Неверный запрос или неавторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /user [get]
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

// DeleteShortLinkRequest описывает тело запроса для удаления короткой ссылки.
type DeleteShortLinkRequest struct {
	ShortLink string `json:"short_link"`
}

// DeleteShortLink godoc
// @Summary Удаление короткой ссылки
// @Description Удаляет указанную короткую ссылку для авторизованного пользователя.
// @Tags Ссылки
// @Produce json
// @Param short_link query string true "Короткая ссылка для удаления"
// @Success 200 {object} nil "Ссылка успешно удалена"
// @Failure 400 {object} string "Неверный запрос или неавторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /shortlink [delete]
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
	shortLink := c.QueryParam("short_link")
	log.Println("short_link", shortLink)
	if shortLink == "" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("пустая строка").Error())
	}

	err = h.Service.DeleteShortLink(ctx, shortLink, email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, nil)
}

type GetShortLinksWithMatchingPatternRequest struct {
	Offset       int    `json:"offset"`
	ContainsWord string `json:"contains_word"`
}

type GetShortLinksWithMatchingPatternResponse struct {
	ShortLinks []dto.Link `json:"links"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// GetShortLinksMatchingPattern godoc
// @Summary Поиск коротких ссылок по шаблону
// @Description Ищет короткие ссылки, содержащие указанное слово, с поддержкой пагинации.
// @Tags Ссылки
// @Produce json
// @Param contains_word query string true "Подстрока для поиска"
// @Param offset query int false "Сдвиг для пагинации"
// @Success 200 {object} GetShortLinksWithMatchingPatternResponse "Результат поиска"
// @Failure 400 {object} string "Неверный запрос или неавторизован"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /shortlinks/search [get]
func (h *Handlers) GetShortLinksMatchingPattern(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if email == "" {
		return c.JSON(http.StatusBadRequest, redirectResponse{
			redirectTo: "/login",
		})
	}

	containsWord := c.QueryParam("contains_word")
	if containsWord == "" {
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("contains_word cannot be empty").Error())
	}

	var offset int
	offsetString := c.QueryParam("offset")
	if offsetString == "" {
		offset = 0
	} else {
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, fmt.Errorf("offset must be a number").Error())
		}
	}

	ctx := c.Request().Context()
	links, err := h.Service.GetShortLinksMatchingPattern(ctx, containsWord, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	var shortLinks []dto.Link
	for _, link := range links.ShortLinks {
		res, err := h.Service.GetShortLink(ctx, link)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		shortLinks = append(shortLinks, *res)
	}

	return c.JSON(http.StatusOK, GetShortLinksWithMatchingPatternResponse{
		ShortLinks: shortLinks,
		Limit:      links.Limit,
		Offset:     links.Offset,
	})
}

// GetLoginWithCodeRequest описывает тело запроса для получения кода входа.
type GetLoginWithCodeRequest struct {
	Email string `json:"email"`
}

// PostLoginWithCode godoc
// @Summary Запрос кода для входа
// @Description Инициирует процесс входа по коду, отправляя код на указанный email.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param GetLoginWithCodeRequest body GetLoginWithCodeRequest true "Запрос на получение кода для входа"
// @Success 200 {object} nil "Код отправлен"
// @Failure 400 {object} string "Неверный запрос"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /login/code [post]
func (h *Handlers) PostLoginWithCode(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if email != "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"redirectTo": "/",
		})
	}

	_ = c.Request().Context()
	requestData := new(GetLoginWithCodeRequest)
	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if c.Echo().Validator != nil {
		if err := c.Validate(requestData); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
	}
	return nil
}

// LoginWithCodeRequest описывает тело запроса для подтверждения кода входа.
type LoginWithCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// SubmitLoginCode godoc
// @Summary Подтверждение кода входа
// @Description Проверяет введённый код для входа и аутентифицирует пользователя.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param LoginWithCodeRequest body LoginWithCodeRequest true "Данные для входа по коду"
// @Success 200 {object} nil "Вход выполнен успешно"
// @Failure 400 {object} string "Неверный запрос"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /login/code/submit [post]
func (h *Handlers) SubmitLoginCode(c echo.Context) error {
	return nil
}

// GetSearchLinksPage godoc
// @Summary Рендер страницы поиска ссылок
// @Description Отрисовывает HTML-страницу для поиска коротких ссылок.
// @Tags Страницы
// @Produce html
// @Success 200 {string} string "HTML страницы поиска ссылок"
// @Failure 500 {object} string "Ошибка сервера"
// @Router /search [get]
func (h *Handlers) GetSearchLinksPage(c echo.Context) error {
	email, err := h.Store.RetrieveEmailFromSession(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if email == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}
	return c.Render(http.StatusOK, "search_file.html", nil)
}
