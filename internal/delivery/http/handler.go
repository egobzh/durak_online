package http

import (
	"errors"
	"net/http"

	"durak-online/internal/delivery/http/middleware"
	"durak-online/internal/domain"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(authService domain.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary Регистрация нового игрока
// @Description Создаёт аккаунт по e-mail, логину и паролю. После регистрации автоматически выдаётся JWT-токен, 5000 кредитов и 1200 очков рейтинга.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} domain.AuthResponse "Успешная регистрация"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 409 {object} map[string]string "Пользователь с таким логином или e-mail уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// Login godoc
// @Summary Вход в аккаунт
// @Description Возвращает JWT-токен по логину и паролю. Токен действителен 24 часа.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Логин и пароль"
// @Success 200 {object} domain.AuthResponse "Успешная авторизация"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 401 {object} map[string]string "Неверный логин или пароль"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ForgotPassword godoc
// @Summary Запрос ссылки для сброса пароля
// @Description Отправляет ссылку для смены пароля на e-mail. На текущем этапе ссылка выводится в консоль сервера. Всегда возвращает одинаковый ответ (защита от перебора аккаунтов).
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body domain.ForgotPasswordRequest true "E-mail пользователя"
// @Success 200 {object} domain.MessageResponse "Статусное сообщение"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ForgotPassword(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{
		Message: "Если указанный e-mail зарегистрирован в системе, ссылка для сброса пароля была отправлена.",
	})
}

// ResetPassword godoc
// @Summary Сброс пароля по токену
// @Description Устанавливает новый пароль по одноразовому токену из письма. После смены пароля все активные сессии аннулируются.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "Токен и новый пароль"
// @Success 200 {object} domain.MessageResponse "Статусное сообщение"
// @Failure 400 {object} map[string]string "Недействительный или просроченный токен"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken),
			errors.Is(err, domain.ErrTokenExpired),
			errors.Is(err, domain.ErrTokenAlreadyUsed):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{
		Message: "Пароль успешно изменён. Войдите с новым паролем.",
	})
}

// GetProfile godoc
// @Summary Профиль текущего игрока
// @Description Возвращает данные авторизованного игрока: рейтинг, баланс, количество матчей и побед.
// @Tags Профиль
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.UserProfileResponse "Профиль игрока"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/profile/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	profile, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}
