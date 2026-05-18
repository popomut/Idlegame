import axios from 'axios';
import { API_BASE_URL } from './api.js';

const api = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  config.withCredentials = true;
  return config;
});

export const miningAPI = {
  // Get player's mining skill progression
  getMiningSkill: () => api.get('/mining/skill'),
  
  // Start mining session
  startMining: (oreId) => api.post('/mining/start', { ore_id: oreId }),
  
  // Stop mining session
  stopMining: () => api.post('/mining/stop'),
  
  // Get mining status (active session, current ores)
  getMiningStatus: () => api.get('/mining/status'),
};
