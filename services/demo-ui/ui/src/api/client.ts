import axios from 'axios';

// The BFF URL.
const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

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
