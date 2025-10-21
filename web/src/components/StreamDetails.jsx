import { useState, useEffect } from 'react';
import { RefreshCw, Mail, HardDrive, ArrowUp, ArrowDown, ToggleLeft, AlertCircle, Loader2 } from 'lucide-react';
import { api } from '../services/api';

export default function StreamDetails({ stream }) {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [autoRefresh, setAutoRefresh] = useState(true);

  useEffect(() => {
    loadStats();

    if (autoRefresh) {
      const interval = setInterval(loadStats, 5000);
      return () => clearInterval(interval);
    }
  }, [stream, autoRefresh]);

  const loadStats = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getStreamStats(stream.connection_id, stream.vhost, stream.name);
      setStats(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const formatNumber = (num) => {
    return num.toLocaleString();
  };

  if (loading && !stats) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-12 h-12 text-blue-500 animate-spin mx-auto mb-4" />
          <p className="text-gray-500 dark:text-gray-400">Loading statistics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="max-w-2xl mx-auto bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-6">
          <div className="flex items-start gap-4">
            <AlertCircle className="w-6 h-6 text-red-600 dark:text-red-400 flex-shrink-0 mt-1" />
            <div className="flex-1">
              <p className="font-semibold text-red-900 dark:text-red-100">
                Error loading stream statistics
              </p>
              <p className="mt-2 text-sm text-red-700 dark:text-red-300">{error}</p>
              <button
                onClick={loadStats}
                className="mt-4 px-4 py-2 bg-red-600 hover:bg-red-700 rounded-lg text-white text-sm font-medium transition-colors inline-flex items-center gap-2"
              >
                <RefreshCw className="w-4 h-4" />
                Retry
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto bg-gray-50 dark:bg-gray-950">
      <div className="p-8 max-w-7xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h2 className="text-3xl font-bold text-gray-900 dark:text-gray-100">
              {stream.name}
            </h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
              VHost: <span className="font-mono">{stream.vhost}</span>
            </p>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              <ToggleLeft
                className={`w-5 h-5 transition-colors ${
                  autoRefresh
                    ? 'text-blue-600 dark:text-blue-400'
                    : 'text-gray-400'
                }`}
              />
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="sr-only"
              />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                Auto-refresh
              </span>
            </label>
            <button
              onClick={loadStats}
              disabled={loading}
              className="p-2.5 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800 rounded-lg transition-colors disabled:opacity-50"
              title="Refresh statistics"
            >
              <RefreshCw className={`w-5 h-5 text-gray-700 dark:text-gray-300 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex items-center justify-between mb-4">
                <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                  <Mail className="w-6 h-6 text-blue-600 dark:text-blue-400" />
                </div>
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
                Total Messages
              </p>
              <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {formatNumber(stats.message_count)}
              </p>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex items-center justify-between mb-4">
                <div className="p-3 bg-green-100 dark:bg-green-900/30 rounded-lg">
                  <HardDrive className="w-6 h-6 text-green-600 dark:text-green-400" />
                </div>
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
                Stream Size
              </p>
              <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {formatBytes(stats.size)}
              </p>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex items-center justify-between mb-4">
                <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                  <ArrowUp className="w-6 h-6 text-purple-600 dark:text-purple-400" />
                </div>
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
                First Offset
              </p>
              <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {formatNumber(stats.first_offset)}
              </p>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex items-center justify-between mb-4">
                <div className="p-3 bg-orange-100 dark:bg-orange-900/30 rounded-lg">
                  <ArrowDown className="w-6 h-6 text-orange-600 dark:text-orange-400" />
                </div>
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
                Last Offset
              </p>
              <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {formatNumber(stats.last_offset)}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
