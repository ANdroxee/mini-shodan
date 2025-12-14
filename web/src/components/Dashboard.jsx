import { useState, useEffect } from 'react';
import axios from 'axios';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Server, Activity, AlertTriangle, TrendingUp } from 'lucide-react';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function Dashboard() {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, []);

  const fetchStats = async () => {
    try {
      const response = await axios.get(`${API_URL}/api/v1/stats`);
      setStats(response.data);
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-400">Loading dashboard...</div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="text-center text-red-400">Failed to load stats</div>
    );
  }

  const COLORS = ['#3b82f6', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981'];

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard
          title="Total Devices"
          value={stats.total_devices}
          icon={<Server className="w-6 h-6" />}
          color="blue"
        />
        <StatCard
          title="Active Devices"
          value={stats.active_devices || 0}
          icon={<Activity className="w-6 h-6" />}
          color="green"
        />
        <StatCard
          title="Total Ports"
          value={stats.total_ports}
          icon={<TrendingUp className="w-6 h-6" />}
          color="purple"
        />
        <StatCard
          title="Risky Ports"
          value={stats.risky_ports?.length || 0}
          icon={<AlertTriangle className="w-6 h-6" />}
          color="red"
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Services Pie Chart */}
        <div className="bg-gray-900 rounded-lg p-6 border border-gray-800">
          <h3 className="text-lg font-semibold mb-4">Top Services</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={stats.top_services?.slice(0, 5) || []}
                dataKey="count"
                nameKey="service"
                cx="50%"
                cy="50%"
                outerRadius={100}
                label
              >
                {stats.top_services?.slice(0, 5).map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip 
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151' }}
              />
            </PieChart>
          </ResponsiveContainer>
        </div>

        {/* Service Distribution Bar Chart */}
        <div className="bg-gray-900 rounded-lg p-6 border border-gray-800">
          <h3 className="text-lg font-semibold mb-4">Service Distribution</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={stats.top_services?.slice(0, 8) || []}>
              <XAxis 
                dataKey="service" 
                stroke="#9ca3af"
                angle={-45}
                textAnchor="end"
                height={100}
              />
              <YAxis stroke="#9ca3af" />
              <Tooltip 
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151' }}
              />
              <Bar dataKey="count" fill="#3b82f6" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Risky Ports Alert */}
      {stats.risky_ports && stats.risky_ports.length > 0 && (
        <div className="bg-red-900/20 border border-red-800 rounded-lg p-6">
          <div className="flex items-center space-x-2 mb-4">
            <AlertTriangle className="w-5 h-5 text-red-400" />
            <h3 className="text-lg font-semibold text-red-400">
              Risky Ports Detected
            </h3>
          </div>
          <div className="space-y-2">
            {stats.risky_ports.slice(0, 5).map((port, index) => (
              <div 
                key={index}
                className="flex items-center justify-between bg-gray-900/50 p-3 rounded"
              >
                <div>
                  <span className="font-mono text-red-400">Port {port.port}</span>
                  <span className="mx-2 text-gray-600">•</span>
                  <span className="text-gray-400">{port.service}</span>
                </div>
                <div className="text-sm text-gray-500">
                  Device ID: {port.device_id}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({ title, value, icon, color }) {
  const colorClasses = {
    blue: 'bg-blue-500/10 text-blue-400 border-blue-800',
    green: 'bg-green-500/10 text-green-400 border-green-800',
    purple: 'bg-purple-500/10 text-purple-400 border-purple-800',
    red: 'bg-red-500/10 text-red-400 border-red-800',
  };

  return (
    <div className={`${colorClasses[color]} rounded-lg p-6 border`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm opacity-80 mb-1">{title}</p>
          <p className="text-3xl font-bold">{value}</p>
        </div>
        <div className="opacity-50">{icon}</div>
      </div>
    </div>
  );
}

export default Dashboard;