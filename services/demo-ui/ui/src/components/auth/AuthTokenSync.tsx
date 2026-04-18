import { useEffect } from "react";
import { setAuthToken } from "../../api/client";
import { useAuth } from "../../auth/context";

export const AuthTokenSync = () => {
    const { getAccessTokenSilently, isAuthenticated } = useAuth();

    useEffect(() => {
        const getToken = async () => {
            if (isAuthenticated) {
                try {
                    const token = await getAccessTokenSilently();
                    setAuthToken(token);
                } catch (error) {
                    console.error("Error getting auth token", error);
                    setAuthToken(null);
                }
            } else {
                setAuthToken(null);
            }
        };
        getToken();
    }, [getAccessTokenSilently, isAuthenticated]);

    return null;
};
