import {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react';
import { useNavigate } from 'react-router-dom';

import {
  ApiError,
  clearStoredToken,
  getStoredToken,
  setStoredToken,
  setUnauthorizedHandler,
} from '../api/client';
import * as systemApi from '../api/system';
import type { CurrentUserResponse } from '../types/system';

interface AuthContextValue {
  token: string | null;
  current: CurrentUserResponse | null;
  loading: boolean;
  sessionError: 'forbidden' | 'unavailable' | null;
  retrySession: () => void;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  hasMenu: (code: string) => boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const [token, setToken] = useState(() => getStoredToken());
  const [current, setCurrent] = useState<CurrentUserResponse | null>(null);
  const [loading, setLoading] = useState(Boolean(token));
  const [sessionError, setSessionError] = useState<AuthContextValue['sessionError']>(null);
  const [retryCount, setRetryCount] = useState(0);
  const recovery = useRef<AbortController | null>(null);

  useEffect(() => {
    return setUnauthorizedHandler(() => {
      recovery.current?.abort();
      setToken(null);
      setCurrent(null);
      setSessionError(null);
      setLoading(false);
      navigate('/login', { replace: true });
    });
  }, [navigate]);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }

    let cancelled = false;
    const controller = new AbortController();
    recovery.current = controller;
    setLoading(true);
    setCurrent(null);
    setSessionError(null);
    systemApi
      .currentUser(controller.signal)
      .then((result) => {
        if (!result?.user || typeof result.user.id !== 'number' || typeof result.user.username !== 'string'
          || !Array.isArray(result.roles) || !result.roles.every((role) => typeof role === 'string')
          || !Array.isArray(result.menuPermissions) || !result.menuPermissions.every((code) => typeof code === 'string')) {
          throw new ApiError('Invalid session response', 200, 'protocol');
        }
        if (!cancelled && !controller.signal.aborted) {
          setCurrent(result);
        }
      })
      .catch((error: unknown) => {
        if (!cancelled && !controller.signal.aborted) {
          setCurrent(null);
          setSessionError(error instanceof ApiError && error.status === 403 ? 'forbidden' : 'unavailable');
        }
      })
      .finally(() => {
        if (!cancelled && !controller.signal.aborted) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [token, retryCount]);

  function retrySession() {
    recovery.current?.abort();
    setLoading(true);
    setSessionError(null);
    setRetryCount((count) => count + 1);
  }

  async function handleLogin(username: string, password: string) {
    const result = await systemApi.login(username, password);
    recovery.current?.abort();
    setLoading(false);
    setSessionError(null);
    setStoredToken(result.token);
    setToken(result.token);
    setCurrent({
      user: result.user,
      roles: result.roles,
      menuPermissions: result.menuPermissions,
    });
  }

  async function handleLogout() {
    recovery.current?.abort();
    try {
      await systemApi.logout();
    } finally {
      clearStoredToken();
      setToken(null);
      setCurrent(null);
      setSessionError(null);
      setLoading(false);
      navigate('/login', { replace: true });
    }
  }

  return (
    <AuthContext.Provider
      value={{
        token,
        current,
        loading,
        sessionError,
        retrySession,
        login: handleLogin,
        logout: handleLogout,
        hasMenu: (code) => current?.menuPermissions.includes(code) ?? false,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error('useAuth must be used inside AuthProvider');
  }
  return value;
}
