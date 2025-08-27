package chat_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/api/http/dto"
	"gopi.com/api/http/handler"
	"gopi.com/internal/app/chat"
	"gopi.com/internal/app/user"
	chatModel "gopi.com/internal/domain/chat/model"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/email"
	"gopi.com/internal/lib/jwt"
	chatMocks "gopi.com/tests/mocks/chat"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper
func setupChatTest(t *testing.T) (*gin.Engine, *chatMocks.MockGroupRepository, *chatMocks.MockMessageRepository, *userMocks.MockUserRepository, *jwt.JWTService) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockGroupRepo := new(chatMocks.MockGroupRepository)
	mockMessageRepo := new(chatMocks.MockMessageRepository)
	mockUserRepo := new(userMocks.MockUserRepository)

	// Create JWT service for testing
	jwtService := jwt.NewJWTService("test-secret", time.Hour, 24*time.Hour, nil)

	// Create real services with mock repositories for integration testing
	chatSvc := chat.NewChatService(mockGroupRepo, mockMessageRepo)

	// Create a concrete email service for testing (minimal config)
	emailConfig := email.EmailConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test@example.com",
		Password: "test-password",
		From:     "test@example.com",
	}
	emailService := email.NewEmailService(emailConfig)
	userSvc := user.NewUserService(mockUserRepo, emailService)

	// Create handler with real services
	chatHandler := handler.NewChatHandler(chatSvc, userSvc)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	protected := router.Group("/chat")
	protected.Use(func(c *gin.Context) {
		// Mock auth middleware - set user_id in context
		c.Set("user_id", "test-user-id")
		c.Next()
	})

	protected.POST("/groups", chatHandler.CreateGroup)
	protected.GET("/groups", chatHandler.GetGroups)
	protected.GET("/groups/:slug", chatHandler.GetGroupBySlug)
	protected.PUT("/groups/:slug", chatHandler.UpdateGroup)
	protected.DELETE("/groups/:slug", chatHandler.DeleteGroup)
	protected.POST("/groups/:slug/join", chatHandler.JoinGroup)
	protected.POST("/groups/:slug/leave", chatHandler.LeaveGroup)

	return router, mockGroupRepo, mockMessageRepo, mockUserRepo, jwtService
}

func TestChatHandler_CreateGroup(t *testing.T) {
	router, mockGroupRepo, _, _, _ := setupChatTest(t)

	tests := []struct {
		name           string
		requestBody    dto.CreateGroupRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful group creation",
			requestBody: dto.CreateGroupRequest{
				Name:      "Test Group",
				Image:     "test-image.jpg",
				MemberIDs: []string{"member1", "member2"},
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				mockGroupRepo.On("Create", mock.MatchedBy(func(g *chatModel.Group) bool {
					return g.Name == "Test Group" &&
						g.CreatorID == "test-user-id" &&
						len(g.MemberIDs) == 3 && // creator + 2 members
						g.Image == "test-image.jpg"
				})).Return(nil)
			},
		},
		{
			name: "group name too long",
			requestBody: dto.CreateGroupRequest{
				Name: "This is a very long group name that exceeds the 20 character limit",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mock setup needed - validation happens before service call
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/chat/groups", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatHandler_GetGroups(t *testing.T) {
	router, mockGroupRepo, _, _, _ := setupChatTest(t)

	testGroups := []*chatModel.Group{
		{
			Base: model.Base{
				ID:        "group1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Group 1",
			CreatorID: "test-user-id",
			MemberIDs: []string{"test-user-id", "member1"},
			Slug:      "group-1-abc123",
		},
		{
			Base: model.Base{
				ID:        "group2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Group 2",
			CreatorID: "other-user",
			MemberIDs: []string{"test-user-id", "member2"},
			Slug:      "group-2-def456",
		},
	}

	mockGroupRepo.On("GetByMemberID", "test-user-id").Return(testGroups, nil)

	req, _ := http.NewRequest(http.MethodGet, "/chat/groups", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.GroupListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Groups, 2)
	assert.Equal(t, 2, response.Total)

	mockGroupRepo.AssertExpectations(t)
}

func TestChatHandler_GetGroupBySlug(t *testing.T) {
	router, mockGroupRepo, _, _, _ := setupChatTest(t)

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful group retrieval",
			slug:           "test-group-abc123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "test-user-id",
					MemberIDs: []string{"test-user-id", "member1"},
					Slug:      "test-group-abc123",
					Image:     "test-image.jpg",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
			},
		},
		{
			name:           "group not found",
			slug:           "nonexistent-group",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockGroupRepo.On("GetBySlug", "nonexistent-group").Return(nil, assert.AnError)
			},
		},
		{
			name:           "not a member",
			slug:           "private-group",
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "private-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Private Group",
					CreatorID: "other-user",
					MemberIDs: []string{"other-user", "member1"},
					Slug:      "private-group",
				}
				mockGroupRepo.On("GetBySlug", "private-group").Return(group, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/chat/groups/"+tt.slug, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatHandler_UpdateGroup(t *testing.T) {
	router, mockGroupRepo, _, _, _ := setupChatTest(t)

	tests := []struct {
		name           string
		slug           string
		requestBody    dto.UpdateGroupRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful group update",
			slug: "test-group-abc123",
			requestBody: dto.UpdateGroupRequest{
				Name:  "Updated Group Name",
				Image: "updated-image.jpg",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Original Group",
					CreatorID: "test-user-id",
					MemberIDs: []string{"test-user-id", "member1"},
					Slug:      "test-group-abc123",
					Image:     "original-image.jpg",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
				mockGroupRepo.On("Update", mock.MatchedBy(func(g *chatModel.Group) bool {
					return g.Name == "Updated Group Name" && g.Image == "updated-image.jpg"
				})).Return(nil)
			},
		},
		{
			name: "group name too long",
			slug: "test-group-abc123",
			requestBody: dto.UpdateGroupRequest{
				Name: "This is a very long group name that exceeds the 20 character limit",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Original Group",
					CreatorID: "test-user-id",
					MemberIDs: []string{"test-user-id", "member1"},
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
			},
		},
		{
			name: "not the creator",
			slug: "test-group-abc123",
			requestBody: dto.UpdateGroupRequest{
				Name: "Updated Group Name",
			},
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Original Group",
					CreatorID: "other-user", // Different from test-user-id
					MemberIDs: []string{"other-user", "member1"},
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
				// Note: Update should NOT be called since authorization fails
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockGroupRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/chat/groups/"+tt.slug, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatHandler_JoinGroup(t *testing.T) {
	_, mockGroupRepo, _, _, _ := setupChatTest(t)

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful join",
			slug:           "test-group-abc123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "other-user",
					MemberIDs: []string{"other-user", "member1"},
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
				mockGroupRepo.On("GetByID", "test-group-id").Return(group, nil)
				mockGroupRepo.On("AddMember", "test-group-id", "test-user-id").Return(nil)
			},
		},
		{
			name:           "already a member",
			slug:           "test-group-abc123",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "other-user",
					MemberIDs: []string{"other-user", "test-user-id", "member1"},
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
				mockGroupRepo.On("GetByID", "test-group-id").Return(group, nil)
				mockGroupRepo.On("AddMember", "test-group-id", "member1").Return(errors.New("member already exists in group"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test
			freshMockGroupRepo := new(chatMocks.MockGroupRepository)
			freshMockMessageRepo := new(chatMocks.MockMessageRepository)
			freshMockUserRepo := new(userMocks.MockUserRepository)

			// Create email service
			emailConfig := email.EmailConfig{
				Host:     "localhost",
				Port:     587,
				Username: "test@example.com",
				Password: "test-password",
				From:     "test@example.com",
			}
			freshEmailService := email.NewEmailService(emailConfig)

			// Create fresh service
			freshChatSvc := chat.NewChatService(freshMockGroupRepo, freshMockMessageRepo)
			freshUserSvc := user.NewUserService(freshMockUserRepo, freshEmailService)
			freshChatHandler := handler.NewChatHandler(freshChatSvc, freshUserSvc)

			// Create fresh router with middleware
			freshRouter := gin.New()
			freshRouter.Use(gin.Recovery())

			protected := freshRouter.Group("/chat")
			protected.Use(func(c *gin.Context) {
				c.Set("user_id", "test-user-id")
				c.Next()
			})

			protected.POST("/groups/:slug/join", freshChatHandler.JoinGroup)

			// Set up expectations on fresh mocks
			group := &chatModel.Group{
				Base: model.Base{
					ID:        "test-group-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Group",
				CreatorID: "other-user",
				MemberIDs: func() []string {
					if tt.name == "already a member" {
						return []string{"other-user", "member1", "test-user-id"}
					}
					return []string{"other-user", "member1"}
				}(),
				Slug: "test-group-abc123",
			}
			freshMockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
			if tt.name != "already a member" {
				freshMockGroupRepo.On("GetByID", "test-group-id").Return(group, nil)
				freshMockGroupRepo.On("AddMember", "test-group-id", "test-user-id").Return(nil)
			}

			req, _ := http.NewRequest(http.MethodPost, "/chat/groups/"+tt.slug+"/join", nil)
			w := httptest.NewRecorder()
			freshRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			freshMockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatHandler_LeaveGroup(t *testing.T) {
	router, mockGroupRepo, _, _, _ := setupChatTest(t)

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful leave",
			slug:           "test-group-abc123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "other-user",
					MemberIDs: []string{"other-user", "test-user-id", "member1"},
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
				mockGroupRepo.On("GetByID", "test-group-id").Return(group, nil)
				mockGroupRepo.On("RemoveMember", "test-group-id", "test-user-id").Return(nil)
			},
		},
		{
			name:           "not a member",
			slug:           "test-group-abc123",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "test-group-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "other-user",
					MemberIDs: []string{"other-user", "member1"}, // test-user-id is NOT a member
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetBySlug", "test-group-abc123").Return(group, nil)
				mockGroupRepo.On("GetByID", "test-group-id").Return(group, nil)
				// Note: RemoveMember should NOT be called since authorization fails
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockGroupRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodPost, "/chat/groups/"+tt.slug+"/leave", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockGroupRepo.AssertExpectations(t)
		})
	}
}
