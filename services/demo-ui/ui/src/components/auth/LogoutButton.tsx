import { LogOut } from "lucide-react";
import { useAuth } from "../../auth/context";

export const LogoutButton = () => {
    const { logout } = useAuth();

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
