'use client';

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

import Alert from '@oxygen-ui/react/src/components/Alert/Alert';
import Box from '@oxygen-ui/react/src/components/Box/Box';
import Button from '@oxygen-ui/react/src/components/Button/Button';
import Checkbox from '@oxygen-ui/react/src/components/Checkbox/Checkbox';
import Divider from '@oxygen-ui/react/src/components/Divider/Divider';
import FormControlLabel from '@oxygen-ui/react/src/components/FormControlLabel/FormControlLabel';
import OutlinedInput from '@oxygen-ui/react/src/components/OutlinedInput/OutlinedInput';
import Link from '@oxygen-ui/react/src/components/Link/Link';
import InputLabel from '@oxygen-ui/react/src/components/InputLabel/InputLabel';
import Typography from '@oxygen-ui/react/src/components/Typography/Typography';
import AppConfig from "@/configs/app.json";
import GoogleIcon from '@/images/google-icon';
import GitHubIcon from '@/images/github-icon';
import React, { useState, useEffect } from 'react';

const LoginPageContent = function () {
  const [sessionDataKey, setSessionDataKey] = useState('');
  const [insecureWarning, setInsecureWarning] = useState(false);

  const [showSignUp] = useState(false);
  const [showGoogleButton] = useState(false);
  const [showGitHubButton] = useState(false);
  const [showRememberMe] = useState(false);
  const [showForgotPassword] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);

    setSessionDataKey(params.get('sessionDataKey') || '');
    setInsecureWarning(params.get('showInsecureWarning') === 'true');
  }, []);

  return (
    <form method="POST" action={ AppConfig.authenticationEndpoint }>

      <Box sx={{ mb: 4 }}>
        <Typography variant="h5" gutterBottom>
          Login to Account
        </Typography>

        { showSignUp &&
          <Typography>
            Don&apos;t have an account{' '}
            <Link href="">Sign up!</Link>
          </Typography>
        }
      </Box>

      {insecureWarning && (
        <Alert
          severity="warning"
          sx={ { my: 2 }}>
            You are about to access a non-secure site. Proceed with caution!
        </Alert>
      )}

      { (showGoogleButton || showGitHubButton) &&
        <>
          <Box>
            { showGoogleButton &&
              <Button
                  fullWidth
                  variant='contained'
                  startIcon={<GoogleIcon />}
                  color='secondary'
                  sx={{ my: 1 }}
                >
                  Continue with Google
              </Button>
            }
            { showGitHubButton &&
              <Button
                  fullWidth
                  variant='contained'
                  startIcon={<GitHubIcon />}
                  color='secondary'
                  sx={{ my: 1 }}
                >
                  Continue with GitHub
              </Button>
            }
          </Box>

          <Divider sx={{ my: 3 }}>or</Divider>
        </>
      }

      <Box display="flex" flexDirection="column" gap={2}>
        <Box display="flex" flexDirection="column" gap={0.5}>
          <InputLabel htmlFor="username">Username</InputLabel>
          <OutlinedInput type='text' id="username" name="username" placeholder="Enter your username" size="small" required />
        </Box>
        <Box display="flex" flexDirection="column" gap={0.5}>
          <InputLabel htmlFor="password">Password</InputLabel>
          <OutlinedInput type='password' id="password" name="password" placeholder="Enter your password" size="small" required />
        </Box>
        { (showRememberMe || showForgotPassword) &&
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            {showRememberMe && (
              <FormControlLabel
                control={<Checkbox name="remember-me-checkbox" />}
                label="Remember me"
              />
            )}
            {showForgotPassword && (
              <Link href="">
                Forgot your password?
              </Link>
            )}
          </Box>
        }
        <input type="hidden" id="sessionDataKey" name="sessionDataKey" value={sessionDataKey} />
        <Button variant="contained" color="primary" type="submit" fullWidth sx={{ mt: 2 }}>
          Sign In
        </Button>
      </Box>
    </form>
  );
}

export default LoginPageContent;
