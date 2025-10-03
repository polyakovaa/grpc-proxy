package service

import "github.com/polyakovaa/grpcproxy/gen/auth"

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
}

// ?? че они возвращают
// internal/auth/service/auth.go
//func (s *AuthServer) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
// Проверяет JWT подпись, expiration, etc.
// Использует секретные ключи
//}

//func (s *AuthServer) RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
// Проверяет refresh token
// Генерирует новые токены
// Обновляет в БД если нужно
//}

//func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
// 1. Хэширует пароль (bcrypt)
// 2. Сохраняет пользователя в БД
// 3. Генерирует JWT токены
// 4. Возвращает готовые токены
// → Вся бизнес-логика здесь!
//}
