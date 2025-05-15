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

import axios from 'axios';
import config from '../config';

const { clientId, clientSecret, redirectUri, authEndpoint, tokenEndpoint } = config;

export const initiateAuth = () => {
    const state = Math.random().toString(36).substring(2, 15); // Generate a random state.
    const url = `${authEndpoint}?response_type=code&client_id=${clientId}&redirect_uri=${redirectUri}&scope=openid&state=${state}`;
    window.location.href = url;
};

export const exchangeCodeForToken = async (code: string) => {
    const headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
    };

    const data = new URLSearchParams();
    data.append('grant_type', 'authorization_code');
    data.append('redirect_uri', redirectUri);
    data.append('code', code);
    data.append('client_id', clientId);
    data.append('client_secret', clientSecret);

    try {
        const response = await axios.post(tokenEndpoint, data, {
            headers,
        });
        return response.data; // This will contain the access token
    } catch (error) {
        console.error('Error exchanging code for token:', error);
        throw error;
    }
};
