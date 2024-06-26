package handler

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sample-go-server/entity"
	handler2 "sample-go-server/module/token/handler"
	usecase2 "sample-go-server/module/token/usecase"
	"sample-go-server/module/user/usecase"
	"sample-go-server/test"
	"testing"
)

const APIPath = "/api/staffs/"

func SetupRouter() *gin.Engine {
	return gin.Default()
}

func Setup(mockData test.MockData) (*gin.Engine, entity.ITokenUseCase, UserHandler) {
	jwtSecret, _ := hex.DecodeString(test.JWTSecret)
	testUserRepo := test.NewTestUserRepo(mockData)
	tokenUC := usecase2.NewTokenUseCase(jwtSecret)
	userUC := usecase.NewUserUseCase(testUserRepo, tokenUC)

	router := SetupRouter()

	thandler := handler2.NewTokenHandler(tokenUC, userUC, jwtSecret)
	router.Use(thandler.Authenticate, thandler.AdminAuthorize)

	return router, tokenUC, NewUserHandler(router, userUC)
}

func TestUserHandler_CreateUser(t *testing.T) {

	mockData := test.NewMockData()
	router, tokenUC, handler := Setup(mockData)

	router.POST(APIPath, handler.CreateUser)

	tests := []struct {
		name   string
		user   entity.User
		new    *entity.NewUserRequest
		status int
	}{
		{
			"Admin Success",
			mockData.Admin,
			&entity.NewUserRequest{
				Username: "newstaff",
				Password: "12345",
				IsAdmin:  false,
			},
			http.StatusOK,
		},
		{
			"Admin add user already exists",
			mockData.Admin,
			&entity.NewUserRequest{
				Username: mockData.Staff.Username,
				Password: "12345",
				IsAdmin:  false,
			},
			http.StatusBadRequest,
		},
		{
			"Admin send incorrect data format",
			mockData.Admin,
			nil,
			http.StatusBadRequest,
		},
		{
			"Staff unauthorized",
			mockData.Staff,
			&entity.NewUserRequest{
				Username: "newstaff",
				Password: "12345",
				IsAdmin:  false,
			},
			http.StatusUnauthorized,
		},
		{
			"Malicious user",
			mockData.InvalidUser,
			&entity.NewUserRequest{
				Username: "newstaff",
				Password: "12345",
				IsAdmin:  false,
			},
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := tokenUC.Create(tt.user)
			token = "Bearer " + token

			var jsonValue []byte
			if tt.new != nil {
				jsonValue, _ = json.Marshal(tt.new)
			}
			req, _ := http.NewRequest("POST", APIPath, bytes.NewBuffer(jsonValue))
			req.Header.Set("Authorization", token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)

			if tt.status == http.StatusOK {
				var newUser entity.User
				json.Unmarshal(w.Body.Bytes(), &newUser)
				assert.NotEmpty(t, newUser)
			}
		})
	}

	testExpiredToken := tests[0]
	t.Run("Expired token", func(t *testing.T) {
		//token, _ := tokenUC.Create(testExpiredToken.user)
		token := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTkwOTIyOTUsImlhdCI6MTcxOTA4ODY5NSwidXNlcm5hbWUiOiJhZG1pbiJ9.mTHMQS_OQC1pbKTMecN0FrIFMxgRnWZzfRBMOoMNVDs"

		var jsonValue []byte
		if testExpiredToken.new != nil {
			jsonValue, _ = json.Marshal(testExpiredToken.new)
		}
		req, _ := http.NewRequest("POST", APIPath, bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	mockData := test.NewMockData()
	router, tokenUC, handler := Setup(mockData)

	router.DELETE(APIPath+":uuid", handler.DeleteUser)

	tests := []struct {
		name       string
		actionUser entity.User
		targetUser entity.User
		status     int
	}{
		{
			"Admin delete an exist user",
			mockData.Admin,
			mockData.Staff,
			http.StatusNoContent,
		},
		{
			"Admin delete an non-exist user",
			mockData.Admin,
			mockData.InvalidUser,
			http.StatusBadRequest,
		},
		{
			"Staff delete staff",
			mockData.Staff,
			mockData.Staff,
			http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := tokenUC.Create(tt.actionUser)
			token = "Bearer " + token

			req, _ := http.NewRequest("DELETE", APIPath+tt.targetUser.Uuid, nil)
			req.Header.Set("Authorization", token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestUserHandler_GetListUsers(t *testing.T) {
	mockData := test.NewMockData()
	router, tokenUC, handler := Setup(mockData)

	router.GET(APIPath, handler.GetListUsers)

	tests := []struct {
		name   string
		user   entity.User
		status int
	}{
		{
			"Admin Success",
			mockData.Admin,
			http.StatusOK,
		},
		{
			"Staff unauthorized",
			mockData.Staff,
			http.StatusUnauthorized,
		},
		{
			"Malicious user",
			mockData.InvalidUser,
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := tokenUC.Create(tt.user)
			token = "Bearer " + token

			req, _ := http.NewRequest("GET", APIPath, nil)
			req.Header.Set("Authorization", token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)

			if tt.status == http.StatusOK {
				var res []entity.User
				json.Unmarshal(w.Body.Bytes(), &res)
				assert.NotEmpty(t, res)
			}
		})
	}
}

func TestUserHandler_GetUserInfo(t *testing.T) {
	mockData := test.NewMockData()
	router, tokenUC, handler := Setup(mockData)

	router.GET(APIPath+":uuid", handler.GetUserInfo)

	tests := []struct {
		name       string
		actionUser entity.User
		targetUuid string
		status     int
	}{
		{
			"Admin Success",
			mockData.Admin,
			mockData.Staff.Uuid,
			http.StatusOK,
		},
		{
			"Admin view non-exist user",
			mockData.Admin,
			mockData.InvalidUser.Uuid,
			http.StatusBadRequest,
		},
		{
			"Staff unauthorized",
			mockData.Staff,
			mockData.Admin.Uuid,
			http.StatusUnauthorized,
		},
		{
			"Malicious user",
			mockData.InvalidUser,
			mockData.Staff.Uuid,
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := tokenUC.Create(tt.actionUser)
			token = "Bearer " + token

			req, _ := http.NewRequest("GET", APIPath+tt.targetUuid, nil)
			req.Header.Set("Authorization", token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)

			if tt.status == http.StatusOK {
				var res entity.User
				json.Unmarshal(w.Body.Bytes(), &res)
				assert.NotEmpty(t, res)
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	mockData := test.NewMockData()
	_, _, handler := Setup(mockData)

	loginpath := "/accounts/login"
	router := SetupRouter()
	router.POST(loginpath, handler.Login)

	tests := []struct {
		name   string
		user   entity.User
		req    *entity.UserLogin
		status int
	}{
		{
			"Admin Success",
			mockData.Admin,
			&entity.UserLogin{
				Username: mockData.Admin.Username,
				Password: mockData.Admin.Password,
			},
			http.StatusOK,
		},
		{
			"Admin wrong password",
			mockData.Admin,
			&entity.UserLogin{
				Username: mockData.Admin.Username,
				Password: "mockData.Admin.Password",
			},
			http.StatusBadRequest,
		},
		{
			"Staff success",
			mockData.Staff,
			&entity.UserLogin{
				Username: mockData.Staff.Username,
				Password: mockData.Staff.Password,
			},
			http.StatusOK,
		},
		{
			"Invalid request body",
			mockData.Admin,
			nil,
			http.StatusBadRequest,
		},
		{
			"Malicious user",
			mockData.InvalidUser,
			&entity.UserLogin{
				Username: "newstaff",
				Password: "12345",
			},
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonValue []byte
			if tt.req != nil {
				jsonValue, _ = json.Marshal(tt.req)
			}
			req, _ := http.NewRequest("POST", loginpath, bytes.NewBuffer(jsonValue))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)

			if tt.status == http.StatusOK {
				var res entity.LoginResponse
				json.Unmarshal(w.Body.Bytes(), &res)
				assert.NotEmpty(t, res)
			}
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	mockData := test.NewMockData()
	router, tokenUC, handler := Setup(mockData)

	router.PUT(APIPath+":uuid", handler.UpdateUser)

	tests := []struct {
		name       string
		user       entity.User
		targetUuid string
		req        *entity.UpdateUserRequest
		status     int
	}{
		{
			"Admin update staff",
			mockData.Admin,
			mockData.Staff.Uuid,
			&entity.UpdateUserRequest{
				Username: mockData.Staff.Username,
				Password: "12345sdfsdf",
				IsAdmin:  true,
			},
			http.StatusOK,
		},
		{
			"Admin update invalid data",
			mockData.Admin,
			mockData.Staff.Uuid,
			nil,
			http.StatusBadRequest,
		},
		{
			"Staff unauthorized",
			mockData.Staff,
			mockData.Staff.Uuid,
			&entity.UpdateUserRequest{
				Username: mockData.Staff.Username,
				Password: "1234sdfsdf5",
				IsAdmin:  true,
			},
			http.StatusUnauthorized,
		},
		{
			"Malicious user",
			mockData.InvalidUser,
			mockData.Staff.Uuid,
			&entity.UpdateUserRequest{
				Username: "newstaff",
				Password: "12345",
				IsAdmin:  false,
			},
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := tokenUC.Create(tt.user)
			token = "Bearer " + token

			var jsonValue []byte
			if tt.req != nil {
				jsonValue, _ = json.Marshal(tt.req)
			}
			req, _ := http.NewRequest("PUT", APIPath+tt.targetUuid, bytes.NewBuffer(jsonValue))
			req.Header.Set("Authorization", token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)
		})
	}
}
