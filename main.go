package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/redis.v3"
	"net/url"
	"os"
	"time"
)

const timeoutInterval time.Duration = 5 * time.Minute
const tickDuration time.Duration = 500 * time.Millisecond

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

func pingPostgres() (bool, error) {
	db, err := sql.Open("postgres", postgresURL())
	defer db.Close()

	if err != nil {
		return false, err
	}

	var ping int
	err = db.QueryRow("SELECT 1").Scan(&ping)

	return ping == 1, err
}

func pingRedis() (bool, error) {
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
	defer client.Close()
	pong, err := client.Ping().Result()
	return pong == "PONG", err
}

func waitPostgres() (bool, error) {
	timeout := time.After(timeoutInterval)
	tick := time.Tick(tickDuration)

	for {
		select {
		case <-timeout:
			return false, errors.New("Timed out while waiting for Postgres")
		case <-tick:
			ready, _ := pingPostgres()

			if ready {
				return true, nil
			}

			printTick()
		}
	}
}

func waitRedis() (bool, error) {
	timeout := time.After(timeoutInterval)
	tick := time.Tick(tickDuration)

	for {
		select {
		case <-timeout:
			return false, errors.New("Timed out while waiting for Redis")
		case <-tick:
			ready, _ := pingRedis()

			if ready {
				return true, nil
			}

			printTick()
		}
	}
}

func printWaiting(service string) {
	fmt.Printf("Waiting for %s to become available...", service)
}

func printTick() {
	fmt.Print(".")
}

func printDone() {
	fmt.Println(" done")
}

func main() {
	if needsPostgres() {
		printWaiting("PostgresSQL")
		ready, err := waitPostgres()

		if !ready {
			fmt.Println(err)
			os.Exit(1)
		}

		printDone()
	}

	if needsRedis() {
		printWaiting("Redis")
		ready, err := waitRedis()

		if !ready {
			fmt.Println(err)
			os.Exit(1)
		}

		printDone()
	}
}
