import { useState } from 'react';
import { Sun, Moon, Database, BarChart3, MessageSquare } from 'lucide-react';
import { ThemeProvider, useTheme } from './context/ThemeContext';
import Sidebar from './components/Sidebar';
import StreamDetails from './components/StreamDetails';
import MessageBrowser from './components/MessageBrowser';

function AppContent() {
  const { theme, toggleTheme } = useTheme();
  const [selectedStream, setSelectedStream] = useState(null);
  const [view, setView] = useState('details');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  const handleStreamSelect = (stream) => {
    setSelectedStream(stream);
    setView('details');
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50 dark:bg-gray-950 text-gray-900 dark:text-gray-100">
      {/* Fixed Header */}
      <header className="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4 shadow-sm z-10">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
              <Database className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                RabbitMQ Stream Viewer
              </h1>
              <p className="text-xs text-gray-500 dark:text-gray-400">
                Browse and inspect stream messages
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {selectedStream && (
              <div className="flex items-center gap-2 bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
                <button
                  onClick={() => setView('details')}
                  className={`px-3 md:px-4 py-2 rounded-md text-xs md:text-sm font-medium transition-all ${
                    view === 'details'
                      ? 'bg-white dark:bg-gray-700 text-blue-600 dark:text-blue-400 shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'
                  }`}
                >
                  <BarChart3 className="w-4 h-4 inline-block mr-1 md:mr-1.5" />
                  <span className="hidden sm:inline">Statistics</span>
                  <span className="sm:hidden">Stats</span>
                </button>
                <button
                  onClick={() => setView('messages')}
                  className={`px-3 md:px-4 py-2 rounded-md text-xs md:text-sm font-medium transition-all ${
                    view === 'messages'
                      ? 'bg-white dark:bg-gray-700 text-blue-600 dark:text-blue-400 shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'
                  }`}
                >
                  <MessageSquare className="w-4 h-4 inline-block mr-1 md:mr-1.5" />
                  Messages
                </button>
              </div>
            )}

            <button
              onClick={toggleTheme}
              className="p-2.5 rounded-lg bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
            >
              {theme === 'dark' ? (
                <Sun className="w-5 h-5 text-yellow-500" />
              ) : (
                <Moon className="w-5 h-5 text-gray-700" />
              )}
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        <Sidebar
          onStreamSelect={handleStreamSelect}
          selectedStream={selectedStream}
          isCollapsed={sidebarCollapsed}
          onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        />

        <main className="flex-1 overflow-hidden">
          {!selectedStream ? (
            <div className="h-full flex items-center justify-center p-8">
              <div className="text-center max-w-md">
                <div className="w-20 h-20 mx-auto mb-6 bg-gradient-to-br from-blue-100 to-purple-100 dark:from-blue-900/20 dark:to-purple-900/20 rounded-2xl flex items-center justify-center">
                  <Database className="w-10 h-10 text-blue-600 dark:text-blue-400" />
                </div>
                <h2 className="text-2xl font-bold mb-3">Welcome to Stream Viewer</h2>
                <p className="text-gray-500 dark:text-gray-400 mb-6">
                  Select a stream from the sidebar to view statistics and browse messages
                </p>
                <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800 text-left shadow-sm">
                  <h3 className="font-semibold mb-3 text-gray-900 dark:text-gray-100">
                    Features
                  </h3>
                  <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                    <li className="flex items-start gap-2">
                      <span className="text-green-500 mt-0.5">✓</span>
                      <span>Real-time stream statistics and metrics</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-green-500 mt-0.5">✓</span>
                      <span>Non-destructive message browsing</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-green-500 mt-0.5">✓</span>
                      <span>Inspect AMQP properties and content</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="text-green-500 mt-0.5">✓</span>
                      <span>Offset-based navigation</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          ) : view === 'details' ? (
            <StreamDetails stream={selectedStream} />
          ) : (
            <MessageBrowser stream={selectedStream} />
          )}
        </main>
      </div>
    </div>
  );
}

function App() {
  return (
    <ThemeProvider>
      <AppContent />
    </ThemeProvider>
  );
}

export default App;
