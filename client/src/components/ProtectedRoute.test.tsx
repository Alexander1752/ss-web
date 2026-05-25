import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import ProtectedRoute from './ProtectedRoute';

const authState = vi.hoisted(() => ({
  value: {
    isLoggedIn: false,
    loading: false,
    token: null,
    isAdmin: false,
    login: vi.fn(),
    logout: vi.fn(),
    register: vi.fn(),
  },
}));

vi.mock('../contexts/AuthContext', () => ({
  useAuth: () => authState.value,
}));

const renderPrivateRoute = () =>
  render(
    <MemoryRouter initialEntries={['/private']}>
      <Routes>
        <Route element={<ProtectedRoute authRequired={true} />}>
          <Route path="/private" element={<div>Private content</div>} />
        </Route>
        <Route path="/login" element={<div>Login page</div>} />
      </Routes>
    </MemoryRouter>,
  );

const renderPublicRoute = () =>
  render(
    <MemoryRouter initialEntries={['/login']}>
      <Routes>
        <Route element={<ProtectedRoute authRequired={false} />}>
          <Route path="/login" element={<div>Login page</div>} />
        </Route>
        <Route path="/" element={<div>Home page</div>} />
      </Routes>
    </MemoryRouter>,
  );

describe('ProtectedRoute', () => {
  beforeEach(() => {
    authState.value = {
      isLoggedIn: false,
      loading: false,
      token: null,
      isAdmin: false,
      login: vi.fn(),
      logout: vi.fn(),
      register: vi.fn(),
    };
  });

  afterEach(() => {
    cleanup();
  });

  it('shows a loading status while auth state is loading', () => {
    authState.value = { ...authState.value, loading: true };

    renderPrivateRoute();

    expect(screen.getByRole('status', { name: 'Loading authentication' })).toBeInTheDocument();
  });

  it('redirects unauthenticated users away from protected routes', () => {
    renderPrivateRoute();

    expect(screen.getByText('Login page')).toBeInTheDocument();
  });

  it('renders protected content for authenticated users', () => {
    authState.value = { ...authState.value, isLoggedIn: true };

    renderPrivateRoute();

    expect(screen.getByText('Private content')).toBeInTheDocument();
  });

  it('redirects authenticated users away from guest-only routes', () => {
    authState.value = { ...authState.value, isLoggedIn: true };

    renderPublicRoute();

    expect(screen.getByText('Home page')).toBeInTheDocument();
  });

  it('renders guest-only content for unauthenticated users', () => {
    renderPublicRoute();

    expect(screen.getByText('Login page')).toBeInTheDocument();
  });
});
