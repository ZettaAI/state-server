package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/labstack/echo"
)

// https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand/

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GetRequestSchemeAndHostURL returns scheme and host as string
// example scheme="https", host="example.com", returns "https://example.com"
func GetRequestSchemeAndHostURL(c echo.Context) string {
	// check headers from load balancer/proxy
	host := c.Request().Header.Get("X-Forwarded-Host")
	scheme := c.Request().Header.Get("X-Forwarded-Proto")

	// if none use host and scheme in request
	if host == "" && scheme == "" {
		host = c.Request().Host
		scheme = c.Scheme()
	}
	return fmt.Sprintf("%v://%v", scheme, host)
}
