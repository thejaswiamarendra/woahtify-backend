package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zmb3/spotify/v2"
)

var (
	state = "abc123"
)

func (a *API) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		log.Printf("Method not allowed, LoginHandler")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	url := a.SpotifyAuthenticator.AuthURL(state)
	log.Println("Redirecting user to:", url)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{RedirectURL: url})
}

func (a *API) SpotifyOAuthHandler(w http.ResponseWriter, r *http.Request) {
	tok, err := a.SpotifyAuthenticator.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	client := spotify.New(a.SpotifyAuthenticator.Client(r.Context(), tok))
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		http.Error(w, "Couldn't get user info", http.StatusForbidden)
		log.Fatal(err)
	}

	userInfo := SpotifyTokenInfo{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		UserName:     user.DisplayName,
		Expiry:       time.Now().Add(24 * time.Hour).Unix(),
	}

	a.AccessTokenMap[userInfo.UserName] = userInfo

	jwt, err := userInfo.GenerateJWTToken()
	if err != nil {
		http.Error(w, "Couldn't generate JWT", http.StatusInternalServerError)
		log.Fatal(err)
	}

	origin := getOrigin(r)

	redirectURL := fmt.Sprintf("%s#token=%s", origin, jwt)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func getOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
