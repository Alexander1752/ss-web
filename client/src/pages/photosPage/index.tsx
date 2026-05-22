import React, { useState, useEffect } from 'react';
import PhotoCard from '../../components/photosCards';
import { useAuth } from '../../contexts/AuthContext';
import { apiFetch } from '../../utils/api';

// Interface for device data
interface Device {
  id: string;
  device_id: string;
  device_name: string;
  device_status: string;
}

// Interface for photo data
interface Photo {
  id: string;
  timestamp: string;
  image_type: string;
  presigned_url: string;
  device_id: string;
  text: string;
}

// Interface for search parameters to store in localStorage
interface SearchParams {
  searchText: string;
  startDate: string;
  endDate: string;
  selectedDevice: string;
}

const STORAGE_KEY = 'photoSearchParams';

const PhotosPage: React.FC = () => {
  // Initialize state with values from localStorage if available
  const getStoredSearchParams = (): SearchParams => {
    const storedParams = localStorage.getItem(STORAGE_KEY);
    if (storedParams) {
      return JSON.parse(storedParams);
    }

    // Default values if nothing is stored
    const today = new Date();
    return {
      searchText: '',
      startDate: `${today.getFullYear()}-01-01`,
      endDate: today.toISOString().slice(0, 10),
      selectedDevice: 'all'
    };
  };

  const storedParams = getStoredSearchParams();

  const [searchText, setSearchText] = useState(storedParams.searchText);
  const [startDate, setStartDate] = useState(storedParams.startDate);
  const [endDate, setEndDate] = useState(storedParams.endDate);
  const [selectedDevice, setSelectedDevice] = useState(storedParams.selectedDevice);

  const [devices, setDevices] = useState<Device[]>([]);
  const [deviceError, setDeviceError] = useState<boolean>(false);
  const [deviceLoading, setDeviceLoading] = useState<boolean>(true);

  // States for photos
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [photosLoading, setPhotosLoading] = useState<boolean>(false);
  const [photosError, setPhotosError] = useState<string | null>(null);
  const [deleteAllConfirm, setDeleteAllConfirm] = useState(false);
  const [deletingAll, setDeletingAll] = useState(false);

  // Command state
  const [commandLoading, setCommandLoading] = useState(false);
  const [commandMessage, setCommandMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  const { token, isAdmin } = useAuth();

  // Clear command message after 3 seconds
  useEffect(() => {
    if (commandMessage) {
      const timer = setTimeout(() => {
        setCommandMessage(null);
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [commandMessage]);

  // Save search parameters to localStorage whenever they change
  useEffect(() => {
    const searchParams: SearchParams = {
      searchText,
      startDate,
      endDate,
      selectedDevice
    };

    localStorage.setItem(STORAGE_KEY, JSON.stringify(searchParams));
  }, [searchText, startDate, endDate, selectedDevice]);

  // Fetch devices from API
  useEffect(() => {
    const fetchDevices = async () => {
      setDeviceLoading(true);
      setDeviceError(false);

      try {
        const response = await apiFetch('/devices', {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error('Failed to fetch devices');
        }

        const data = await response.json();
        setDevices(data);

      } catch (error) {
        console.error('Error fetching devices:', error);
        setDeviceError(true);
      } finally {
        setDeviceLoading(false);
      }
    };

    fetchDevices();
  }, [token]);

  // Initial search on page load
  useEffect(() => {
    handleSearch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSearch = async () => {
    setPhotosLoading(true);
    setPhotosError(null);

    try {
      // Convert dates to Unix timestamps (seconds)
      const startTimestamp = Math.floor(new Date(startDate).getTime() / 1000);
      const endTimestamp = Math.floor(new Date(endDate).getTime() / 1000) + 86399; // End of day

      // Build query parameters
      const queryParams = new URLSearchParams();
      queryParams.append('start', startTimestamp.toString());
      queryParams.append('end', endTimestamp.toString());

      if (searchText.trim()) {
        queryParams.append('text', searchText.trim());
      }

      // Only add device_id if a specific device (not "all") is selected
      if (selectedDevice !== 'all') {
        queryParams.append('device_id', selectedDevice);
      }

      // Make API request
      const response = await apiFetch(`/photos?${queryParams.toString()}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch photos: ${response.status} ${response.statusText}`);
      }

      const data = await response.json();
      // Ensure data is an array, otherwise use empty array
      setPhotos(Array.isArray(data) ? data : []);

    } catch (error) {
      console.error('Error fetching photos:', error);
      setPhotosError((error as Error).message || 'Failed to load photos');
      setPhotos([]);
    } finally {
      setPhotosLoading(false);
    }
  };

  const handleDeletePhoto = async (photoId: string) => {
    try {
      const response = await apiFetch(`/photos/${photoId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to delete photo');
      }

      // Remove the deleted photo from state
      setPhotos(photos.filter(p => p.id !== photoId));
    } catch (error) {
      console.error('Error deleting photo:', error);
      alert('Failed to delete photo');
    }
  };

  const handleDeleteAllPhotos = async () => {
    setDeletingAll(true);
    try {
      const response = await apiFetch('/photos/all', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to delete all photos');
      }

      const result = await response.json();
      setPhotos([]);
      setDeleteAllConfirm(false);
      alert(`Deleted ${result.deleted} photos`);
    } catch (error) {
      console.error('Error deleting all photos:', error);
      alert('Failed to delete all photos');
    } finally {
      setDeletingAll(false);
    }
  };

  const sendCommand = async (command: 'CAPTURE' | 'START-LIVE' | 'STOP-LIVE') => {
    // Determine device ID
    // If a specific device is selected, use it. Otherwise, default to "camera_stream" (the unknown device default)
    const targetDeviceId = selectedDevice !== 'all' ? selectedDevice : 'camera_stream';

    setCommandLoading(true);
    setCommandMessage(null);

    try {
      const response = await apiFetch('/devices/command', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          device_id: targetDeviceId,
          command: command
        })
      });

      if (!response.ok) {
        throw new Error(`Failed to send command: ${response.status} ${response.statusText}`);
      }

      setCommandMessage({ type: 'success', text: `Command ${command} sent successfully` });
    } catch (error) {
      console.error(`Error sending command ${command}:`, error);
      setCommandMessage({ type: 'error', text: (error as Error).message || 'Failed to send command' });
    } finally {
      setCommandLoading(false);
    }
  };

  return (
    <div className="container mx-auto">
      <h1 className="text-2xl font-semibold text-sky-700 dark:text-sky-300 mb-6">Photos</h1>

      {/* Search and filter section */}
      <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm mb-6">
        <div className="flex flex-wrap items-end gap-4">
          {/* Text search */}
          <div className="flex-1 min-w-[200px]">
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Search Text
            </label>
            <input
              id="search"
              type="text"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              placeholder="Search text in photos..."
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent"
            />
          </div>

          {/* Start date */}
          <div>
            <label htmlFor="start-date" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Start Date
            </label>
            <input
              id="start-date"
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent"
            />
          </div>

          {/* End date */}
          <div>
            <label htmlFor="end-date" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              End Date
            </label>
            <input
              id="end-date"
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent"
            />
          </div>

          {/* Device dropdown - always shown, with or without device data */}
          {!deviceLoading && (
            <div>
              <label htmlFor="device" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Device
              </label>
              <select
                id="device"
                value={selectedDevice}
                onChange={(e) => setSelectedDevice(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent"
              >
                <option value="all">All</option>
                {!deviceError && devices.map(device => (
                  <option key={device.id} value={device.device_id}>
                    {device.device_id} - {device.device_name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Loading indicator for devices */}
          {deviceLoading && (
            <div className="flex items-end">
              <div className="h-10 flex items-center">
                <div className="animate-spin rounded-full h-5 w-5 border-t-2 border-b-2 border-sky-500 mr-2"></div>
                <span className="text-sm text-gray-500 dark:text-gray-400">Loading devices...</span>
              </div>
            </div>
          )}

          {/* Search button */}
          <div>
            <button
              onClick={handleSearch}
              disabled={photosLoading}
              className="px-4 py-2 bg-sky-600 text-white rounded-md hover:bg-sky-700 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {photosLoading ? 'Searching...' : 'Search'}
            </button>
          </div>

          {/* Separator */}
          <div className="w-px h-10 bg-gray-300 dark:bg-gray-600 mx-2 hidden md:block"></div>

          {/* ESP Camera Controls */}
          <div className="flex items-center gap-3 bg-gray-50 dark:bg-gray-700 p-2 rounded-md border border-gray-200 dark:border-gray-600">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300 whitespace-nowrap">ESP Camera:</span>

            <div className="flex gap-2">
              <button
                onClick={() => sendCommand('CAPTURE')}
                disabled={commandLoading}
                className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors disabled:opacity-50"
              >
                Capture
              </button>
              <button
                onClick={() => sendCommand('START-LIVE')}
                disabled={commandLoading}
                className="px-3 py-1.5 bg-emerald-600 text-white text-sm rounded hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition-colors disabled:opacity-50"
              >
                Start Live
              </button>
              <button
                onClick={() => sendCommand('STOP-LIVE')}
                disabled={commandLoading}
                className="px-3 py-1.5 bg-red-600 text-white text-sm rounded hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 transition-colors disabled:opacity-50"
              >
                Stop Live
              </button>
            </div>
          </div>

          {/* Delete All button */}
          <div>
            <button
              onClick={() => setDeleteAllConfirm(true)}
              className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors"
            >
              Delete All
            </button>
          </div>
        </div>
      </div>

      {/* Command Status Message */}
      {commandMessage && (
        <div className={`mb-6 p-4 rounded-md shadow-sm ${commandMessage.type === 'success'
          ? 'bg-green-50 dark:bg-green-900/30 text-green-800 dark:text-green-300 border border-green-200 dark:border-green-700'
          : 'bg-red-50 dark:bg-red-900/30 text-red-800 dark:text-red-300 border border-red-200 dark:border-red-700'
          }`}>
          <div className="flex items-center">
            <span className="text-lg mr-2">{commandMessage.type === 'success' ? '✅' : '❌'}</span>
            {commandMessage.text}
          </div>
        </div>
      )}

      {/* Delete All Confirmation Modal */}
      {deleteAllConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md shadow-xl">
            <h3 className="text-lg font-semibold text-red-600 dark:text-red-400 mb-4">⚠️ Delete All Photos</h3>
            <p className="text-gray-700 dark:text-gray-300 mb-6">Are you sure you want to delete ALL photos? This action cannot be undone.</p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setDeleteAllConfirm(false)}
                className="px-4 py-2 bg-gray-300 hover:bg-gray-400 dark:bg-gray-600 dark:hover:bg-gray-500 dark:text-gray-100 rounded-md transition-colors"
                disabled={deletingAll}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteAllPhotos}
                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
                disabled={deletingAll}
              >
                {deletingAll ? 'Deleting...' : 'Yes, Delete All'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Photos section with fixed height and scroll */}
      <div className="bg-gray-50 dark:bg-gray-900 p-4 rounded-lg shadow-sm overflow-y-auto max-h-[60vh]">
        {/* Loading state */}
        {photosLoading && (
          <div className="flex justify-center items-center h-40">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-sky-500"></div>
          </div>
        )}

        {/* Error state */}
        {!photosLoading && photosError && (
          <div className="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 text-red-700 dark:text-red-300 p-4 rounded-md">
            <p className="font-medium">Error loading photos</p>
            <p className="mt-1">{photosError}</p>
          </div>
        )}

        {/* Results grid */}
        {!photosLoading && !photosError && (
          <>
            {(photos || []).length === 0 ? (
              <div className="text-center text-gray-500 dark:text-gray-400 py-10">
                No photos found matching your search criteria
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {(photos || []).map(photo => (
                  <PhotoCard
                    key={photo.id}
                    photoId={photo.id}
                    imageUrl={photo.presigned_url}
                    extractedText={photo.text}
                    altText={`Photo from ${new Date(photo.timestamp).toLocaleDateString()}`}
                    isAdmin={isAdmin}
                    onDelete={handleDeletePhoto}
                  />
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default PhotosPage;