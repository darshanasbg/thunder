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

import { decodeJwt } from '../services/jwtService';

const HomePage = ({ token }: { token: string }) => {
    const handleBackToLogin = () => {
        window.location.href = '/'; // Redirect to the login page and refresh
    };

    // Decode the token if available.
    let decodedToken = null;

    if (token) {
        decodedToken = decodeJwt(token);
    }

    return (
        <div className="home-container">
            {token ? (
                <div className="token-container">
                    <h2>Access Token:</h2>
                    <pre style={{ margin: 0 }}>
                        <code>{token}</code>
                    </pre>
                    <hr />
                    {decodedToken && (
                        <div>
                            <h2>Decoded Token:</h2>
                            <div className="decoded-token-container">
                                <div className="decoded-token-section">
                                    <h3>Header:</h3>
                                    <pre className="decoded-token">
                                        {JSON.stringify(decodedToken.header, null, 2)}
                                    </pre>
                                    <h3>Payload:</h3>
                                    <pre className="decoded-token">
                                        {JSON.stringify(decodedToken.payload, null, 2)}
                                    </pre>
                                </div>
                                <div className="decoded-token-section">
                                    <h3>Signature:</h3>
                                    <pre className="decoded-token">
                                        <code>{decodedToken.signature}</code>
                                    </pre>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            ) : (
                <p>No token available. Please log in.</p>
            )}
            <button onClick={handleBackToLogin} className="back-to-login-button">
                Back to Login
            </button>
        </div>
    );
};

export default HomePage;
