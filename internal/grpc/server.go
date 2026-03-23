// Package grpc реализует gRPC сервер для сервиса сокращения URL.
// Предоставляет те же возможности, что и HTTP сервер, через протокол gRPC.
package grpc

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	pb "github.com/rebaxis/urlshrter/internal/grpc/pb"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

type contextKey string

const userIDKey contextKey = "userID"

// Server реализует ShortenerServiceServer — gRPC-фасад над бизнес-логикой сервиса.
type Server struct {
	pb.UnimplementedShortenerServiceServer
	urlService service.URLService
	opts       shortener.Opts
}

// NewServer создаёт новый экземпляр gRPC сервера.
func NewServer(urlService service.URLService, opts shortener.Opts) *Server {
	return &Server{
		urlService: urlService,
		opts:       opts,
	}
}

// ShortenURL обрабатывает RPC ShortenURL (аналог POST /api/shorten).
func (s *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	urlEnt := model.URLEnt{
		OriginalURL: req.GetUrl(),
		UserID:      userID,
	}

	data, saveErr := s.urlService.SaveURL(urlEnt, s.opts)
	if saveErr != nil && !errors.Is(saveErr, service.ErrExistID) {
		return nil, status.Errorf(codes.Internal, "internal error: %v", saveErr)
	}

	return &pb.URLShortenResponse{Result: data.Result}, nil
}

// ExpandURL обрабатывает RPC ExpandURL (аналог GET /<id>).
func (s *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	url, err := s.urlService.GetURL(req.GetId())
	if err != nil {
		if errors.Is(err, service.ErrDeletedID) {
			return nil, status.Errorf(codes.NotFound, "URL was deleted")
		}
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}
	if url == "" {
		return nil, status.Errorf(codes.NotFound, "URL not found for id: %s", req.GetId())
	}

	return &pb.URLExpandResponse{Result: url}, nil
}

// ListUserURLs обрабатывает RPC ListUserURLs (аналог GET /api/user/urls).
func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	data, err := s.urlService.GetURLByUser(userID, s.opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	resp := &pb.UserURLsResponse{}
	for _, v := range data.URLS {
		shortURL := s.opts.BaseURL + "/" + v.ShortURL
		resp.Url = append(resp.Url, &pb.URLData{
			ShortUrl:    shortURL,
			OriginalUrl: v.OriginalURL,
		})
	}

	return resp, nil
}

// TokenParser определяет интерфейс для парсинга JWT токенов.
type TokenParser interface {
	ParseToken(tokenString string) (*service.Claims, error)
}

// AuthInterceptor возвращает унарный серверный перехватчик для авторизации через JWT.
// Читает токен из метаданных заголовка "authorization".
// Если заголовок отсутствует — генерирует новый userID (как HTTP AuthorizationMw).
// Если заголовок присутствует, но токен невалиден — возвращает Unauthenticated.
func AuthInterceptor(parser TokenParser) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		values := md.Get("authorization")
		if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
			userID, err := lib.GenerateRandomAlphabetString(6)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to generate user ID: %v", err)
			}
			ctx = context.WithValue(ctx, userIDKey, userID)
			return handler(ctx, req)
		}

		claims, err := parser.ParseToken(values[0])
		if err != nil || claims == nil || claims.UserID == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		return handler(ctx, req)
	}
}

// userIDFromContext извлекает userID из контекста, установленного AuthInterceptor.
func userIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}
