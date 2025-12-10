interface DiceProps {
    value: number;
    size?: number;
    used?: boolean;
    className?: string;
}

export function Dice({ value, size = 50, used = false, className = "" }: DiceProps) {
    // Dice pip positions for each value (1-6)
    const getPipPositions = (val: number): Array<{ x: number; y: number }> => {
        const positions: Record<number, Array<{ x: number; y: number }>> = {
            1: [{ x: 0.5, y: 0.5 }],
            2: [
                { x: 0.25, y: 0.25 },
                { x: 0.75, y: 0.75 },
            ],
            3: [
                { x: 0.25, y: 0.25 },
                { x: 0.5, y: 0.5 },
                { x: 0.75, y: 0.75 },
            ],
            4: [
                { x: 0.25, y: 0.25 },
                { x: 0.75, y: 0.25 },
                { x: 0.25, y: 0.75 },
                { x: 0.75, y: 0.75 },
            ],
            5: [
                { x: 0.25, y: 0.25 },
                { x: 0.75, y: 0.25 },
                { x: 0.5, y: 0.5 },
                { x: 0.25, y: 0.75 },
                { x: 0.75, y: 0.75 },
            ],
            6: [
                { x: 0.25, y: 0.25 },
                { x: 0.75, y: 0.25 },
                { x: 0.25, y: 0.5 },
                { x: 0.75, y: 0.5 },
                { x: 0.25, y: 0.75 },
                { x: 0.75, y: 0.75 },
            ],
        };
        return positions[val] || [];
    };

    const pips = getPipPositions(value);
    const pipRadius = size * 0.1;

    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 100 100"
            className={className}
            style={{
                filter: used ? "grayscale(0.5) opacity(0.6)" : "drop-shadow(2px 2px 3px rgba(0, 0, 0, 0.4))",
            }}
        >
            {/* Dice background with rounded corners */}
            <rect
                x="2"
                y="2"
                width="96"
                height="96"
                rx="15"
                ry="15"
                fill={used ? "#666" : "url(#diceGradient)"}
                stroke={used ? "#444" : "#d4af37"}
                strokeWidth="2"
            />

            {/* Gradient definition for shiny dice effect */}
            <defs>
                <linearGradient id="diceGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stopColor="#ffffff" />
                    <stop offset="50%" stopColor="#f5f5f5" />
                    <stop offset="100%" stopColor="#e0e0e0" />
                </linearGradient>
            </defs>

            {/* Pips (dots) */}
            {pips.map((pip, index) => (
                <circle
                    key={index}
                    cx={pip.x * 100}
                    cy={pip.y * 100}
                    r={pipRadius}
                    fill={used ? "#333" : "#000"}
                />
            ))}
        </svg>
    );
}

interface DiceDisplayProps {
    dice: number[];
    used?: boolean[];
    size?: number;
    className?: string;
}

export function DiceDisplay({ dice, used = [], size = 50, className = "" }: DiceDisplayProps) {
    // Check if we have doubles (all dice are the same value)
    const isDoubles = dice.length > 2 && dice.every((die) => die === dice[0]);

    // For doubles, only show 2 dice
    const displayDice = isDoubles ? dice.slice(0, 2) : dice;
    const displayUsed = isDoubles ? used.slice(0, 2) : used;

    return (
        <div className={`flex gap-2 items-center ${className}`}>
            <div className="flex gap-2">
                {displayDice.map((die, index) => (
                    <Dice key={index} value={die} size={size} used={displayUsed[index] || false} />
                ))}
            </div>
            {isDoubles && (
                <span className="inline-flex items-center justify-center px-2 py-1 text-xs font-bold text-white bg-gold rounded-full border border-gold-light">
                    ×2
                </span>
            )}
        </div>
    );
}
