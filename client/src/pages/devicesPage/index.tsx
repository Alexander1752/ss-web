import React, { useState, useEffect } from 'react';
import DeviceCard from '../../components/devicesCards';
import { useAuth } from '../../contexts/AuthContext';
import { apiFetch } from '../../utils/api';

// Interface for device data
interface Device {
  id: string;
  device_id: string;
  device_name: string;
  device_status: string;
  ip_address?: string;
  port?: string;
  last_seen?: string;
}

// Interface for tracking device action states
interface DeviceActionState {
  [deviceId: string]: {
    loading: boolean;
    success: boolean;
    error: string | null;
    lastMode: string | null;
  };
}

const DevicesPage: React.FC = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [brokerInfo, setBrokerInfo] = useState<{ ip: string; port: string }>({ ip: '', port: '' });

  // Track status of actions for individual devices
  const [deviceActionStates, setDeviceActionStates] = useState<DeviceActionState>({});

  const { token } = useAuth();

  // Fetch broker info
  useEffect(() => {
    const fetchBrokerInfo = async () => {
      try {
        const response = await apiFetch('/broker-info');
        if (response.ok) {
          const data = await response.json();
          setBrokerInfo(data);
        }
      } catch (err) {
        console.error('Failed to fetch broker info:', err);
      }
    };
    fetchBrokerInfo();
  }, []);

  // Fetch devices from API
  useEffect(() => {
    const fetchDevices = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await apiFetch('/devices', {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch devices: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();

        // Filter devices to only include active ones
        const activeDevices = Array.isArray(data)
          ? data.filter(device => device.device_status === "active")
          : [];

        setDevices(activeDevices);

        // Initialize action states for all active devices
        const initialActionStates: DeviceActionState = {};
        activeDevices.forEach(device => {
          initialActionStates[device.device_id] = {
            loading: false,
            success: false,
            error: null,
            lastMode: null,
          };
        });
        setDeviceActionStates(initialActionStates);

      } catch (error) {
        console.error('Error fetching devices:', error);
        setError((error as Error).message || 'Failed to load devices');
        setDevices([]);
      } finally {
        setLoading(false);
      }
    };

    fetchDevices();
  }, [token]);

  // Clear success/error messages after delay
  useEffect(() => {
    const timeout = setTimeout(() => {
      // Clear success/error messages but keep the last mode for displaying status
      setDeviceActionStates(prevStates => {
        const newStates = { ...prevStates };
        Object.keys(newStates).forEach(deviceId => {
          if (newStates[deviceId].success || newStates[deviceId].error) {
            newStates[deviceId] = {
              ...newStates[deviceId],
              success: false,
              error: null,
            };
          }
        });
        return newStates;
      });
    }, 3000);

    return () => clearTimeout(timeout);
  }, [deviceActionStates]);

  const sendCommand = async (deviceId: string, command: 'CAPTURE' | 'START-LIVE' | 'STOP-LIVE') => {
    // Set loading state
    setDeviceActionStates(prev => ({
      ...prev,
      [deviceId]: {
        ...prev[deviceId],
        loading: true,
        success: false,
        error: null,
      }
    }));

    try {
      const response = await apiFetch('/devices/command', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          device_id: deviceId,
          command: command
        })
      });

      if (!response.ok) {
        throw new Error(`Failed to send command: ${response.status} ${response.statusText}`);
      }

      // Set success state
      setDeviceActionStates(prev => ({
        ...prev,
        [deviceId]: {
          ...prev[deviceId],
          loading: false,
          success: true,
          error: null,
        }
      }));

    } catch (error) {
      console.error(`Error sending command ${command} to device ${deviceId}:`, error);

      // Set error state
      setDeviceActionStates(prev => ({
        ...prev,
        [deviceId]: {
          ...prev[deviceId],
          loading: false,
          success: false,
          error: (error as Error).message || `Failed to send command`,
        }
      }));
    }
  };



  return (
    <div className="container mx-auto">
      <h1 className="text-2xl font-semibold text-sky-700 dark:text-sky-400 mb-6">Devices</h1>

      {/* MQTT Broker Connection Info */}
      <div className="bg-gradient-to-r from-sky-500 to-blue-600 text-white p-4 rounded-lg shadow-md mb-6">
        <div className="flex items-center gap-3">
          <div className="bg-white/20 rounded-full p-2">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
            </svg>
          </div>
          <div>
            <p className="text-sm opacity-80">MQTT Broker Connection</p>
            <p className="text-xl font-bold font-mono">
              {brokerInfo.ip ? `${brokerInfo.ip}:${brokerInfo.port}` : 'Loading...'}
            </p>
          </div>
          <button
            onClick={() => navigator.clipboard.writeText(`${brokerInfo.ip}:${brokerInfo.port}`)}
            className="ml-auto bg-white/20 hover:bg-white/30 px-3 py-1 rounded-md text-sm transition-colors"
            disabled={!brokerInfo.ip}
          >
            Copy
          </button>
        </div>
      </div>

      {/* Devices grid with fixed height and scroll */}
      <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg shadow-sm overflow-y-auto max-h-[60vh]">
        {/* Loading state */}
        {loading && (
          <div className="flex justify-center items-center h-40">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-sky-500"></div>
          </div>
        )}

        {/* Error state */}
        {!loading && error && (
          <div className="bg-red-50 dark:bg-red-900/40 border border-red-200 dark:border-red-700 text-red-700 dark:text-red-300 p-4 rounded-md">
            <p className="font-medium">You don't have permission to view connected devices</p>
            <p className="mt-1 text-sm">Please contact your administrator for access</p>
          </div>
        )}

        {/* Success state - display devices */}
        {!loading && !error && (
          <>
            {devices.length === 0 ? (
              <div className="text-center text-gray-500 dark:text-gray-400 py-10">
                No devices found
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {devices.map(device => {
                  const actionState = deviceActionStates[device.device_id] || {
                    loading: false,
                    success: false,
                    error: null,
                    lastMode: null,
                  };

                  return (
                    <div key={device.id} className="relative">
                      <DeviceCard
                        deviceId={device.device_id}
                        deviceName={device.device_name}
                        onCaptureClick={() => sendCommand(device.device_id, 'CAPTURE')}
                        onStartLiveClick={() => sendCommand(device.device_id, 'START-LIVE')}
                        onStopLiveClick={() => sendCommand(device.device_id, 'STOP-LIVE')}
                      />

                      {/* Loading overlay */}
                      {actionState.loading && (
                        <div className="absolute inset-0 bg-white dark:bg-gray-900 bg-opacity-70 dark:bg-opacity-70 flex items-center justify-center rounded-lg">
                          <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-sky-500"></div>
                        </div>
                      )}

                      {/* Success notification */}
                      {actionState.success && (
                        <div className="absolute top-0 right-0 left-0 bg-green-100 dark:bg-green-900/60 text-green-800 dark:text-green-300 text-sm p-2 rounded-t-lg text-center">
                          Command sent successfully
                        </div>
                      )}

                      {/* Error notification */}
                      {actionState.error && (
                        <div className="absolute top-0 right-0 left-0 bg-red-100 dark:bg-red-900/60 text-red-800 dark:text-red-300 text-sm p-2 rounded-t-lg text-center">
                          {actionState.error}
                        </div>
                      )}

                      {/* Mode indicator */}

                    </div>
                  );
                })}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default DevicesPage; 