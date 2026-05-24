import React, { useState, useEffect, useRef } from "react";
import PhotoCard from "../../components/photosCards";
import { useAuth } from "../../contexts/AuthContext";
import { apiFetch } from "../../utils/api";

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

const STORAGE_KEY = "photoSearchParams";

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
      searchText: "",
      startDate: `${today.getFullYear()}-01-01`,
      endDate: today.toISOString().slice(0, 10),
      selectedDevice: "all",
    };
  };

  const storedParams = getStoredSearchParams();

  const [searchText, setSearchText] = useState(storedParams.searchText);
  const [startDate, setStartDate] = useState(storedParams.startDate);
  const [endDate, setEndDate] = useState(storedParams.endDate);
  const [selectedDevice, setSelectedDevice] = useState(
    storedParams.selectedDevice,
  );

  const [devices, setDevices] = useState<Device[]>([]);
  const [deviceError, setDeviceError] = useState<boolean>(false);
  const [deviceLoading, setDeviceLoading] = useState<boolean>(true);

  // States for photos
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [photosLoading, setPhotosLoading] = useState<boolean>(false);
  const [photosError, setPhotosError] = useState<string | null>(null);
  const [deleteAllConfirm, setDeleteAllConfirm] = useState(false);
  const [deletingAll, setDeletingAll] = useState(false);

  // Upload States
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadPreview, setUploadPreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  // Camera States
  const [isCameraOpen, setIsCameraOpen] = useState(false);
  const [stream, setStream] = useState<MediaStream | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  // Command state
  const [commandLoading, setCommandLoading] = useState(false);
  const [commandMessage, setCommandMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

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
      selectedDevice,
    };

    localStorage.setItem(STORAGE_KEY, JSON.stringify(searchParams));
  }, [searchText, startDate, endDate, selectedDevice]);

  // Handle Video Stream attachment
  useEffect(() => {
    if (isCameraOpen && videoRef.current && stream) {
      videoRef.current.srcObject = stream;
    }
  }, [isCameraOpen, stream]);

  // Cleanup camera stream on unmount
  useEffect(() => {
    return () => {
      if (stream) {
        stream.getTracks().forEach((track) => track.stop());
      }
    };
  }, [stream]);

  // Fetch devices from API
  useEffect(() => {
    const fetchDevices = async () => {
      setDeviceLoading(true);
      setDeviceError(false);

      try {
        const response = await apiFetch("/devices", {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        });

        if (!response.ok) {
          throw new Error("Failed to fetch devices");
        }

        const data = await response.json();
        setDevices(data);
      } catch (error) {
        console.error("Error fetching devices:", error);
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
      const startTimestamp = Math.floor(new Date(startDate).getTime() / 1000);
      const endTimestamp =
        Math.floor(new Date(endDate).getTime() / 1000) + 86399;

      const queryParams = new URLSearchParams();
      queryParams.append("start", startTimestamp.toString());
      queryParams.append("end", endTimestamp.toString());

      if (searchText.trim()) {
        queryParams.append("text", searchText.trim());
      }

      if (selectedDevice !== "all") {
        queryParams.append("device_id", selectedDevice);
      }

      const response = await apiFetch(`/photos?${queryParams.toString()}`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        throw new Error(
          `Failed to fetch photos: ${response.status} ${response.statusText}`,
        );
      }

      const data = await response.json();
      setPhotos(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Error fetching photos:", error);
      setPhotosError((error as Error).message || "Failed to load photos");
      setPhotos([]);
    } finally {
      setPhotosLoading(false);
    }
  };

  const handleDeletePhoto = async (photoId: string) => {
    try {
      const response = await apiFetch(`/photos/${photoId}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        throw new Error("Failed to delete photo");
      }

      setPhotos(photos.filter((p) => p.id !== photoId));
    } catch (error) {
      console.error("Error deleting photo:", error);
      alert("Failed to delete photo");
    }
  };

  const handleDeleteAllPhotos = async () => {
    setDeletingAll(true);
    try {
      const response = await apiFetch("/photos/all", {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        throw new Error("Failed to delete all photos");
      }

      const result = await response.json();
      setPhotos([]);
      setDeleteAllConfirm(false);
      alert(`Deleted ${result.deleted} photos`);
    } catch (error) {
      console.error("Error deleting all photos:", error);
      alert("Failed to delete all photos");
    } finally {
      setDeletingAll(false);
    }
  };

  const sendCommand = async (
    command: "CAPTURE" | "START-LIVE" | "STOP-LIVE",
  ) => {
    const targetDeviceId =
      selectedDevice !== "all" ? selectedDevice : "camera_stream";

    setCommandLoading(true);
    setCommandMessage(null);

    try {
      const response = await apiFetch("/devices/command", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          device_id: targetDeviceId,
          command: command,
        }),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to send command: ${response.status} ${response.statusText}`,
        );
      }

      setCommandMessage({
        type: "success",
        text: `Command ${command} sent successfully`,
      });
    } catch (error) {
      console.error(`Error sending command ${command}:`, error);
      setCommandMessage({
        type: "error",
        text: (error as Error).message || "Failed to send command",
      });
    } finally {
      setCommandLoading(false);
    }
  };

  // --- Upload & Camera Handlers ---
  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      processFile(e.dataTransfer.files[0]);
    }
  };

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      processFile(e.target.files[0]);
    }
  };

  const processFile = (file: File) => {
    if (!file.type.match("image.*")) {
      alert("Please select an image file");
      return;
    }
    setSelectedFile(file);
    const reader = new FileReader();
    reader.onload = (e) => setUploadPreview(e.target?.result as string);
    reader.readAsDataURL(file);
  };

  // Live Web Camera API Logic
  const startCamera = async () => {
    try {
      const mediaStream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: "environment" }, // Prefers back camera if available
      });
      setStream(mediaStream);
      setIsCameraOpen(true);
    } catch (err) {
      console.error("Error accessing camera", err);
      alert(
        "Could not access the camera. Please check your browser permissions.",
      );
    }
  };

  const stopCamera = () => {
    if (stream) {
      stream.getTracks().forEach((track) => track.stop());
      setStream(null);
    }
    setIsCameraOpen(false);
  };

  const capturePhoto = () => {
    if (videoRef.current && canvasRef.current) {
      const video = videoRef.current;
      const canvas = canvasRef.current;

      canvas.width = video.videoWidth;
      canvas.height = video.videoHeight;

      const context = canvas.getContext("2d");
      if (context) {
        context.drawImage(video, 0, 0, canvas.width, canvas.height);

        canvas.toBlob((blob) => {
          if (blob) {
            const file = new File([blob], "camera-capture.jpg", {
              type: "image/jpeg",
            });
            processFile(file);
            stopCamera();
          }
        }, "image/jpeg");
      }
    }
  };

  const resetUploadState = () => {
    setIsUploadModalOpen(false);
    setSelectedFile(null);
    setUploadPreview(null);
    stopCamera();
  };

  const handleUploadSubmit = async () => {
    if (!selectedFile) return;

    setIsUploading(true);
    try {
      const formData = new FormData();
      formData.append("file", selectedFile);

      if (selectedDevice !== "all") {
        formData.append("device_id", selectedDevice);
      }

      const response = await apiFetch("/photos/upload", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        throw new Error("Failed to upload photo");
      }

      setCommandMessage({
        type: "success",
        text: "Photo uploaded successfully",
      });
      resetUploadState();
      handleSearch();
    } catch (error) {
      console.error("Error uploading photo:", error);
      setCommandMessage({ type: "error", text: "Failed to upload photo" });
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="container mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-sky-700 dark:text-sky-300">
          Photos
        </h1>
        <button
          onClick={() => setIsUploadModalOpen(true)}
          className="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-md cursor-pointer hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
            />
          </svg>
          Upload Photo
        </button>
      </div>

      {/* Search and filter section */}
      <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm mb-6">
        <div className="flex flex-wrap items-end gap-4">
          {/* Text search */}
          <div className="flex-1 min-w-[200px]">
            <label
              htmlFor="search"
              className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >
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
            <label
              htmlFor="start-date"
              className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >
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
            <label
              htmlFor="end-date"
              className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >
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

          {/* Device dropdown */}
          {!deviceLoading && (
            <div>
              <label
                htmlFor="device"
                className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >
                Device
              </label>
              <select
                id="device"
                value={selectedDevice}
                onChange={(e) => setSelectedDevice(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent"
              >
                <option value="all">All</option>
                {!deviceError &&
                  devices.map((device) => (
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
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Loading devices...
                </span>
              </div>
            </div>
          )}

          {/* Search button */}
          <div>
            <button
              onClick={handleSearch}
              disabled={photosLoading}
              className="px-4 py-2 bg-sky-600 text-white rounded-md hover:bg-sky-700 cursor-pointer focus:outline-none focus:ring-2 focus:ring-sky-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {photosLoading ? "Searching..." : "Search"}
            </button>
          </div>

          {/* Separator */}
          <div className="w-px h-10 bg-gray-300 dark:bg-gray-600 mx-2 hidden md:block"></div>

          {/* ESP Camera Controls */}
          <div className="flex items-center gap-3 bg-gray-50 dark:bg-gray-700 px-3 py-1 rounded-md border border-gray-200 dark:border-gray-600">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300 whitespace-nowrap">
              ESP Camera:
            </span>

            <div className="flex gap-2">
              <button
                onClick={() => sendCommand("CAPTURE")}
                disabled={commandLoading}
                className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors disabled:opacity-50 cursor-pointer"
              >
                Capture
              </button>
              <button
                onClick={() => sendCommand("START-LIVE")}
                disabled={commandLoading}
                className="px-3 py-1.5 bg-emerald-600 text-white text-sm rounded hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition-colors disabled:opacity-50 cursor-pointer"
              >
                Start Live
              </button>
              <button
                onClick={() => sendCommand("STOP-LIVE")}
                disabled={commandLoading}
                className="px-3 py-1.5 bg-red-600 text-white text-sm rounded hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 transition-colors disabled:opacity-50 cursor-pointer"
              >
                Stop Live
              </button>
            </div>
          </div>

          {/* Delete All button */}
          <div>
            <button
              onClick={() => setDeleteAllConfirm(true)}
              className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors cursor-pointer"
            >
              Delete All
            </button>
          </div>
        </div>
      </div>

      {/* Command Status Message */}
      {commandMessage && (
        <div
          className={`mb-6 p-4 rounded-md shadow-sm ${
            commandMessage.type === "success"
              ? "bg-green-50 dark:bg-green-900/30 text-green-800 dark:text-green-300 border border-green-200 dark:border-green-700"
              : "bg-red-50 dark:bg-red-900/30 text-red-800 dark:text-red-300 border border-red-200 dark:border-red-700"
          }`}
        >
          <div className="flex items-center">
            <span className="text-lg mr-2">
              {commandMessage.type === "success" ? "✅" : "❌"}
            </span>
            {commandMessage.text}
          </div>
        </div>
      )}

      {/* Delete All Confirmation Modal */}
      {deleteAllConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md shadow-xl">
            <h3 className="text-lg font-semibold text-red-600 dark:text-red-400 mb-4">
              ⚠️ Delete All Photos
            </h3>
            <p className="text-gray-700 dark:text-gray-300 mb-6">
              Are you sure you want to delete ALL photos? This action cannot be
              undone.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setDeleteAllConfirm(false)}
                className="px-4 py-2 bg-gray-300 hover:bg-gray-400 dark:bg-gray-600 dark:hover:bg-gray-500 dark:text-gray-100 rounded-md transition-colors cursor-pointer"
                disabled={deletingAll}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteAllPhotos}
                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors cursor-pointer"
                disabled={deletingAll}
              >
                {deletingAll ? "Deleting..." : "Yes, Delete All"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Upload Photo Modal */}
      {isUploadModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg w-full max-w-lg shadow-xl flex flex-col">
            <div className="flex justify-between items-center p-4 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Upload Photo
              </h3>
              <button
                onClick={resetUploadState}
                className="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 focus:outline-none cursor-pointer"
              >
                <svg
                  className="w-6 h-6"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <div className="p-6 overflow-y-auto">
              {/* State 1: Active Web Camera */}
              {isCameraOpen ? (
                <div className="flex flex-col items-center">
                  <div className="relative w-full max-h-64 mb-4 rounded-lg overflow-hidden bg-black flex justify-center">
                    <video
                      ref={videoRef}
                      autoPlay
                      playsInline
                      muted
                      className="object-contain max-h-64 w-full"
                    />
                  </div>
                  {/* Hidden canvas to process the image frame */}
                  <canvas ref={canvasRef} className="hidden" />

                  <div className="flex gap-3 mt-2">
                    <button
                      onClick={stopCamera}
                      className="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-md hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors cursor-pointer"
                    >
                      Cancel Camera
                    </button>
                    <button
                      onClick={capturePhoto}
                      className="px-4 py-2 bg-emerald-600 text-white rounded-md hover:bg-emerald-700 transition-colors flex items-center gap-2 cursor-pointer"
                    >
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
                        />
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
                        />
                      </svg>
                      Take Photo
                    </button>
                  </div>
                </div>
              ) : !selectedFile ? (
                /* State 2: Drag & Drop Area */
                <div
                  className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                    dragActive
                      ? "border-sky-500 bg-sky-50 dark:bg-sky-900/20"
                      : "border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700/50"
                  }`}
                  onDragEnter={handleDrag}
                  onDragOver={handleDrag}
                  onDragLeave={handleDrag}
                  onDrop={handleDrop}
                >
                  <svg
                    className="mx-auto h-12 w-12 text-gray-400 mb-4"
                    stroke="currentColor"
                    fill="none"
                    viewBox="0 0 48 48"
                  >
                    <path
                      d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
                      strokeWidth={2}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  <p className="text-gray-600 dark:text-gray-300 mb-2">
                    Drag and drop your image here, or
                  </p>

                  <div className="flex flex-col sm:flex-row justify-center gap-3 mt-4">
                    {/* Browse Files Button */}
                    <button
                      onClick={() => fileInputRef.current?.click()}
                      className="px-4 py-2 bg-white dark:bg-gray-700 text-sky-600 dark:text-sky-400 border border-sky-600 dark:border-sky-500 rounded-md hover:bg-sky-50 dark:hover:bg-gray-600 transition-colors cursor-pointer"
                    >
                      Browse Files
                    </button>

                    {/* Native API Web Camera Button */}
                    <button
                      onClick={startCamera}
                      className="px-4 py-2 bg-white dark:bg-gray-700 text-emerald-600 dark:text-emerald-400 border border-emerald-600 dark:border-emerald-500 rounded-md hover:bg-emerald-50 dark:hover:bg-gray-600 transition-colors flex justify-center items-center gap-2 cursor-pointer"
                    >
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
                        />
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
                        />
                      </svg>
                      Open Camera
                    </button>

                    {/* Hidden input to trigger file dialogs */}
                    <input
                      type="file"
                      accept="image/*"
                      ref={fileInputRef}
                      onChange={handleFileInput}
                      className="hidden"
                    />
                  </div>
                </div>
              ) : (
                /* State 3: Selected Image Preview */
                <div className="flex flex-col items-center">
                  {uploadPreview && (
                    <div className="relative w-full max-h-64 mb-4 rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-900 flex justify-center">
                      <img
                        src={uploadPreview}
                        alt="Upload preview"
                        className="object-contain max-h-64"
                      />
                    </div>
                  )}
                  <div className="flex justify-between w-full items-center mb-2">
                    <p className="text-sm font-medium text-gray-700 dark:text-gray-300 truncate mr-4">
                      {selectedFile.name}
                    </p>
                    <button
                      onClick={() => {
                        setSelectedFile(null);
                        setUploadPreview(null);
                      }}
                      className="text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 whitespace-nowrap cursor-pointer"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              )}
            </div>

            <div className="p-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3 bg-gray-50 dark:bg-gray-800/50 rounded-b-lg">
              <button
                onClick={resetUploadState}
                disabled={isUploading}
                className="px-4 py-2 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 rounded-md transition-colors cursor-pointer"
              >
                Cancel
              </button>
              <button
                onClick={handleUploadSubmit}
                disabled={!selectedFile || isUploading || isCameraOpen}
                className="px-4 py-2 bg-sky-600 hover:bg-sky-700 text-white rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center cursor-pointer"
              >
                {isUploading ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-white mr-2"></div>
                    Uploading...
                  </>
                ) : (
                  "Upload"
                )}
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
                {(photos || []).map((photo) => (
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
