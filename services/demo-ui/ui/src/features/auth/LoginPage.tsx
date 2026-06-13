import { Building2 } from "lucide-react";
import { useAuth } from "../../auth/context";
import "./LoginPage.css";

export function LoginPage() {
    const { loginWithRedirect, enabled } = useAuth();

    return (
        <div className="login">
            <div className="login-card">
                <div className="login-card__glow" />

                <div className="login-logo">
                    <Building2 size={36} />
                </div>

                <h1 className="login-title">TMF Demo</h1>
                <p className="login-subtitle">
                    Secure dashboard for Customer &amp; Party Management.
                    <br />
                    Please authenticate to continue.
                </p>

                <button
                    disabled={!enabled}
                    onClick={() => loginWithRedirect({
                        authorizationParams: { screen_hint: 'login' }
                    })}
                    className="btn btn-primary btn--block btn--lg"
                >
                    {enabled ? 'Log In' : 'Auth Requires HTTPS or localhost'}
                </button>

                {!enabled && (
                    <p className="login-note">
                        Auth0 browser login requires a secure origin. Open this app via HTTPS or localhost.
                    </p>
                )}
            </div>
        </div>
    );
}
