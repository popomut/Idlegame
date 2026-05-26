import axios from 'axios';

const API_ROOT =
  import.meta.env.VITE_API_BASE_URL ||
  (import.meta.env.DEV
    ? ''
    : '');
export const API_BASE_URL = API_ROOT.replace(/\/$/, '');
const API_BASE = `${API_BASE_URL}/api`;

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
  getOreTypes: () => axios.get(`${API_BASE_URL}/api/ore-types`), // public — no auth cookie needed
};

export const userAPI = {
  getUser: () => api.get('/user'),
  updateUser: (playerName, playerClass) =>
    api.post('/user/update', { player_name: playerName, player_class: playerClass }),
};

export const characterAPI = {
  getCharacter: () => api.get('/character'),
  heal: (hp) => api.post('/character/heal', { hp }),
};

export const mapAPI = {
  getContinents: () => axios.get(`${API_BASE_URL}/api/map/continents`), // public
  enterArea:     (areaKey) => api.post('/map/enter', { area_key: areaKey }),
  getSession:    () => api.get('/map/session'),
  resume:        () => api.post('/map/resume'),
  advance:       () => api.post('/map/advance'),
  flee:          () => api.post('/map/flee'),
};

export const equipmentAPI = {
  getTypes:   () => axios.get(`${API_BASE_URL}/api/equipment/types`), // public
  getBag:     () => api.get('/equipment/bag'),
  getSlots:   () => api.get('/equipment/slots'),
  equip:      (userEquipmentId, slot) => api.post('/equipment/equip', { user_equipment_id: userEquipmentId, slot }),
  unequip:    (slot) => api.post('/equipment/unequip', { slot }),
  give:       (equipmentKey) => api.post('/equipment/give', { equipment_key: equipmentKey }),
};

export const combatAPI = {
  startCombat:  (zoneKey) => api.post('/combat/start', { zone_key: zoneKey }),
  getStatus:    () => api.get('/combat/status'),
  flee:         () => api.post('/combat/flee'),
};
