import React, { useState } from 'react';
import { logsAPI } from '../services/api';
import { DocumentTextIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

interface LogStats {
  level_counts: Record<string, number>;
  top_errors: Array<{
    message: string;
    count: number;
  }>;
  total_entries: number;
}

const Logs: React.FC = () => {
  const [filePath, setFilePath] = useState('');
  const [logStats, setLogStats] = useState<LogStats | null>(null);
  const [loading, setLoading] = useState(false);

  const handleAnalyzeLogs = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!filePath.trim()) {
      toast.error('Please enter a file path');
      return;
    }

    setLoading(true);
    try {
      const response = await logsAPI.analyze(filePath);
      setLogStats(response.data.stats);
      toast.success('Log analysis completed');
    } catch (error: any) {
      console.error('Failed to analyze logs:', error);
      toast.error(error.response?.data?.error || 'Failed to analyze logs');
    } finally {
      setLoading(false);
    }
  };

  const getLevelColor = (level: string) => {
    switch (level.toUpperCase()) {
      case 'ERROR':
        return 'bg-red-100 text-red-800';
      case 'WARN':
        return 'bg-yellow-100 text-yellow-800';
      case 'INFO':
        return 'bg-blue-100 text-blue-800';
      case 'DEBUG':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Log Analysis</h1>
        <p className="text-gray-600">Analyze log files for patterns and errors</p>
      </div>

      {/* Log Analysis Form */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Analyze Log File</h3>
        </div>
        <div className="p-6">
          <form onSubmit={handleAnalyzeLogs} className="space-y-4">
            <div>
              <label htmlFor="filePath" className="block text-sm font-medium text-gray-700">
                Log File Path
              </label>
              <div className="mt-1 relative rounded-md shadow-sm">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <DocumentTextIcon className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="text"
                  id="filePath"
                  value={filePath}
                  onChange={(e) => setFilePath(e.target.value)}
                  className="focus:ring-indigo-500 focus:border-indigo-500 block w-full pl-10 sm:text-sm border-gray-300 rounded-md"
                  placeholder="e.g., /var/log/app.log or ./data/sample.log"
                />
              </div>
              <p className="mt-2 text-sm text-gray-500">
                Enter the path to the log file you want to analyze. For testing, you can use: <code className="bg-gray-100 px-1 rounded">./data/sample.log</code>
              </p>
            </div>
            
            <div>
              <button
                type="submit"
                disabled={loading}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? (
                  <>
                    <div className="animate-spin -ml-1 mr-3 h-4 w-4 border-2 border-white border-t-transparent rounded-full"></div>
                    Analyzing...
                  </>
                ) : (
                  <>
                    <MagnifyingGlassIcon className="-ml-1 mr-2 h-4 w-4" />
                    Analyze Logs
                  </>
                )}
              </button>
            </div>
          </form>
        </div>
      </div>

      {/* Log Analysis Results */}
      {logStats && (
        <div className="space-y-6">
          {/* Summary Stats */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Analysis Summary</h3>
            </div>
            <div className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-gray-900">{logStats.total_entries}</div>
                  <div className="text-sm text-gray-500">Total Entries</div>
                </div>
                {Object.entries(logStats.level_counts).map(([level, count]) => (
                  <div key={level} className="text-center">
                    <div className="text-2xl font-bold text-gray-900">{count}</div>
                    <div className={`text-sm px-2 py-1 rounded-full inline-block ${getLevelColor(level)}`}>
                      {level}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Log Level Distribution */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Log Level Distribution</h3>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                {Object.entries(logStats.level_counts).map(([level, count]) => {
                  const percentage = (count / logStats.total_entries) * 100;
                  return (
                    <div key={level} className="flex items-center">
                      <div className="w-20 text-sm font-medium text-gray-900">{level}</div>
                      <div className="flex-1 mx-4">
                        <div className="bg-gray-200 rounded-full h-2">
                          <div
                            className={`h-2 rounded-full ${
                              level === 'ERROR' ? 'bg-red-500' :
                              level === 'WARN' ? 'bg-yellow-500' :
                              level === 'INFO' ? 'bg-blue-500' : 'bg-gray-500'
                            }`}
                            style={{ width: `${percentage}%` }}
                          ></div>
                        </div>
                      </div>
                      <div className="w-16 text-sm text-gray-500 text-right">
                        {count} ({percentage.toFixed(1)}%)
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Top Errors */}
          {logStats.top_errors.length > 0 && (
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">Top 5 Most Frequent Errors</h3>
              </div>
              <div className="divide-y divide-gray-200">
                {logStats.top_errors.map((error, index) => (
                  <div key={index} className="px-6 py-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <div className="flex-shrink-0">
                          <span className="inline-flex items-center justify-center h-8 w-8 rounded-full bg-red-100 text-red-800 text-sm font-medium">
                            {index + 1}
                          </span>
                        </div>
                        <div className="flex-1">
                          <p className="text-sm font-medium text-gray-900">{error.message}</p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                          {error.count} occurrences
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Instructions */}
      {!logStats && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
          <div className="flex">
            <div className="flex-shrink-0">
              <DocumentTextIcon className="h-5 w-5 text-blue-400" />
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">How to use Log Analysis</h3>
              <div className="mt-2 text-sm text-blue-700">
                <p>
                  Enter the path to a log file to analyze its contents. The analyzer will:
                </p>
                <ul className="list-disc list-inside mt-2 space-y-1">
                  <li>Count log entries by level (INFO, WARN, ERROR, DEBUG)</li>
                  <li>Identify the top 5 most frequent error messages</li>
                  <li>Provide statistics about your log file</li>
                </ul>
                <p className="mt-2">
                  <strong>Sample file:</strong> Use <code className="bg-blue-100 px-1 rounded">./data/sample.log</code> to test with the provided sample data.
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Logs;