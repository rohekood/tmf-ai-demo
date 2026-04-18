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
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-slate-500 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
            Log In
        </button>
    );
};
