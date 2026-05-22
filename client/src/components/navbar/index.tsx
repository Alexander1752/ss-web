import React from 'react';
import { useNavigate } from 'react-router-dom';
import Button from '../button';
import { useTheme } from '../../contexts/ThemeContext';

interface ButtonProps {
  text: string;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'outline';
  size?: 'sm' | 'md' | 'lg';
}

interface NavbarProps {
  title: string;
  leftButtons?: ButtonProps[];
  rightButtons?: ButtonProps[];
}

const ThemeToggle: React.FC = () => {
  const { theme, toggleTheme } = useTheme();
  const isDark = theme === 'dark';

  return (
    <button
      onClick={toggleTheme}
      aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      className={`relative ml-2 flex items-center h-8 w-16 rounded-full cursor-pointer transition-colors duration-300 focus:outline-none focus:ring-2 focus:ring-sky-400 focus:ring-offset-1 flex-shrink-0 ${
        isDark ? 'bg-indigo-900' : 'bg-amber-100'
      }`}
    >
      {/* Sliding thumb — rendered first so icons appear above it */}
      <span
        className={`absolute top-1 h-6 w-6 rounded-full shadow-md transition-all duration-300 pointer-events-none ${
          isDark ? 'translate-x-9 bg-indigo-500' : 'translate-x-1 bg-white'
        }`}
      />

      {/* Sun icon — left half */}
      <span className="relative z-10 flex items-center justify-center w-8 h-8 pointer-events-none">
        <svg xmlns="http://www.w3.org/2000/svg" className={`h-4 w-4 transition-colors duration-300 ${isDark ? 'text-indigo-400' : 'text-amber-500'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707M17.657 17.657l-.707-.707M6.343 6.343l-.707-.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      </span>

      {/* Moon icon — right half */}
      <span className="relative z-10 flex items-center justify-center w-8 h-8 pointer-events-none">
        <svg xmlns="http://www.w3.org/2000/svg" className={`h-4 w-4 transition-colors duration-300 ${isDark ? 'text-indigo-200' : 'text-slate-400'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
        </svg>
      </span>
    </button>
  );
};

const Navbar: React.FC<NavbarProps> = ({ 
  title, 
  leftButtons = [], 
  rightButtons = [] 
}) => {
  const navigate = useNavigate();

  const handleTitleClick = () => {
    navigate('/');
  };

  return (
    <nav className="fixed top-0 left-0 right-0 bg-sky-50 dark:bg-gray-900 shadow-sm dark:shadow-gray-800 z-50 transition-colors duration-300">
      <div className="px-4 py-3 flex items-center justify-between">
        <div className="flex-1 flex space-x-2">
          {leftButtons.map((button, index) => (
            <Button 
              key={index}
              text={button.text}
              onClick={button.onClick}
              variant={button.variant || 'outline'}
              size="sm"
            />
          ))}
        </div>
        
        <h1 
          className="text-xl font-semibold text-sky-700 dark:text-sky-300 cursor-pointer hover:text-sky-800 dark:hover:text-sky-200 transition-colors"
          onClick={handleTitleClick}
        >
          {title}
        </h1>
        
        <div className="flex-1 flex items-center space-x-2 justify-end">
          {rightButtons.map((button, index) => (
            <Button 
              key={index}
              text={button.text}
              onClick={button.onClick}
              variant={button.variant || 'outline'}
              size="sm"
            />
          ))}
          <ThemeToggle />
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
