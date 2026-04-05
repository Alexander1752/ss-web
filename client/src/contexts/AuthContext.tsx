import React, { createContext, useState, useContext, useEffect, useRef } from 'react';
import Keycloak from 'keycloak-js';

interface AuthContextType {
  isLoggedIn: boolean;
  token: string | null;
  loading: boolean;
  isAdmin: boolean;
  login: () => void;
  logout: () => void;
  register: () => void;
}

const AuthContext = createContext<AuthContextType>({
  isLoggedIn: false,
  token: null,
  loading: true,
  isAdmin: false,
  login: () => {},
  logout: () => {},
  register: () => {},
});

export const useAuth = () => useContext(AuthContext);

const keycloak = new Keycloak({
  url: import.meta.env.VITE_KEYCLOAK_URL ?? 'http://localhost:8081',
  realm: import.meta.env.VITE_KEYCLOAK_REALM ?? 'ss-project',
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID ?? 'ss-web',
});

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [isAdmin, setIsAdmin] = useState(false);
  const initialized = useRef(false);

  useEffect(() => {
    if (initialized.current) return;
    initialized.current = true;

    keycloak
      .init({
        onLoad: 'check-sso',
        silentCheckSsoRedirectUri: window.location.origin + '/silent-check-sso.html',
        pkceMethod: 'S256',
      })
      .then((authenticated) => {
        setIsLoggedIn(authenticated);
        if (authenticated && keycloak.token) {
          setToken(keycloak.token);
          const roles = keycloak.realmAccess?.roles ?? [];
          setIsAdmin(roles.includes('admin'));
        }
      })
      .finally(() => setLoading(false));

    keycloak.onTokenExpired = () => {
      keycloak.updateToken(30).then((refreshed) => {
        if (refreshed && keycloak.token) {
          setToken(keycloak.token);
        }
      });
    };
  }, []);

  const login = () => keycloak.login();
  const logout = () => keycloak.logout({ redirectUri: window.location.origin });
  const register = () => keycloak.register({ redirectUri: window.location.origin });

  return (
    <AuthContext.Provider value={{ isLoggedIn, token, loading, isAdmin, login, logout, register }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContext;
