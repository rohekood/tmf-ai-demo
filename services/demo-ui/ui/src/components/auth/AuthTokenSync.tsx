import { useAuth0 } from "@auth0/auth0-react";
import { useEffect } from "react";
import { setAuthToken } from "../../api/client";

export const AuthTokenSync = () => {
    const { getAccessTokenSilently, isAuthenticated } = useAuth0();

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
