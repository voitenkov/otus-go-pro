package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

//go:generate easyjson -all stats.go
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (result users, err error) {
	var (
		line []byte
		i    int
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line = scanner.Bytes()
		var user User
		if err = user.UnmarshalJSON(line); err != nil {
			return
		}
		result[i] = user
		i++
	}

	if err = scanner.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
		return
	}
	return result, nil
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	for _, user := range u {
		if strings.HasSuffix(user.Email, "."+domain) {
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]++
		}
	}
	return result, nil
}
