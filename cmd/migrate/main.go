package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	var dir string
	flag.StringVar(&dir, "path", "migrations", "каталог с SQL-файлами миграций")
	flag.Parse()

	cmd := "up"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	_ = godotenv.Load()

	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASSWORD", "postgres")
	dbname := getenv("DB_NAME", "mobile_engineer")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)

	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}
	srcURL := "file://" + filepath.ToSlash(abs)

	m, err := migrate.New(srcURL, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("migrate: up OK")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("migrate: down OK")
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
	default:
		fmt.Println("usage: migrate [-path migrations] [up|down|version]")
		os.Exit(2)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
