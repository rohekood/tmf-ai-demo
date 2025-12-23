import axios from 'axios';

// The BFF URL. In local dev with Docker, the browser accesses Nginx on localhost:80.
// But the BFF is on localhost:8080. 
// If running via clean docker-compose, both are exposed.
// We should check env var or default to locahost:8080.
const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const apiClient = axios.create({
    baseURL: baseURL,
    withCredentials: true, // Send cookies for Auth
    headers: {
        'Content-Type': 'application/json',
    },
});

// Response interceptor for global error handling
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        // Handle 401 (Unauthorized) -> Redirect to login?
        if (error.response?.status === 401) {
            console.warn('Unauthorized access. Redirecting to login...');
            // In a real app with Okta BFF, 401 might trigger a redirect flow
        }
        return Promise.reject(error);
    }
);
