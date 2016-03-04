package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/redis.v3"
	"net/url"
	"os"
)

func postgresURL() string {
	return os.Getenv("DATABASE_URL")
}

func redisURL() string {
	return os.Getenv("REDIS_URL")
}

func needsPostgres() bool {
	return len(postgresURL()) > 0
}

func needsRedis() bool {
	return len(redisURL()) > 0
}

func checkPostgres() (bool, error) {
	db, err := sql.Open("postgres", postgresURL())

	if err != nil {
		return false, err
	}

	var ping int
	err = db.QueryRow("SELECT 1").Scan(&ping)

	return ping == 1, err
}

func checkRedis() (bool, error) {
	userInfo, err := url.Parse(redisURL())
	if err != nil {
		return false, err
	}

	options := redis.Options{
		Addr: userInfo.Host,
	}

	if userInfo.User != nil {
		password, _ := userInfo.User.Password()
		options.Password = password
	}

	client := redis.NewClient(&options)
	pong, err := client.Ping().Result()
	return pong == "PONG", err
}

func main() {
	if needsPostgres() {
		ready, err := checkPostgres()
		fmt.Println(ready, err)
	}

	if needsRedis() {
		ready, err := checkRedis()
		fmt.Println(ready, err)
	}
}
