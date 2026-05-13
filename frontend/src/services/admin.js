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

export const adminAPI = {
  // Equipment endpoints
  getAllEquipment: () => api.get('/admin/equipment'),
  createEquipment: (data) => api.post('/admin/equipment', data),
  updateEquipment: (id, data) => api.put(`/admin/equipment/${id}`, data),
  deleteEquipment: (id) => api.delete(`/admin/equipment/${id}`),

  // Monster endpoints
  getAllMonsters: () => api.get('/admin/monsters'),
  createMonster: (data) => api.post('/admin/monsters', data),
  updateMonster: (id, data) => api.put(`/admin/monsters/${id}`, data),
  deleteMonster: (id) => api.delete(`/admin/monsters/${id}`),
};


