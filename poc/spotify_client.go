package main

// import (
// "context"
// 	"log"

// 	"github.com/davecgh/go-spew/spew"
// 	spotifyauth "github.com/zmb3/spotify/v2/auth"

// 	"golang.org/x/oauth2/clientcredentials"

// 	"github.com/zmb3/spotify/v2"
// )

// const (
// 	SPOTIFY_ID     = "caf2f8782e4a4a538a916b9d9e1ad427"
// 	SPOTIFY_SECRET = "f9744e59f7164ab89797534011469a9c"
// )

// func main() {
// 	ctx := context.Background()
// 	config := &clientcredentials.Config{
// 		// ClientID:     os.Getenv("SPOTIFY_ID"),
// 		// ClientSecret: os.Getenv("SPOTIFY_SECRET"),
// 		ClientID:     SPOTIFY_ID,
// 		ClientSecret: SPOTIFY_SECRET,
// 		TokenURL:     spotifyauth.TokenURL,
// 	}
// 	token, err := config.Token(ctx)
// 	if err != nil {
// 		log.Fatalf("couldn't get token: %v", err)
// 	}

// 	httpClient := spotifyauth.New().Client(ctx, token)
// 	client := spotify.New(httpClient)
// 	// search for playlists and albums containing "holiday"
// 	results, err := client.Search(ctx, "holiday", spotify.SearchTypePlaylist|spotify.SearchTypeAlbum|spotify.SearchTypeTrack)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	spew.Dump(results.Tracks.Tracks)
// 	// // handle album results
// 	// if results.Albums != nil {
// 	// 	fmt.Println("Albums:")
// 	// 	for _, item := range results.Albums.Albums {
// 	// 		fmt.Println("   ", item.Name)
// 	// 	}
// 	// }
// 	// // handle playlist results
// 	// if results.Playlists != nil {
// 	// 	fmt.Println("Playlists:")
// 	// 	for _, item := range results.Playlists.Playlists {
// 	// 		fmt.Println("   ", item.Name)
// 	// 	}
// 	// }
// }
