package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type stubValidator struct {
	fn func(token string) (*domain.Claims, error)
}

func (s *stubValidator) ValidateToken(token string) (*domain.Claims, error) {
	if s.fn != nil {
		return s.fn(token)
	}
	return &domain.Claims{
		UserID: uuid.New().String(),
		Email:  "test@example.com",
		Role:   domain.RoleOperator,
	}, nil
}

func newTestRouter(mw gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func do(r *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(rec, req)
	return rec
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	r := newTestRouter(RequestID())
	rec := do(r, http.MethodGet, "/test", nil)
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestRequestID_UsesExistingHeader(t *testing.T) {
	r := newTestRouter(RequestID())
	rec := do(r, http.MethodGet, "/test", map[string]string{"X-Request-ID": "my-id"})
	assert.Equal(t, "my-id", rec.Header().Get("X-Request-ID"))
}

func TestStructuredLogger_DoesNotPanic(t *testing.T) {
	r := newTestRouter(StructuredLogger())
	require.NotPanics(t, func() {
		do(r, http.MethodGet, "/test", nil)
	})
}

func TestRecovery_CatchesPanic(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) { panic("boom") })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestSecurityHeaders(t *testing.T) {
	r := newTestRouter(SecurityHeaders())
	rec := do(r, http.MethodGet, "/test", nil)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, rec.Header().Get("Strict-Transport-Security"))
	assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
}

func TestCORS_Development(t *testing.T) {
	r := newTestRouter(CORS("development"))
	rec := do(r, http.MethodGet, "/test", map[string]string{"Origin": "http://localhost:3000"})
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_Production(t *testing.T) {
	r := newTestRouter(CORS("production"))
	rec := do(r, http.MethodGet, "/test", map[string]string{"Origin": "http://evil.com"})
	assert.Equal(t, "https://app.nexus-logistics.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_Options_Preflight(t *testing.T) {
	r := newTestRouter(CORS("development"))
	rec := do(r, http.MethodOptions, "/test", map[string]string{"Origin": "http://localhost"})
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRateLimit_AllowsNormalTraffic(t *testing.T) {
	r := newTestRouter(RateLimit(10, time.Second))
	for i := 0; i < 5; i++ {
		rec := do(r, http.MethodGet, "/test", nil)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRateLimit_BlocksExcessiveTraffic(t *testing.T) {

	r := gin.New()
	r.Use(RateLimit(1, time.Minute))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

func TestBucketCleanupLoop_RemovesStaleEntries(t *testing.T) {
	ch := make(chan time.Time, 1)

	staleIP := "192.0.2.99"
	buckets.Store(staleIP, &bucket{
		tokens:   10,
		lastSeen: time.Now().Add(-24 * time.Hour),
	})

	done := make(chan struct{})
	go func() {
		bucketCleanupLoop(1*time.Millisecond, ch)
		close(done)
	}()

	ch <- time.Now()
	close(ch)
	<-done

	_, still := buckets.Load(staleIP)
	assert.False(t, still, "stale bucket should have been deleted")
}

func TestBucketCleanupLoop_KeepsFreshEntries(t *testing.T) {
	ch := make(chan time.Time, 1)

	freshIP := "192.0.2.50"
	buckets.Store(freshIP, &bucket{
		tokens:   10,
		lastSeen: time.Now(),
	})

	done := make(chan struct{})
	go func() {
		bucketCleanupLoop(1*time.Minute, ch)
		close(done)
	}()

	ch <- time.Now()
	close(ch)
	<-done

	_, still := buckets.Load(freshIP)
	assert.True(t, still, "fresh bucket should NOT have been deleted")
	buckets.Delete(freshIP)
}

func TestStartBucketCleanup_DoesNotPanic(t *testing.T) {
	require.NotPanics(t, func() {
		startBucketCleanup(time.Minute)
	})
}

func TestMinFloat(t *testing.T) {
	assert.Equal(t, 1.0, minFloat(1.0, 2.0))
	assert.Equal(t, 1.0, minFloat(3.0, 1.0))
	assert.Equal(t, 5.0, minFloat(5.0, 5.0))
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	r := newTestRouter(JWTAuth(&stubValidator{}))
	rec := do(r, http.MethodGet, "/test", nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_MalformedHeader_NoBearer(t *testing.T) {
	r := newTestRouter(JWTAuth(&stubValidator{}))
	rec := do(r, http.MethodGet, "/test", map[string]string{"Authorization": "Basic abc"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_MalformedHeader_MissingToken(t *testing.T) {
	r := newTestRouter(JWTAuth(&stubValidator{}))
	rec := do(r, http.MethodGet, "/test", map[string]string{"Authorization": "Bearer"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	v := &stubValidator{fn: func(token string) (*domain.Claims, error) {
		return nil, assert.AnError
	}}
	r := newTestRouter(JWTAuth(v))
	rec := do(r, http.MethodGet, "/test", map[string]string{"Authorization": "Bearer bad-token"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_ValidToken(t *testing.T) {
	userID := uuid.New().String()
	v := &stubValidator{fn: func(token string) (*domain.Claims, error) {
		return &domain.Claims{UserID: userID, Email: "u@test.com", Role: domain.RoleAdmin}, nil
	}}
	r := gin.New()
	r.Use(JWTAuth(v))
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get(ContextUserID)
		c.String(http.StatusOK, id.(string))
	})
	rec := do(r, http.MethodGet, "/test", map[string]string{"Authorization": "Bearer valid"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, userID, rec.Body.String())
}

func TestRequireRole_Allowed(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextRole, string(domain.RoleAdmin))
		c.Next()
	})
	r.Use(RequireRole(string(domain.RoleAdmin), string(domain.RoleOperator)))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := do(r, http.MethodGet, "/test", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_Forbidden(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextRole, string(domain.RoleViewer))
		c.Next()
	})
	r.Use(RequireRole(string(domain.RoleAdmin)))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := do(r, http.MethodGet, "/test", nil)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
