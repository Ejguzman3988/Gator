package Command

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"time"
	"uuid"

	. "github.com/ejguzman3988/Gator/internal/config"
	. "github.com/ejguzman3988/Gator/internal/database"
)

type State struct {
	Con *Config
	DB  *Queries
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.Handlers[cmd.Name]

	if !exists {
		return fmt.Errorf("Handler does not exist")
	}

	// run the Command
	err := handler(s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Handlers[name] = f
}

type Command struct {
	Name string
	Args []string
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("No arguments passed to register Command")
	}

	var arg CreateUserParams
	arg.ID = uuid.New().String()
	arg.Name = cmd.Args[2]
	now := time.Now()
	arg.CreatedAt = now
	arg.UpdatedAt = now

	user, err := s.DB.CreateUser(context.Background(), arg)
	if err != nil {
		fmt.Printf("%v", err)
		return err
	}

	fmt.Printf("User has been created \n %v \n", user)

	HandlerLogin(s, cmd)

	return nil

}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("No arguments passed to login Command")
	}

	user, err := s.DB.FindUserByName(context.Background(), cmd.Args[2])
	if err != nil {
		os.Exit(1)
	}

	err = s.Con.SetUser(user.Name)

	if err != nil {
		return err
	}

	fmt.Printf("%v has been set as the current user \n", cmd.Args[2])

	return nil

}

func HandlerReset(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	err := s.DB.ResetUsers(context.Background())

	if err != nil {
		return err
	}

	fmt.Printf("Database Reset")

	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	users, err := s.DB.GetUsers(context.Background())

	if err != nil {
		return err
	}

	for _, u := range users {
		name := u.Name
		if s.Con.CurrentUserName == u.Name {
			name += " (current)"
		}
		fmt.Printf("* %v %v", name, "\n")
	}

	return nil
}

func HandlerFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 3 {

		return fmt.Errorf("Not enough arguments passed to %v\n", cmd.Name)
	}

	time_between_reqs := cmd.Args[2]
	t, err := time.ParseDuration(time_between_reqs)

	if err != nil {

		fmt.Println("Error")
		return err
	}

	fmt.Printf("Collecting feeds every %v\n", time_between_reqs)

	ticker := time.NewTicker(t)

	for ; ; <-ticker.C {
		err := ScrapeFeeds(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func HandlerAddFeed(s *State, cmd Command, user User) error {
	if len(cmd.Args) < 2 {
		os.Exit(1)
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	if len(cmd.Args) < 3 {

		os.Exit(1)
		return fmt.Errorf("Missing argument needs args: name and url for %v\n", cmd.Name)
	}

	// current_user := s.Con.CurrentUserName
	// user, err := s.DB.FindUserByName(context.Background(), current_user)
	//
	// if err != nil {
	// 	return err
	// }

	user_id := user.ID
	current_time := time.Now()
	feed_name := cmd.Args[2]
	feed_url := cmd.Args[3]

	var args CreateFeedParams

	args.ID = uuid.New().String()
	args.CreatedAt = current_time
	args.UpdatedAt = current_time
	args.Name = feed_name
	args.Url = feed_url
	args.UserID = user_id

	feed, err := s.DB.CreateFeed(context.Background(), args)

	if err != nil {
		return err
	}

	var feedFollowArgs CreateFeedFollowParams

	feedFollowArgs.ID = uuid.New().String()
	feedFollowArgs.UserID = user.ID
	feedFollowArgs.FeedID = feed.ID
	feedFollowArgs.CreatedAt = current_time
	feedFollowArgs.UpdatedAt = current_time

	_, err = s.DB.CreateFeedFollow(context.Background(), feedFollowArgs)

	if err != nil {
		return err
	}

	prettyXML, _ := xml.MarshalIndent(&feed, "", "    ")

	fmt.Println(string(prettyXML))

	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	feeds, err := s.DB.ListFeeds(context.Background())

	if err != nil {
		return err
	}

	for _, f := range feeds {

		fmt.Println("\n------------------")
		fmt.Printf("Name:        %v %v", f.FeedName, "\n")
		fmt.Printf("URL:         %v %v", f.FeedUrl, "\n")
		fmt.Printf("Created By:  %v %v", f.UserName, "\n")
		fmt.Println("------------------ \n")
	}

	return nil
}

func HandlerFollow(s *State, cmd Command, user User) error {
	if len(cmd.Args) < 3 {
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	url := cmd.Args[2]

	// user, err := s.DB.FindUserByName(context.Background(), s.Con.CurrentUserName)
	//
	// if err != nil {
	// 	os.Exit(1)
	// 	return err
	// }

	feed, err := s.DB.FindFeedByUrl(context.Background(), url)

	if err != nil {
		fmt.Printf("%v", fmt.Errorf("Feed not found %v\n", cmd.Name))
		os.Exit(1)
		return err
	}

	var arg CreateFeedFollowParams
	arg.ID = uuid.New().String()
	now := time.Now()
	arg.CreatedAt = now
	arg.UpdatedAt = now
	arg.UserID = user.ID
	arg.FeedID = feed.ID

	ff, err := s.DB.CreateFeedFollow(context.Background(), arg)

	if err != nil {
		return err
	}

	fmt.Println("\n------------------")
	fmt.Printf("%v is now following: %v %v", ff.UserName, ff.FeedName, "\n")
	fmt.Println("------------------")

	return nil
}

func HandlerFollowing(s *State, cmd Command, user User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	// user, err := s.DB.FindUserByName(context.Background(), s.Con.CurrentUserName)
	//
	// if err != nil {
	// 	os.Exit(1)
	// 	return err
	// }

	feeds, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)

	if err != nil {
		fmt.Printf("%v", fmt.Errorf("Error fetching feeds %v\n", cmd.Name))
		os.Exit(1)
		return err
	}

	fmt.Printf("\n%v follows:        %v", user.Name, "\n")
	fmt.Println("\n------------------")
	for idx, f := range feeds {
		fmt.Printf("  %v. %v followed by %v%v", idx+1, f.FeedName, f.UserName, "\n")
	}
	fmt.Println("\n------------------")

	return nil
}

func HandlerUnfollow(s *State, cmd Command, user User) error {
	if len(cmd.Args) < 3 {
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	url := cmd.Args[2]

	// find the feed
	feed, err := s.DB.FindFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	feeds, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)

	isFollowing := false
	for _, f := range feeds {
		if feed.ID == f.FeedID {
			isFollowing = true
			break
		}
	}

	if !isFollowing {
		err = fmt.Errorf("User %v is not following any feed with this url: %v", user.Name, feed.Url)
		fmt.Println(err)
		os.Exit(1)
		return err
	}

	var arg UnfollowFeedParams

	arg.FeedID = feed.ID
	arg.UserID = user.ID

	err = s.DB.UnfollowFeed(context.Background(), arg)
	if err != nil {
		return err
	}

	fmt.Printf("%v unfollowed %v\n", user.Name, feed.Name)
	return nil
}

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.DB.FindUserByName(context.Background(), s.Con.CurrentUserName)
		if err != nil {
			fmt.Printf("%v", fmt.Errorf("User not registered %v\n", s.Con.CurrentUserName))
			os.Exit(1)
			return err
		}

		handler(s, cmd, user)
		return nil
	}
}

func ScrapeFeeds(s *State) error {
	// Get the next feed to fetch frm the Database
	// Mark it as fetched
	// Fetch the feed using the URL
	// Iterate over the items in the fed and print their titles to the console

	feed, err := s.DB.GetNextFeedToFetch(context.Background())

	fmt.Printf("Attempting to collect feed: %v\n At url: %v\n", feed.Name, feed.Url)
	if err != nil {
		return err
	}

	var args UpdateLastFetchedAtParams

	args.ID = feed.ID
	args.LastFetchedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	err = s.DB.UpdateLastFetchedAt(context.Background(), args)
	fmt.Printf("Updated last_fetched_at for %v\n", feed.Name)

	if err != nil {
		return err
	}

	rssFeed, err := FetchFeed(context.Background(), feed.Url)
	fmt.Printf("Fetching feed for %v\n", feed.Name)

	if err != nil {

		return err
	}

	fmt.Printf("Printing Rss Feed - %v \n\n", rssFeed.Channel.Title)
	for _, f := range rssFeed.Channel.Item {

		var postArgs CreatePostParams

		postArgs.ID = uuid.New().String()
		now := time.Now()
		postArgs.CreatedAt = now
		postArgs.UpdatedAt = now
		postArgs.Title = f.Title
		postArgs.Url = f.Link
		postArgs.Description = sql.NullString{
			String: f.Description,
			Valid:  true,
		}
		publishedAt, err := mail.ParseDate(f.PubDate)
		if err != nil {
			return err
		}

		postArgs.PublishedAt = publishedAt
		postArgs.FeedID = feed.ID

		post, err := s.DB.CreatePost(context.Background(), postArgs)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		fmt.Printf("- %v \n", post.Title)
	}

	return nil
}

func HandlerBrowse(s *State, cmd Command, user User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("No arguments passed to %v\n", cmd.Name)
	}

	var limit int32
	limit = 2

	if len(cmd.Args) > 3 {
		num, err := strconv.ParseInt(cmd.Args[2], 10, 32)
		if err != nil {
			return err
		}

		limit = int32(num)
	}

	var args GetPostsForUserParams

	args.Limit = limit
	args.UserID = user.ID

	foundPosts, err := s.DB.GetPostsForUser(context.Background(), args)

	if err != nil {
		return err
	}

	fmt.Printf("%v followed posts: \n", user.Name)

	for idx, p := range foundPosts {
		fmt.Printf("%v. %v \n", idx+1, p.Title)
		fmt.Printf("   %v \n", p.PublishedAt)
		fmt.Printf("   %v \n", p.Description)
	}

	return nil
}
