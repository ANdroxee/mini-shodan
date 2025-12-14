import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import { Search, ExternalLink, Wifi, WifiOff } from 'lucide-react';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function DeviceTable() {
  const [devices, setDevices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [serviceFilter, setServiceFilter] = useState('');

  useEffect(() => {
    fetchDevices();
  }, [serviceFilter]);

  const fetchDevices = async () => {
    try {
      let url = `${API_URL}/api/v1/devices?limit=100`;
      if (serviceFilter) {
        url += `&service=${serviceFilter}`;
      }
      
      const response = await axios.get(url);
      setDevices(response.data.devices || []);
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch devices:', error);
      setLoading(false);
    }
  };

  const filteredDevices = devices.filter(device => {
    if (!search) return true;
    const searchLower = search.toLowerCase();
    return (
      device.ip.includes(searchLower) ||
      device.hostname?.toLowerCase().includes(searchLower) ||
      device.vendor?.toLowerCase().includes(searchLower)
    );
  });

  const isRecent = (lastSeen) => {
    const hoursDiff = (new Date() - new Date(lastSeen)) / (1000 * 60 * 60);
    return hoursDiff < 24;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-400">Loading devices...</div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Network Devices</h2>
        <button 
          onClick={fetchDevices}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition"
        >
          Refresh
        </button>
      </div>

      {/* Filters */}
      <div className="bg-gray-900 rounded-lg p-4 border border-gray-800">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-500" />
            <input
              type="text"
              placeholder="Search IP, hostname, vendor..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-blue-500"
            />
          </div>

          {/* Service Filter */}
          <select
            value={serviceFilter}
            onChange={(e) => setServiceFilter(e.target.value)}
            className="px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-blue-500"
          >
            <option value="">All Services</option>
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
            <option value="ssh">SSH</option>
            <option value="ftp">FTP</option>
            <option value="telnet">Telnet</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="bg-gray-900 rounded-lg border border-gray-800 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-800 border-b border-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  IP Address
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Hostname
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Vendor
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  OS
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Last Seen
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800">
              {filteredDevices.map((device) => (
                <tr key={device.id} className="hover:bg-gray-800/50 transition">
                  <td className="px-6 py-4 whitespace-nowrap">
                    {isRecent(device.last_seen) ? (
                      <Wifi className="w-5 h-5 text-green-400" />
                    ) : (
                      <WifiOff className="w-5 h-5 text-gray-600" />
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="font-mono text-blue-400">{device.ip}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-gray-300">
                    {device.hostname || '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-gray-400 text-sm">
                    {device.vendor || '-'}
                  </td>
                  <td className="px-6 py-4 text-gray-400 text-sm">
                    {device.os ? (
                      <span className="truncate max-w-xs block">
                        {device.os.substring(0, 40)}...
                      </span>
                    ) : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-gray-500 text-sm">
                    {new Date(device.last_seen).toLocaleString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      to={`/devices/${device.ip}`}
                      className="text-blue-400 hover:text-blue-300 transition flex items-center space-x-1"
                    >
                      <span>Details</span>
                      <ExternalLink className="w-4 h-4" />
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredDevices.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            No devices found
          </div>
        )}
      </div>

      {/* Stats */}
      <div className="text-sm text-gray-500">
        Showing {filteredDevices.length} of {devices.length} devices
      </div>
    </div>
  );
}

export default DeviceTable;