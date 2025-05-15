/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package utils provides utility functions for authentication related operations.
package utils

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// GetLoginPageRedirectURI returns the login page URL with the given query parameters.
func GetLoginPageRedirectURI(queryParams map[string]string) (string, error) {
	gateClientConfig := config.GetThunderRuntime().Config.GateClient
	loginPageURL := (&url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", gateClientConfig.Hostname, gateClientConfig.Port),
		Path:   "login",
	}).String()

	return utils.GetURIWithQueryParams(loginPageURL, queryParams)
}

// GetErrorPageURL returns the server error page URL.
func GetErrorPageURL(queryParams map[string]string) (string, error) {
	gateClientConfig := config.GetThunderRuntime().Config.GateClient
	errorPageURL := (&url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", gateClientConfig.Hostname, gateClientConfig.Port),
		Path:   "error",
	}).String()

	return utils.GetURIWithQueryParams(errorPageURL, queryParams)
}

// RedirectToErrorPage redirects the user to the error page with the given error details.
func RedirectToErrorPage(w http.ResponseWriter, r *http.Request, code, msg string) {
	if w == nil || r == nil {
		log.GetLogger().Error("Response writer or request is nil. Cannot redirect to error page.")
		return
	}

	queryParams := map[string]string{
		constants.OAuthErrorCode:    code,
		constants.OAuthErrorMessage: msg,
	}
	redirectURL, err := GetErrorPageURL(queryParams)
	if err != nil {
		log.GetLogger().Error("Failed to construct error page URL: " + err.Error())
		return
	}
	log.GetLogger().Info("Redirecting to error page: " + redirectURL)

	// Redirect with the request object.
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
