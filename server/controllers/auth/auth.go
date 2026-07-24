package auth

import (
	"net/http"
	"strings"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/middleware"
	"ark-server-commander/models"
	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Check Init YesNoInitializeUser
// @Summary Initialization status
// @Description YesNoInitializeUser
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} map[string]bool "Initialization status"
// @Router /auth/check-init [get]
func CheckInit(c *gin.Context) {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"initialized": count > 0,
	})
}

// InitUser InitializeUser
// @Summary InitializeUser
// @Description CreateUser，Initialize
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.UserRequest true "User"
// @Success 200 {object} map[string]interface{} "InitializeSuccess"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/init [post]
func InitUser(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// YesNoUser
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": " Initialize"})
		return
	}

	// Password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password "})
		return
	}

	// CreateUser
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UserCreate "})
		return
	}

	// token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": " "})
		return
	}

	middleware.Log.Log(user.ID, "auth.init", "user", user.Username, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "InitializeSuccess",
		"token":   token,
		"user": models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}


// Login UserLogin
// @Summary UserLogin
// @Description UserPasswordLogin
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body models.UserRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "LoginSuccess"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Authentication"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// User
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User PasswordError"})
		return
	}

	// Password
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User PasswordError"})
		return
	}

	// token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": " "})
		return
	}

	// refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": " "})
		return
	}

	middleware.Log.Log(user.ID, "auth.login", "user", user.Username, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":       "LoginSuccess",
		"token":         token,
		"refresh_token": refreshToken,
		"user": models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}

// RefreshToken refreshes the JWT access token
// @Summary JWT
// @Description refresh_tokenaccess_tokenrefresh_token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true ""
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "None"
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// refresh token
	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "None "})
		return
	}

	// refresh token
	utils.BlacklistToken(req.RefreshToken, time.Now().Add(24*time.Hour))

	// access tokenrefresh token
	accessToken, err := utils.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": " "})
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken(claims.UserID, claims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": " "})
		return
	}

	middleware.Log.Log(claims.UserID, "auth.refresh", "token", "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":       " Success",
		"token":         accessToken,
		"refresh_token": newRefreshToken,
		"user": models.UserResponse{
			ID:       claims.UserID,
			Username: claims.Username,
		},
	})
}

// Logout User
// @Summary User
// @Description JWT
// @Tags Authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]string "Success"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": " "})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		c.JSON(http.StatusBadRequest, gin.H{"error": " Error"})
		return
	}

	// token
	utils.BlacklistToken(parts[1], time.Now().Add(24*time.Hour))

	middleware.Log.Log(c.GetUint("user_id"), "auth.logout", "token", parts[1], c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": " Success",
	})
}

// GetProfile User
// @Summary User
// @Description LoginUser
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]models.UserResponse "User"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /profile [get]
func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	c.JSON(http.StatusOK, gin.H{
		"user": models.UserResponse{
			ID:       userID,
			Username: username,
		},
	})
}
