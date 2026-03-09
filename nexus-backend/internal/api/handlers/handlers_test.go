package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/api/middleware"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type stubAuthSvc struct {
	loginFn         func(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error)
	validateTokenFn func(token string) (*domain.Claims, error)
	refreshTokenFn  func(ctx context.Context, token string) (*domain.TokenPair, error)
}

func (s *stubAuthSvc) Login(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error) {
	if s.loginFn != nil {
		return s.loginFn(ctx, req)
	}
	return &domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}, nil
}

func (s *stubAuthSvc) ValidateToken(token string) (*domain.Claims, error) {
	if s.validateTokenFn != nil {
		return s.validateTokenFn(token)
	}
	return &domain.Claims{UserID: uuid.New().String(), Email: "test@example.com", Role: domain.RoleOperator}, nil
}

func (s *stubAuthSvc) RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error) {
	if s.refreshTokenFn != nil {
		return s.refreshTokenFn(ctx, token)
	}
	return &domain.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh", ExpiresIn: 900}, nil
}

type stubShipmentSvc struct {
	createFn       func(ctx context.Context, req domain.CreateShipmentRequest, senderID uuid.UUID) (*domain.Shipment, error)
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*domain.Shipment, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, req domain.UpdateStatusRequest, actorID uuid.UUID) (*domain.Shipment, error)
	listFn         func(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error)
}

func (s *stubShipmentSvc) Create(ctx context.Context, req domain.CreateShipmentRequest, id uuid.UUID) (*domain.Shipment, error) {
	if s.createFn != nil {
		return s.createFn(ctx, req, id)
	}
	return &domain.Shipment{ID: uuid.New(), TrackingNumber: "NX-abc12345", Status: domain.StatusPending}, nil
}

func (s *stubShipmentSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return &domain.Shipment{ID: id, Status: domain.StatusPending}, nil
}

func (s *stubShipmentSvc) UpdateStatus(ctx context.Context, id uuid.UUID, req domain.UpdateStatusRequest, actorID uuid.UUID) (*domain.Shipment, error) {
	if s.updateStatusFn != nil {
		return s.updateStatusFn(ctx, id, req, actorID)
	}
	return &domain.Shipment{ID: id, Status: req.Status}, nil
}

func (s *stubShipmentSvc) List(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, senderID, limit, offset)
	}
	return []domain.Shipment{}, 0, nil
}

func buildShipmentRouter(svc shipmentService) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {

		c.Set(middleware.ContextUserID, uuid.New().String())
		c.Set(middleware.ContextRole, string(domain.RoleOperator))
		c.Next()
	})
	h := NewShipmentHandler(svc)
	r.POST("/shipments", h.Create)
	r.GET("/shipments", h.List)
	r.GET("/shipments/:id", h.GetByID)
	r.PUT("/shipments/:id/status", h.UpdateStatus)
	return r
}

func buildAuthRouter(svc authService) *gin.Engine {
	r := gin.New()
	h := NewAuthHandler(svc)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := &stubAuthSvc{}
	r := buildAuthRouter(svc)

	body := jsonBody(t, domain.LoginRequest{Email: "op@nexus.com", Password: "correctpassword"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var pair domain.TokenPair
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pair))
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestAuthHandler_Login_BadCredentials(t *testing.T) {
	svc := &stubAuthSvc{
		loginFn: func(_ context.Context, _ domain.LoginRequest) (*domain.TokenPair, error) {
			return nil, fmt.Errorf("invalid credentials")
		},
	}
	r := buildAuthRouter(svc)

	body := jsonBody(t, domain.LoginRequest{Email: "op@nexus.com", Password: "wrongpassword"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	svc := &stubAuthSvc{}
	r := buildAuthRouter(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"notanemail"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_Create_Success(t *testing.T) {
	stub := &stubShipmentSvc{}
	r := buildShipmentRouter(stub)

	payload := domain.CreateShipmentRequest{
		RecipientName: "Alice",
		Origin:        domain.Address{City: "Madrid", Country: "ES"},
		Destination:   domain.Address{City: "Barcelona", Country: "ES"},
		WeightKg:      2.5,
		EstimatedAt:   time.Now().Add(48 * time.Hour),
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/shipments", jsonBody(t, payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var s domain.Shipment
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &s))
	assert.NotEmpty(t, s.TrackingNumber)
}

func TestShipmentHandler_Create_InvalidBody(t *testing.T) {
	r := buildShipmentRouter(&stubShipmentSvc{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/shipments", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_Create_ServiceError(t *testing.T) {
	stub := &stubShipmentSvc{
		createFn: func(_ context.Context, _ domain.CreateShipmentRequest, _ uuid.UUID) (*domain.Shipment, error) {
			return nil, fmt.Errorf("db down")
		},
	}
	r := buildShipmentRouter(stub)

	payload := domain.CreateShipmentRequest{
		RecipientName: "Bob",
		Origin:        domain.Address{City: "Sevilla", Country: "ES"},
		Destination:   domain.Address{City: "Malaga", Country: "ES"},
		WeightKg:      1.0,
		EstimatedAt:   time.Now().Add(24 * time.Hour),
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/shipments", jsonBody(t, payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestShipmentHandler_GetByID_NotFound(t *testing.T) {
	stub := &stubShipmentSvc{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Shipment, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	r := buildShipmentRouter(stub)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/shipments/"+uuid.New().String(), nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestShipmentHandler_GetByID_InvalidUUID(t *testing.T) {
	r := buildShipmentRouter(&stubShipmentSvc{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/shipments/not-a-uuid", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_UpdateStatus_InvalidTransition(t *testing.T) {
	stub := &stubShipmentSvc{
		updateStatusFn: func(_ context.Context, _ uuid.UUID, _ domain.UpdateStatusRequest, _ uuid.UUID) (*domain.Shipment, error) {
			return nil, fmt.Errorf("invalid status transition from PENDING to DELIVERED")
		},
	}
	r := buildShipmentRouter(stub)

	body := jsonBody(t, domain.UpdateStatusRequest{Status: domain.StatusDelivered})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/shipments/"+uuid.New().String()+"/status", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_List_OK(t *testing.T) {
	stub := &stubShipmentSvc{
		listFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]domain.Shipment, int64, error) {
			return []domain.Shipment{{ID: uuid.New(), Status: domain.StatusPending}}, 1, nil
		},
	}
	r := buildShipmentRouter(stub)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/shipments?limit=10&offset=0", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float64(1), resp["total"])
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	r := buildAuthRouter(&stubAuthSvc{})
	body := jsonBody(t, map[string]string{"refresh_token": "valid-refresh"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var pair domain.TokenPair
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pair))
	assert.NotEmpty(t, pair.AccessToken)
}

func TestAuthHandler_Refresh_BadBody(t *testing.T) {
	r := buildAuthRouter(&stubAuthSvc{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Refresh_Error(t *testing.T) {
	svc := &stubAuthSvc{
		refreshTokenFn: func(_ context.Context, _ string) (*domain.TokenPair, error) {
			return nil, errors.New("invalid refresh token")
		},
	}
	r := buildAuthRouter(svc)
	body := jsonBody(t, map[string]string{"refresh_token": "bad"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHealth(t *testing.T) {
	r := gin.New()
	r.GET("/health", Health)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestReady(t *testing.T) {
	r := gin.New()
	r.GET("/ready", Ready)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAnalyticsHandler_DemandForecast(t *testing.T) {
	h := NewAnalyticsHandler()
	r := gin.New()
	r.GET("/analytics/forecast", h.DemandForecast)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/analytics/forecast", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAnalyticsHandler_KPIs(t *testing.T) {
	h := NewAnalyticsHandler()
	r := gin.New()
	r.GET("/analytics/kpis", h.KPIs)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/analytics/kpis", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Contains(t, resp, "total_shipments")
}

func TestOptimizerHandler_OptimizeRoute_Success(t *testing.T) {

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"route":"optimized"}`))
	}))
	defer upstream.Close()

	h := &OptimizerHandler{baseURL: upstream.URL, client: upstream.Client()}
	r := gin.New()
	r.POST("/optimize/route", h.OptimizeRoute)

	body := bytes.NewBufferString(`{"waypoints":[]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/optimize/route", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOptimizerHandler_OptimizeRoute_Unavailable(t *testing.T) {

	h := NewOptimizerHandler("127.0.0.1", "19999")
	r := gin.New()
	r.POST("/optimize/route", h.OptimizeRoute)

	body := bytes.NewBufferString(`{"waypoints":[]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/optimize/route", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestOptimizerHandler_OptimizeRoute_BodyReadError(t *testing.T) {
	h := NewOptimizerHandler("localhost", "9090")
	r := gin.New()
	r.POST("/optimize/route", h.OptimizeRoute)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/optimize/route", &errReader{})
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestOptimizerHandler_proxyPost_BadURL(t *testing.T) {
	h := &OptimizerHandler{baseURL: "::invalid::", client: &http.Client{}}
	_, err := h.proxyPost("/path", []byte("body"))
	assert.Error(t, err)
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

func TestByteReader_NonEmpty(t *testing.T) {
	r := byteReader("hello")
	p := make([]byte, 10)
	n, err := r.Read(p)
	assert.Equal(t, 5, n)
	assert.NoError(t, err)
}

func TestByteReader_Empty(t *testing.T) {
	r := byteReader(nil)
	p := make([]byte, 10)
	n, err := r.Read(p)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)
}

func buildShipmentRouterBadUser(svc shipmentService) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserID, "not-a-uuid")
		c.Set(middleware.ContextRole, string(domain.RoleOperator))
		c.Next()
	})
	h := NewShipmentHandler(svc)
	r.POST("/shipments", h.Create)
	return r
}

func TestShipmentHandler_Create_InvalidUserID(t *testing.T) {
	r := buildShipmentRouterBadUser(&stubShipmentSvc{})
	payload := domain.CreateShipmentRequest{
		RecipientName: "Alice",
		Origin:        domain.Address{City: "Madrid", Country: "ES"},
		Destination:   domain.Address{City: "Barcelona", Country: "ES"},
		WeightKg:      1.0,
		EstimatedAt:   time.Now().Add(24 * time.Hour),
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/shipments", jsonBody(t, payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_List_ServiceError(t *testing.T) {
	stub := &stubShipmentSvc{
		listFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]domain.Shipment, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	r := buildShipmentRouter(stub)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/shipments", nil))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestShipmentHandler_UpdateStatus_Success(t *testing.T) {
	r := buildShipmentRouter(&stubShipmentSvc{})
	body := jsonBody(t, domain.UpdateStatusRequest{Status: domain.StatusPickedUp})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/shipments/"+uuid.New().String()+"/status", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShipmentHandler_UpdateStatus_InvalidUUID(t *testing.T) {
	r := buildShipmentRouter(&stubShipmentSvc{})
	body := jsonBody(t, domain.UpdateStatusRequest{Status: domain.StatusPickedUp})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/shipments/not-a-uuid/status", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_UpdateStatus_BadBody(t *testing.T) {
	r := buildShipmentRouter(&stubShipmentSvc{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/shipments/"+uuid.New().String()+"/status",
		bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_GetBlockchainTrace_Success(t *testing.T) {
	id := uuid.New()
	stub := &stubShipmentSvc{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, TrackingNumber: "NX-abc123", Status: domain.StatusInTransit}, nil
		},
	}
	r := buildShipmentRouterWithTrace(stub)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/shipments/"+id.String()+"/trace", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShipmentHandler_GetBlockchainTrace_InvalidUUID(t *testing.T) {
	r := buildShipmentRouterWithTrace(&stubShipmentSvc{})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/shipments/not-a-uuid/trace", nil))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestShipmentHandler_GetBlockchainTrace_NotFound(t *testing.T) {
	stub := &stubShipmentSvc{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Shipment, error) {
			return nil, errors.New("not found")
		},
	}
	r := buildShipmentRouterWithTrace(stub)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/shipments/"+uuid.New().String()+"/trace", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func buildShipmentRouterWithTrace(svc shipmentService) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserID, uuid.New().String())
		c.Set(middleware.ContextRole, string(domain.RoleOperator))
		c.Next()
	})
	h := NewShipmentHandler(svc)
	r.GET("/shipments/:id/trace", h.GetBlockchainTrace)
	return r
}

func TestShipmentHandler_GetByID_Success(t *testing.T) {
	id := uuid.New()
	r := buildShipmentRouter(&stubShipmentSvc{})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/shipments/"+id.String(), nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	var s domain.Shipment
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &s))
	assert.Equal(t, id, s.ID)
}
