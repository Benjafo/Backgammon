import { useAuth } from "@/contexts/AuthContext";
import { Navigate } from "react-router-dom";

export function PublicRoute({ children }: { children: React.ReactNode }) {
    const { user, loading } = useAuth();

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <p className="text-muted-foreground">Loading...</p>
            </div>
        );
    }

    if (user) {
        return <Navigate to="/lobby" replace />;
    }

    return <>{children}</>;
}
