package config

import "os"

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type PersonConfig struct {
	FirstName string
	LastName  string
	Phone     string
	URACF     string
}

type CoupleConfig struct {
	Groom PersonConfig
	Bride PersonConfig
}

type Config struct {
	DB         DBConfig
	Port       string
	CORSOrigin string
	Couple     CoupleConfig
}

func Load() Config {
	return Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE"),
		},
		Port:       getEnv("PORT"),
		CORSOrigin: getEnv("CORS_ORIGIN"),
		Couple: CoupleConfig{
			Groom: PersonConfig{
				FirstName: getEnv("GROOM_FIRST_NAME"),
				LastName:  getEnv("GROOM_LAST_NAME"),
				Phone:     getEnv("GROOM_PHONE"),
				URACF:     getEnv("GROOM_URACF"),
			},
			Bride: PersonConfig{
				FirstName: getEnv("BRIDE_FIRST_NAME"),
				LastName:  getEnv("BRIDE_LAST_NAME"),
				Phone:     getEnv("BRIDE_PHONE"),
				URACF:     getEnv("BRIDE_URACF"),
			},
		},
	}
}

func getEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return ""
}
