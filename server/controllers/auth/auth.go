package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/middleware"
	"ark-server-commander/models"
	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Password policy for accounts created through this package. bcrypt silently
// truncates its input at 72 bytes, so longer passwords are rejected instead of
// being quietly weakened.
const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

// errAlreadyInitialized is returned from the init transaction when a user
// already exists, so the handler can answer 400 instead of 500
var errAlreadyInitialized = errors.New("already initialized")

// validatePassword enforces the account password policy
func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("Password must be at least %d characters", minPasswordLength)
	}
	if len(password) > maxPasswordLength {
		return fmt.Errorf("Password must be at most %d bytes", maxPasswordLength)
	}
	return nil
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
	// Unscoped: soft-deleted users still count as initialized, so this stays
	// consistent with the check InitUser performs
	var count int64
	if err := database.DB.Unscoped().Model(&models.User{}).Count(&count).Error; err != nil {
		utils.Error("count users failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

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

	if err := validatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Password, hashed outside the transaction so bcrypt does not hold the
	// write lock for the duration of the hash
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

	// The "no user exists yet" check and the insert must be atomic, otherwise
	// two concurrent requests can both observe an empty table and both create
	// an admin. The unique index on username makes the loser fail cleanly.
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// YesNoUser
		var count int64
		if countErr := tx.Unscoped().Model(&models.User{}).Count(&count).Error; countErr != nil {
			return countErr
		}
		if count > 0 {
			return errAlreadyInitialized
		}

		return tx.Create(&user).Error
	})
	if err != nil {
		if errors.Is(err, errAlreadyInitialized) || strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			c.JSON(http.StatusBadRequest, gin.H{"error": " Initialize"})
			return
		}
		utils.Error("init user failed", zap.Error(err))
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

	// ParseRefreshToken, not ParseToken: only a refresh token may be exchanged
	// here, otherwise an access token can be rolled over into a fresh 30-day pair
	claims, err := utils.ParseRefreshToken(req.RefreshToken)
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

// tokenFingerprint returns a short non-reversible fingerprint of a token,
// safe to persist in audit logs for correlation
func tokenFingerprint(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])[:8]
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

	middleware.Log.Log(c.GetUint("user_id"), "auth.logout", "token", tokenFingerprint(parts[1]), c.ClientIP())

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
