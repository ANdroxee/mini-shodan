import { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Dashboard from './components/Dashboard';
import DeviceTable from './components/DeviceTable';
import DeviceDetail from './components/DeviceDetail';
import { Network, Shield, Activity } from 'lucide-react';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-950 text-gray-100">
        {/* Header */}
        <header className="bg-gray-900 border-b border-gray-800">
          <div className="container mx-auto px-4 py-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <Shield className="w-8 h-8 text-blue-500" />
                <h1 className="text-2xl font-bold">Shodan@Home</h1>
              </div>
              
              <nav className="flex space-x-6">
                <Link 
                  to="/" 
                  className="flex items-center space-x-2 hover:text-blue-400 transition"
                >
                  <Activity className="w-4 h-4" />
                  <span>Dashboard</span>
                </Link>
                <Link 
                  to="/devices" 
                  className="flex items-center space-x-2 hover:text-blue-400 transition"
                >
                  <Network className="w-4 h-4" />
                  <span>Devices</span>
                </Link>
              </nav>
            </div>
          </div>
        </header>

        {/* Main Content */}
        <main className="container mx-auto px-4 py-8">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/devices" element={<DeviceTable />} />
            <Route path="/devices/:ip" element={<DeviceDetail />} />
          </Routes>
        </main>

        {/* Footer */}
        <footer className="bg-gray-900 border-t border-gray-800 mt-12">
          <div className="container mx-auto px-4 py-6 text-center text-gray-500 text-sm">
            Shodan@Home v1.0 - Network Scanner & Monitor
          </div>
        </footer>
      </div>
    </Router>
  );
}

export default App;