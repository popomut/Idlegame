import axios from 'axios';

const API_BASE = 'http://localhost:3000/api';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Always send cookies
});

// Ensure credentials are sent with every request
api.interceptors.request.use((config) => {
  // Explicitly enable credentials for this request
  config.withCredentials = true;
  return config;
});

// Better error handling for auth issues
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      console.error('Auth failed (401): Token invalid or expired. Please log in again.');
    }
    return Promise.reject(error);
  }
);

export const authAPI = {
  register: (username, email, password) =>
    api.post('/auth/register', { username, email, password }),
  login: (username, password) =>
    api.post('/auth/login', { username, password }),
  guestLogin: () =>
    api.post('/auth/guest', {}),
};

export const miningAPI = {
  startMining: (oreId) => api.post('/mining/start', { ore_id: oreId }),
  stopMining: () => api.post('/mining/stop'),
  getMiningStatus: () => api.get('/mining/status'),
};

export const inventoryAPI = {
  getOreInventory: () => api.get('/inventory/ores'),
};

export const userAPI = {
  getUser: () => api.get('/user'),
  updateUser: (playerName, playerClass) =>
    api.post('/user/update', { player_name: playerName, player_class: playerClass }),
};
