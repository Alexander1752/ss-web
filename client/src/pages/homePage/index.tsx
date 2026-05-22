import React from 'react';
import logo from '../../assets/logo.svg';

const HomePage: React.FC = () => {
  return (
    <div className="container mx-auto px-4 flex flex-col items-center justify-center min-h-[80vh]">
      <div className="flex flex-col items-center text-center max-w-3xl">
        <img src={logo} alt="Security of Systems - First Force Logo" className="w-48 h-48 mb-8" />
        
        <h1 className="text-4xl font-bold text-sky-700 dark:text-sky-400 mb-4">
          Security of Systems - First Force
        </h1>
        
        <p className="text-lg text-gray-600 dark:text-gray-300 mb-8">
          An advanced system for capturing images, extracting text content, and providing powerful search capabilities.
        </p>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 w-full mt-8">
          <div className="bg-sky-50 dark:bg-gray-800 p-6 rounded-lg shadow-sm flex flex-col items-center">
            <div className="rounded-full bg-sky-100 dark:bg-gray-700 p-4 mb-4">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-sky-700 dark:text-sky-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M4 5a2 2 0 00-2 2v8a2 2 0 002 2h12a2 2 0 002-2V7a2 2 0 00-2-2h-1.586a1 1 0 01-.707-.293l-1.121-1.121A2 2 0 0011.172 3H8.828a2 2 0 00-1.414.586L6.293 4.707A1 1 0 015.586 5H4zm6 9a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-sky-700 dark:text-sky-400 mb-2">Capture</h3>
            <p className="text-gray-600 dark:text-gray-300 text-center">Take photos from your mobile device with our dedicated Android app</p>
          </div>
          
          <div className="bg-sky-50 dark:bg-gray-800 p-6 rounded-lg shadow-sm flex flex-col items-center">
            <div className="rounded-full bg-sky-100 dark:bg-gray-700 p-4 mb-4">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-sky-700 dark:text-sky-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clipRule="evenodd" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-sky-700 dark:text-sky-400 mb-2">Extract</h3>
            <p className="text-gray-600 dark:text-gray-300 text-center">Our server automatically extracts text from your images</p>
          </div>
          
          <div className="bg-sky-50 dark:bg-gray-800 p-6 rounded-lg shadow-sm flex flex-col items-center">
            <div className="rounded-full bg-sky-100 dark:bg-gray-700 p-4 mb-4">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-sky-700 dark:text-sky-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-sky-700 dark:text-sky-400 mb-2">Search</h3>
            <p className="text-gray-600 dark:text-gray-300 text-center">Easily find photos by searching the extracted text content</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HomePage; 