import React, { useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';

const RegisterPage: React.FC = () => {
  const { register } = useAuth();

  useEffect(() => {
    register();
  }, []);

  return (
    <div className="flex justify-center items-center min-h-[80vh]">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-sky-500 mx-auto mb-4"></div>
        <p className="text-gray-600 dark:text-gray-400">Redirecting to registration...</p>
      </div>
    </div>
  );
};

export default RegisterPage;
