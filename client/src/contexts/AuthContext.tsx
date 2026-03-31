// TODO: Implement authentication - See docs/AUTH_IMPLEMENTATION.md
// This file currently bypasses authentication. To implement real auth:
// 1. Remove the auto-login in useEffect
// 2. Validate JWT tokens from the backend
// 3. Store and use real tokens for API requests

import React, { createContext, useState, useContext, useEffect } from 'react';

interface AuthContextType {
  isLoggedIn: boolean;
  token: string | null;
  loading: boolean;
  isAdmin: boolean;
  login: (token: string) => void;
  logout: () => void;
}



const AuthContext = createContext<AuthContextType>({
  isLoggedIn: false,
  token: null,
  loading: true,
  isAdmin: true,
  login: () => { },
  logout: () => { },
});

export const useAuth = () => useContext(AuthContext);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(true);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    if (storedToken) {
      setToken(storedToken);
      setIsLoggedIn(true);
      setIsAdmin(true);
    }
    setLoading(false);
  }, []);

  const login = (newToken: string) => {
    localStorage.setItem('token', newToken);
    setToken(newToken);
    setIsLoggedIn(true);

    setIsAdmin(true);
  };

  const logout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setIsLoggedIn(false);
    setIsAdmin(false);
  };

  const value = {
    token,
    isLoggedIn,
    loading,
    isAdmin,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export default AuthContext; 