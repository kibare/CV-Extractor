package utils

import (
    "os"
    "time"
    "github.com/golang-jwt/jwt/v4"
    "github.com/joho/godotenv"
)

var jwtSecret []byte

func init() {
    err := godotenv.Load()
    if err != nil {
        panic("Error loading .env file")
    }

    jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))
    if len(jwtSecret) == 0 {
        panic("JWT_SECRET_KEY environment variable is not set")
    }
}

// Claims represents the JWT claims
type Claims struct {
    UserID    uint `json:"user_id"`
    CompanyID uint `json:"company_id"`
    jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(userID uint, companyID uint) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID:    userID,
        CompanyID: companyID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// ParseJWT parses and validates a JWT token
func ParseJWT(tokenString string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, jwt.ErrSignatureInvalid
    }
    return claims, nil
}
