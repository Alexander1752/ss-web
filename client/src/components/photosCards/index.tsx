import React, { useState } from 'react';
import fallbackImage from '../../assets/photo-fallback.svg';

interface PhotoCardProps {
  photoId: string;
  imageUrl: string;
  altText?: string;
  extractedText?: string;
  isAdmin?: boolean;
  onDelete?: (photoId: string) => void;
}

const PhotoCard: React.FC<PhotoCardProps> = ({
  photoId,
  imageUrl,
  altText = 'Photo',
  extractedText = '',
  isAdmin = false,
  onDelete
}) => {
  const [isZoomed, setIsZoomed] = useState(false);
  const [imageError, setImageError] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleImageError = () => {
    setImageError(true);
  };

  const toggleZoom = () => {
    setIsZoomed(!isZoomed);
  };

  // Handle click outside the zoomed image to close it
  const handleModalClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      setIsZoomed(false);
    }
  };

  const handleDeleteClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowDeleteConfirm(true);
  };

  const handleConfirmDelete = async () => {
    setIsDeleting(true);
    if (onDelete) {
      await onDelete(photoId);
    }
    setShowDeleteConfirm(false);
    setIsDeleting(false);
  };

  return (
    <>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden transition-all hover:shadow-lg relative">
        <div className="relative h-48 cursor-pointer" onClick={toggleZoom}>
          <img
            src={imageError ? fallbackImage : imageUrl}
            alt={altText}
            onError={handleImageError}
            className="w-full h-full object-cover"
          />
          {isAdmin && (
            <button
              onClick={handleDeleteClick}
              className="absolute top-2 right-2 bg-red-500 hover:bg-red-600 text-white rounded-full p-2 shadow-lg transition-all duration-200 opacity-80 hover:opacity-100"
              title="Delete photo"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          )}
        </div>

        {extractedText && (
          <div className="p-3 border-t border-gray-100 dark:border-gray-700">
            <p className="text-sm text-gray-600 dark:text-gray-300 truncate">{extractedText}</p>
          </div>
        )}

        {/* Delete confirmation dialog */}
        {showDeleteConfirm && (
          <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <div className="bg-white dark:bg-gray-800 rounded-lg p-4 m-4 shadow-xl">
              <p className="text-gray-800 dark:text-gray-100 mb-4">Delete this photo?</p>
              <div className="flex gap-2 justify-center">
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  className="px-4 py-2 bg-gray-300 hover:bg-gray-400 dark:bg-gray-600 dark:hover:bg-gray-500 dark:text-gray-100 rounded-md transition-colors"
                  disabled={isDeleting}
                >
                  Cancel
                </button>
                <button
                  onClick={handleConfirmDelete}
                  className="px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-md transition-colors"
                  disabled={isDeleting}
                >
                  {isDeleting ? 'Deleting...' : 'Delete'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Zoom modal */}
      {isZoomed && (
        <div
          style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.7)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50, padding: '1rem' }}
          onClick={handleModalClick}
        >
          <div
            className="bg-white dark:bg-gray-900"
            style={{ borderRadius: '12px', width: '100%', maxWidth: '56rem', height: 'calc(100vh - 2rem)', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}
          >
            {/* Header */}
            <div style={{ flexShrink: 0, display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem 1.25rem', borderBottom: '1px solid rgba(128,128,128,0.3)' }}>
              <span className="text-gray-900 dark:text-gray-100 font-medium truncate" style={{ paddingRight: '1rem' }}>{altText}</span>
              <button onClick={toggleZoom} className="flex-shrink-0 text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* Scrollable body */}
            <div style={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
              <div style={{ padding: '1rem' }}>
                <img
                  src={imageError ? fallbackImage : imageUrl}
                  alt={altText}
                  onError={handleImageError}
                  style={{ maxWidth: '100%', display: 'block', margin: '0 auto', borderRadius: '8px' }}
                />
              </div>
              {extractedText && (
                <div className="bg-gray-50 dark:bg-gray-800" style={{ padding: '1.5rem', borderTop: '1px solid rgba(128,128,128,0.2)' }}>
                  <h3 className="text-gray-500 dark:text-gray-400" style={{ fontSize: '0.75rem', fontWeight: 600, marginBottom: '0.5rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Extracted Text</h3>
                  <p className="text-gray-800 dark:text-gray-200" style={{ whiteSpace: 'pre-wrap', lineHeight: 1.6 }}>{extractedText}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default PhotoCard;