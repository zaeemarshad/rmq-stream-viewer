import { useState, useEffect } from 'react';
import { RefreshCw, ChevronRight, ChevronLeft, Database, Folder, AlertCircle, Loader2 } from 'lucide-react';

export default function Sidebar({ onStreamSelect, selectedStream, isCollapsed, onToggleCollapse }) {
  const [vhosts, setVHosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [expandedVHosts, setExpandedVHosts] = useState({});

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const vhostsResponse = await fetch('/api/vhosts');

      if (!vhostsResponse.ok) {
        throw new Error('Failed to fetch vhosts');
      }

      const vhosts = await vhostsResponse.json();
      setVHosts(vhosts);

      // Auto-expand vhosts that have streams
      const expanded = {};
      vhosts.forEach(vhost => {
        if (vhost.streams && vhost.streams.length > 0) {
          expanded[`${vhost.connection_id}-${vhost.name}`] = true;
        }
      });
      setExpandedVHosts(expanded);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const toggleVHost = (vhostKey) => {
    setExpandedVHosts(prev => ({
      ...prev,
      [vhostKey]: !prev[vhostKey],
    }));
  };

  if (isCollapsed) {
    return (
      <div className="w-14 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col items-center py-4">
        <button
          onClick={onToggleCollapse}
          className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          title="Expand sidebar"
        >
          <ChevronRight className="w-5 h-5" />
        </button>
      </div>
    );
  }

  return (
    <div className="w-80 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col">
      <div className="p-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">VHosts & Streams</h2>
        <div className="flex gap-2">
          <button
            onClick={loadData}
            disabled={loading}
            className="p-1.5 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors disabled:opacity-50"
            title="Refresh"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={onToggleCollapse}
            className="p-1.5 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            title="Collapse sidebar"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="p-8 text-center">
            <Loader2 className="w-8 h-8 text-blue-500 animate-spin mx-auto mb-3" />
            <p className="text-sm text-gray-500 dark:text-gray-400">Loading vhosts...</p>
          </div>
        ) : error ? (
          <div className="p-4">
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <p className="font-semibold text-red-900 dark:text-red-100 text-sm">
                    Error loading vhosts
                  </p>
                  <p className="mt-1 text-sm text-red-700 dark:text-red-300">{error}</p>
                  <button
                    onClick={loadData}
                    className="mt-3 px-3 py-1.5 bg-red-600 hover:bg-red-700 rounded-lg text-white text-xs font-medium transition-colors"
                  >
                    Retry
                  </button>
                </div>
              </div>
            </div>
          </div>
        ) : vhosts.length === 0 ? (
          <div className="p-8 text-center">
            <Database className="w-12 h-12 text-gray-300 dark:text-gray-700 mx-auto mb-3" />
            <p className="text-gray-600 dark:text-gray-400 text-sm">No vhosts found</p>
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-500">
              Check your RabbitMQ configuration
            </p>
          </div>
        ) : (
          <div className="py-2">
            {vhosts.map(vhost => {
              const vhostKey = `${vhost.connection_id}-${vhost.name}`;
              const isExpanded = expandedVHosts[vhostKey];
              const streamCount = vhost.streams ? vhost.streams.length : 0;

              return (
                <div key={vhostKey} className="mb-1">
                  <button
                    onClick={() => toggleVHost(vhostKey)}
                    className="w-full px-4 py-3 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors group"
                  >
                    <div className="flex items-center gap-3 flex-1 min-w-0">
                      <ChevronRight
                        className={`w-4 h-4 text-gray-400 transition-transform flex-shrink-0 ${
                          isExpanded ? 'rotate-90' : ''
                        }`}
                      />
                      <Database className="w-4 h-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
                      <span className="font-medium text-gray-900 dark:text-gray-100 truncate">
                        {vhost.name}
                      </span>
                    </div>
                    <span className="text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded-full ml-2">
                      {streamCount}
                    </span>
                  </button>

                  {isExpanded && (
                    <div className="ml-4 border-l-2 border-gray-200 dark:border-gray-800">
                      {streamCount === 0 ? (
                        <div className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                          No streams
                        </div>
                      ) : (
                        vhost.streams.map(stream => {
                          const isSelected =
                            selectedStream?.connection_id === stream.connection_id &&
                            selectedStream?.name === stream.name &&
                            selectedStream?.vhost === stream.vhost;

                          return (
                            <button
                              key={`${stream.connection_id}-${stream.vhost}-${stream.name}`}
                              onClick={() => onStreamSelect(stream)}
                              className={`w-full px-4 py-3 text-left transition-all ${
                                isSelected
                                  ? 'bg-blue-50 dark:bg-blue-900/20 border-l-2 border-blue-500'
                                  : 'hover:bg-gray-50 dark:hover:bg-gray-800 border-l-2 border-transparent'
                              }`}
                            >
                              <div className="flex items-center gap-3">
                                <Folder
                                  className={`w-4 h-4 flex-shrink-0 ${
                                    isSelected
                                      ? 'text-blue-600 dark:text-blue-400'
                                      : 'text-gray-400'
                                  }`}
                                />
                                <span
                                  className={`text-sm truncate ${
                                    isSelected
                                      ? 'text-blue-700 dark:text-blue-300 font-medium'
                                      : 'text-gray-700 dark:text-gray-300'
                                  }`}
                                >
                                  {stream.name}
                                </span>
                              </div>
                            </button>
                          );
                        })
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
