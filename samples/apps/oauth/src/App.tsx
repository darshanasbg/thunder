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

import { useEffect, useState, useMemo, useRef } from 'react';
import { Route, BrowserRouter as Router, Switch, useLocation } from 'react-router-dom';
import ErrorPage from './components/ErrorPage';
import HomePage from './components/HomePage';
import LoginPage from './components/LoginPage';
import { exchangeCodeForToken } from './services/authService';
import './App.css';

interface TokenErrorInterface {
  error: string;
  error_description: string;
}

const App = () => {
  const [token, setToken] = useState(null);
  const [error, setError] = useState<TokenErrorInterface | null>(null); // State to store token error

  const hasFetched = useRef(false);

  const urlParams = useMemo(() => new URLSearchParams(window.location.search), []);
  const location = useLocation(); // Get the current location

  useEffect(() => {
    const error = urlParams.get('error');
    const errorDescription = urlParams.get('error_description');

    if (error) {
      setError({
        error: error,
        error_description: errorDescription || 'No description available',
      });

      return;
    }
  }, [urlParams]);

  useEffect(() => {
    // Prevent double fetch calls
    if (hasFetched.current) return;
    hasFetched.current = true;

    const code = urlParams.get('code');
    
    if (code && !token) {
      exchangeCodeForToken(code)
        .then(response => {
          setToken(response.access_token);
        })
        .catch(error => {
          console.error('Error fetching access token:', error);
          setError(
            error.response && error.response.data
              ? error.response.data
              : { message: 'Unknown error' }
          );
        });
    }
  }, [token, urlParams]);

  return (
    <Switch>
      <Route path="/" exact key={location.key}>
        {error ? (
          <ErrorPage
            errorCode={error.error || 'Unknown Error'}
            errorMessage={error.error_description || 'No description available'}
          />
        ) : token ? (
          <HomePage token={token} />
        ) : (
          <LoginPage />
        )}
      </Route>
    </Switch>
  );
};

const AppWrapper = () => (
  <Router>
    <App />
  </Router>
);

export default AppWrapper;
