import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";

export default function LobbyPage() {
    const { user, logout } = useAuth();

    return (
        <div className="min-h-screen bg-background p-8">
            <div className="max-w-7xl mx-auto">
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h1 className="text-3xl font-bold">Backgammon Lobby</h1>
                        <p className="text-muted-foreground">Welcome, {user?.username}!</p>
                    </div>
                    <Button onClick={logout} variant="outline">
                        Logout
                    </Button>
                </div>

                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    <Card>
                        <CardHeader>
                            <CardTitle>Quick Match</CardTitle>
                            <CardDescription>Find a random opponent</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Button className="w-full">Find Game</Button>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Active Games</CardTitle>
                            <CardDescription>Your ongoing matches</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground">No active games</p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Online Players</CardTitle>
                            <CardDescription>Players in the lobby</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground">Loading...</p>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
