import React, { useState, useEffect } from 'react';
import { alertsAPI, apiUtils } from '../services/api';
import { CogIcon, ServerIcon, BellIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

const Settings: React.FC = () => {
  const [apiStatus, setApiStatus] = useState<{ status: string; message: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [testAlert, setTestAlert] = useState({
    type: 'cpu_usage' as 'cpu_usage' | 'memory_usage',
    value: 85,
    threshold: 80,
  });

  useEffect(() => {
    checkApiStatus();
  }, []);

  const checkApiStatus = async () => {
    setLoading(true);
    try {
      const status = await apiUtils.getStatus();
      setApiStatus(status);
    } catch (error) {
      console.error('Failed to check API status:', error);
    } finally {
      setLoading(false);
    }
  };

  const createTestAlert = async () => {
    try {
      await alertsAPI.createAlert(testAlert.type, testAlert.value, testAlert.threshold);
      toast.success('Test alert created successfully');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to create test alert');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-600">System configuration and testing tools</p>
      </div>

      {/* API Status */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900 flex items-center">
              <ServerIcon className="h-5 w-5 mr-2" />
              API Status
            </h3>
            <button
              onClick={checkApiStatus}
              disabled={loading}
              className="inline-flex items-center px-3 py-1 border border-transparent text-sm font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 disabled:opacity-50"
            >
              {loading ? 'Checking...' : 'Refresh'}
            </button>
          </div>
        </div>
        <div className="p-6">
          {apiStatus ? (
            <div className="flex items-center space-x-3">
              <div
                className={`w-3 h-3 rounded-full ${
                  apiStatus.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'
                }`}
              />
              <div>
                <p className="text-sm font-medium text-gray-900">
                  Status: {apiStatus.status === 'healthy' ? 'Connected' : 'Disconnected'}
                </p>
                <p className="text-sm text-gray-500">{apiStatus.message}</p>
              </div>
            </div>
          ) : (
            <p className="text-gray-500">Click refresh to check API status</p>
          )}
        </div>
      </div>

      {/* Test Alert Creation */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900 flex items-center">
            <BellIcon className="h-5 w-5 mr-2" />
            Test Alert Creation
          </h3>
        </div>
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Alert Type</label>
              <select
                value={testAlert.type}
                onChange={(e) => setTestAlert({ ...testAlert, type: e.target.value as 'cpu_usage' | 'memory_usage' })}
                className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              >
                <option value="cpu_usage">CPU Usage</option>
                <option value="memory_usage">Memory Usage</option>
              </select>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Value (%)</label>
                <input
                  type="number"
                  value={testAlert.value}
                  onChange={(e) => setTestAlert({ ...testAlert, value: parseFloat(e.target.value) })}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  min="0"
                  max="100"
                  step="0.1"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Threshold (%)</label>
                <input
                  type="number"
                  value={testAlert.threshold}
                  onChange={(e) => setTestAlert({ ...testAlert, threshold: parseFloat(e.target.value) })}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  min="0"
                  max="100"
                  step="0.1"
                />
              </div>
            </div>

            <button
              onClick={createTestAlert}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Create Test Alert
            </button>
          </div>
        </div>
      </div>

      {/* System Information */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900 flex items-center">
            <CogIcon className="h-5 w-5 mr-2" />
            System Information
          </h3>
        </div>
        <div className="p-6">
          <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-gray-500">Frontend Version</dt>
              <dd className="mt-1 text-sm text-gray-900">1.0.0</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">API Base URL</dt>
              <dd className="mt-1 text-sm text-gray-900">http://localhost:8080/api/v1</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Authentication</dt>
              <dd className="mt-1 text-sm text-gray-900">JWT Token Based</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Auto Refresh</dt>
              <dd className="mt-1 text-sm text-gray-900">30 seconds</dd>
            </div>
          </dl>
        </div>
      </div>

      {/* Available Endpoints */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Available API Endpoints</h3>
        </div>
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-gray-900">Authentication</h4>
              <ul className="mt-2 text-sm text-gray-600 space-y-1">
                <li>• POST /auth/register - User registration</li>
                <li>• POST /auth/login - User login</li>
                <li>• POST /auth/validate - Token validation</li>
                <li>• POST /auth/refresh - Token refresh</li>
                <li>• POST /auth/logout - User logout</li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-900">Metrics</h4>
              <ul className="mt-2 text-sm text-gray-600 space-y-1">
                <li>• GET /metrics/current - Current system metrics</li>
                <li>• GET /metrics/history/:type - Historical metrics</li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-900">Alerts</h4>
              <ul className="mt-2 text-sm text-gray-600 space-y-1">
                <li>• GET /alerts - List alerts</li>
                <li>• POST /alerts - Create alert</li>
                <li>• PUT /alerts/:id/resolve - Resolve alert</li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-900">Other</h4>
              <ul className="mt-2 text-sm text-gray-600 space-y-1">
                <li>• GET /logs/analyze - Analyze log files</li>
                <li>• GET /summary - System summary</li>
                <li>• GET /health - Health check</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;