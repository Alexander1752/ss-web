import { BrowserRouter, Routes, Route, Navigate, Outlet, useNavigate } from 'react-router-dom';
import Navbar from './components/navbar';
import HomePage from './pages/homePage';
import LoginPage from './pages/loginPage';
import RegisterPage from './pages/register';
import PhotosPage from './pages/photosPage';
import DevicesPage from './pages/devicesPage';
import StatisticsPage from './pages/statisticsPage';
import ProtectedRoute from './components/ProtectedRoute';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';

const Layout = () => {
  const navigate = useNavigate();
  const { isLoggedIn, logout } = useAuth();

  // Left-side buttons (only shown when logged in)
  const leftButtons = isLoggedIn
    ? [
      {
        text: 'Photos',
        variant: 'secondary' as const,
        onClick: () => navigate('/photos')
      },
      {
        text: 'Devices',
        variant: 'secondary' as const,
        onClick: () => navigate('/devices')
      },
      {
        text: 'Statistics',
        variant: 'secondary' as const,
        onClick: () => navigate('/statistics')
      }
    ]
    : [];

  // Right-side buttons (different based on login status)
  const rightButtons = isLoggedIn
    ? [
      {
        text: 'Logout',
        variant: 'outline' as const,
        onClick: () => {
          logout();
          navigate('/');
        }
      }
    ]
    : [
      {
        text: 'Login',
        variant: 'outline' as const,
        onClick: () => navigate('/login')
      },
      {
        text: 'Register',
        variant: 'primary' as const,
        onClick: () => navigate('/register')
      }
    ];

  return (
    <>
      <Navbar
        title="Security of Systems - First Force"
        leftButtons={leftButtons}
        rightButtons={rightButtons}
      />
      <div className="pt-16 px-4">
        <Outlet />
      </div>
    </>
  );
};

const App = () => {
  return (
    <BrowserRouter>
      <ThemeProvider>
      <AuthProvider>
        <Routes>
          {/* Common layout for all routes */}
          <Route element={<Layout />}>
            {/* Public routes - accessible to everyone */}
            <Route path="/" element={<HomePage />} />

            {/* Auth routes - only for non-authenticated users */}
            <Route element={<ProtectedRoute authRequired={false} />}>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
            </Route>

            {/* Protected routes - only for authenticated users */}
            <Route element={<ProtectedRoute authRequired={true} />}>
              <Route path="/photos" element={<PhotosPage />} />
              <Route path="/devices" element={<DevicesPage />} />
              <Route path="/statistics" element={<StatisticsPage />} />
            </Route>

            {/* Fallback route */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
};

export default App;
