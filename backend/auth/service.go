package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	repo      *Repository
	jwtSecret []byte
	isPostgres bool
}

func NewService(db *sql.DB, jwtSecret string, isPostgres bool) *Service {
	return &Service{
		repo:      NewRepository(db),
		jwtSecret: []byte(jwtSecret),
		isPostgres: isPostgres,
	}
}

func (s *Service) InitializeDatabase() error {
	var err error
	if s.isPostgres {
		err = s.repo.CreateUsersTablePostgres()
	} else {
		err = s.repo.CreateUsersTable()
	}
	
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	count, err := s.repo.CountUsers()
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count == 0 {
		password := s.generateRandomPassword(16)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		if s.isPostgres {
			err = s.repo.CreateUserPostgres("admin", string(hashedPassword))
		} else {
			err = s.repo.CreateUser("admin", string(hashedPassword))
		}
		
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		log.Infof("========================================")
		log.Infof("ADMIN USER CREATED")
		log.Infof("Username: admin")
		log.Infof("Password: %s", password)
		log.Infof("PLEASE SAVE THIS PASSWORD - IT WILL NOT BE SHOWN AGAIN")
		log.Infof("========================================")
	}

	return nil
}

func (s *Service) generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "admin123"
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

func (s *Service) Login(username, password string) (*LoginResponse, error) {
	var user *User
	var err error
	
	if s.isPostgres {
		user, err = s.repo.GetUserByUsernamePostgres(username)
	} else {
		user, err = s.repo.GetUserByUsername(username)
	}
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, expiresIn, err := s.generateToken(user.Username, 2*time.Hour)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.generateToken(user.Username, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Username:     user.Username,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *Service) generateToken(username string, duration time.Duration) (string, int64, error) {
	expiresAt := time.Now().Add(duration)
	claims := jwt.MapClaims{
		"username": username,
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}
	
	return tokenString, expiresAt.Unix(), nil
}

func (s *Service) RefreshToken(refreshTokenString string) (*LoginResponse, error) {
	username, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	token, expiresIn, err := s.generateToken(username, 2*time.Hour)
	if err != nil {
		return nil, err
	}

	newRefreshToken, _, err := s.generateToken(username, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		Username:     username,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", errors.New("invalid token claims")
		}
		return username, nil
	}

	return "", errors.New("invalid token")
}

func (s *Service) ChangePassword(username, oldPassword, newPassword string) error {
	var user *User
	var err error
	
	if s.isPostgres {
		user, err = s.repo.GetUserByUsernamePostgres(username)
	} else {
		user, err = s.repo.GetUserByUsername(username)
	}
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if s.isPostgres {
		return s.repo.UpdatePasswordPostgres(username, string(hashedPassword))
	}
	return s.repo.UpdatePassword(username, string(hashedPassword))
}
