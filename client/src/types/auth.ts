export interface User {
    id: number;
    username: string;
}

export interface LoginRequest {
    username: string;
    password: string;
}

export interface RegisterRequest {
    username: string;
    password: string;
    token: string;
}

export interface AuthResponse {
    user: User;
    message?: string;
}
