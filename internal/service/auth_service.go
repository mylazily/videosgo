package service

import (
	"errors"
	"time"

	"videosgo/internal/config"
	"videosgo/internal/model"
	"videosgo/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
}

// NewAuthService 创建认证服务
func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求（支持用户名或邮箱登录）
type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*model.User, error) {
	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Status:    "active",
		Role:      "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录（返回 tokens 和 user）
func (s *AuthService) Login(req *LoginRequest) (*TokenResponse, *model.User, error) {
	var user *model.User
	var err error

	// 支持用户名或邮箱登录
	if req.Username != "" {
		user, err = s.userRepo.GetByUsername(req.Username)
	} else if req.Email != "" {
		user, err = s.userRepo.GetByEmail(req.Email)
	} else {
		return nil, nil, errors.New("请输入用户名或邮箱")
	}

	if err != nil {
		return nil, nil, errors.New("用户名或密码错误")
	}
	if user == nil {
		return nil, nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, nil, errors.New("用户名或密码错误")
	}

	// 生成 Token
	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

// generateTokens 生成 JWT Token
func (s *AuthService) generateTokens(user *model.User) (*TokenResponse, error) {
	expiresIn := s.cfg.JWT.ExpireHours * 3600

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: uuid.New().String(),
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}
