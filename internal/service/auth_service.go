package service

import (
	"errors"
	"fmt"
	"time"

	"release-manager/internal/config"
	"release-manager/internal/model"
	"release-manager/internal/repository"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
	logger   *config.Logger
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, logger *config.Logger) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		logger:   logger,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// LDAP认证
	ldapUser, err := s.ldapAuth(req.Username, req.Password)
	if err != nil {
		s.logger.Warnw("LDAP auth failed", "username", req.Username, "error", err)
		return nil, errors.New("用户名或密码错误")
	}

	// 查找或创建用户
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新用户
			isAdmin := s.isAdminUser(req.Username)
			user = &model.User{
				Username:    req.Username,
				DisplayName: ldapUser.DisplayName,
				Email:       ldapUser.Email,
				IsAdmin:     isAdmin,
			}
			if err := s.userRepo.Create(user); err != nil {
				s.logger.Errorw("Failed to create user", "error", err)
				return nil, errors.New("创建用户失败")
			}
		} else {
			return nil, errors.New("系统错误")
		}
	}

	// 生成JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	s.logger.Infow("User logged in", "username", user.Username, "userId", user.ID)

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

type ldapUserInfo struct {
	DisplayName string
	Email       string
}

func (s *AuthService) ldapAuth(username, password string) (*ldapUserInfo, error) {
	// 开发环境或LDAP未配置时模拟认证
	if s.cfg.Env == "development" || s.cfg.LDAP.Host == "ldap.example.com" || s.cfg.LDAP.BindPass == "" {
		s.logger.Infow("Development mode: skipping LDAP auth", "username", username)
		return &ldapUserInfo{
			DisplayName: username,
			Email:       username + "@example.com",
		}, nil
	}

	// 连接LDAP服务器
	conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", s.cfg.LDAP.Host, s.cfg.LDAP.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect LDAP: %w", err)
	}
	defer conn.Close()

	// 绑定管理员账号搜索用户
	if err := conn.Bind(s.cfg.LDAP.BindDN, s.cfg.LDAP.BindPass); err != nil {
		return nil, fmt.Errorf("failed to bind admin: %w", err)
	}

	// 搜索用户
	searchFilter := fmt.Sprintf(s.cfg.LDAP.UserFilter, ldap.EscapeFilter(username))
	searchReq := ldap.NewSearchRequest(
		s.cfg.LDAP.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn", "cn", "mail", "displayName"},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search user: %w", err)
	}

	if len(result.Entries) != 1 {
		return nil, errors.New("user not found")
	}

	userDN := result.Entries[0].DN

	// 使用用户凭证验证
	if err := conn.Bind(userDN, password); err != nil {
		return nil, errors.New("invalid password")
	}

	return &ldapUserInfo{
		DisplayName: result.Entries[0].GetAttributeValue("displayName"),
		Email:       result.Entries[0].GetAttributeValue("mail"),
	}, nil
}

func (s *AuthService) isAdminUser(username string) bool {
	for _, admin := range s.cfg.LDAP.AdminUsers {
		if admin == username {
			return true
		}
	}
	return false
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}
