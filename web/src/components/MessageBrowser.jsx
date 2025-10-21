import { useState, useEffect } from 'react';
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  RefreshCw,
  ArrowRight,
  Loader2,
  AlertCircle,
  MessageSquare,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { api } from '../services/api';
import MessageViewer from './MessageViewer';

export default function MessageBrowser({ stream }) {
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [currentOffset, setCurrentOffset] = useState(0);
  const [limit, setLimit] = useState(100);
  const [expandedMessageOffset, setExpandedMessageOffset] = useState(null);
  const [offsetInput, setOffsetInput] = useState('0');
  const [offsetType, setOffsetType] = useState('first');
  const [timestampInput, setTimestampInput] = useState('');
  const [stats, setStats] = useState(null);

  useEffect(() => {
    loadStreamStats();
  }, [stream]);

  useEffect(() => {
    loadMessages();
  }, [stream, currentOffset, limit]);

  const loadStreamStats = async () => {
    try {
      const data = await api.getStreamStats(stream.connection_id, stream.vhost, stream.name);
      setStats(data);
      // Start from first offset by default
      setCurrentOffset(data.first_offset);
      setOffsetInput(String(data.first_offset));
    } catch (err) {
      console.error('Failed to load stream stats:', err);
    }
  };

  const loadMessages = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getMessages(stream.connection_id, stream.vhost, stream.name, currentOffset, limit);
      setMessages(data.messages || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleJumpToFirst = () => {
    if (stats) {
      setCurrentOffset(stats.first_offset);
      setOffsetInput(String(stats.first_offset));
    }
  };

  const handleJumpToLast = () => {
    if (stats) {
      const offset = Math.max(0, stats.last_offset - limit + 1);
      setCurrentOffset(offset);
      setOffsetInput(String(offset));
    }
  };

  const handlePrevious = () => {
    const newOffset = Math.max(stats?.first_offset || 0, currentOffset - limit);
    setCurrentOffset(newOffset);
    setOffsetInput(String(newOffset));
  };

  const handleNext = () => {
    const newOffset = currentOffset + limit;
    if (!stats || newOffset <= stats.last_offset) {
      setCurrentOffset(newOffset);
      setOffsetInput(String(newOffset));
    }
  };

  const handleOffsetSubmit = (e) => {
    e.preventDefault();
    
    if (offsetType === 'first' && stats) {
      setCurrentOffset(stats.first_offset);
      setOffsetInput(String(stats.first_offset));
    } else if (offsetType === 'last' && stats) {
      setCurrentOffset(Math.max(stats.first_offset, stats.last_offset - limit + 1));
      setOffsetInput(String(stats.last_offset));
    } else if (offsetType === 'offset') {
      const offset = parseInt(offsetInput);
      if (!isNaN(offset) && offset >= 0) {
        setCurrentOffset(offset);
      }
    } else if (offsetType === 'timestamp') {
      // For timestamp, we'd need backend support - for now, just use first offset
      // TODO: Implement timestamp-based offset lookup
      if (stats) {
        setCurrentOffset(stats.first_offset);
      }
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'ArrowLeft' && !e.target.matches('input')) {
      handlePrevious();
    } else if (e.key === 'ArrowRight' && !e.target.matches('input')) {
      handleNext();
    }
  };

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentOffset, limit, stats]);

  return (
    <div className="flex-1 flex flex-col bg-gray-50 dark:bg-gray-950">
      {/* Controls Bar */}
      <div className="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 p-4 shadow-sm">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div className="flex items-center gap-2">
            <button
              onClick={handleJumpToFirst}
              disabled={loading || !stats}
              className="px-3 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-gray-700 dark:text-gray-300 text-sm font-medium transition-colors inline-flex items-center gap-2"
              title="Jump to first messages"
            >
              <ChevronsLeft className="w-4 h-4" />
              First
            </button>
            <button
              onClick={handlePrevious}
              disabled={loading || !stats || currentOffset <= (stats?.first_offset || 0)}
              className="px-3 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-gray-700 dark:text-gray-300 text-sm font-medium transition-colors inline-flex items-center gap-2"
              title="Previous page"
            >
              <ChevronLeft className="w-4 h-4" />
              Previous
            </button>
            <button
              onClick={handleNext}
              disabled={loading || (stats && currentOffset >= stats.last_offset)}
              className="px-3 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-gray-700 dark:text-gray-300 text-sm font-medium transition-colors inline-flex items-center gap-2"
              title="Next page"
            >
              Next
              <ChevronRight className="w-4 h-4" />
            </button>
            <button
              onClick={handleJumpToLast}
              disabled={loading || !stats}
              className="px-3 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-gray-700 dark:text-gray-300 text-sm font-medium transition-colors inline-flex items-center gap-2"
              title="Jump to last messages"
            >
              Last
              <ChevronsRight className="w-4 h-4" />
            </button>
          </div>

          <form onSubmit={handleOffsetSubmit} className="flex items-center gap-2">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Start from:
            </label>
            <select
              value={offsetType}
              onChange={(e) => setOffsetType(e.target.value)}
              className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="first">First</option>
              <option value="last">Last</option>
              <option value="offset">Offset</option>
              <option value="timestamp">Timestamp</option>
            </select>
            {offsetType === 'offset' && (
              <input
                type="number"
                value={offsetInput}
                onChange={(e) => setOffsetInput(e.target.value)}
                className="w-32 px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                min="0"
                placeholder="Offset"
              />
            )}
            {offsetType === 'timestamp' && (
              <input
                type="datetime-local"
                value={timestampInput}
                onChange={(e) => setTimestampInput(e.target.value)}
                className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Timestamp"
              />
            )}
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white text-sm font-medium transition-colors inline-flex items-center gap-2"
            >
              Go
              <ArrowRight className="w-4 h-4" />
            </button>
          </form>

          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Limit:
            </label>
            <select
              value={limit}
              onChange={(e) => setLimit(Number(e.target.value))}
              className="px-3 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="10">10</option>
              <option value="25">25</option>
              <option value="50">50</option>
              <option value="100">100</option>
              <option value="200">200</option>
              <option value="500">500</option>
            </select>
          </div>

          <button
            onClick={loadMessages}
            disabled={loading}
            className="p-2.5 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors disabled:opacity-50"
            title="Refresh messages"
          >
            <RefreshCw className={`w-5 h-5 text-gray-700 dark:text-gray-300 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>

        {stats && (
          <div className="mt-3 text-xs text-gray-500 dark:text-gray-400">
            Range: {stats.first_offset.toLocaleString()} - {stats.last_offset.toLocaleString()} (
            {stats.message_count.toLocaleString()} messages)
          </div>
        )}
      </div>

      {/* Messages List */}
      <div className="flex-1 overflow-y-auto bg-gray-50 dark:bg-gray-950">
        <div className="max-w-7xl mx-auto p-4">
          <div className="mb-4">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
              Messages {messages.length > 0 && `(${messages.length})`}
            </h3>
          </div>
          <div className="space-y-2">
            {loading ? (
              <div className="p-8 text-center bg-white dark:bg-gray-900 rounded-lg">
                <Loader2 className="w-8 h-8 text-blue-500 animate-spin mx-auto mb-3" />
                <p className="text-sm text-gray-500 dark:text-gray-400">Loading messages...</p>
              </div>
            ) : error ? (
              <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <p className="font-semibold text-red-900 dark:text-red-100 text-sm">Error</p>
                    <p className="mt-1 text-sm text-red-700 dark:text-red-300">{error}</p>
                    <button
                      onClick={loadMessages}
                      className="mt-3 px-3 py-1.5 bg-red-600 hover:bg-red-700 rounded-lg text-white text-xs font-medium transition-colors"
                    >
                      Retry
                    </button>
                  </div>
                </div>
              </div>
            ) : messages.length === 0 ? (
              <div className="p-8 text-center bg-white dark:bg-gray-900 rounded-lg">
                <MessageSquare className="w-12 h-12 text-gray-300 dark:text-gray-700 mx-auto mb-3" />
                <p className="text-sm text-gray-600 dark:text-gray-400">No messages at this offset</p>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-500">
                  Try a different offset
                </p>
              </div>
            ) : (
              messages.map((msg, index) => {
                const isExpanded = expandedMessageOffset === msg.offset;
                const uuid = msg.properties?.message_id || '-';
                const routingKey = msg.properties?.routing_key || msg.properties?.subject || '-';
                
                return (
                  <div
                    key={`${msg.offset}-${index}`}
                    className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm hover:shadow-md transition-shadow"
                  >
                    {/* Message Summary - Always Visible */}
                    <button
                      onClick={() => setExpandedMessageOffset(isExpanded ? null : msg.offset)}
                      className="w-full p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                    >
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1 space-y-2">
                          <div className="grid grid-cols-2 gap-x-4 gap-y-2">
                            <div>
                              <span className="text-xs font-medium text-gray-500 dark:text-gray-400">UUID:</span>
                              <p className="text-sm font-mono text-gray-900 dark:text-gray-100 truncate">{uuid}</p>
                            </div>
                            <div>
                              <span className="text-xs font-medium text-gray-500 dark:text-gray-400">Routing Key:</span>
                              <p className="text-sm font-mono text-gray-900 dark:text-gray-100 truncate">{routingKey}</p>
                            </div>
                            <div>
                              <span className="text-xs font-medium text-gray-500 dark:text-gray-400">Timestamp:</span>
                              <p className="text-sm text-gray-900 dark:text-gray-100">
                                {new Date(msg.timestamp).toLocaleString()}
                              </p>
                            </div>
                            <div>
                              <span className="text-xs font-medium text-gray-500 dark:text-gray-400">Offset:</span>
                              <p className="text-sm font-mono text-gray-900 dark:text-gray-100">{msg.offset.toLocaleString()}</p>
                            </div>
                          </div>
                        </div>
                        <div className="flex-shrink-0">
                          {isExpanded ? (
                            <ChevronUp className="w-5 h-5 text-gray-400" />
                          ) : (
                            <ChevronDown className="w-5 h-5 text-gray-400" />
                          )}
                        </div>
                      </div>
                    </button>

                    {/* Expanded Details */}
                    {isExpanded && (
                      <div className="border-t border-gray-200 dark:border-gray-800">
                        <MessageViewer message={msg} compact={true} />
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
