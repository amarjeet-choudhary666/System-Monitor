import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api/v1';
const HEALTH_URL = 'http://localhost:8080/health';

// Create axios instance
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const refreshToken = localStorage.getItem('refreshToken');
      if (refreshToken) {
        try {
          const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
            refresh_token: refreshToken,
          });
          const newToken = response.data.token;
          localStorage.setItem('token', newToken);
          
          // Retry the original request
          error.config.headers.Authorization = `Bearer ${newToken}`;
          return api.request(error.config);
        } catch (refreshError) {
          // Refresh failed, redirect to login
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

// Health API
export const healthAPI = {
  check: () => axios.get(HEALTH_URL),
};

// Auth API - Complete implementation
export const authAPI = {
  // Public routes
  register: (username: string, email: string, password: string) =>
    api.post('/auth/register', { username, email, password }),
  
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  
  validateToken: (token: string) =>
    api.post('/auth/validate', { token }),
  
  refreshToken: (refreshToken: string) =>
    api.post('/auth/refresh', { refresh_token: refreshToken }),
  
  // Protected routes
  logout: () => api.post('/auth/logout'),
};

// Metrics API - Complete implementation
export const metricsAPI = {
  // Get current system metrics
  getCurrent: () => api.get('/metrics/current'),
  
  // Get historical metrics by type
  getHistory: (type: 'cpu_usage' | 'memory_usage', limit?: number) => {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit.toString());
    return api.get(`/metrics/history/${type}?${params.toString()}`);
  },
};

// Alerts API - Complete implementation
export const alertsAPI = {
  // Get alerts with optional filtering
  getAlerts: (status?: 'active' | 'resolved', limit?: number) => {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (limit) params.append('limit', limit.toString());
    return api.get(`/alerts?${params.toString()}`);
  },
  
  // Create a new alert (for testing)
  createAlert: (type: 'cpu_usage' | 'memory_usage', value: number, threshold: number) =>
    api.post('/alerts', { type, value, threshold }),
  
  // Resolve an alert
  resolveAlert: (id: number) =>
    api.put(`/alerts/${id}/resolve`),
};

// Logs API - Complete implementation
export const logsAPI = {
  // Analyze log file
  analyze: (filePath: string) =>
    api.get(`/logs/analyze?file=${encodeURIComponent(filePath)}`),
};

// Summary API - Complete implementation
export const summaryAPI = {
  // Get comprehensive system summary
  getSummary: (limit?: number) => {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit.toString());
    return api.get(`/summary?${params.toString()}`);
  },
};

// Additional utility functions
export const apiUtils = {
  // Check if API is reachable
  ping: async () => {
    try {
      await healthAPI.check();
      return true;
    } catch {
      return false;
    }
  },
  
  // Get API status
  getStatus: async () => {
    try {
      const response = await healthAPI.check();
      return {
        status: 'healthy',
        message: response.data.message || 'API is running',
      };
    } catch (error: any) {
      return {
        status: 'error',
        message: error.message || 'API is not reachable',
      };
    }
  },
};

export default api;