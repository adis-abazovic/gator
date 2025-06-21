package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/adis-abazovic/gator/internal/config"
	"github.com/adis-abazovic/gator/internal/database"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func main() {

	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading configuration file!")
		os.Exit(-1)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Println("error: failed to open database")
		os.Exit(-1)
	}

	dbQueries := database.New(db)

	s := state{
		db:     dbQueries,
		config: &cfg,
	}

	commands := commands{
		cmds: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerGetUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerGetFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browser", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("error: Less than 2 arguments provided")
		os.Exit(-1)
	}

	cmdName := args[1]
	cmdArgs := args[2:]

	if cmdName == "login" {
		loginCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, loginCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "register" {
		registerCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, registerCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "reset" {
		resetCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, resetCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "users" {
		getUsersCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, getUsersCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "agg" {
		aggCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, aggCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "addfeed" {
		addFeedCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, addFeedCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "feeds" {
		getFeedsCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, getFeedsCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "follow" {
		followCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, followCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "following" {
		followingCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, followingCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "unfollow" {
		unfollowCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, unfollowCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	} else if cmdName == "browser" {
		browseCmd := command{
			name: cmdName,
			args: cmdArgs,
		}

		err = commands.run(&s, browseCmd)
		if err != nil {
			fmt.Println("Error executing command")
			fmt.Println(err)
			os.Exit(-1)
		}
	}

}

func handlerBrowse(s *state, cmd command, user database.User) error {

	limit := 2
	if len(cmd.args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.Name)
		fmt.Printf("--- %s ---\n", post.Title.String)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("follow handler expects a single argument, the URL")
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	deleteParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), deleteParams)
	if err != nil {
		return err
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("follow handler expects a single argument, the URL")
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	feedCreated, err := s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return err
	}

	fmt.Println(feedCreated.FeedName)
	fmt.Println(feedCreated.UserName)
	fmt.Println()

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, f := range feeds {
		fmt.Println(f.FeedName)
	}

	fmt.Println()

	return nil
}

func handlerLogin(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("long handler expects a single argument, the username")
	}

	userName := cmd.args[0]

	foundUser, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User DON'T exists")
			fmt.Println(foundUser)
			os.Exit(1)
		}
	}

	s.config.SetUser(userName)

	fmt.Printf("user '%s' has been set\t", userName)
	return nil
}

func handlerRegister(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("register handler expects a single argument, the username")
	}

	userName := cmd.args[0]
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	}

	foundUser, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("User already exists")
			fmt.Println(foundUser)
			os.Exit(1)
		}
	}

	user, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		fmt.Println("error: inserting user in DB failed")
		fmt.Println(err)
		os.Exit(-1)
	}

	s.config.SetUser(user.Name)
	fmt.Printf("\nUser '%s' has been created.\n", user.Name)
	fmt.Println(user)

	return nil
}

func handlerReset(s *state, cmd command) error {

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Reset command was execute successuly.")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, u := range users {
		if s.config.CurrentUserName == u.Name {
			fmt.Printf("* %s (current) \n", u.Name)
		} else {
			fmt.Printf("* %s\n", u.Name)
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("agg handler expects a single argument, the time between request")
	}

	arg := cmd.args[0]
	timeBetweenReqs, err := time.ParseDuration(arg)
	if err != nil {
		return err
	}

	fmt.Printf("\nCollecting feeds every %s\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Printf("\nerror: \n")
			fmt.Println(err)
		}

	}

}

func handlerAddFeed(s *state, cmd command, user database.User) error {

	if len(cmd.args) != 2 {
		return fmt.Errorf("addfeed handler expects two arguments, the username and URL")
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("error: failed to create feed")
	}

	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	f, err := s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		fmt.Println(f)
		return err
	}

	fmt.Println(feed)
	return nil
}

func handlerGetFeeds(s *state, cmd command) error {

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	getUserName := func(userId uuid.UUID) string {
		u, _ := s.db.GetUserById(context.Background(), userId)
		return u.Name
	}

	for _, f := range feeds {

		fmt.Println(f.Name)
		fmt.Println(f.Url)
		fmt.Println(getUserName(f.UserID))
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func (c *commands) run(s *state, cmd command) error {

	foundCmd, ok := c.cmds[cmd.name]
	if !ok {
		return fmt.Errorf("command '%s' not found", cmd.name)
	}

	err := foundCmd(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{Timeout: 60 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rssFeed := RSSFeed{}
	err = xml.Unmarshal(data, &rssFeed)
	if err != nil {
		return nil, err
	}

	rssFeed.DecodeEscapedStrings()

	return &rssFeed, nil
}

func (rssFeed *RSSFeed) DecodeEscapedStrings() {

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for _, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}

func scrapeFeeds(s *state) error {

	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("*****************************")
	fmt.Printf("Next Feed to fetch: %s\n", feed.Url)
	fmt.Println("*****************************")

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, f := range rssFeed.Channel.Item {

		params := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PublishedAt: stringtoSqlNullTime(f.PubDate),
			Title:       stringToSqlNullString(f.Title),
			Url:         f.Link,
			Description: stringToSqlNullString(f.Description),
			FeedID:      feed.ID,
		}

		p, err := s.db.CreatePost(context.Background(), params)
		if err != nil {
			fmt.Println("error saving post!!!")
			fmt.Println(err)
			fmt.Println(p)
		} else {
			fmt.Printf("\nCreated post: %s\n", p.Url)
		}
	}

	return nil
}

func stringtoSqlNullTime(s string) sql.NullTime {

	if s == "" {
		return sql.NullTime{}
	} else {
		layout := time.RFC3339
		t, err := time.Parse(layout, s)
		if err != nil {
			return sql.NullTime{}
		}

		return sql.NullTime{
			Time:  t,
			Valid: true,
		}
	}
}

func stringToSqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{String: "", Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
