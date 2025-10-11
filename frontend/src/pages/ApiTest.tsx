import React, { useState } from 'react';
import { 
  healthAPI, 
  authAPI, 
  metricsAPI, 
  alertsAPI, 
  logsAPI, 
  summaryAPI 
} from '../services/api';
import { 
  PlayIcon, 
  CheckCircleIcon, 
  XCircleIcon,
  ClockIcon 
} from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

interface TestResult {
  endpoint: string;
  method: string;
  status: 'pending' | 'success' | 'error';
  response?: any;
  error?: string;
  duration?: number;
}

const ApiTest: React.FC = () => {
  const [testResults, setTestResults] = useState<TestResult[]>([]);
  const [testing, setTesting] = useState(false);

  const updateTestResult = (endpoint: string, result: Partial<TestResult>) => {
    setTestResults(prev => {
      const existing = prev.find(r => r.endpoint === endpoint);
      if (existing) {
        return prev.map(r => r.endpoint === endpoint ? { ...r, ...result } : r);
      } else {
        return [...prev, { endpoint, method: 'GET', status: 'pending', ...result }];
      }
    });
  };

  const runTest = async (
    endpoint: string, 
    method: string, 
    testFn: () => Promise<any>,
    requiresAuth: boolean = true
  ) => {
    const startTime = Date.now();
    updateTestResult(endpoint, { method, status: 'pending' });

    try {
      const response = await testFn();
      const duration = Date.now() - startTime;
      updateTestResult(endpoint, { 
        status: 'success', 
        response: response.data,
        duration 
      });
    } catch (error: any) {
      const duration = Date.now() - startTime;
      updateTestResult(endpoint, { 
        status: 'error', 
        error: error.response?.data?.error || error.message,
        duration 
      });
    }
  };

  const runAllTests = async () => {
    setTesting(true);
    setTestResults([]);

    // Health Check (Public)
    await runTest('/health', 'GET', () => healthAPI.check(), false);

    // Auth Tests (Public)
    await runTest('/auth/validate', 'POST', () => 
      authAPI.validateToken(localStorage.getItem('token') || ''), false);

    // Protected Routes (Require Auth)
    const token = localStorage.getItem('token');
    if (token) {
      // Metrics Tests
      await runTest('/metrics/current', 'GET', () => metricsAPI.getCurrent());
      await runTest('/metrics/history/cpu_usage', 'GET', () => 
        metricsAPI.getHistory('cpu_usage', 10));
      await runTest('/metrics/history/memory_usage', 'GET', () => 
        metricsAPI.getHistory('memory_usage', 10));

      // Alerts Tests
      await runTest('/alerts', 'GET', () => alertsAPI.getAlerts());
      await runTest('/alerts?status=active', 'GET', () => 
        alertsAPI.getAlerts('active'));
      await runTest('/alerts?status=resolved', 'GET', () => 
        alertsAPI.getAlerts('resolved'));

      // Logs Test
      await runTest('/logs/analyze', 'GET', () => 
        logsAPI.analyze('./data/sample.log'));

      // Summary Test
      await runTest('/summary', 'GET', () => summaryAPI.getSummary(10));

      // Test Alert Creation
      await runTest('/alerts', 'POST', () => 
        alertsAPI.createAlert('cpu_usage', 85, 80));
    } else {
      toast.error('Please login first to test protected routes');
    }

    setTesting(false);
    toast.success('API tests completed');
  };

  const getStatusIcon = (status: TestResult['status']) => {
    switch (status) {
      case 'pending':
        return <ClockIcon className="h-5 w-5 text-yellow-500 animate-spin" />;
      case 'success':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />;
      case 'error':
        return <XCircleIcon className="h-5 w-5 text-red-500" />;
    }
  };

  const getStatusColor = (status: TestResult['status']) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-50 border-yellow-200';
      case 'success':
        return 'bg-green-50 border-green-200';
      case 'error':
        return 'bg-red-50 border-red-200';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">API Testing</h1>
          <p className="text-gray-600">Test all backend API endpoints</p>
        </div>
        
        <button
          onClick={runAllTests}
          disabled={testing}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
        >
          <PlayIcon className="h-4 w-4 mr-2" />
          {testing ? 'Running Tests...' : 'Run All Tests'}
        </button>
      </div>

      {/* Test Results */}
      {testResults.length > 0 && (
        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Test Results</h3>
          </div>
          <div className="divide-y divide-gray-200">
            {testResults.map((result, index) => (
              <div key={index} className={`p-6 ${getStatusColor(result.status)} border-l-4`}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    {getStatusIcon(result.status)}
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {result.method} {result.endpoint}
                      </p>
                      {result.duration && (
                        <p className="text-xs text-gray-500">
                          {result.duration}ms
                        </p>
                      )}
                    </div>
                  </div>
                  
                  <div className="text-right">
                    {result.status === 'success' && (
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                        Success
                      </span>
                    )}
                    {result.status === 'error' && (
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                        Error
                      </span>
                    )}
                    {result.status === 'pending' && (
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800">
                        Testing...
                      </span>
                    )}
                  </div>
                </div>

                {result.error && (
                  <div className="mt-2">
                    <p className="text-sm text-red-600">Error: {result.error}</p>
                  </div>
                )}

                {result.response && result.status === 'success' && (
                  <div className="mt-2">
                    <details className="text-sm">
                      <summary className="cursor-pointer text-gray-600 hover:text-gray-900">
                        View Response
                      </summary>
                      <pre className="mt-2 p-2 bg-gray-100 rounded text-xs overflow-x-auto">
                        {JSON.stringify(result.response, null, 2)}
                      </pre>
                    </details>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* API Documentation */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Available Endpoints</h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">Public Endpoints</h4>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">GET /health</span>
                  <span className="text-green-600">✓ Public</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">POST /auth/register</span>
                  <span className="text-green-600">✓ Public</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">POST /auth/login</span>
                  <span className="text-green-600">✓ Public</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">POST /auth/validate</span>
                  <span className="text-green-600">✓ Public</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">POST /auth/refresh</span>
                  <span className="text-green-600">✓ Public</span>
                </div>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">Protected Endpoints</h4>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">GET /metrics/current</span>
                  <span className="text-orange-600">🔒 Auth Required</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">GET /metrics/history/:type</span>
                  <span className="text-orange-600">🔒 Auth Required</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">GET /alerts</span>
                  <span className="text-orange-600">🔒 Auth Required</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">POST /alerts</span>
                  <span className="text-orange-600">🔒 Auth Required</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">PUT /alerts/:id/resolve</span>
                  <span className="text-orange-600">🔒 Auth Required</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">GET /logs/analyze</span>
                  <span className="text-orange-600">🔒 Auth Required</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">GET /summary</span>
                  <span className="text-orange-600">🔒 Auth Required</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ApiTest;