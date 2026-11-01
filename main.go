package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	. "github.com/ejguzman3988/Gator/internal/command"
	. "github.com/ejguzman3988/Gator/internal/config"
	"github.com/ejguzman3988/Gator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	config, err := Read()
	if err != nil {
		log.Fatalf(": Error reading gatorconfig\n %v\n", err)
	}
	fmt.Println(config.DBUrl)

	db, err := sql.Open("postgres", config.DBUrl)

	if err != nil {
		log.Fatalf(": Error connecting to database\n %v\n", err)
	}

	dbQueries := database.New(db)

	s := State{Con: &config, DB: dbQueries}

	handlers := make(map[string]func(*State, Command) error)

	commands := Commands{Handlers: handlers}

	commands.Register(
		"login",
		HandlerLogin,
	)

	commands.Register(
		"register",
		HandlerRegister,
	)

	commands.Register(
		"reset",
		HandlerReset,
	)

	commands.Register(
		"users",
		HandlerUsers,
	)

	commands.Register(
		"agg",
		HandlerFeed,
	)

	commands.Register(
		"addfeed",
		MiddlewareLoggedIn(HandlerAddFeed),
	)

	commands.Register(
		"feeds",
		HandlerFeeds,
	)

	commands.Register(
		"follow",
		MiddlewareLoggedIn(HandlerFollow),
	)

	commands.Register(
		"unfollow",
		MiddlewareLoggedIn(HandlerUnfollow),
	)

	commands.Register(
		"following",
		MiddlewareLoggedIn(HandlerFollowing),
	)

	commands.Register(
		"browse",
		MiddlewareLoggedIn(HandlerBrowse),
	)

	args := os.Args
	if len(args) < 2 {
		log.Fatalf("Fewer than 2 arguments. Exiting...")
		os.Exit(1)
	}

	command := Command{Name: args[1], Args: args}

	err = commands.Run(&s, command)

	if err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}
