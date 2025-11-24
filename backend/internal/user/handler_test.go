/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package user

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	syscontext "github.com/asgardeo/thunder/internal/system/context"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

type UserHandlerTestSuite struct {
	suite.Suite
}

const (
	defaultUserID              = "user-123"
	defaultUserAttributes      = `{"username":"alice"}`
	defaultEmailAttributes     = `{"email":"alice@example.com"}`
	defaultCredentialAttribute = `{"password":"Secret123!"}`
)

func TestUserHandler_UserHandlerTestSuite_Run(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

type flakyResponseWriter struct {
	*httptest.ResponseRecorder
	failNext bool
}

func newFlakyResponseWriter() *flakyResponseWriter {
	return &flakyResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		failNext:         true,
	}
}

func (w *flakyResponseWriter) Write(b []byte) (int, error) {
	if w.failNext {
		w.failNext = false
		return 0, errors.New("write failure")
	}
	return w.ResponseRecorder.Write(b)
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.ensureRuntime()
}

func (suite *UserHandlerTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (suite *UserHandlerTestSuite) ensureRuntime() {
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("", &config.Config{})
	suite.Require().NoError(err)
}

type userHandlerTestCase struct {
	name          string
	method        string
	url           string
	body          string
	useFlaky      bool
	withAuth      bool
	authUserID    string
	setup         func(*UserServiceInterfaceMock)
	assert        func(*httptest.ResponseRecorder)
	assertService func(*UserServiceInterfaceMock)
}

func (suite *UserHandlerTestSuite) runHandlerTestCases(
	testCases []userHandlerTestCase,
	invoke func(*userHandler, http.ResponseWriter, *http.Request),
) {
	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			serviceMock := NewUserServiceInterfaceMock(suite.T())
			handler := newUserHandler(serviceMock)

			method := tc.method
			if method == "" {
				method = http.MethodGet
			}

			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			}

			req := httptest.NewRequest(method, tc.url, bodyReader)
			if tc.body != "" {
				req.Header.Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			}

			if tc.withAuth {
				userID := tc.authUserID
				if userID == "" {
					userID = defaultUserID
				}
				authCtx := syscontext.NewAuthenticationContext(userID, "", "", "", nil)
				req = req.WithContext(syscontext.WithAuthenticationContext(req.Context(), authCtx))
			}

			var writer http.ResponseWriter
			var recorder *httptest.ResponseRecorder
			if tc.useFlaky {
				flaky := newFlakyResponseWriter()
				writer = flaky
				recorder = flaky.ResponseRecorder
			} else {
				recorder = httptest.NewRecorder()
				writer = recorder
			}

			if tc.setup != nil {
				tc.setup(serviceMock)
			}

			invoke(handler, writer, req)

			if tc.assert != nil {
				tc.assert(recorder)
			}

			if tc.assertService != nil {
				tc.assertService(serviceMock)
			} else {
				serviceMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *UserHandlerTestSuite) TestHandleSelfUserGetRequest() {
	testCases := []userHandlerTestCase{
		{
			name:     "success",
			url:      "/users/me",
			withAuth: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("GetUser", defaultUserID).
					Return(&User{
						ID:         defaultUserID,
						Attributes: json.RawMessage(defaultUserAttributes),
					}, nil).
					Once()
			},
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusOK, recorder.Code)
				suite.Equal(serverconst.ContentTypeJSON, recorder.Header().Get(serverconst.ContentTypeHeaderName))

				var respUser User
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &respUser))
				suite.Equal(defaultUserID, respUser.ID)
				suite.JSONEq(defaultUserAttributes, string(respUser.Attributes))
			},
		},
		{
			name: "unauthorized",
			url:  "/users/me",
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusUnauthorized, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorAuthenticationFailed.Code, errResp.Code)
			},
			assertService: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.AssertNotCalled(suite.T(), "GetUser", mock.Anything)
			},
		},
		{
			name:     "encode error",
			url:      "/users/me",
			withAuth: true,
			useFlaky: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("GetUser", defaultUserID).
					Return(&User{ID: defaultUserID}, nil).
					Once()
			},
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusOK, recorder.Code)
				suite.Equal("Failed to encode response\n", recorder.Body.String())
			},
		},
	}

	suite.runHandlerTestCases(testCases,
		func(handler *userHandler, writer http.ResponseWriter, req *http.Request) {
			handler.HandleSelfUserGetRequest(writer, req)
		})
}

func (suite *UserHandlerTestSuite) TestHandleSelfUserPutRequest() {
	testCases := []userHandlerTestCase{
		{
			name:     "success",
			method:   http.MethodPut,
			url:      "/users/me",
			body:     `{"attributes":{"email":"alice@example.com"}}`,
			withAuth: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("UpdateUserAttributes", defaultUserID, mock.MatchedBy(func(attr json.RawMessage) bool {
						return string(attr) == defaultEmailAttributes
					})).
					Return(&User{
						ID:         defaultUserID,
						Attributes: json.RawMessage(defaultEmailAttributes),
						Type:       "employee",
					}, nil).
					Once()
			},
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusOK, recorder.Code)
				suite.Equal(serverconst.ContentTypeJSON, recorder.Header().Get(serverconst.ContentTypeHeaderName))

				var respUser User
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &respUser))
				suite.Equal(defaultUserID, respUser.ID)
				suite.JSONEq(defaultEmailAttributes, string(respUser.Attributes))
			},
		},
		{
			name:   "unauthorized",
			method: http.MethodPut,
			url:    "/users/me",
			body:   `{"attributes":{"email":"alice@example.com"}}`,
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusUnauthorized, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorAuthenticationFailed.Code, errResp.Code)
			},
			assertService: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.AssertNotCalled(suite.T(), "UpdateUserAttributes", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "invalid body",
			method:   http.MethodPut,
			url:      "/users/me",
			body:     `{"attributes":`,
			withAuth: true,
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusBadRequest, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorInvalidRequestFormat.Code, errResp.Code)
			},
			assertService: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.AssertNotCalled(suite.T(), "UpdateUserAttributes", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "missing attributes",
			method:   http.MethodPut,
			url:      "/users/me",
			body:     `{}`,
			withAuth: true,
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusBadRequest, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorInvalidRequestFormat.Code, errResp.Code)
			},
			assertService: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.AssertNotCalled(suite.T(), "UpdateUserAttributes", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "service error",
			method:   http.MethodPut,
			url:      "/users/me",
			body:     `{"attributes":{"email":"alice@example.com"}}`,
			withAuth: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("UpdateUserAttributes", defaultUserID, mock.AnythingOfType("json.RawMessage")).
					Return((*User)(nil), &ErrorUserNotFound).
					Once()
			},
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusNotFound, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorUserNotFound.Code, errResp.Code)
			},
		},
		{
			name:     "encode error",
			method:   http.MethodPut,
			url:      "/users/me",
			body:     `{"attributes":{"email":"alice@example.com"}}`,
			withAuth: true,
			useFlaky: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("UpdateUserAttributes", defaultUserID, mock.AnythingOfType("json.RawMessage")).
					Return(&User{ID: defaultUserID, Attributes: json.RawMessage(defaultEmailAttributes)}, nil).
					Once()
			},
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusOK, recorder.Code)
				suite.Equal("Failed to encode response\n", recorder.Body.String())
			},
		},
	}

	suite.runHandlerTestCases(testCases,
		func(handler *userHandler, writer http.ResponseWriter, req *http.Request) {
			handler.HandleSelfUserPutRequest(writer, req)
		})
}

func (suite *UserHandlerTestSuite) TestHandleSelfUserCredentialUpdateRequest() {
	testCases := []userHandlerTestCase{
		{
			name:     "success",
			method:   http.MethodPost,
			url:      "/users/me/update-credentials",
			body:     `{"attributes":{"password":"Secret123!"}}`,
			withAuth: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("UpdateUserCredentials", defaultUserID, mock.MatchedBy(func(attr json.RawMessage) bool {
						return string(attr) == defaultCredentialAttribute
					})).
					Return((*serviceerror.ServiceError)(nil)).
					Once()
			},
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusNoContent, recorder.Code)
				suite.Zero(recorder.Body.Len())
			},
		},
		{
			name:   "unauthorized",
			method: http.MethodPost,
			url:    "/users/me/update-credentials",
			body:   `{"attributes":{"password":"Secret123!"}}`,
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusUnauthorized, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorAuthenticationFailed.Code, errResp.Code)
			},
			assertService: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.AssertNotCalled(suite.T(), "UpdateUserCredentials", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "invalid body",
			method:   http.MethodPost,
			url:      "/users/me/update-credentials",
			body:     `{"attributes":`,
			withAuth: true,
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusBadRequest, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorInvalidRequestFormat.Code, errResp.Code)
			},
			assertService: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.AssertNotCalled(suite.T(), "UpdateUserCredentials", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "missing credentials",
			method:   http.MethodPost,
			url:      "/users/me/update-credentials",
			body:     `{}`,
			withAuth: true,
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusBadRequest, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorMissingCredentials.Code, errResp.Code)
			},
			assertService: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.AssertNotCalled(suite.T(), "UpdateUserCredentials", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "service error",
			method:   http.MethodPost,
			url:      "/users/me/update-credentials",
			body:     `{"attributes":{"password":"Secret123!"}}`,
			withAuth: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("UpdateUserCredentials", defaultUserID, mock.AnythingOfType("json.RawMessage")).
					Return(&ErrorInternalServerError).
					Once()
			},
			assert: func(recorder *httptest.ResponseRecorder) {
				suite.Equal(http.StatusInternalServerError, recorder.Code)

				var errResp apierror.ErrorResponse
				suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &errResp))
				suite.Equal(ErrorInternalServerError.Code, errResp.Code)
			},
		},
	}

	suite.runHandlerTestCases(testCases,
		func(handler *userHandler, writer http.ResponseWriter, req *http.Request) {
			handler.HandleSelfUserCredentialUpdateRequest(writer, req)
		})
}
