package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rayda/rayda-service/internal/api/handlers"
	"github.com/rayda/rayda-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditLogRepository is a mock implementation of AuditLogRepository
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, filter *model.AuditLogFilter, page, pageSize int) ([]*model.AuditLog, int64, error) {
	args := m.Called(ctx, tenantID, filter, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.AuditLog), int64(args.Int(1)), args.Error(2)
}

func (m *MockAuditLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*model.AuditLog, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).([]*model.AuditLog), int64(args.Int(1)), args.Error(2)
}

func (m *MockAuditLogRepository) FindByResource(ctx context.Context, resourceType model.ResourceType, resourceID string, page, pageSize int) ([]*model.AuditLog, int64, error) {
	args := m.Called(ctx, resourceType, resourceID, page, pageSize)
	return args.Get(0).([]*model.AuditLog), int64(args.Int(1)), args.Error(2)
}

func TestAuditLogHandler_ListAuditLogs(t *testing.T) {
	// Setup
	mockRepo := new(MockAuditLogRepository)
	handler := handlers.NewAuditLogHandler(mockRepo)

	// Test cases
	tests := []struct {
		name           string
		setupMock     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			setupMock: func() {
				testLogs := []*model.AuditLog{{
					Base: model.Base{
						ID:        uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Action:       model.ActionCreate,
					UserID:       uuid.New(),
					TenantID:     uuid.New(),
					ResourceType: model.ResourceUser,
					ResourceID:   uuid.New().String(),
				}}
				mockRepo.On("FindByTenantID", mock.Anything, mock.Anything, mock.Anything, 1, 1000).
					Return(testLogs, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid tenant ID",
			setupMock:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid tenant ID in context"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			if tt.setupMock != nil {
				tt.setupMock()
			}

			// Setup test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Add tenant ID to context for success case
			if tt.expectedStatus != http.StatusBadRequest {
				c.Set("tenant_id", uuid.New().String())
			}

			// Add query parameters
			c.Request, _ = http.NewRequest("GET", "/audit-logs?page=1&page_size=10", nil)

			// Call handler
			handler.ListAuditLogs(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}
