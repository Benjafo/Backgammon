import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/contexts/AuthContext";
import { formatAuthError, type FormattedError } from "@/lib/errors";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

export default function RegisterPage() {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<FormattedError | null>(null);
    const [loading, setLoading] = useState(false);
    const [token, setToken] = useState<string | null>(null);
    const [tokenError, setTokenError] = useState<FormattedError | null>(null);
    const { register, fetchRegistrationToken } = useAuth();
    const navigate = useNavigate();

    useEffect(() => {
        const getToken = async () => {
            try {
                const newToken = await fetchRegistrationToken();
                setToken(newToken);
            } catch (err) {
                const formattedError = formatAuthError(err);
                setTokenError(formattedError);
                console.error("Failed to fetch registration token:", err);
            }
        };

        getToken();
    }, [fetchRegistrationToken]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        if (!token) {
            setError({
                message: "Registration token not loaded",
                type: "server",
                suggestion: "Please refresh the page"
            });
            return;
        }

        setLoading(true);

        try {
            await register({ username, password, token });
            navigate("/lobby");
        } catch (err) {
            const formattedError = formatAuthError(err);
            setError(formattedError);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex items-center justify-center min-h-screen bg-felt felt-texture">
            <Card className="w-full max-w-md bg-black/60 backdrop-blur-sm border-2 border-gold shadow-lg">
                <CardHeader>
                    <CardTitle className="font-display text-2xl text-gold-light">
                        Create an Account
                    </CardTitle>
                </CardHeader>
                <form onSubmit={handleSubmit}>
                    <CardContent className="space-y-4">
                        {tokenError && (
                            <div className="p-3 rounded-lg bg-destructive/10 border border-destructive/50 space-y-1">
                                <p className="text-sm font-semibold text-destructive">
                                    {tokenError.message}
                                </p>
                                {tokenError.suggestion && (
                                    <p className="text-xs text-destructive/80">
                                        {tokenError.suggestion}
                                    </p>
                                )}
                            </div>
                        )}
                        {!token && !tokenError && (
                            <p className="text-sm text-muted-foreground">
                                Loading registration form...
                            </p>
                        )}
                        <div className="space-y-2">
                            <Label htmlFor="username">Username</Label>
                            <Input
                                id="username"
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                placeholder="Choose a username"
                                required
                                disabled={loading || !token}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="password">Password</Label>
                            <Input
                                id="password"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="Create a password"
                                required
                                disabled={loading || !token}
                            />
                        </div>
                        {error && (
                            <div className="p-3 rounded-lg bg-destructive/10 border border-destructive/50 space-y-1">
                                <p className="text-sm font-semibold text-destructive">
                                    {error.message}
                                </p>
                                {error.suggestion && (
                                    <p className="text-xs text-destructive/80">
                                        {error.suggestion}
                                    </p>
                                )}
                            </div>
                        )}
                    </CardContent>
                    <CardFooter className="flex flex-col space-y-4">
                        <Button
                            type="submit"
                            variant="casino"
                            className="w-full"
                            disabled={loading || !token}
                        >
                            {!token ? "Loading..." : loading ? "Creating account..." : "Register"}
                        </Button>
                        <p className="text-sm text-center text-muted-foreground">
                            Already have an account?{" "}
                            <Link
                                to="/login"
                                className="text-gold-light hover:underline font-semibold"
                            >
                                Login
                            </Link>
                        </p>
                    </CardFooter>
                </form>
            </Card>
        </div>
    );
}
