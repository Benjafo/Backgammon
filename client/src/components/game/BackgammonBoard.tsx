import type { GameState, LegalMove } from "@/types/game";

interface BackgammonBoardProps {
    gameState: GameState;
    myColor: "white" | "black";
    isMyTurn: boolean;
    legalMoves: LegalMove[];
    selectedPoint: number | null;
    onPointClick: (point: number) => void;
}

export default function BackgammonBoard({
    gameState,
    myColor,
    isMyTurn,
    legalMoves,
    selectedPoint,
    onPointClick,
}: BackgammonBoardProps) {
    const BOARD_WIDTH = 800;
    const BOARD_HEIGHT = 600;
    const POINT_WIDTH = 50;
    const POINT_HEIGHT = 200;
    const CHECKER_RADIUS = 20;

    // Helper to get checker color for a point
    const getCheckerColor = (count: number): "white" | "black" | null => {
        if (count > 0) return "white";
        if (count < 0) return "black";
        return null;
    };

    // Helper to determine if a point is a valid destination
    const isValidDestination = (point: number): boolean => {
        if (!isMyTurn || selectedPoint === null) return false;
        return legalMoves.some((m) => m.fromPoint === selectedPoint && m.toPoint === point);
    };

    // Render a single triangular point
    const renderPoint = (pointNum: number, isTop: boolean, xPosition: number) => {
        const checkerCount = gameState.board[pointNum - 1];
        const checkerColor = getCheckerColor(checkerCount);
        const absCount = Math.abs(checkerCount);

        const isSelected = selectedPoint === pointNum;
        const isDestination = isValidDestination(pointNum);
        const hasMyChecker = (myColor === "white" && checkerCount > 0) || (myColor === "black" && checkerCount < 0);
        const isClickable = isMyTurn && hasMyChecker;

        // Calculate triangle points
        const y = isTop ? 50 : BOARD_HEIGHT - 50;
        const trianglePoints = isTop
            ? `${xPosition},${y} ${xPosition + POINT_WIDTH / 2},${y + POINT_HEIGHT} ${xPosition + POINT_WIDTH},${y}`
            : `${xPosition},${y} ${xPosition + POINT_WIDTH / 2},${y - POINT_HEIGHT} ${xPosition + POINT_WIDTH},${y}`;

        // Point color (alternating)
        const pointColor = pointNum % 2 === 0 ? "#8B4513" : "#D2691E";

        return (
            <g key={`point-${pointNum}`}>
                {/* Triangle */}
                <polygon
                    points={trianglePoints}
                    fill={isSelected ? "#FFD700" : isDestination ? "#90EE90" : pointColor}
                    stroke="#000"
                    strokeWidth="1"
                    onClick={() => isClickable && onPointClick(pointNum)}
                    style={{ cursor: isClickable ? "pointer" : "default" }}
                    opacity={isDestination ? 0.7 : 1}
                />

                {/* Checkers */}
                {checkerColor && absCount > 0 && (
                    <>
                        {Array.from({ length: Math.min(absCount, 5) }).map((_, i) => {
                            const cy = isTop ? y + 30 + i * (CHECKER_RADIUS * 2 + 2) : y - 30 - i * (CHECKER_RADIUS * 2 + 2);
                            return (
                                <circle
                                    key={`checker-${pointNum}-${i}`}
                                    cx={xPosition + POINT_WIDTH / 2}
                                    cy={cy}
                                    r={CHECKER_RADIUS}
                                    fill={checkerColor}
                                    stroke="#000"
                                    strokeWidth="2"
                                    onClick={() => isClickable && onPointClick(pointNum)}
                                    style={{ cursor: isClickable ? "pointer" : "default" }}
                                />
                            );
                        })}
                        {/* Show count if more than 5 */}
                        {absCount > 5 && (
                            <text
                                x={xPosition + POINT_WIDTH / 2}
                                y={isTop ? y + 30 + 4 * (CHECKER_RADIUS * 2 + 2) : y - 30 - 4 * (CHECKER_RADIUS * 2 + 2)}
                                textAnchor="middle"
                                fill={checkerColor === "white" ? "#000" : "#FFF"}
                                fontSize="14"
                                fontWeight="bold"
                            >
                                {absCount}
                            </text>
                        )}
                    </>
                )}

                {/* Point number */}
                <text
                    x={xPosition + POINT_WIDTH / 2}
                    y={isTop ? y - 5 : y + 15}
                    textAnchor="middle"
                    fill="#666"
                    fontSize="12"
                >
                    {pointNum}
                </text>
            </g>
        );
    };

    // Render bar checkers
    const renderBar = () => {
        const barX = BOARD_WIDTH / 2 - 25;
        const whiteCount = gameState.barWhite;
        const blackCount = gameState.barBlack;

        const whiteClickable = isMyTurn && myColor === "white" && whiteCount > 0;
        const blackClickable = isMyTurn && myColor === "black" && blackCount > 0;
        const barSelected = selectedPoint === 0;
        const barIsDestination = isValidDestination(0);

        return (
            <g>
                {/* Bar background */}
                <rect
                    x={barX}
                    y={50}
                    width={50}
                    height={BOARD_HEIGHT - 100}
                    fill={barSelected ? "#FFD700" : barIsDestination ? "#90EE90" : "#654321"}
                    stroke="#000"
                    strokeWidth="2"
                />

                <text x={barX + 25} y={30} textAnchor="middle" fill="#666" fontSize="12">
                    Bar
                </text>

                {/* White checkers on bar */}
                {whiteCount > 0 && (
                    <>
                        {Array.from({ length: Math.min(whiteCount, 3) }).map((_, i) => (
                            <circle
                                key={`bar-white-${i}`}
                                cx={barX + 25}
                                cy={BOARD_HEIGHT / 2 - 100 + i * (CHECKER_RADIUS * 2 + 2)}
                                r={CHECKER_RADIUS}
                                fill="white"
                                stroke="#000"
                                strokeWidth="2"
                                onClick={() => whiteClickable && onPointClick(0)}
                                style={{ cursor: whiteClickable ? "pointer" : "default" }}
                            />
                        ))}
                        {whiteCount > 3 && (
                            <text x={barX + 25} y={BOARD_HEIGHT / 2 - 40} textAnchor="middle" fontSize="14" fontWeight="bold">
                                {whiteCount}
                            </text>
                        )}
                    </>
                )}

                {/* Black checkers on bar */}
                {blackCount > 0 && (
                    <>
                        {Array.from({ length: Math.min(blackCount, 3) }).map((_, i) => (
                            <circle
                                key={`bar-black-${i}`}
                                cx={barX + 25}
                                cy={BOARD_HEIGHT / 2 + 40 + i * (CHECKER_RADIUS * 2 + 2)}
                                r={CHECKER_RADIUS}
                                fill="black"
                                stroke="#000"
                                strokeWidth="2"
                                onClick={() => blackClickable && onPointClick(0)}
                                style={{ cursor: blackClickable ? "pointer" : "default" }}
                            />
                        ))}
                        {blackCount > 3 && (
                            <text x={barX + 25} y={BOARD_HEIGHT / 2 + 120} textAnchor="middle" fill="#FFF" fontSize="14" fontWeight="bold">
                                {blackCount}
                            </text>
                        )}
                    </>
                )}
            </g>
        );
    };

    // Render borne-off areas
    const renderBorneOff = () => {
        return (
            <g>
                {/* White borne off (right side) */}
                <rect x={BOARD_WIDTH - 80} y={BOARD_HEIGHT / 2 - 60} width={70} height={120} fill="#E0E0E0" stroke="#000" strokeWidth="2" rx="5" />
                <text x={BOARD_WIDTH - 45} y={BOARD_HEIGHT / 2 - 70} textAnchor="middle" fill="#666" fontSize="12">
                    White Off
                </text>
                <text x={BOARD_WIDTH - 45} y={BOARD_HEIGHT / 2} textAnchor="middle" fontSize="24" fontWeight="bold">
                    {gameState.bornedOffWhite}
                </text>

                {/* Black borne off (left side) */}
                <rect x={10} y={BOARD_HEIGHT / 2 - 60} width={70} height={120} fill="#E0E0E0" stroke="#000" strokeWidth="2" rx="5" />
                <text x={45} y={BOARD_HEIGHT / 2 - 70} textAnchor="middle" fill="#666" fontSize="12">
                    Black Off
                </text>
                <text x={45} y={BOARD_HEIGHT / 2} textAnchor="middle" fontSize="24" fontWeight="bold">
                    {gameState.bornedOffBlack}
                </text>
            </g>
        );
    };

    // Calculate x positions for points
    const getPointX = (pointNum: number): number => {
        const leftMargin = 100;
        const barWidth = 50;

        if (pointNum >= 13 && pointNum <= 18) {
            // Right quadrant, top
            return leftMargin + (18 - pointNum) * POINT_WIDTH;
        } else if (pointNum >= 19 && pointNum <= 24) {
            // Right quadrant, bottom
            return leftMargin + (24 - pointNum) * POINT_WIDTH;
        } else if (pointNum >= 7 && pointNum <= 12) {
            // Left quadrant, top
            return leftMargin + barWidth + (12 - pointNum) * POINT_WIDTH;
        } else {
            // Left quadrant, bottom (1-6)
            return leftMargin + barWidth + (6 - pointNum) * POINT_WIDTH;
        }
    };

    return (
        <svg width={BOARD_WIDTH} height={BOARD_HEIGHT} className="border-2 border-gray-800 rounded-lg bg-amber-100">
            {/* Board background */}
            <rect x="0" y="0" width={BOARD_WIDTH} height={BOARD_HEIGHT} fill="#DEB887" />

            {/* Borne off areas */}
            {renderBorneOff()}

            {/* Top points (13-24) */}
            {[13, 14, 15, 16, 17, 18].map((p) => renderPoint(p, true, getPointX(p)))}
            {[19, 20, 21, 22, 23, 24].map((p) => renderPoint(p, true, getPointX(p)))}

            {/* Bottom points (1-12) */}
            {[7, 8, 9, 10, 11, 12].map((p) => renderPoint(p, false, getPointX(p)))}
            {[1, 2, 3, 4, 5, 6].map((p) => renderPoint(p, false, getPointX(p)))}

            {/* Bar */}
            {renderBar()}

            {/* Dice display */}
            {gameState.diceRoll && (
                <g>
                    <text x={BOARD_WIDTH / 2} y={BOARD_HEIGHT - 20} textAnchor="middle" fontSize="18" fontWeight="bold">
                        Dice: {gameState.diceRoll[0]}
                        {gameState.diceUsed && gameState.diceUsed[0] && " (used)"}, {gameState.diceRoll[1]}
                        {gameState.diceUsed && gameState.diceUsed[1] && " (used)"}
                    </text>
                </g>
            )}
        </svg>
    );
}
