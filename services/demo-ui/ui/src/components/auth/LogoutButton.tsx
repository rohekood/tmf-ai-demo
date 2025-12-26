import { useAuth0 } from "@auth0/auth0-react";
import { LogOut } from "lucide-react";

export const LogoutButton = () => {
    const { logout } = useAuth0();

    return (
        <button
            onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}
            className="sidebar-logout-btn"
            title="Log Out"
            aria-label="Log out"
        >
            <LogOut size={16} />
        </button>
    );
};
