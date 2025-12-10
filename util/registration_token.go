package util

import (
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"math"
	"net"
	"net/http"
	"strings"
	"time"
)

// Token encoding configuration constants
const (
	IP_BASE                  = 20  // Base for IP address encoding
	TIMESTAMP_BASE           = 18  // Base for timestamp encoding
	TOKEN_LENGTH             = 36  // Final token length (including checksum)
	CHECKSUM_LENGTH          = 4   // Checksum length in characters
	TOKEN_EXPIRATION         = 15 * time.Minute
	IP_ENCODED_LENGTH        = 12  // Fixed length for encoded IP
	TIMESTAMP_ENCODED_LENGTH = 10  // Fixed length for encoded timestamp
	USER_AGENT_CHARS         = 6   // Number of chars to extract from User-Agent
	RANDOM_PADDING_LENGTH    = 8   // Random padding length
)

var interleavePositions = map[string][]int{
	"ip":        {2, 7, 11, 15, 19, 23, 27, 31, 4, 8, 12, 16},
	"timestamp": {0, 5, 9, 13, 17, 21, 25, 29, 3, 6},
	"useragent": {1, 10, 14, 18, 22, 26},
	"random":    {20, 24, 28, 30, 32, 33, 34, 35},
}

type TokenData struct {
	EncodedIP        string
	EncodedTimestamp string
	Timestamp        time.Time
	UserAgentChars   string
	IsValid          bool
}

func convertToBase(num int64, base int) string {
	if num == 0 {
		return "0"
	}

	if base < 2 || base > 36 {
		return ""
	}

	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	result := ""

	for num > 0 {
		remainder := num % int64(base)
		result = string(digits[remainder]) + result
		num = num / int64(base)
	}

	return result
}

// Convert string in the given base back to base-10
func convertFromBase(str string, base int) (int64, error) {
	if base < 2 || base > 36 {
		return 0, fmt.Errorf("invalid base: %d", base)
	}

	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	str = strings.ToLower(str)
	var result int64 = 0

	for i, char := range str {
		digitValue := strings.IndexRune(digits, char)
		if digitValue == -1 || digitValue >= base {
			return 0, fmt.Errorf("invalid character '%c' for base %d", char, base)
		}
		power := len(str) - i - 1
		result += int64(digitValue) * int64(math.Pow(float64(base), float64(power)))
	}

	return result, nil
}

// Pad string with leading zeros to reach the target length
func normalizeLength(str string, length int) string {
	if len(str) >= length {
		return str[:length]
	}
	return strings.Repeat("0", length-len(str)) + str
}

// Convert IPv4 address to base-X string
func encodeIPAddress(ipAddr string, base int) (string, error) {
	// Parse IP address
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipAddr)
	}

	// Convert to IPv4
	ip = ip.To4()
	if ip == nil {
		// If IPv6, try to extract IPv4-mapped address
		ipv6 := net.ParseIP(ipAddr).To16()
		if ipv6 != nil && ipv6[10] == 0xff && ipv6[11] == 0xff {
			// IPv4-mapped IPv6 address
			ip = ipv6[12:16]
		} else {
			// Pure IPv6 - use hash of first 32 bits
			var hashVal uint32 = 0
			for i := 0; i < 4; i++ {
				hashVal = (hashVal << 8) | uint32(ipv6[i])
			}
			encoded := convertToBase(int64(hashVal), base)
			return normalizeLength(encoded, IP_ENCODED_LENGTH), nil
		}
	}

	ipInt := int64(ip[0])<<24 | int64(ip[1])<<16 | int64(ip[2])<<8 | int64(ip[3])
	encoded := convertToBase(ipInt, base)
	return normalizeLength(encoded, IP_ENCODED_LENGTH), nil
}

// decodeIPAddress converts a base-X string back to an IPv4 address
func decodeIPAddress(encoded string, base int) (string, error) {
	// Convert from target base to integer
	ipInt, err := convertFromBase(encoded, base)
	if err != nil {
		return "", fmt.Errorf("failed to decode IP: %w", err)
	}

	// Extract octets
	octet1 := (ipInt >> 24) & 0xFF
	octet2 := (ipInt >> 16) & 0xFF
	octet3 := (ipInt >> 8) & 0xFF
	octet4 := ipInt & 0xFF

	// Validate octets
	if octet1 < 0 || octet1 > 255 || octet2 < 0 || octet2 > 255 ||
		octet3 < 0 || octet3 > 255 || octet4 < 0 || octet4 > 255 {
		return "", fmt.Errorf("invalid IP octets")
	}

	return fmt.Sprintf("%d.%d.%d.%d", octet1, octet2, octet3, octet4), nil
}

func encodeTimestamp(timestamp int64, base int) string {
	encoded := convertToBase(timestamp, base)
	return normalizeLength(encoded, TIMESTAMP_ENCODED_LENGTH)
}

func decodeTimestamp(encoded string, base int) (int64, error) {
	return convertFromBase(encoded, base)
}

func processUserAgent(userAgent string) string {
	if userAgent == "" {
		userAgent = "unknown"
	}

	// Extract specific character positions for consistency
	positions := []int{0, 5, 10, 20, 30, 50}
	result := ""

	for _, pos := range positions {
		if pos < len(userAgent) {
			result += string(userAgent[pos])
		} else {
			result += "0"
		}
	}

	return normalizeLength(result, USER_AGENT_CHARS)
}

func generateRandomChars(length int) (string, error) {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(bytes), nil
}

func interleaveComponents(ipEncoded, timestampEncoded, userAgentChars, randomChars string) (string, error) {
	// Validate input lengths
	if len(ipEncoded) != IP_ENCODED_LENGTH {
		return "", fmt.Errorf("invalid IP encoded length: %d", len(ipEncoded))
	}
	if len(timestampEncoded) != TIMESTAMP_ENCODED_LENGTH {
		return "", fmt.Errorf("invalid timestamp encoded length: %d", len(timestampEncoded))
	}
	if len(userAgentChars) != USER_AGENT_CHARS {
		return "", fmt.Errorf("invalid user agent chars length: %d", len(userAgentChars))
	}
	if len(randomChars) != RANDOM_PADDING_LENGTH {
		return "", fmt.Errorf("invalid random chars length: %d", len(randomChars))
	}

	// Create result array
	result := make([]byte, TOKEN_LENGTH-CHECKSUM_LENGTH)

	// Place IP characters
	for i, pos := range interleavePositions["ip"] {
		if i < len(ipEncoded) {
			result[pos] = ipEncoded[i]
		}
	}

	// Place timestamp characters
	for i, pos := range interleavePositions["timestamp"] {
		if i < len(timestampEncoded) {
			result[pos] = timestampEncoded[i]
		}
	}

	// Place user agent characters
	for i, pos := range interleavePositions["useragent"] {
		if i < len(userAgentChars) {
			result[pos] = userAgentChars[i]
		}
	}

	// Place random characters
	for i, pos := range interleavePositions["random"] {
		if i < len(randomChars) {
			result[pos] = randomChars[i]
		}
	}

	return string(result), nil
}

// Extract components from interleaved token
func deinterleaveComponents(token string) (ipEncoded, timestampEncoded, userAgentChars string, err error) {
	// Validate token length (without checksum)
	tokenWithoutChecksum := token[:len(token)-CHECKSUM_LENGTH]
	if len(tokenWithoutChecksum) != TOKEN_LENGTH-CHECKSUM_LENGTH {
		return "", "", "", fmt.Errorf("invalid token length: %d", len(token))
	}

	// Extract IP characters
	ipBytes := make([]byte, IP_ENCODED_LENGTH)
	for i, pos := range interleavePositions["ip"] {
		if i < IP_ENCODED_LENGTH {
			ipBytes[i] = tokenWithoutChecksum[pos]
		}
	}
	ipEncoded = string(ipBytes)

	// Extract timestamp characters
	timestampBytes := make([]byte, TIMESTAMP_ENCODED_LENGTH)
	for i, pos := range interleavePositions["timestamp"] {
		if i < TIMESTAMP_ENCODED_LENGTH {
			timestampBytes[i] = tokenWithoutChecksum[pos]
		}
	}
	timestampEncoded = string(timestampBytes)

	// Extract user agent characters
	uaBytes := make([]byte, USER_AGENT_CHARS)
	for i, pos := range interleavePositions["useragent"] {
		if i < USER_AGENT_CHARS {
			uaBytes[i] = tokenWithoutChecksum[pos]
		}
	}
	userAgentChars = string(uaBytes)

	return ipEncoded, timestampEncoded, userAgentChars, nil
}

func calculateChecksum(data string) string {
	crc := crc32.ChecksumIEEE([]byte(data))
	// Convert to hex and take first 4 characters
	checksum := fmt.Sprintf("%08x", crc)
	return checksum[:CHECKSUM_LENGTH]
}

func validateChecksum(token string) bool {
	if len(token) < CHECKSUM_LENGTH {
		return false
	}

	tokenData := token[:len(token)-CHECKSUM_LENGTH]
	providedChecksum := token[len(token)-CHECKSUM_LENGTH:]
	calculatedChecksum := calculateChecksum(tokenData)

	return providedChecksum == calculatedChecksum
}

func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from load balancers/proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take first IP (original client)
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func GenerateRegistrationToken(ipAddr, userAgent string, timestamp time.Time) (string, error) {
	encodedIP, err := encodeIPAddress(ipAddr, IP_BASE)
	if err != nil {
		return "", fmt.Errorf("failed to encode IP: %w", err)
	}

	encodedTimestamp := encodeTimestamp(timestamp.Unix(), TIMESTAMP_BASE)

	userAgentChars := processUserAgent(userAgent)

	randomChars, err := generateRandomChars(RANDOM_PADDING_LENGTH)
	if err != nil {
		return "", fmt.Errorf("failed to generate random chars: %w", err)
	}

	interleaved, err := interleaveComponents(encodedIP, encodedTimestamp, userAgentChars, randomChars)
	if err != nil {
		return "", fmt.Errorf("failed to interleave components: %w", err)
	}

	checksum := calculateChecksum(interleaved)
	token := interleaved + checksum

	return token, nil
}

func ValidateRegistrationTokenStructure(token, requestIP, requestUserAgent string) (*TokenData, error) {
	if len(token) != TOKEN_LENGTH {
		return nil, fmt.Errorf("invalid token length: expected %d, got %d", TOKEN_LENGTH, len(token))
	}

	if !validateChecksum(token) {
		return nil, fmt.Errorf("token integrity check failed")
	}

	encodedIP, encodedTimestamp, userAgentChars, err := deinterleaveComponents(token)
	if err != nil {
		return nil, fmt.Errorf("failed to deinterleave token: %w", err)
	}

	decodedIP, err := decodeIPAddress(encodedIP, IP_BASE)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IP: %w", err)
	}

	// Normalize IPs for comparison
	requestIPNormalized := net.ParseIP(requestIP)
	decodedIPNormalized := net.ParseIP(decodedIP)

	if requestIPNormalized == nil || decodedIPNormalized == nil ||
		!requestIPNormalized.Equal(decodedIPNormalized) {
		return nil, fmt.Errorf("token IP mismatch")
	}

	timestampUnix, err := decodeTimestamp(encodedTimestamp, TIMESTAMP_BASE)
	if err != nil {
		return nil, fmt.Errorf("failed to decode timestamp: %w", err)
	}

	tokenTimestamp := time.Unix(timestampUnix, 0)
	now := time.Now()

	if now.Sub(tokenTimestamp) > TOKEN_EXPIRATION {
		return nil, fmt.Errorf("token expired")
	}

	requestUserAgentChars := processUserAgent(requestUserAgent)
	if requestUserAgentChars != userAgentChars {
		return nil, fmt.Errorf("token User-Agent mismatch")
	}

	// Return successful validation data
	return &TokenData{
		EncodedIP:        encodedIP,
		EncodedTimestamp: encodedTimestamp,
		Timestamp:        tokenTimestamp,
		UserAgentChars:   userAgentChars,
		IsValid:          true,
	}, nil
}
