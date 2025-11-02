import type { GameState, LegalMove } from "@/types/game";
import { useState } from "react";

interface BackgammonBoardProps {
    gameState: GameState;
    myColor: "white" | "black";
    isMyTurn: boolean;
    legalMoves: LegalMove[];
    draggedPoint: number | null;
    onDragStart: (point: number) => void;
    onDrop: (point: number) => void;
}

export default function BackgammonBoard({
    gameState,
    myColor,
    isMyTurn,
    legalMoves,
    draggedPoint,
    onDragStart,
    onDrop,
}: BackgammonBoardProps) {
    const BOARD_WIDTH = 800;
    const BOARD_HEIGHT = 600;
    const POINT_WIDTH = 50;
    const POINT_HEIGHT = 200;
    const CHECKER_RADIUS = 20;

    const [isDragging, setIsDragging] = useState(false);
    const [dragX, setDragX] = useState(0);
    const [dragY, setDragY] = useState(0);

    // Handle mouse move on SVG - update dragged checker position
    const handleMouseMove = (e: React.MouseEvent<SVGSVGElement>) => {
        if (isDragging && draggedPoint !== null) {
            const svg = e.currentTarget;
            const rect = svg.getBoundingClientRect();
            setDragX(e.clientX - rect.left);
            setDragY(e.clientY - rect.top);
        }
    };

    // Handle mouse up on SVG - drop the checker
    const handleMouseUp = (e: React.MouseEvent<SVGSVGElement>) => {
        if (isDragging && draggedPoint !== null) {
            // Determine which point we're over and call onDrop
            const svg = e.currentTarget;
            const rect = svg.getBoundingClientRect();
            const mouseX = e.clientX - rect.left;
            const mouseY = e.clientY - rect.top;

            // Find which point is under the mouse
            const targetPoint = findPointAtPosition(mouseX, mouseY);
            if (targetPoint !== null) {
                onDrop(targetPoint);
            }

            setIsDragging(false);
        }
    };

    // Handle checker mousedown - start dragging
    const handleCheckerMouseDown = (point: number, x: number, y: number) => {
        onDragStart(point);
        setIsDragging(true);
        setDragX(x);
        setDragY(y);
    };

    // Find which point is at a given position
    const findPointAtPosition = (x: number, y: number): number | null => {
        // Check bear-off zones (point 25)
        // White bear-off should be near white's home (points 1-6, right side)
        // Black bear-off should be near black's home (points 19-24, left side)

        // Left side bear-off zone (for white)
        if (myColor === "white" && x >= 10 && x <= 80 && y >= BOARD_HEIGHT / 2 - 60 && y <= BOARD_HEIGHT / 2 + 60) {
            return 25;
        }
        // Right side bear-off zone (for black)
        if (myColor === "black" && x >= BOARD_WIDTH - 80 && x <= BOARD_WIDTH - 10 && y >= BOARD_HEIGHT / 2 - 60 && y <= BOARD_HEIGHT / 2 + 60) {
            return 25;
        }

        // Check bar
        const barX = BOARD_WIDTH / 2 - 25;
        if (x >= barX && x <= barX + 50 && y >= 50 && y <= BOARD_HEIGHT - 50) {
            return 0;
        }

        // Check all 24 points
        for (let pointNum = 1; pointNum <= 24; pointNum++) {
            const pointX = getPointX(pointNum);
            const isTop = pointNum >= 13;
            const pointY = isTop ? 50 : BOARD_HEIGHT - 50;

            // Check if mouse is within the triangle bounds
            if (x >= pointX && x <= pointX + POINT_WIDTH) {
                if (isTop && y >= pointY && y <= pointY + POINT_HEIGHT) {
                    return pointNum;
                } else if (!isTop && y >= pointY - POINT_HEIGHT && y <= pointY) {
                    return pointNum;
                }
            }
        }

        return null;
    };

    // Helper to get checker color for a point
    const getCheckerColor = (count: number): "white" | "black" | null => {
        if (count > 0) return "white";
        if (count < 0) return "black";
        return null;
    };

    // Helper to determine if a point is a valid destination
    const isValidDestination = (point: number): boolean => {
        if (!isMyTurn || draggedPoint === null) return false;
        return legalMoves.some((m) => m.fromPoint === draggedPoint && m.toPoint === point);
    };

    // Render a single triangular point
    const renderPoint = (pointNum: number, isTop: boolean, xPosition: number) => {
        const checkerCount = gameState.board[pointNum - 1];
        const checkerColor = getCheckerColor(checkerCount);
        const absCount = Math.abs(checkerCount);

        const isDragged = draggedPoint === pointNum;
        const isDestination = isValidDestination(pointNum);
        const hasMyChecker =
            (myColor === "white" && checkerCount > 0) || (myColor === "black" && checkerCount < 0);
        const isDraggable = isMyTurn && hasMyChecker;

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
                    fill={isDragged ? "#FFD700" : isDestination ? "#90EE90" : pointColor}
                    stroke="#000"
                    strokeWidth="1"
                    style={{ cursor: isDestination ? "pointer" : "default" }}
                    opacity={isDestination ? 0.7 : 1}
                />

                {/* Checkers */}
                {checkerColor && absCount > 0 && (
                    <>
                        {Array.from({ length: Math.min(absCount, 5) }).map((_, i) => {
                            const cy = isTop
                                ? y + 30 + i * (CHECKER_RADIUS * 2 + 2)
                                : y - 30 - i * (CHECKER_RADIUS * 2 + 2);
                            const cx = xPosition + POINT_WIDTH / 2;
                            return (
                                <circle
                                    key={`checker-${pointNum}-${i}`}
                                    cx={cx}
                                    cy={cy}
                                    r={CHECKER_RADIUS}
                                    fill={checkerColor}
                                    stroke="#000"
                                    strokeWidth="2"
                                    onMouseDown={() => {
                                        if (isDraggable) {
                                            handleCheckerMouseDown(pointNum, cx, cy);
                                        }
                                    }}
                                    style={{ cursor: isDraggable ? "grab" : "default" }}
                                />
                            );
                        })}
                        {/* Show count if more than 5 */}
                        {absCount > 5 && (
                            <text
                                x={xPosition + POINT_WIDTH / 2}
                                y={
                                    isTop
                                        ? y + 30 + 4 * (CHECKER_RADIUS * 2 + 2)
                                        : y - 30 - 4 * (CHECKER_RADIUS * 2 + 2)
                                }
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

        const whiteDraggable = isMyTurn && myColor === "white" && whiteCount > 0;
        const blackDraggable = isMyTurn && myColor === "black" && blackCount > 0;
        const barDragged = draggedPoint === 0;
        const barIsDestination = isValidDestination(0);

        return (
            <g>
                {/* Bar background */}
                <rect
                    x={barX}
                    y={50}
                    width={50}
                    height={BOARD_HEIGHT - 100}
                    fill={barDragged ? "#FFD700" : barIsDestination ? "#90EE90" : "#654321"}
                    stroke="#000"
                    strokeWidth="2"
                />

                <text x={barX + 25} y={30} textAnchor="middle" fill="#666" fontSize="12">
                    Bar
                </text>

                {/* White checkers on bar */}
                {whiteCount > 0 && (
                    <>
                        {Array.from({ length: Math.min(whiteCount, 3) }).map((_, i) => {
                            const cx = barX + 25;
                            const cy = BOARD_HEIGHT / 2 - 100 + i * (CHECKER_RADIUS * 2 + 2);
                            return (
                                <circle
                                    key={`bar-white-${i}`}
                                    cx={cx}
                                    cy={cy}
                                    r={CHECKER_RADIUS}
                                    fill="white"
                                    stroke="#000"
                                    strokeWidth="2"
                                    onMouseDown={() => {
                                        if (whiteDraggable) {
                                            handleCheckerMouseDown(0, cx, cy);
                                        }
                                    }}
                                    style={{ cursor: whiteDraggable ? "grab" : "default" }}
                                />
                            );
                        })}
                        {whiteCount > 3 && (
                            <text
                                x={barX + 25}
                                y={BOARD_HEIGHT / 2 - 40}
                                textAnchor="middle"
                                fontSize="14"
                                fontWeight="bold"
                            >
                                {whiteCount}
                            </text>
                        )}
                    </>
                )}

                {/* Black checkers on bar */}
                {blackCount > 0 && (
                    <>
                        {Array.from({ length: Math.min(blackCount, 3) }).map((_, i) => {
                            const cx = barX + 25;
                            const cy = BOARD_HEIGHT / 2 + 40 + i * (CHECKER_RADIUS * 2 + 2);
                            return (
                                <circle
                                    key={`bar-black-${i}`}
                                    cx={cx}
                                    cy={cy}
                                    r={CHECKER_RADIUS}
                                    fill="black"
                                    stroke="#000"
                                    strokeWidth="2"
                                    onMouseDown={() => {
                                        if (blackDraggable) {
                                            handleCheckerMouseDown(0, cx, cy);
                                        }
                                    }}
                                    style={{ cursor: blackDraggable ? "grab" : "default" }}
                                />
                            );
                        })}
                        {blackCount > 3 && (
                            <text
                                x={barX + 25}
                                y={BOARD_HEIGHT / 2 + 120}
                                textAnchor="middle"
                                fill="#FFF"
                                fontSize="14"
                                fontWeight="bold"
                            >
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
        const isBearOffDestination = isValidDestination(25);

        return (
            <g>
                {/* White borne off (left side) */}
                <rect
                    x={10}
                    y={BOARD_HEIGHT / 2 - 60}
                    width={70}
                    height={120}
                    fill={isBearOffDestination && myColor === "white" ? "#90EE90" : "#E0E0E0"}
                    stroke="#000"
                    strokeWidth="2"
                    rx="5"
                    opacity={isBearOffDestination && myColor === "white" ? 0.7 : 1}
                    style={{ cursor: isBearOffDestination && myColor === "white" ? "pointer" : "default" }}
                />
                <text
                    x={45}
                    y={BOARD_HEIGHT / 2 - 70}
                    textAnchor="middle"
                    fill="#666"
                    fontSize="12"
                >
                    White Off
                </text>
                <text
                    x={45}
                    y={BOARD_HEIGHT / 2}
                    textAnchor="middle"
                    fontSize="24"
                    fontWeight="bold"
                >
                    {gameState.bornedOffWhite}
                </text>

                {/* Black borne off (right side) */}
                <rect
                    x={BOARD_WIDTH - 80}
                    y={BOARD_HEIGHT / 2 - 60}
                    width={70}
                    height={120}
                    fill={isBearOffDestination && myColor === "black" ? "#90EE90" : "#E0E0E0"}
                    stroke="#000"
                    strokeWidth="2"
                    rx="5"
                    opacity={isBearOffDestination && myColor === "black" ? 0.7 : 1}
                    style={{ cursor: isBearOffDestination && myColor === "black" ? "pointer" : "default" }}
                />
                <text
                    x={BOARD_WIDTH - 45}
                    y={BOARD_HEIGHT / 2 - 70}
                    textAnchor="middle"
                    fill="#666"
                    fontSize="12"
                >
                    Black Off
                </text>
                <text
                    x={BOARD_WIDTH - 45}
                    y={BOARD_HEIGHT / 2}
                    textAnchor="middle"
                    fontSize="24"
                    fontWeight="bold"
                >
                    {gameState.bornedOffBlack}
                </text>
            </g>
        );
    };

    const getPointX = (pointNum: number): number => {
        const barX = BOARD_WIDTH / 2 - 25;
        const barWidth = 50;

        if (pointNum >= 13 && pointNum <= 18) {
            // Left side, top - 13, 14, 15, 16, 17, 18 from left to right
            const offset = 18 - pointNum;
            return barX - POINT_WIDTH - offset * POINT_WIDTH;
        } else if (pointNum >= 19 && pointNum <= 24) {
            // Right side, top - 19, 20, 21, 22, 23, 24 from left to right
            const offset = pointNum - 19;
            return barX + barWidth + offset * POINT_WIDTH;
        } else if (pointNum >= 7 && pointNum <= 12) {
            // Left side, bottom - 12, 11, 10, 9, 8, 7 from left to right
            const offset = pointNum - 7;
            return barX - POINT_WIDTH - offset * POINT_WIDTH;
        } else {
            // Right side, bottom - 6, 5, 4, 3, 2, 1 from left to right
            const offset = 6 - pointNum;
            return barX + barWidth + offset * POINT_WIDTH;
        }
    };

    // Get the color of the dragged checker
    const getDraggedCheckerColor = (): "white" | "black" | null => {
        if (draggedPoint === null) return null;
        if (draggedPoint === 0) {
            // Bar
            return myColor;
        }
        const checkerCount = gameState.board[draggedPoint - 1];
        return getCheckerColor(checkerCount);
    };

    return (
        <svg
            width={BOARD_WIDTH}
            height={BOARD_HEIGHT}
            className="border-2 border-gray-800 rounded-lg bg-amber-100"
            onMouseMove={handleMouseMove}
            onMouseUp={handleMouseUp}
        >
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
                    <text
                        x={BOARD_WIDTH / 2}
                        y={BOARD_HEIGHT - 20}
                        textAnchor="middle"
                        fontSize="18"
                        fontWeight="bold"
                    >
                        Dice: {gameState.diceRoll[0]}
                        {gameState.diceUsed && gameState.diceUsed[0] && " (used)"},{" "}
                        {gameState.diceRoll[1]}
                        {gameState.diceUsed && gameState.diceUsed[1] && " (used)"}
                    </text>
                </g>
            )}

            {/* Dragged checker (ghost) */}
            {isDragging && draggedPoint !== null && (
                <circle
                    cx={dragX}
                    cy={dragY}
                    r={CHECKER_RADIUS}
                    fill={getDraggedCheckerColor() || "gray"}
                    stroke="#000"
                    strokeWidth="2"
                    opacity={0.7}
                    style={{ pointerEvents: "none" }}
                />
            )}
        </svg>
    );
}
