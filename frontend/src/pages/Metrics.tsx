import React, { useState, useEffect } from 'react';
import { metricsAPI } from '../services/api';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { format } from 'date-fns';
import { CpuChipIcon, CircleStackIcon } from '@heroicons/react/24/outline';

interface Metric {
  id: number;
  type: string;
  value: number;
  unit: string;
  timestamp: string;
}

const Metrics: React.FC = () => {
  const [currentMetrics, setCurrentMetrics] = useState<any>(null);
  const [cpuHistory, setCpuHistory] = useState<Metric[]>([]);
  const [memoryHistory, setMemoryHistory] = useState<Metric[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchMetrics = async () => {
    try {
      const [current, cpu, memory] = await Promise.all([
        metricsAPI.getCurrent(),
        metricsAPI.getHistory('cpu_usage', 50),
        metricsAPI.getHistory('memory_usage', 50),
      ]);

      setCurrentMetrics(current.data.metrics);
      setCpuHistory(cpu.data.history || []);
      setMemoryHistory(memory.data.history || []);
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  const formatChartData = (data: Metric[]) => {
    return data.map(item => ({
      timestamp: format(new Date(item.timestamp), 'HH:mm:ss'),
      value: item.value,
    }));
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">System Metrics</h1>
        <p className="text-gray-600">Real-time system performance monitoring</p>
      </div>

      {/* Current Metrics Cards */}
      {currentMetrics && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <CpuChipIcon className="h-8 w-8 text-blue-600" />
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">CPU Usage</dt>
                    <dd className="text-3xl font-bold text-gray-900">
                      {currentMetrics.cpu_usage.toFixed(1)}%
                    </dd>
                    <dd className="text-sm text-gray-500">
                      Last updated: {format(new Date(currentMetrics.timestamp), 'HH:mm:ss')}
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <CircleStackIcon className="h-8 w-8 text-green-600" />
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">Memory Usage</dt>
                    <dd className="text-3xl font-bold text-gray-900">
                      {currentMetrics.memory_usage.toFixed(1)}%
                    </dd>
                    <dd className="text-sm text-gray-500">
                      Last updated: {format(new Date(currentMetrics.timestamp), 'HH:mm:ss')}
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* CPU Usage Chart */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900 mb-4">CPU Usage History</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={formatChartData(cpuHistory)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="timestamp" />
              <YAxis domain={[0, 100]} />
              <Tooltip formatter={(value) => [`${value}%`, 'CPU Usage']} />
              <Line type="monotone" dataKey="value" stroke="#3B82F6" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Memory Usage Chart */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Memory Usage History</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={formatChartData(memoryHistory)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="timestamp" />
              <YAxis domain={[0, 100]} />
              <Tooltip formatter={(value) => [`${value}%`, 'Memory Usage']} />
              <Line type="monotone" dataKey="value" stroke="#10B981" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Metrics Table */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Recent Metrics</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Value
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Timestamp
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {[...cpuHistory.slice(0, 10), ...memoryHistory.slice(0, 10)]
                .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
                .slice(0, 20)
                .map((metric) => (
                  <tr key={`${metric.type}-${metric.id}`}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {metric.type.replace('_', ' ').toUpperCase()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {metric.value.toFixed(1)}{metric.unit}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {format(new Date(metric.timestamp), 'MMM dd, yyyy HH:mm:ss')}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default Metrics;