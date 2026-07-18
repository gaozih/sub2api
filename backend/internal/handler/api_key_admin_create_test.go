package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func TestCreateAPIKeyForUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		userID     string
		body       string
		create     createAPIKeyForUserFunc
		wantStatus int
		wantUserID int64
	}{
		{
			name:       "非法用户 ID",
			userID:     "bad",
			body:       `{"name":"agentos"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "非法请求",
			userID:     "42",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "服务错误",
			userID: "42",
			body:   `{"name":"agentos"}`,
			create: func(context.Context, int64, service.CreateAPIKeyRequest) (*service.APIKey, error) {
				return nil, service.ErrGroupNotAllowed
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "成功创建",
			userID: "42",
			body:   `{"name":"agentos","group_id":7,"custom_key":"sk-agentos-managed-key"}`,
			create: func(_ context.Context, userID int64, req service.CreateAPIKeyRequest) (*service.APIKey, error) {
				if userID != 42 || req.GroupID == nil || *req.GroupID != 7 || req.CustomKey == nil {
					return nil, errors.New("请求映射错误")
				}
				return &service.APIKey{ID: 9, UserID: userID, Key: *req.CustomKey, Name: req.Name, GroupID: req.GroupID, Status: service.StatusActive}, nil
			},
			wantStatus: http.StatusOK,
			wantUserID: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/api/v1/admin/users/:id/api-keys", func(c *gin.Context) {
				create := tt.create
				if create == nil {
					create = func(context.Context, int64, service.CreateAPIKeyRequest) (*service.APIKey, error) {
						t.Fatal("不应调用 service")
						return nil, nil
					}
				}
				createAPIKeyForUser(c, create)
			})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+tt.userID+"/api-keys", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("状态码：期望 %d，实际 %d；响应 %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}
			if tt.wantUserID != 0 {
				var body struct {
					Data struct {
						UserID int64 `json:"user_id"`
					} `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatalf("解析响应：%v", err)
				}
				if body.Data.UserID != tt.wantUserID {
					t.Fatalf("用户 ID：期望 %d，实际 %d", tt.wantUserID, body.Data.UserID)
				}
			}
		})
	}
}
