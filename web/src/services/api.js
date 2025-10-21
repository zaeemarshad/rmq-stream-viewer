const API_BASE = '/api';

export const api = {
  async getConnections() {
    const response = await fetch(`${API_BASE}/connections`);
    if (!response.ok) {
      throw new Error('Failed to fetch connections');
    }
    return response.json();
  },

  async getVHosts() {
    const response = await fetch(`${API_BASE}/vhosts`);
    if (!response.ok) {
      throw new Error('Failed to fetch vhosts');
    }
    return response.json();
  },

  async getStreams() {
    const response = await fetch(`${API_BASE}/streams`);
    if (!response.ok) {
      throw new Error('Failed to fetch streams');
    }
    return response.json();
  },

  async getStreamStats(connectionId, vhost, streamName) {
    const response = await fetch(
      `${API_BASE}/streams/${encodeURIComponent(connectionId)}/${encodeURIComponent(vhost)}/${encodeURIComponent(streamName)}/stats`
    );
    if (!response.ok) {
      throw new Error('Failed to fetch stream stats');
    }
    return response.json();
  },

  async getMessages(connectionId, vhost, streamName, offset = 0, limit = 100) {
    const response = await fetch(
      `${API_BASE}/streams/${encodeURIComponent(connectionId)}/${encodeURIComponent(vhost)}/${encodeURIComponent(streamName)}/messages?offset=${offset}&limit=${limit}`
    );
    if (!response.ok) {
      throw new Error('Failed to fetch messages');
    }
    return response.json();
  },
};

