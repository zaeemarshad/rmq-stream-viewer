import { useState } from 'react';
import { Copy, Check, FileJson, Tag, Clock, HardDrive } from 'lucide-react';

export default function MessageViewer({ message, compact = false }) {
  const [copiedProps, setCopiedProps] = useState(false);
  const [copiedContent, setCopiedContent] = useState(false);

  const copyToClipboard = async (text, type = 'properties') => {
    try {
      await navigator.clipboard.writeText(text);
      if (type === 'properties') {
        setCopiedProps(true);
        setTimeout(() => setCopiedProps(false), 2000);
      } else {
        setCopiedContent(true);
        setTimeout(() => setCopiedContent(false), 2000);
      }
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  const formatData = (data) => {
    try {
      if (data && data.length > 0) {
        const text = new TextDecoder().decode(data);
        try {
          const json = JSON.parse(text);
          return JSON.stringify(json, null, 2);
        } catch {
          return text;
        }
      }
      return '';
    } catch {
      return String(data);
    }
  };

  const formattedData = formatData(message.data);
  const isJSON = (() => {
    try {
      JSON.parse(formattedData);
      return true;
    } catch {
      return false;
    }
  })();

  // Separate properties by type
  const {
    application_properties,
    message_annotations,
    delivery_annotations,
    footer,
    ...standardProperties
  } = message.properties;

  const renderPropertySection = (title, props, icon, colorClass) => {
    if (!props || (typeof props === 'object' && Object.keys(props).length === 0)) {
      return null;
    }

    return (
      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
          <div className="flex items-center gap-2">
            {icon}
            <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 uppercase tracking-wide">
              {title}
            </h4>
          </div>
        </div>
        <div className="p-4">
          <div className="space-y-3">
            {Object.entries(props).map(([key, value]) => (
              <div
                key={key}
                className="flex items-start gap-4 py-2 border-b border-gray-100 dark:border-gray-800 last:border-0"
              >
                <div className={`text-sm font-medium ${colorClass} w-1/3 flex-shrink-0 break-words`}>
                  {key}
                </div>
                <div className="text-sm text-gray-700 dark:text-gray-300 font-mono break-all flex-1">
                  {typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className={compact ? '' : 'h-full flex flex-col'}>
      {!compact && (
        <div className="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 p-6">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
              Message Details
            </h3>
            <button
              onClick={() => copyToClipboard(JSON.stringify(message.properties, null, 2), 'properties')}
              className="px-3 py-1.5 text-xs bg-blue-600 hover:bg-blue-700 rounded-lg text-white font-medium transition-colors inline-flex items-center gap-2"
            >
              {copiedProps ? (
                <>
                  <Check className="w-3 h-3" />
                  Copied All
                </>
              ) : (
                <>
                  <Copy className="w-3 h-3" />
                  Copy All Properties
                </>
              )}
            </button>
          </div>
          <div className="flex flex-wrap items-center gap-4 text-sm">
            <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
              <Tag className="w-4 h-4" />
              <span className="font-medium">Offset:</span>
              <span className="font-mono text-gray-900 dark:text-gray-100">
                {message.offset.toLocaleString()}
              </span>
            </div>
            {message.properties?.routing_key && (
              <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
                <Tag className="w-4 h-4" />
                <span className="font-medium">Routing Key:</span>
                <span className="font-mono text-gray-900 dark:text-gray-100">
                  {message.properties.routing_key}
                </span>
              </div>
            )}
            {message.properties?.subject && (
              <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
                <Tag className="w-4 h-4" />
                <span className="font-medium">Subject:</span>
                <span className="font-mono text-gray-900 dark:text-gray-100">
                  {message.properties.subject}
                </span>
              </div>
            )}
          </div>
        </div>
      )}

      <div className={`${compact ? 'p-4' : 'flex-1 overflow-y-auto p-6'} space-y-6`}>
        {/* Standard AMQP Properties */}
        {renderPropertySection(
          'AMQP Properties',
          standardProperties,
          <Tag className="w-5 h-5 text-blue-600 dark:text-blue-400" />,
          'text-blue-600 dark:text-blue-400'
        )}

        {/* Application Properties */}
        {renderPropertySection(
          'Application Properties',
          application_properties,
          <FileJson className="w-5 h-5 text-green-600 dark:text-green-400" />,
          'text-green-600 dark:text-green-400'
        )}

        {/* Message Annotations */}
        {renderPropertySection(
          'Message Annotations',
          message_annotations,
          <Tag className="w-5 h-5 text-purple-600 dark:text-purple-400" />,
          'text-purple-600 dark:text-purple-400'
        )}

        {/* Delivery Annotations */}
        {renderPropertySection(
          'Delivery Annotations',
          delivery_annotations,
          <Tag className="w-5 h-5 text-orange-600 dark:text-orange-400" />,
          'text-orange-600 dark:text-orange-400'
        )}

        {/* Footer */}
        {renderPropertySection(
          'Footer',
          footer,
          <Tag className="w-5 h-5 text-pink-600 dark:text-pink-400" />,
          'text-pink-600 dark:text-pink-400'
        )}

        {/* Message Content Section */}
        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
            <div className="flex items-center gap-2">
              <FileJson className="w-5 h-5 text-green-600 dark:text-green-400" />
              <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 uppercase tracking-wide">
                Message Content
              </h4>
              {isJSON && (
                <span className="px-2 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-medium rounded">
                  JSON
                </span>
              )}
            </div>
            <button
              onClick={() => copyToClipboard(formattedData, 'content')}
              className="px-3 py-1.5 text-xs bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 rounded-lg text-gray-700 dark:text-gray-300 font-medium transition-colors inline-flex items-center gap-2"
            >
              {copiedContent ? (
                <>
                  <Check className="w-3 h-3" />
                  Copied!
                </>
              ) : (
                <>
                  <Copy className="w-3 h-3" />
                  Copy
                </>
              )}
            </button>
          </div>
          <div className="p-6 bg-gray-50 dark:bg-gray-950">
            <pre className="text-sm text-gray-700 dark:text-gray-300 font-mono overflow-x-auto">
              {formattedData || '(empty)'}
            </pre>
          </div>
        </div>

        {/* Metadata Section */}
        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-6 shadow-sm">
          <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 uppercase tracking-wide mb-4">
            Metadata
          </h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <Clock className="w-4 h-4 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">Timestamp</p>
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {new Date(message.timestamp).toLocaleString()}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                <HardDrive className="w-4 h-4 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">Size</p>
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {message.data?.length || 0} bytes
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
