import { Building2 } from "lucide-react";
import { useAuth } from "../../auth/context";

export function LoginPage() {
    const { loginWithRedirect, enabled } = useAuth();

    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800 p-4 font-sans">
            <div className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl shadow-2xl max-w-md w-full p-8 text-center text-white relative overflow-hidden">
                {/* Decorative background glow */}
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-40 h-40 bg-blue-500/20 rounded-full blur-3xl -z-10"></div>

                <div className="flex justify-center mb-8">
                    <div className="p-4 bg-gradient-to-tr from-blue-500 to-indigo-600 rounded-2xl shadow-lg ring-1 ring-white/20">
                        <Building2 size={40} className="text-white" />
                    </div>
                </div>

                <h1 className="text-3xl font-bold mb-3 tracking-tight">TMF Demo</h1>
                <p className="text-slate-400 mb-8 leading-relaxed">
                    Secure Dashboard for Customer & Party Management.
                    <br />
                    Please authenticate to continue.
                </p>

                <button
                    disabled={!enabled}
                    onClick={() => loginWithRedirect({
                        authorizationParams: { screen_hint: 'login' }
                    })}
                    className="w-full bg-blue-600 hover:bg-blue-500 disabled:bg-slate-600 disabled:cursor-not-allowed text-white font-semibold py-3.5 px-6 rounded-xl transition-all duration-200 shadow-lg shadow-blue-900/40 hover:shadow-blue-600/30 hover:-translate-y-0.5 active:translate-y-0"
                >
                    {enabled ? 'Log In' : 'Auth Requires HTTPS or localhost'}
                </button>

                {!enabled && (
                    <p className="mt-4 text-xs text-slate-400">
                        Auth0 browser login requires a secure origin. Open this app via HTTPS or localhost.
                    </p>
                )}
            </div>
        </div>
    )
}
