package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env        string
	ServerAddr string
	LogLevel   string
	JWTSecret  string
	Database   DatabaseConfig
	LDAP       LDAPConfig
	BOS        BOSConfig
	BuildTool  string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type LDAPConfig struct {
	Host       string
	Port       int
	BaseDN     string
	BindDN     string
	BindPass   string
	UserFilter string
	AdminUsers []string
}

type BOSConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	InternalDomain  string
	ExternalDomain  string
}

func Load() (*Config, error) {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "3306"))
	ldapPort, _ := strconv.Atoi(getEnv("LDAP_PORT", "389"))

	return &Config{
		Env:        getEnv("ENV", "development"),
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		JWTSecret:  getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "mysql"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "root123"),
			DBName:   getEnv("DB_NAME", "release_manager"),
		},
		LDAP: LDAPConfig{
			Host:       getEnv("LDAP_HOST", "ldap.example.com"),
			Port:       ldapPort,
			BaseDN:     getEnv("LDAP_BASE_DN", "dc=example,dc=com"),
			BindDN:     getEnv("LDAP_BIND_DN", "cn=admin,dc=example,dc=com"),
			BindPass:   getEnv("LDAP_BIND_PASS", ""),
			UserFilter: getEnv("LDAP_USER_FILTER", "(uid=%s)"),
			AdminUsers: parseAdminUsers(getEnv("LDAP_ADMIN_USERS", "admin")),
		},
		BOS: BOSConfig{
			Endpoint:        getEnv("BOS_ENDPOINT", "bj.bcebos.com"),
			AccessKeyID:     getEnv("BOS_ACCESS_KEY", ""),
			SecretAccessKey: getEnv("BOS_SECRET_KEY", ""),
			Bucket:          getEnv("BOS_BUCKET", "release-bucket"),
			InternalDomain:  getEnv("BOS_INTERNAL_DOMAIN", ""),
			ExternalDomain:  getEnv("BOS_EXTERNAL_DOMAIN", ""),
		},
		BuildTool: getEnv("BUILD_TOOL_PATH", "/usr/local/bin/build-tool"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseAdminUsers(s string) []string {
	if s == "" {
		return []string{}
	}
	var users []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if start < i {
				users = append(users, s[start:i])
			}
			start = i + 1
		}
	}
	return users
}
