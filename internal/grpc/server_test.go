package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	pb "github.com/rebaxis/urlshrter/internal/grpc/pb"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

// --- stubs ---

type stubRepo struct {
	urls map[string]model.URLEnt
}

func newStubRepo() *stubRepo {
	return &stubRepo{urls: make(map[string]model.URLEnt)}
}

func (r *stubRepo) Save(ent model.URLEnt) error {
	r.urls[ent.ShortURL] = ent
	return nil
}

func (r *stubRepo) SaveBatch(batch model.URLBatch) error {
	for _, e := range batch.URLS {
		r.urls[e.ShortURL] = e
	}
	return nil
}

func (r *stubRepo) Get(id string) (model.URLEnt, error) {
	ent, ok := r.urls[id]
	if !ok {
		return model.URLEnt{}, nil
	}
	return ent, nil
}

func (r *stubRepo) GetByURL(url string) (model.URLEnt, error) {
	for _, e := range r.urls {
		if e.OriginalURL == url {
			return e, nil
		}
	}
	return model.URLEnt{}, nil
}

func (r *stubRepo) GetEntByURL(url string) (model.URLEnt, error) {
	return r.GetByURL(url)
}

func (r *stubRepo) GetEntByUser(userID string) (model.URLBatch, error) {
	var batch model.URLBatch
	for _, e := range r.urls {
		if e.UserID == userID {
			batch.URLS = append(batch.URLS, e)
		}
	}
	return batch, nil
}

func (r *stubRepo) DeleteURLByUser(_ context.Context, ch chan model.DeleteURLRecord) error {
	for range ch {
	}
	return nil
}

func (r *stubRepo) GetStats() (int, int, error) {
	return len(r.urls), 1, nil
}

// stubTokenParser simulates JWT parsing results.
type stubTokenParser struct {
	claims *service.Claims
	err    error
}

func (p *stubTokenParser) ParseToken(_ string) (*service.Claims, error) {
	return p.claims, p.err
}

// --- helpers ---

func ctxWithUserID(userID string) context.Context {
	return context.WithValue(context.Background(), userIDKey, userID)
}

func newTestServer() *Server {
	repo := newStubRepo()
	urlSvc := service.NewURLService(repo)
	opts := shortener.Opts{
		BaseURL: "http://localhost:8080",
	}
	return NewServer(urlSvc, opts)
}

// --- tests ---

func TestShortenURL_Success(t *testing.T) {
	srv := newTestServer()

	ctx := ctxWithUserID("user1")
	resp, err := srv.ShortenURL(ctx, &pb.URLShortenRequest{Url: "https://example.com"})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetResult())
	assert.Contains(t, resp.GetResult(), "http://localhost:8080/")
}

func TestShortenURL_Unauthorized(t *testing.T) {
	srv := newTestServer()

	_, err := srv.ShortenURL(context.Background(), &pb.URLShortenRequest{Url: "https://example.com"})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestExpandURL_NotFound(t *testing.T) {
	srv := newTestServer()

	ctx := ctxWithUserID("user1")
	_, err := srv.ExpandURL(ctx, &pb.URLExpandRequest{Id: "nonexistent"})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestExpandURL_Success(t *testing.T) {
	repo := newStubRepo()
	repo.urls["abc12345"] = model.URLEnt{ShortURL: "abc12345", OriginalURL: "https://example.com", UserID: "user1"}
	urlSvc := service.NewURLService(repo)
	srv := NewServer(urlSvc, shortener.Opts{BaseURL: "http://localhost:8080"})

	ctx := ctxWithUserID("user1")
	resp, err := srv.ExpandURL(ctx, &pb.URLExpandRequest{Id: "abc12345"})

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", resp.GetResult())
}

func TestListUserURLs_Empty(t *testing.T) {
	srv := newTestServer()

	ctx := ctxWithUserID("user99")
	resp, err := srv.ListUserURLs(ctx, &emptypb.Empty{})

	require.NoError(t, err)
	assert.Empty(t, resp.GetUrl())
}

func TestAuthInterceptor_NoMetadata_GeneratesUserID(t *testing.T) {
	parser := &stubTokenParser{}
	interceptor := AuthInterceptor(parser)

	var capturedCtx context.Context
	handler := func(ctx context.Context, _ interface{}) (interface{}, error) {
		capturedCtx = ctx
		return nil, nil
	}

	_, err := interceptor(context.Background(), nil, nil, handler)
	require.NoError(t, err)

	userID, ok := capturedCtx.Value(userIDKey).(string)
	assert.True(t, ok)
	assert.NotEmpty(t, userID)
}

func TestAuthInterceptor_ValidToken_SetsUserID(t *testing.T) {
	parser := &stubTokenParser{
		claims: &service.Claims{UserID: "user42"},
	}
	interceptor := AuthInterceptor(parser)

	md := metadata.Pairs("authorization", "some-jwt-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var capturedCtx context.Context
	handler := func(ctx context.Context, _ interface{}) (interface{}, error) {
		capturedCtx = ctx
		return nil, nil
	}

	_, err := interceptor(ctx, nil, nil, handler)
	require.NoError(t, err)

	userID, ok := capturedCtx.Value(userIDKey).(string)
	assert.True(t, ok)
	assert.Equal(t, "user42", userID)
}

func TestAuthInterceptor_InvalidToken_ReturnsUnauthenticated(t *testing.T) {
	parser := &stubTokenParser{
		err: errors.New("bad token"),
	}
	interceptor := AuthInterceptor(parser)

	md := metadata.Pairs("authorization", "bad-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, nil, func(context.Context, interface{}) (interface{}, error) {
		return nil, nil
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}
