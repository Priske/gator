package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Priske/gator/internal/config"
	"github.com/Priske/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()

	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("error opening db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	st := state{cfg: &cfg, db: database.New(db)}
	cmds := commands{handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", middlewareLoggedIn(handlerAgg))
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "not enough arguments")
		os.Exit(1)
	}
	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}
	if err := cmds.run(&st, cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
