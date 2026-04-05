import React from 'react';
import { useNavigate } from 'react-router-dom';
import Button from '../button';

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

  const handleTitleClick = () => {
    navigate('/');
  };

  return (
    <nav className="fixed top-0 left-0 right-0 bg-sky-50 shadow-sm z-50">
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
          className="text-xl font-semibold text-sky-700 cursor-pointer hover:text-sky-800 transition-colors"
          onClick={handleTitleClick}
        >
          {title}
        </h1>
        
        <div className="flex-1 flex space-x-2 justify-end">
          {rightButtons.map((button, index) => (
            <Button 
              key={index}
              text={button.text}
              onClick={button.onClick}
              variant={button.variant || 'outline'}
              size="sm"
            />
          ))}
        </div>
      </div>
    </nav>
  );
};

export default Navbar; 