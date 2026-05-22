import React, { useState, useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { apiFetch } from '../../utils/api';
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer,
    PieChart,
    Pie,
    Cell,
} from 'recharts';

// Interface for photo data, including the new boolean fields
interface Photo {
    id: string;
    timestamp: string;

    // Boolean fields from backend
    control_angajare?: boolean;
    control_periodic?: boolean;
    control_adaptare?: boolean;
    control_reluare?: boolean;
    control_supraveghere?: boolean;
    control_alte?: boolean;

    aviz_apt?: boolean;
    aviz_apt_conditionat?: boolean;
    aviz_inapt_temporar?: boolean;
    aviz_inapt?: boolean;
}

const StatisticsPage: React.FC = () => {
    const { token } = useAuth();
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [photos, setPhotos] = useState<Photo[]>([]);

    const [controlChartType, setControlChartType] = useState<'bar' | 'pie'>('bar');
    const [avizChartType, setAvizChartType] = useState<'bar' | 'pie'>('pie');

    // Default date range: Last 30 days
    const today = new Date();
    const thirtyDaysAgo = new Date();
    thirtyDaysAgo.setDate(today.getDate() - 30);

    const [startDate, setStartDate] = useState(thirtyDaysAgo.toISOString().slice(0, 10));
    const [endDate, setEndDate] = useState(today.toISOString().slice(0, 10));

    useEffect(() => {
        fetchData();
    }, [startDate, endDate]);

    const fetchData = async () => {
        setLoading(true);
        setError(null);
        try {
            // Calculate timestamps
            const startTimestamp = Math.floor(new Date(startDate).getTime() / 1000);
            const endTimestamp = Math.floor(new Date(endDate).getTime() / 1000) + 86399;

            const queryParams = new URLSearchParams();
            queryParams.append('start', startTimestamp.toString());
            queryParams.append('end', endTimestamp.toString());

            const response = await apiFetch(`/photos?${queryParams.toString()}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error('Failed to fetch data');
            }

            const data = await response.json();
            setPhotos(Array.isArray(data) ? data : []);
        } catch (err) {
            console.error('Error fetching stats data:', err);
            setError('Failed to load statistics data');
        } finally {
            setLoading(false);
        }
    };

    // Process data for charts
    const getControlStats = () => {
        const stats = {
            'Angajare': 0,
            'Periodic': 0,
            'Adaptare': 0,
            'Reluare': 0,
            'Supraveghere': 0,
            'Alte': 0,
        };

        photos.forEach(photo => {
            if (photo.control_angajare) stats['Angajare']++;
            if (photo.control_periodic) stats['Periodic']++;
            if (photo.control_adaptare) stats['Adaptare']++;
            if (photo.control_reluare) stats['Reluare']++;
            if (photo.control_supraveghere) stats['Supraveghere']++;
            if (photo.control_alte) stats['Alte']++;
        });

        return Object.entries(stats).map(([name, value]) => ({ name, value }));
    };

    const getAvizStats = () => {
        const stats = {
            'APT': 0,
            'APT Conditionat': 0,
            'Inapt Temporar': 0,
            'Inapt': 0,
        };

        photos.forEach(photo => {
            if (photo.aviz_apt) stats['APT']++;
            if (photo.aviz_apt_conditionat) stats['APT Conditionat']++;
            if (photo.aviz_inapt_temporar) stats['Inapt Temporar']++;
            if (photo.aviz_inapt) stats['Inapt']++;
        });

        return Object.entries(stats).map(([name, value]) => ({ name, value }));
    };

    const controlData = getControlStats();
    const avizData = getAvizStats();

    const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8', '#82ca9d'];

    const renderChartToggle = (currentType: 'bar' | 'pie', setType: (t: 'bar' | 'pie') => void) => (
        <div className="flex bg-gray-100 dark:bg-gray-700 p-1 rounded-md">
            <button
                onClick={() => setType('bar')}
                className={`px-3 py-1 text-sm rounded-sm transition-colors ${currentType === 'bar'
                    ? 'bg-white dark:bg-gray-600 shadow-sm text-sky-600 dark:text-sky-300 font-medium'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                    }`}
            >
                Bar
            </button>
            <button
                onClick={() => setType('pie')}
                className={`px-3 py-1 text-sm rounded-sm transition-colors ${currentType === 'pie'
                    ? 'bg-white dark:bg-gray-600 shadow-sm text-sky-600 dark:text-sky-300 font-medium'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                    }`}
            >
                Pie
            </button>
        </div>
    );

    const renderBarChart = (data: { name: string, value: number }[]) => (
        <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="value" fill="#0ea5e9" name="Count" />
            </BarChart>
        </ResponsiveContainer>
    );

    const renderPieChart = (data: { name: string, value: number }[]) => (
        <ResponsiveContainer width="100%" height="100%">
            <PieChart>
                <Pie
                    data={data}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }) => `${name} ${(percent ? percent * 100 : 0).toFixed(0)}%`}
                    outerRadius={100}
                    fill="#8884d8"
                    dataKey="value"
                >
                    {data.map((_, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                </Pie>
                <Tooltip />
            </PieChart>
        </ResponsiveContainer>
    );

    return (
        <div className="container mx-auto pb-10">
            <h1 className="text-2xl font-semibold text-sky-700 dark:text-sky-300 mb-6">Statistics</h1>

            {/* Date Filter */}
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm mb-6 flex gap-4 items-end">
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Start Date</label>
                    <input
                        type="date"
                        value={startDate}
                        onChange={(e) => setStartDate(e.target.value)}
                        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-sky-500"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">End Date</label>
                    <input
                        type="date"
                        value={endDate}
                        onChange={(e) => setEndDate(e.target.value)}
                        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-sky-500"
                    />
                </div>
                <button
                    onClick={fetchData}
                    className="px-4 py-2 bg-sky-600 text-white rounded-md hover:bg-sky-700 mb-[1px]"
                >
                    Refresh
                </button>
            </div>

            {loading ? (
                <div className="flex justify-center h-40 items-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-sky-500"></div>
                </div>
            ) : error ? (
                <div className="bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-4 rounded-md">{error}</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">

                    {/* Control Type Chart */}
                    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
                        <div className="flex justify-between items-center mb-4">
                            <h3 className="text-lg font-medium text-gray-800 dark:text-gray-100">Control Type Distribution</h3>
                            {renderChartToggle(controlChartType, setControlChartType)}
                        </div>
                        <div className="h-[300px]">
                            {controlChartType === 'bar' ? renderBarChart(controlData) : renderPieChart(controlData)}
                        </div>
                    </div>

                    {/* Aviz Medical Chart */}
                    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
                        <div className="flex justify-between items-center mb-4">
                            <h3 className="text-lg font-medium text-gray-800 dark:text-gray-100">Medical Opinion Results</h3>
                            {renderChartToggle(avizChartType, setAvizChartType)}
                        </div>
                        <div className="h-[300px]">
                            {avizChartType === 'bar' ? renderBarChart(avizData) : renderPieChart(avizData)}
                        </div>
                    </div>

                    {/* Simple Summary Cards */}
                    <div className="col-span-1 md:col-span-2 grid grid-cols-1 sm:grid-cols-3 gap-4">
                        <div className="bg-blue-50 dark:bg-blue-900/30 p-4 rounded-lg border border-blue-100 dark:border-blue-800">
                            <span className="block text-sm text-blue-600 dark:text-blue-400 font-medium">Total Files</span>
                            <span className="block text-2xl font-bold text-blue-900 dark:text-blue-200">{photos.length}</span>
                        </div>
                        <div className="bg-green-50 dark:bg-green-900/30 p-4 rounded-lg border border-green-100 dark:border-green-800">
                            <span className="block text-sm text-green-600 dark:text-green-400 font-medium">FIT (APT)</span>
                            <span className="block text-2xl font-bold text-green-900 dark:text-green-200">
                                {avizData.find(d => d.name === 'APT')?.value || 0}
                            </span>
                        </div>
                        <div className="bg-orange-50 dark:bg-orange-900/30 p-4 rounded-lg border border-orange-100 dark:border-orange-800">
                            <span className="block text-sm text-orange-600 dark:text-orange-400 font-medium">Periodic Checks</span>
                            <span className="block text-2xl font-bold text-orange-900 dark:text-orange-200">
                                {controlData.find(d => d.name === 'Periodic')?.value || 0}
                            </span>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default StatisticsPage;
