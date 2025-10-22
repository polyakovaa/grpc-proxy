package service_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenRepo struct {
	mock.Mock
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockTokenRepo) CreateRefreshToken(rt *model.RefreshToken) error {
	args := m.Called(rt)
	return args.Error(0)
}

func (m *MockTokenRepo) FindByTokenHash(token string) (*model.RefreshToken, error) {
	return nil, nil
}

func (m *MockTokenRepo) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepo) FindByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) CreateUser(user *model.User) (*model.User, error) {
	args := m.Called(user)
	return nil, args.Error(0)
}

func TestGenerateTokens_Success(t *testing.T) {
	trepo := new(MockTokenRepo)
	urepo := new(MockUserRepo)

	urepo.On("FindByID", mock.Anything).Return(&model.User{ID: uuid.New().String()}, nil)
	trepo.On("CreateRefreshToken", mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	svc := service.NewAuthService(urepo, trepo, "secret", time.Minute*15, time.Hour*24)

	userID := uuid.New().String()
	token, err := svc.GenerateTokens(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.RefreshToken)
	trepo.AssertCalled(t, "CreateRefreshToken", mock.AnythingOfType("*model.RefreshToken"))

}

func TestGenerateTokens_JWTClaims(t *testing.T) {
	trepo := new(MockTokenRepo)
	urepo := new(MockUserRepo)

	urepo.On("FindByID", mock.Anything).Return(&model.User{ID: uuid.New().String()}, nil)
	trepo.On("CreateRefreshToken", mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	svc := service.NewAuthService(urepo, trepo, "secret", time.Minute*15, time.Hour*24)
	userID := "user-123"
	tokens, err := svc.GenerateTokens(userID)
	assert.NoError(t, err)
	parsed, err := jwt.Parse(tokens.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, userID, claims["user_id"])
	assert.NotEmpty(t, claims["expires_at"])

}

func TestValidateAccessToken_Success(t *testing.T) {
	trepo := new(MockTokenRepo)
	urepo := new(MockUserRepo)

	urepo.On("FindByID", mock.Anything).Return(&model.User{ID: uuid.New().String()}, nil)
	trepo.On("CreateRefreshToken", mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	svc := service.NewAuthService(urepo, trepo, "secret123", time.Minute*15, time.Hour*24)

	claims := jwt.MapClaims{
		"user_id":    uuid.New().String(),
		"token_id":   uuid.New().String(),
		"expires_at": time.Now().Add(-time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("secret"))

	assert.NoError(t, err)

	u, exp, ok := svc.ValidateAccessToken(tokenStr)
	assert.False(t, ok, "invalid")
	assert.True(t, exp.Before(time.Now()), "expired")
	assert.Nil(t, u, "invalid")

}
