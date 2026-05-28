import {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useState,
} from 'react';
import { useNavigate } from 'react-router-dom';

import {
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

  useEffect(() => {
    setUnauthorizedHandler(() => {
      setToken(null);
      setCurrent(null);
      navigate('/login', { replace: true });
    });
  }, [navigate]);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    systemApi
      .currentUser()
      .then((result) => {
        if (!cancelled) {
          setCurrent(result);
        }
      })
      .catch(() => {
        if (!cancelled) {
          clearStoredToken();
          setToken(null);
          setCurrent(null);
          navigate('/login', { replace: true });
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [token]);

  async function handleLogin(username: string, password: string) {
    const result = await systemApi.login(username, password);
    setStoredToken(result.token);
    setToken(result.token);
    setCurrent({
      user: result.user,
      roles: result.roles,
      menuPermissions: result.menuPermissions,
    });
  }

  async function handleLogout() {
    try {
      await systemApi.logout();
    } finally {
      clearStoredToken();
      setToken(null);
      setCurrent(null);
      navigate('/login', { replace: true });
    }
  }

  return (
    <AuthContext.Provider
      value={{
        token,
        current,
        loading,
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
