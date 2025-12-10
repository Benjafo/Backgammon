export type ErrorType = "validation" | "auth" | "network" | "server" | "unknown";

export interface FormattedError {
    message: string;
    type: ErrorType;
    suggestion?: string;
}

/**
 * Parse and format error messages for better user experience
 * Matches EXACT error messages from the backend
 */
export function formatAuthError(error: unknown): FormattedError {
    const originalMessage = error instanceof Error ? error.message : String(error);
    const errorMessage = originalMessage.toLowerCase();

    // ===== EXACT Username validation errors from backend =====
    if (errorMessage === "username must be at least 3 characters") {
        return {
            message: "Username is too short",
            type: "validation",
            suggestion: "Username must be at least 3 characters long",
        };
    }

    if (errorMessage === "username must not exceed 255 characters") {
        return {
            message: "Username is too long",
            type: "validation",
            suggestion: "Username must not exceed 255 characters",
        };
    }

    if (
        errorMessage === "username must contain only letters, numbers, underscores, and hyphens"
    ) {
        return {
            message: "Invalid username format",
            type: "validation",
            suggestion:
                "Username can only contain letters, numbers, underscores (_), and hyphens (-)",
        };
    }

    if (errorMessage === "username is reserved") {
        return {
            message: "This username is reserved",
            type: "validation",
            suggestion: "Please choose a different username",
        };
    }

    if (errorMessage === "username already exists") {
        return {
            message: "Username already taken",
            type: "validation",
            suggestion: "This username is already registered. Please choose another.",
        };
    }

    // ===== EXACT Password validation errors from backend =====
    if (errorMessage === "password must be at least 8 characters") {
        return {
            message: "Password is too short",
            type: "validation",
            suggestion: "Password must be at least 8 characters long",
        };
    }

    if (errorMessage === "password must not exceed 256 characters") {
        return {
            message: "Password is too long",
            type: "validation",
            suggestion: "Password must not exceed 256 characters",
        };
    }

    if (
        errorMessage ===
        "password must contain at least 2 of: uppercase letter, number, special character"
    ) {
        return {
            message: "Password is too weak",
            type: "validation",
            suggestion:
                "Password must contain at least 2 of: uppercase letter, number, or special character (!@#$%^&*)",
        };
    }

    // ===== Token errors =====
    if (errorMessage === "invalid or expired registration token") {
        return {
            message: "Your registration session has expired",
            type: "auth",
            suggestion: "Please refresh the page and try again",
        };
    }

    if (errorMessage.includes("invalid registration token:")) {
        return {
            message: "Invalid registration session",
            type: "auth",
            suggestion: "Please refresh the page and try again",
        };
    }

    if (errorMessage === "registration token is required") {
        return {
            message: "Unable to load registration form",
            type: "server",
            suggestion: "Please refresh the page",
        };
    }

    // ===== Login errors =====
    if (errorMessage === "invalid credentials") {
        return {
            message: "Invalid username or password",
            type: "auth",
            suggestion: "Double-check your credentials and try again",
        };
    }

    // ===== General errors =====
    if (errorMessage === "username and password are required") {
        return {
            message: "Please fill in all required fields",
            type: "validation",
            suggestion: "Both username and password are required",
        };
    }

    if (errorMessage === "failed to create account") {
        return {
            message: "Registration failed",
            type: "server",
            suggestion: "Please try again in a few moments",
        };
    }

    if (errorMessage === "registration successful but login failed") {
        return {
            message: "Account created successfully, but automatic login failed",
            type: "server",
            suggestion: "Please go to the login page to sign in",
        };
    }

    if (errorMessage === "login failed") {
        return {
            message: "Login failed due to a server error",
            type: "server",
            suggestion: "Please try again",
        };
    }

    if (errorMessage === "database not initialized") {
        return {
            message: "Server is temporarily unavailable",
            type: "server",
            suggestion: "Please try again in a few moments",
        };
    }

    if (errorMessage.includes("failed to") || errorMessage.includes("server")) {
        return {
            message: "Server error",
            type: "server",
            suggestion: "Please try again in a few moments",
        };
    }

    // Network errors
    if (
        errorMessage.includes("network") ||
        errorMessage.includes("fetch") ||
        errorMessage.includes("connection")
    ) {
        return {
            message: "Network connection error",
            type: "network",
            suggestion: "Check your internet connection and try again",
        };
    }

    // Default - return the original message
    return {
        message: originalMessage || "An unexpected error occurred",
        type: "unknown",
        suggestion: "Please try again",
    };
}

/**
 * Get icon for error type
 */
export function getErrorIcon(type: ErrorType): string {
    switch (type) {
        case "validation":
            return "⚠️";
        case "auth":
            return "🔒";
        case "network":
            return "📡";
        case "server":
            return "🔧";
        default:
            return "❌";
    }
}
