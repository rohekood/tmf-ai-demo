import { useAuth } from "../../auth/context";

export const LoginButton = () => {
    const { loginWithRedirect, enabled } = useAuth();

    return (
        <button
            disabled={!enabled}
            onClick={() => loginWithRedirect({
                authorizationParams: {
                    screen_hint: 'login'
                }
            })}
            className="btn btn-primary btn--block"
        >
            Log In
        </button>
    );
};
