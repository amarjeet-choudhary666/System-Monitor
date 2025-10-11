import React, { useState, useEffect } from 'react';
import { summaryAPI, metricsAPI } from '../services/api';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { format } from 'date-fns';
import {
  CpuChipIcon,
  CircleStackIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
} from '@heroicons/react/24/outline';

interface SystemMetrics {
  cpu_usage: number;
  memory_usage: number;
  timestamp: string;
}

interface AlertSummary {
  total_alerts: number;
  active_alerts: number;
  resolved_alerts: number;
  alerts_by_type: Record<string, number>;
  alerts_by_severity: Record<string, number>;
  recent_alerts: Array<{
    id: number;
    type: string;
    message: string;
    severity: string;
    status: string;
    triggered_at: string;
  }>;
}

interface Summary {
  current_metrics: SystemMetrics;
  alerts: AlertSummary;
  metric_averages: {
    cpu: { average: number; min: number; max: number; count: number };
    memory: { average: number; min: number; max: number; count: number };
  };
}

const Dashboard: React.FC = () => {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [metricsHistory, setMetricsHistory] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchDashboardData = async () => {
    try {
      const [summaryResponse, cpuHistory, memoryHistory] = await Promise.all([
        summaryAPI.getSummary(10),
        metricsAPI.getHistory('cpu_usage', 20),
        metricsAPI.getHistory('memory_usage', 20),
      ]);

      setSummary(summaryResponse.data.summary);

      // Combine CPU and memory history for chart
      const cpuData = cpuHistory.data.history || [];
      const memoryData = memoryHistory.data.history || [];
      
      const combinedData = cpuData.map((cpu: any, index: number) => ({
        timestamp: format(new Date(cpu.timestamp), 'HH:mm'),
        cpu: cpu.value,
        memory: memoryData[index]?.value || 0,
      }));

      setMetricsHistory(combinedData);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
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

  if (!summary) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Failed to load dashboard data</p>
      </div>
    );
  }

  const alertTypeData = Object.entries(summary.alerts.alerts_by_type).map(([type, count]) => ({
    name: type.replace('_', ' ').toUpperCase(),
    value: count,
  }));

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600">System observability and monitoring overview</p>
      </div>

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricCard
          title="CPU Usage"
          value={`${summary.current_metrics.cpu_usage.toFixed(1)}%`}
          icon={CpuChipIcon}
          color="blue"
          subtitle={`Avg: ${summary.metric_averages.cpu.average.toFixed(1)}%`}
        />
        <MetricCard
          title="Memory Usage"
          value={`${summary.current_metrics.memory_usage.toFixed(1)}%`}
          icon={CircleStackIcon}
          color="green"
          subtitle={`Avg: ${summary.metric_averages.memory.average.toFixed(1)}%`}
        />
        <MetricCard
          title="Active Alerts"
          value={summary.alerts.active_alerts.toString()}
          icon={ExclamationTriangleIcon}
          color="red"
          subtitle={`Total: ${summary.alerts.total_alerts}`}
        />
        <MetricCard
          title="Resolved Alerts"
          value={summary.alerts.resolved_alerts.toString()}
          icon={CheckCircleIcon}
          color="green"
          subtitle="All time"
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Metrics Timeline */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Metrics Timeline</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={metricsHistory}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="timestamp" />
              <YAxis />
              <Tooltip />
              <Line type="monotone" dataKey="cpu" stroke="#3B82F6" name="CPU %" />
              <Line type="monotone" dataKey="memory" stroke="#10B981" name="Memory %" />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Alert Distribution */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Alert Distribution</h3>
          {alertTypeData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={alertTypeData}
                  cx="50%"
                  cy="50%"
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                  label={({ name, value }) => `${name}: ${value}`}
                >
                  {alertTypeData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={index % 2 === 0 ? '#3B82F6' : '#10B981'} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-64 text-gray-500">
              No alerts data available
            </div>
          )}
        </div>
      </div>

      {/* Recent Alerts */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Recent Alerts</h3>
        </div>
        <div className="divide-y divide-gray-200">
          {summary.alerts.recent_alerts.length > 0 ? (
            summary.alerts.recent_alerts.map((alert) => (
              <div key={alert.id} className="px-6 py-4 flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div
                    className={`w-3 h-3 rounded-full ${
                      alert.status === 'active' ? 'bg-red-500' : 'bg-green-500'
                    }`}
                  />
                  <div>
                    <p className="text-sm font-medium text-gray-900">{alert.message}</p>
                    <p className="text-sm text-gray-500">
                      {format(new Date(alert.triggered_at), 'MMM dd, yyyy HH:mm')}
                    </p>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <span
                    className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      alert.severity === 'critical'
                        ? 'bg-red-100 text-red-800'
                        : alert.severity === 'high'
                        ? 'bg-orange-100 text-orange-800'
                        : alert.severity === 'medium'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-green-100 text-green-800'
                    }`}
                  >
                    {alert.severity}
                  </span>
                  <span
                    className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      alert.status === 'active'
                        ? 'bg-red-100 text-red-800'
                        : 'bg-green-100 text-green-800'
                    }`}
                  >
                    {alert.status}
                  </span>
                </div>
              </div>
            ))
          ) : (
            <div className="px-6 py-8 text-center text-gray-500">
              No recent alerts
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

interface MetricCardProps {
  title: string;
  value: string;
  icon: React.ComponentType<any>;
  color: 'blue' | 'green' | 'red' | 'yellow';
  subtitle?: string;
}

const MetricCard: React.FC<MetricCardProps> = ({ title, value, icon: Icon, color, subtitle }) => {
  const colorClasses = {
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    red: 'bg-red-500',
    yellow: 'bg-yellow-500',
  };

  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <Icon className={`h-6 w-6 text-white p-1 rounded ${colorClasses[color]}`} />
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="text-sm font-medium text-gray-500 truncate">{title}</dt>
              <dd className="text-lg font-medium text-gray-900">{value}</dd>
              {subtitle && <dd className="text-sm text-gray-500">{subtitle}</dd>}
            </dl>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;