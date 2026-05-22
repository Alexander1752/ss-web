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

const Navbar: React.FC<NavbarProps> = ({ 
  title, 
  leftButtons = [], 
  rightButtons = [] 
}) => {
  const navigate = useNavigate();
  const { isDark, toggleTheme } = useTheme();

  const handleTitleClick = () => {
    navigate('/');
  };

  return (
    <nav className="fixed top-0 left-0 right-0 bg-sky-50 dark:bg-gray-900 shadow-sm dark:shadow-gray-800 z-50 transition-colors duration-200">
      <div className="container mx-auto px-4 py-3 flex items-center justify-between">
        <div className="flex space-x-2">
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
          className="text-xl font-semibold text-sky-700 dark:text-sky-400 cursor-pointer hover:text-sky-800 dark:hover:text-sky-300 transition-colors"
          onClick={handleTitleClick}
        >
          {title}
        </h1>
        
        <div className="flex items-center space-x-2">
          {rightButtons.map((button, index) => (
            <Button 
              key={index}
              text={button.text}
              onClick={button.onClick}
              variant={button.variant || 'outline'}
              size="sm"
            />
          ))}

          {/* Dark mode slider toggle */}
          <button
            onClick={toggleTheme}
            aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
            role="switch"
            aria-checked={isDark}
            className={`ml-2 relative inline-flex items-center w-14 h-7 rounded-full transition-colors duration-300 focus:outline-none focus:ring-2 focus:ring-sky-400 focus:ring-offset-2 dark:focus:ring-offset-gray-900 ${
              isDark ? 'bg-indigo-900' : 'bg-gray-300'
            }`}
          >
            {/* Sun icon — left side */}
            <span className="absolute left-1 flex items-center justify-center w-5 h-5 text-yellow-400 pointer-events-none">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-3.5 h-3.5">
                <path d="M12 2.25a.75.75 0 01.75.75v2.25a.75.75 0 01-1.5 0V3a.75.75 0 01.75-.75zM7.5 12a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM18.894 6.166a.75.75 0 00-1.06-1.06l-1.591 1.59a.75.75 0 101.06 1.061l1.591-1.59zM21.75 12a.75.75 0 01-.75.75h-2.25a.75.75 0 010-1.5H21a.75.75 0 01.75.75zM17.834 18.894a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 10-1.061 1.06l1.59 1.591zM12 18a.75.75 0 01.75.75V21a.75.75 0 01-1.5 0v-2.25A.75.75 0 0112 18zM7.758 17.303a.75.75 0 00-1.061-1.06l-1.591 1.59a.75.75 0 001.06 1.061l1.591-1.59zM6 12a.75.75 0 01-.75.75H3a.75.75 0 010-1.5h2.25A.75.75 0 016 12zM6.697 7.757a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 00-1.061 1.06l1.59 1.591z" />
              </svg>
            </span>

            {/* Moon icon — right side */}
            <span className="absolute right-1 flex items-center justify-center w-5 h-5 text-sky-200 pointer-events-none">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-3.5 h-3.5">
                <path fillRule="evenodd" d="M9.528 1.718a.75.75 0 01.162.819A8.97 8.97 0 009 6a9 9 0 009 9 8.97 8.97 0 003.463-.69.75.75 0 01.981.98 10.503 10.503 0 01-9.694 6.46c-5.799 0-10.5-4.701-10.5-10.5 0-4.368 2.667-8.112 6.46-9.694a.75.75 0 01.818.162z" clipRule="evenodd" />
              </svg>
            </span>

            {/* Sliding knob */}
            <span
              className={`absolute top-0.5 w-6 h-6 bg-white rounded-full shadow-md transform transition-transform duration-300 ${
                isDark ? 'translate-x-7' : 'translate-x-0.5'
              }`}
            />
          </button>
        </div>
      </div>
    </nav>
  );
};

export default Navbar; 