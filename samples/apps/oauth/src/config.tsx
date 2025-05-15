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

const config = {
    authEndpoint: import.meta.env.VITE_REACT_APP_SERVER_AUTHENTICATION_ENDPOINT,
    tokenEndpoint: import.meta.env.VITE_REACT_APP_SERVER_TOKEN_ENDPOINT,
    clientId: import.meta.env.VITE_REACT_APP_CLIENT_ID,
    clientSecret: import.meta.env.VITE_REACT_APP_CLIENT_SECRET,
    redirectUri: import.meta.env.VITE_REACT_APP_REDIRECT_URI,
    scope: import.meta.env.VITE_REACT_APP_SCOPE
};

export default config;
