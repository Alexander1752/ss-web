package routes

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"mqtt-streaming-server/domain"
)

// ErrUserAlreadyExists is returned when registering an email that is already in use.
var ErrUserAlreadyExists = errors.New("user already exists")

// UserService contains the business logic for user operations.
type UserService struct {
	Repo domain.UserRepository
}

// NewUserService creates a new UserService backed by the given repository.
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

// Register creates a new user account with a bcrypt-hashed password.
// Returns ErrUserAlreadyExists if the email is already taken.
func (s *UserService) Register(ctx context.Context, email, password string) error {
	existing, err := s.Repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return ErrUserAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.Repo.Save(ctx, email, string(hashed))
}

// Login verifies credentials and returns a signed JWT along with the matched user.
func (s *UserService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	claims := jwt.MapClaims{
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	return tokenStr, user, nil
}

// GetProfile returns the user's profile with the password field cleared.
func (s *UserService) GetProfile(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}
