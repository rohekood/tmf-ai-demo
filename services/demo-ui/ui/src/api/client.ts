import axios from 'axios';
import { getRuntimeConfig } from '../config/runtime';

// The BFF URL.
const baseURL = getRuntimeConfig().apiUrl || '';

let authToken: string | null = null;

export const setAuthToken = (token: string | null) => {
    authToken = token;
};

export const apiClient = axios.create({
    baseURL: baseURL,
    withCredentials: true, // Send cookies for Auth (if needed along with headers)
    headers: {
        'Content-Type': 'application/json',
    },
});

apiClient.interceptors.request.use(
    (config) => {
        if (typeof config.url === 'string' && config.url.startsWith('/api/api/')) {
            config.url = config.url.replace('/api/api/', '/api/');
        }
        if (authToken) {
            config.headers.Authorization = `Bearer ${authToken}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Response interceptor for global error handling
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        // Handle 401 (Unauthorized)
        if (error.response?.status === 401) {
            console.warn('Unauthorized access. Token might be expired or invalid.');
        }
        return Promise.reject(error);
    }
);
