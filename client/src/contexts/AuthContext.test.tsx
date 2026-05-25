import { act } from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

type KeycloakMock = {
  init: ReturnType<typeof vi.fn>;
  login: ReturnType<typeof vi.fn>;
  logout: ReturnType<typeof vi.fn>;
  register: ReturnType<typeof vi.fn>;
  updateToken: ReturnType<typeof vi.fn>;
  token?: string;
  realmAccess?: { roles?: string[] };
  onTokenExpired?: () => void;
};

const createKeycloakMock = (): KeycloakMock => ({
  init: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  register: vi.fn(),
  updateToken: vi.fn(),
});

const keycloakMock = vi.hoisted(() => ({
  instance: {
    init: vi.fn(),
    login: vi.fn(),
    logout: vi.fn(),
    register: vi.fn(),
    updateToken: vi.fn(),
  } as KeycloakMock,
}));

vi.mock('keycloak-js', () => ({
  default: vi.fn(() => keycloakMock.instance),
}));

const renderAuthProbe = async () => {
  const { AuthProvider, useAuth } = await import('./AuthContext');

  const Probe = () => {
    const auth = useAuth();

    return (
      <div>
        <div data-testid="loading">{String(auth.loading)}</div>
        <div data-testid="logged-in">{String(auth.isLoggedIn)}</div>
        <div data-testid="token">{auth.token ?? ''}</div>
        <div data-testid="admin">{String(auth.isAdmin)}</div>
        <button type="button" onClick={auth.login}>
          login
        </button>
        <button type="button" onClick={auth.logout}>
          logout
        </button>
        <button type="button" onClick={auth.register}>
          register
        </button>
      </div>
    );
  };

  await act(async () => {
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );
  });
};

describe('AuthProvider', () => {
  beforeEach(() => {
    vi.resetModules();
    keycloakMock.instance = createKeycloakMock();
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it('initializes Keycloak with silent SSO and exposes unauthenticated state', async () => {
    keycloakMock.instance.init.mockResolvedValue(false);

    await renderAuthProbe();

    await waitFor(() => expect(screen.getByTestId('loading')).toHaveTextContent('false'));
    expect(keycloakMock.instance.init).toHaveBeenCalledWith({
      onLoad: 'check-sso',
      silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
      pkceMethod: 'S256',
    });
    expect(screen.getByTestId('logged-in')).toHaveTextContent('false');
    expect(screen.getByTestId('token')).toHaveTextContent('');
    expect(screen.getByTestId('admin')).toHaveTextContent('false');
  });

  it('stores token and admin role when Keycloak authenticates the user', async () => {
    keycloakMock.instance.token = 'access-token';
    keycloakMock.instance.realmAccess = { roles: ['user', 'admin'] };
    keycloakMock.instance.init.mockResolvedValue(true);

    await renderAuthProbe();

    await waitFor(() => expect(screen.getByTestId('loading')).toHaveTextContent('false'));
    expect(screen.getByTestId('logged-in')).toHaveTextContent('true');
    expect(screen.getByTestId('token')).toHaveTextContent('access-token');
    expect(screen.getByTestId('admin')).toHaveTextContent('true');
  });

  it('delegates login logout and register actions to Keycloak', async () => {
    keycloakMock.instance.init.mockResolvedValue(false);

    await renderAuthProbe();
    await waitFor(() => expect(screen.getByTestId('loading')).toHaveTextContent('false'));

    fireEvent.click(screen.getByRole('button', { name: 'login' }));
    fireEvent.click(screen.getByRole('button', { name: 'logout' }));
    fireEvent.click(screen.getByRole('button', { name: 'register' }));

    expect(keycloakMock.instance.login).toHaveBeenCalledTimes(1);
    expect(keycloakMock.instance.logout).toHaveBeenCalledWith({ redirectUri: window.location.origin });
    expect(keycloakMock.instance.register).toHaveBeenCalledWith({ redirectUri: window.location.origin });
  });

  it('refreshes the stored token when Keycloak reports token expiry', async () => {
    keycloakMock.instance.token = 'old-token';
    keycloakMock.instance.realmAccess = { roles: ['user'] };
    keycloakMock.instance.init.mockResolvedValue(true);
    keycloakMock.instance.updateToken.mockResolvedValue(true);

    await renderAuthProbe();
    await waitFor(() => expect(screen.getByTestId('token')).toHaveTextContent('old-token'));

    keycloakMock.instance.token = 'new-token';
    await act(async () => {
      keycloakMock.instance.onTokenExpired?.();
    });

    expect(keycloakMock.instance.updateToken).toHaveBeenCalledWith(30);
    await waitFor(() => expect(screen.getByTestId('token')).toHaveTextContent('new-token'));
  });
});
