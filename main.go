package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/redis.v3"
	"net/url"
	"os"
	"time"
)

const timeoutInterval time.Duration = 5 * time.Minute
const tickDuration time.Duration = 500 * time.Millisecond

type PingFunc func() (bool, error)

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

func waitService(name string, pingFn PingFunc) error {
	timeout := time.After(timeoutInterval)
	tick := time.Tick(tickDuration)

	printWaiting(name)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out waiting for %s", name)
		case <-tick:
			ready, _ := pingFn()
			if ready {
				printDone()
				return nil
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

func printTimeout() {
	fmt.Println(" timed out")
}

func main() {
	if needsPostgres() {
		err := waitService("PostgresSQL", pingPostgres)
		if err != nil {
			printTimeout()
			os.Exit(1)
		}
	}

	if needsRedis() {
		err := waitService("Redis", pingRedis)
		if err != nil {
			printTimeout()
			os.Exit(1)
		}
	}
}
