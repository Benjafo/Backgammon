import type { GameState, LegalMove } from "@/types/game";
import { useState } from "react";
import { DiceDisplay } from "./Dice";

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
    const BOARD_WIDTH = 870;
    const BOARD_HEIGHT = 640;
    const POINT_WIDTH = 50;
    const POINT_HEIGHT = 200;
    const CHECKER_RADIUS = 20;
    const GAMEPLAY_BOTTOM = 550; // Fixed bottom position for gameplay area

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
        // White bear-off is always on the right
        if (myColor === "white" && x >= 780 && x <= 850 && y >= 250 && y <= 350) {
            return 25;
        }
        // Black bear-off is always on the left
        if (myColor === "black" && x >= 20 && x <= 90 && y >= 250 && y <= 350) {
            return 25;
        }

        // Check bar
        const barX = BOARD_WIDTH / 2 - 25;
        if (x >= barX && x <= barX + 50 && y >= 50 && y <= GAMEPLAY_BOTTOM) {
            return 0;
        }

        // Check all 24 points
        for (let pointNum = 1; pointNum <= 24; pointNum++) {
            const pointX = getPointX(pointNum);
            const isTop = isPointOnTop(pointNum);
            const pointY = isTop ? 50 : GAMEPLAY_BOTTOM;

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

    // Determine if a point should be rendered on top based on player perspective
    const isPointOnTop = (pointNum: number): boolean => {
        if (myColor === "white") {
            return pointNum >= 13;
        } else {
            return pointNum <= 12;
        }
    };

    // Render a single triangular point
    const renderPoint = (pointNum: number, xPosition: number) => {
        const checkerCount = gameState.board[pointNum - 1];
        const checkerColor = getCheckerColor(checkerCount);
        const absCount = Math.abs(checkerCount);
        const isTop = isPointOnTop(pointNum);

        const isDragged = draggedPoint === pointNum;
        const isDestination = isValidDestination(pointNum);
        const hasMyChecker =
            (myColor === "white" && checkerCount > 0) || (myColor === "black" && checkerCount < 0);
        const isDraggable = isMyTurn && hasMyChecker;

        // Calculate triangle points
        const y = isTop ? 50 : GAMEPLAY_BOTTOM;
        const trianglePoints = isTop
            ? `${xPosition},${y} ${xPosition + POINT_WIDTH / 2},${y + POINT_HEIGHT} ${xPosition + POINT_WIDTH},${y}`
            : `${xPosition},${y} ${xPosition + POINT_WIDTH / 2},${y - POINT_HEIGHT} ${xPosition + POINT_WIDTH},${y}`;

        // Point color (alternating) - Mahogany and dark gold
        const pointColor = pointNum % 2 === 0 ? "hsl(18 52% 22%)" : "hsl(43 40% 35%)";

        return (
            <g key={`point-${pointNum}`}>
                {/* Triangle */}
                <polygon
                    points={trianglePoints}
                    fill={
                        isDragged
                            ? "hsl(43 70% 70%)"
                            : isDestination
                              ? "hsl(43 60% 58%)"
                              : pointColor
                    }
                    stroke="hsl(43 60% 58%)"
                    strokeWidth="2"
                    style={{ cursor: isDestination ? "pointer" : "default" }}
                    opacity={isDestination ? 0.6 : 1}
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
                                    stroke="hsl(43 60% 58%)"
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
                    fill="hsl(var(--gold-light))"
                    fontSize="12"
                    fontWeight="bold"
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
                    height={GAMEPLAY_BOTTOM - 50}
                    fill={barDragged ? "#FFD700" : barIsDestination ? "#90EE90" : "#654321"}
                    stroke="hsl(43 60% 58%)"
                    strokeWidth="2"
                />

                <text
                    x={barX + 25}
                    y={30}
                    textAnchor="middle"
                    fill="hsl(var(--gold-light))"
                    fontSize="12"
                    fontWeight="bold"
                >
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
        const boxWidth = 70;
        const boxHeight = 100;
        const blackOffX = 20; // 20px from left edge
        const whiteOffX = 780; // 20px from playing area right edge
        const boxY = (50 + GAMEPLAY_BOTTOM) / 2 - boxHeight / 2; // Vertically centered in gameplay area

        return (
            <g>
                {/* Black borne off (always left side) */}
                <rect
                    x={blackOffX}
                    y={boxY}
                    width={boxWidth}
                    height={boxHeight}
                    fill={
                        isBearOffDestination && myColor === "black"
                            ? "hsl(43 60% 58%)"
                            : "hsl(18 52% 22%)"
                    }
                    stroke="hsl(43 60% 58%)"
                    strokeWidth="3"
                    rx="8"
                    opacity={isBearOffDestination && myColor === "black" ? 0.8 : 1}
                    style={{
                        cursor: isBearOffDestination && myColor === "black" ? "pointer" : "default",
                    }}
                />
                <text
                    x={blackOffX + boxWidth / 2}
                    y={boxY - 8}
                    textAnchor="middle"
                    fill="hsl(var(--gold-light))"
                    fontSize="11"
                    fontWeight="bold"
                >
                    Black Off
                </text>
                <text
                    x={blackOffX + boxWidth / 2}
                    y={boxY + boxHeight / 2 + 10}
                    textAnchor="middle"
                    fill="hsl(var(--gold-light))"
                    fontSize="32"
                    fontWeight="bold"
                >
                    {gameState.bornedOffBlack}
                </text>

                {/* White borne off (always right side) */}
                <rect
                    x={whiteOffX}
                    y={boxY}
                    width={boxWidth}
                    height={boxHeight}
                    fill={
                        isBearOffDestination && myColor === "white"
                            ? "hsl(43 60% 58%)"
                            : "hsl(18 52% 22%)"
                    }
                    stroke="hsl(43 60% 58%)"
                    strokeWidth="3"
                    rx="8"
                    opacity={isBearOffDestination && myColor === "white" ? 0.8 : 1}
                    style={{
                        cursor: isBearOffDestination && myColor === "white" ? "pointer" : "default",
                    }}
                />
                <text
                    x={whiteOffX + boxWidth / 2}
                    y={boxY - 8}
                    textAnchor="middle"
                    fill="hsl(var(--gold-light))"
                    fontSize="11"
                    fontWeight="bold"
                >
                    White Off
                </text>
                <text
                    x={whiteOffX + boxWidth / 2}
                    y={boxY + boxHeight / 2 + 10}
                    textAnchor="middle"
                    fill="hsl(var(--gold-light))"
                    fontSize="32"
                    fontWeight="bold"
                >
                    {gameState.bornedOffWhite}
                </text>
            </g>
        );
    };

    const getPointX = (pointNum: number): number => {
        const barX = BOARD_WIDTH / 2 - 25;
        const barWidth = 50;

        if (myColor === "white") {
            // White's perspective
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
        } else {
            // Black's perspective (180 degree rotation)
            if (pointNum >= 1 && pointNum <= 6) {
                // Left side, top - 1, 2, 3, 4, 5, 6 from left to right
                const offset = 6 - pointNum;
                return barX - POINT_WIDTH - offset * POINT_WIDTH;
            } else if (pointNum >= 7 && pointNum <= 12) {
                // Right side, top - 7, 8, 9, 10, 11, 12 from left to right
                const offset = pointNum - 7;
                return barX + barWidth + offset * POINT_WIDTH;
            } else if (pointNum >= 13 && pointNum <= 18) {
                // Right side, bottom - 18, 17, 16, 15, 14, 13 from left to right
                const offset = 18 - pointNum;
                return barX + barWidth + offset * POINT_WIDTH;
            } else {
                // Left side, bottom - 24, 23, 22, 21, 20, 19 from left to right
                const offset = pointNum - 19;
                return barX - POINT_WIDTH - offset * POINT_WIDTH;
            }
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

    // Render directional arrows showing movement direction
    const renderDirectionalArrows = () => {
        const arrowColor = "hsl(43 60% 58%)"; // Gold color
        const arrowWidth = 3;

        // Create arrow marker definitions
        const arrowMarker = (id: string) => (
            <marker
                id={id}
                markerWidth="10"
                markerHeight="10"
                refX="8"
                refY="3"
                orient="auto"
                markerUnits="strokeWidth"
            >
                <path d="M0,0 L0,6 L9,3 z" fill={arrowColor} />
            </marker>
        );

        // For white: moves from 24 -> 1 (high to low)
        // For black: moves from 1 -> 24 (low to high)

        // Arrows on the top half
        const topArrows = [];
        const bottomArrows = [];

        if (myColor === "white") {
            // White's perspective: 24 -> 1
            // Top: points 13-24, arrows point right to left
            // Bottom: points 12-1, arrows point left to right

            // Top left arrow (points 13-18)
            topArrows.push(
                <path
                    key="arrow-top-left"
                    d={`M ${BOARD_WIDTH / 2 - 80} 15 L ${120} 15`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-left)"
                    opacity="0.7"
                />
            );

            // Top right arrow (points 19-24)
            topArrows.push(
                <path
                    key="arrow-top-right"
                    d={`M ${BOARD_WIDTH - 120} 15 L ${BOARD_WIDTH / 2 + 80} 15`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-left)"
                    opacity="0.7"
                />
            );

            // Bottom left arrow (points 12-7)
            bottomArrows.push(
                <path
                    key="arrow-bottom-left"
                    d={`M ${120} ${GAMEPLAY_BOTTOM + 45} L ${BOARD_WIDTH / 2 - 80} ${GAMEPLAY_BOTTOM + 45}`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-right)"
                    opacity="0.7"
                />
            );

            // Bottom right arrow (points 6-1)
            bottomArrows.push(
                <path
                    key="arrow-bottom-right"
                    d={`M ${BOARD_WIDTH / 2 + 80} ${GAMEPLAY_BOTTOM + 45} L ${BOARD_WIDTH - 120} ${GAMEPLAY_BOTTOM + 45}`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-right)"
                    opacity="0.7"
                />
            );
        } else {
            // Black's perspective: 1 -> 24
            // Top: points 1-12, arrows point left to right
            // Bottom: points 24-13, arrows point right to left

            // Top left arrow (points 1-6)
            topArrows.push(
                <path
                    key="arrow-top-left"
                    d={`M ${120} 15 L ${BOARD_WIDTH / 2 - 80} 15`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-right)"
                    opacity="0.7"
                />
            );

            // Top right arrow (points 7-12)
            topArrows.push(
                <path
                    key="arrow-top-right"
                    d={`M ${BOARD_WIDTH / 2 + 80} 15 L ${BOARD_WIDTH - 120} 15`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-right)"
                    opacity="0.7"
                />
            );

            // Bottom left arrow (points 24-19)
            bottomArrows.push(
                <path
                    key="arrow-bottom-left"
                    d={`M ${BOARD_WIDTH / 2 - 80} ${GAMEPLAY_BOTTOM + 45} L ${120} ${GAMEPLAY_BOTTOM + 45}`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-left)"
                    opacity="0.7"
                />
            );

            // Bottom right arrow (points 18-13)
            bottomArrows.push(
                <path
                    key="arrow-bottom-right"
                    d={`M ${BOARD_WIDTH - 120} ${GAMEPLAY_BOTTOM + 45} L ${BOARD_WIDTH / 2 + 80} ${GAMEPLAY_BOTTOM + 45}`}
                    stroke={arrowColor}
                    strokeWidth={arrowWidth}
                    fill="none"
                    markerEnd="url(#arrow-left)"
                    opacity="0.7"
                />
            );
        }

        return (
            <g>
                <defs>
                    {arrowMarker("arrow-right")}
                    {arrowMarker("arrow-left")}
                </defs>
                {topArrows}
                {bottomArrows}
            </g>
        );
    };

    return (
        <svg
            width={BOARD_WIDTH}
            height={BOARD_HEIGHT}
            className="border-4 border-gold rounded-lg shadow-gold-lg"
            onMouseMove={handleMouseMove}
            onMouseUp={handleMouseUp}
        >
            {/* Board background - Casino felt */}
            <rect x="0" y="0" width={BOARD_WIDTH} height={BOARD_HEIGHT} fill="hsl(var(--felt))" />

            {/* Directional arrows */}
            {renderDirectionalArrows()}

            {/* Borne off areas */}
            {renderBorneOff()}

            {/* All points */}
            {Array.from({ length: 24 }, (_, i) => i + 1).map((p) => renderPoint(p, getPointX(p)))}

            {/* Bar */}
            {renderBar()}

            {/* Dice display */}
            {gameState.diceRoll && (
                <foreignObject
                    x={BOARD_WIDTH / 2 - 90}
                    y={GAMEPLAY_BOTTOM + 15}
                    width="180"
                    height="60"
                >
                    <div
                        style={{
                            display: "flex",
                            justifyContent: "center",
                            alignItems: "center",
                            height: "100%",
                        }}
                    >
                        <DiceDisplay
                            dice={gameState.diceRoll}
                            used={gameState.diceUsed || []}
                            size={41}
                        />
                    </div>
                </foreignObject>
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
