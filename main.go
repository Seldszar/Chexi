package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
)

type State struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Creator     string `json:"creator"`
}

type Map = map[string]any

var (
	client = http.Client{
		Timeout: 5 * time.Second,
	}

	state = new(State)
)

func fetch(token, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	if len(token) > 0 {
		req.Header.Add("Cookie", fmt.Sprintf(".ROBLOSECURITY=%s", token))
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return io.ReadAll(resp.Body)
	}

	return nil, nil
}

func getPresence(token, userID string) ([]byte, error) {
	return fetch(token, http.MethodPost, "https://presence.roblox.com/v1/presence/users", strings.NewReader(fmt.Sprintf(`{"userIds":[%s]}`, userID)))
}

func getUniverse(token string, universeID int64) ([]byte, error) {
	return fetch(token, http.MethodGet, fmt.Sprintf(`https://games.roblox.com/v1/games?universeIds=%d`, universeID), nil)
}

func refresh(token, userID string) (*State, error) {
	presenceData, err := getPresence(token, userID)

	if err != nil {
		log.Err(err).
			Msg("An error occured while fetching presence")

		return nil, err
	}

	presenceResult := gjson.ParseBytes(presenceData)

	log.Info().
		Bytes("data", presenceData).
		Msg("Fetched presence")

	universeID := presenceResult.Get("userPresences.0.universeId").
		Int()

	if universeID == 0 {
		return nil, nil
	}

	universeData, err := getUniverse(token, universeID)

	if err != nil {
		log.Err(err).
			Msg("An error occured while fetching universe")

		return nil, err
	}

	universeResult := gjson.ParseBytes(universeData)

	log.Info().
		Bytes("data", universeData).
		Msg("Fetched universe")

	universe := universeResult.Get("data.0")

	s := &State{
		ID:          universe.Get("id").Int(),
		Name:        universe.Get("name").String(),
		Description: universe.Get("description").String(),
		Creator:     universe.Get("creator.name").String(),
	}

	return s, nil
}

func startWebServer(port int) error {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().
			Set("Access-Control-Allow-Origin", "*")

		w.Header().
			Set("Content-Type", "application/json")

		json.NewEncoder(w).
			Encode(Map{
				"data": state,
			})
	})

	log.Info().
		Int("port", port).
		Msgf("Server is running: http://localhost:%d", port)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
	})

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				EnvVars:  []string{"USER_ID"},
				Usage:    "The Roblox user ID to check",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				EnvVars: []string{"TOKEN"},
				Usage:   "The Roblox security token",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				EnvVars: []string{"PORT"},
				Usage:   "The server port to use",
				Value:   3000,
			},
		},
		Action: func(ctx *cli.Context) error {
			token := ctx.String("token")
			user := ctx.String("user")
			port := ctx.Int("port")

			go startWebServer(port)

			for {
				if s, err := refresh(token, user); err == nil {
					state = s
				}

				time.Sleep(time.Minute)
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err)
	}
}
