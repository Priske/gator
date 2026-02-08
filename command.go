package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Priske/gator/internal/config"
	"github.com/Priske/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string   //probably
	args []string // maybe no clue not explained well

}
type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {

	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}

	return handler(s, cmd)
}
func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f

}

// ////// HANDLERS SECTION  //////////
func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2 // default

	if len(cmd.args) > 0 {
		n, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("limit must be a number")
		}
		if n <= 0 {
			return fmt.Errorf("limit must be greater than 0")
		}
		limit = n
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit)})
	if err != nil {
		return err
	}
	for _, p := range posts {
		fmt.Printf(
			"%s\n%s\n\n",
			p.Title,
			p.Url,
		)
	}

	return nil
}
func handerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("You must give a URL to unfollow")
	}

	usrid := user.ID
	ctx := context.Background()
	feed, err := s.db.GetFeedByURL(ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("Feed you are trying to unfollow doesn't exist")
	}

	err = s.db.RemoveFeedFollow(ctx, database.RemoveFeedFollowParams{UserID: usrid, FeedID: feed.ID})
	if err != nil {
		return fmt.Errorf("unfollow error occured")
	}

	return nil
}
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("You must give a URL to follow")
	}
	url := cmd.args[0]
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
	feed, err := s.db.GetFeedByURL(ctx, url)
	if err != nil {
		return fmt.Errorf("Feed not found for url: %s(add it first with the addfeed command)", url)
	}

	ff, err := createFollow(ctx, s, user.ID, feed.ID)
	if err != nil {
		return err
	}
	fmt.Printf("%s followed by %s\n", ff.FeedName, ff.UserName)
	return nil

}
func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	follows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}

	for _, f := range follows {
		fmt.Printf("%s follows %s\n", f.UserName, f.FeedName)
	}

	return nil
}
func handlerFeeds(s *state, cmd command) error {
	fds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	if len(fds) == 0 {
		return fmt.Errorf("No feeds found")
	}

	for i, f := range fds {
		fmt.Printf("Feed %d:\n", i+1)
		fmt.Printf("  Name: %s\n", f.FeedName)
		fmt.Printf("  URL: %s\n", f.FeedUrl)
		fmt.Printf("  Created by: %s\n\n", f.UserName)
	}

	return nil
}
func handlerAgg(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Time between reqs needs to be given")
	}
	time_between_reqs := cmd.args[0]
	ctx := context.Background()
	duration, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		fmt.Printf("Collecting feeds every: %s\n", duration)
		scrapeFeeds(s, ctx)
	}

}
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username is required")
	}

	username := cmd.args[0]

	u, err := findUser(context.Background(), s, username)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("user %q does not exist", username)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User %s has been set\n", username)
	return nil
}
func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Database reset successfully")
	return nil
}
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Was unable to fetch users")
	}
	for _, user := range users {
		fmt.Printf("* %s", user)
		if user == s.cfg.CurrentUserName {
			fmt.Printf(" (current)\n")

		} else {
			fmt.Println()
		}
	}
	return nil
}
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	if s.cfg.CurrentUserName == "" {
		return fmt.Errorf("no current user set (run login/register first)")
	}

	ctx := context.Background()
	now := time.Now().UTC()

	nameFeed := cmd.args[0]
	urlFeed := cmd.args[1]

	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      nameFeed,
		CreatedAt: now,
		UpdatedAt: now,
		Url:       urlFeed,
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}

	_, err = createFollow(ctx, s, user.ID, feed.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Feed added: %s (%s)\n", feed.Name, feed.Url)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("A name is required to register")
	}

	username := cmd.args[0]

	u, err := findUser(context.Background(), s, username)
	if err != nil {
		return err
	}
	if u != nil {
		return fmt.Errorf("user %q already exists", username)
	}

	usr, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return err
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf(
		"User has been registered:\n- name: %s\n- id: %s\n- created_at: %s\n- updated_at: %s\n",
		usr.Name,
		usr.ID,
		usr.CreatedAt.Format(time.RFC3339),
		usr.UpdatedAt.Format(time.RFC3339),
	)
	fmt.Printf("User %s has been set\n", usr.Name)

	return nil
}

// ///////////////////Helper Function //////////////////////
func findUser(ctx context.Context, s *state, name string) (*database.User, error) {
	u, err := s.db.GetUser(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // user doesn't exist
		}
		return nil, err // real DB error
	}
	return &u, nil // user exists
}

func createFollow(ctx context.Context, s *state, userID, feedID uuid.UUID) (database.CreateFeedFollowRow, error) {
	return s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    userID,
		FeedID:    feedID,
	})
}
func scrapeFeeds(s *state, ctx context.Context) error {
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return err
	}
	rss, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return err
	}

	for _, item := range rss.Channel.Item {
		publishedAt, err := parsePubDate(item.PubDate)
		if err != nil {
			log.Printf("could not parse pubDate %q for feed %s: %v",
				item.PubDate,
				feed.Url,
				err,
			)
			continue
		}
		s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})

	}

	return nil
}
func parsePubDate(s string) (sql.NullTime, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullTime{Valid: false}, nil
	}

	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return sql.NullTime{Time: t, Valid: true}, nil
		}
	}

	return sql.NullTime{Valid: false}, fmt.Errorf("unable to parse pubDate: %q", s)
}

// //////////////////////MiddleWare /////////////////////
func middlewareLoggedIn(handler func(*state, command, database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		ctx := context.Background()
		name := s.cfg.CurrentUserName
		if name == "" {
			return fmt.Errorf("no user logged in")
		}
		user, err := findUser(ctx, s, name)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("current user %q does not exist", name)
		}

		return handler(s, cmd, *user)
	}
}
