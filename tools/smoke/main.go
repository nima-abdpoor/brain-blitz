// The smoke client exercises the public gateway against real infrastructure.
// It creates two disposable users; run it against a test deployment.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

var baseURL = strings.TrimRight(os.Getenv("BASE_URL"), "/")
var client = &http.Client{Timeout: 10 * time.Second}

type message struct {
	Success    bool     `json:"success"`
	Event      string   `json:"event"`
	Message    string   `json:"message"`
	Categories []string `json:"categories"`
	MetaData   struct {
		GameID    string `json:"gameId"`
		Questions []struct {
			ID      string   `json:"id"`
			Choices []string `json:"choices"`
		} `json:"questions"`
		LeaderBoard struct {
			Players []struct {
				PlayerID string `json:"playerId"`
				Point    int    `json:"point"`
			} `json:"playerPoint"`
		} `json:"leaderBoard"`
	} `json:"metaData"`
}

type player struct {
	token string
	conn  net.Conn
	rw    io.ReadWriter
}

func request(method, path, token string, body any, expected int, result any) error {
	var data io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		data = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, baseURL+path, data)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// An untrusted identity header must never bypass the gateway's auth check.
	if expected == http.StatusUnauthorized {
		req.Header.Set("X-User-ID", "1")
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != expected {
		return fmt.Errorf("%s %s: expected HTTP %d, got %d", method, path, expected, res.StatusCode)
	}
	if result != nil {
		return json.NewDecoder(res.Body).Decode(result)
	}
	_, err = io.Copy(io.Discard, res.Body)
	return err
}

func newPlayer() (*player, error) {
	credentials := map[string]string{"email": "smoke-" + uuid.NewString() + "@example.com", "password": uuid.NewString()}
	if err := request("POST", "/user-service/public/api/v1/signup", "", credentials, 200, nil); err != nil {
		return nil, err
	}
	var login struct {
		AccessToken string `json:"accessToken"`
	}
	if err := request("POST", "/user-service/public/api/v1/login", "", credentials, 200, &login); err != nil {
		return nil, err
	}
	if login.AccessToken == "" {
		return nil, fmt.Errorf("login returned an empty access token")
	}
	if err := request("GET", "/user-service/api/v1/profile", login.AccessToken, nil, 200, nil); err != nil {
		return nil, err
	}
	return connectPlayer(login.AccessToken)
}

func connectPlayer(token string) (*player, error) {
	url := strings.Replace(baseURL, "http", "ws", 1) + "/game-service/api/v1/process-game"
	dialer := ws.Dialer{Header: ws.HandshakeHeaderHTTP(http.Header{"Authorization": {"Bearer " + token}})}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, buffered, _, err := dialer.Dial(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("WebSocket upgrade: %w", err)
	}
	var reader io.Reader = conn
	if buffered != nil {
		reader = buffered
	}
	// Reads can receive data buffered during the handshake; writes go straight to
	// the socket so no explicit buffered flush is required.
	p := &player{token: token, conn: conn, rw: struct {
		io.Reader
		io.Writer
	}{reader, conn}}
	msg, err := p.read()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if len(msg.Categories) == 0 {
		conn.Close()
		return nil, fmt.Errorf("initial WebSocket message has no categories")
	}
	return p, nil
}

func (p *player) read() (message, error) {
	p.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	data, _, err := wsutil.ReadServerData(p.rw)
	var msg message
	if err != nil {
		return msg, err
	}
	err = json.Unmarshal(data, &msg)
	return msg, err
}

func (p *player) send(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return wsutil.WriteClientMessage(p.rw, ws.OpText, data)
}

func (p *player) until(event string) (message, error) {
	for i := 0; i < 20; i++ {
		msg, err := p.read()
		if err != nil {
			return msg, fmt.Errorf("waiting for %s: %w", event, err)
		}
		if msg.Event == "ERROR" {
			return msg, fmt.Errorf("game error while waiting for %s: %s", event, msg.Message)
		}
		if msg.Event == event {
			if !msg.Success {
				return msg, fmt.Errorf("%s was unsuccessful: %s", event, msg.Message)
			}
			return msg, nil
		}
	}
	return message{}, fmt.Errorf("did not receive %s", event)
}

func run() error {
	if baseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}
	for _, path := range []string{
		"/auth-service/api/v1/health-check", "/user-service/public/api/v1/health-check",
		"/match-service/api/v1/health-check", "/game-service/api/v1/health-check",
		"/question-service/api/v1/question/health-check",
	} {
		if err := request("GET", path, "", nil, 200, nil); err != nil {
			return err
		}
	}
	for _, token := range []string{"", "invalid.token.value"} {
		if err := request("GET", "/user-service/api/v1/profile", token, nil, 401, nil); err != nil {
			return err
		}
	}
	if err := request("POST", "/auth-service/api/v1/access-token", "", map[string]any{"data": []any{}}, 404, nil); err != nil {
		return err
	}
	fmt.Println("PASS: service health and gateway authentication")
	players := make([]*player, 0, 2)
	defer func() {
		for _, p := range players {
			p.conn.Close()
		}
	}()
	for i := 0; i < 2; i++ {
		p, err := newPlayer()
		if err != nil {
			return fmt.Errorf("player %d: %w", i+1, err)
		}
		players = append(players, p)
	}
	fmt.Println("PASS: signup, login, profile, and two authenticated WebSocket clients")
	for _, p := range players {
		if err := p.send(map[string]any{"command": "ADD_TO_WAITING_LIST", "category": "TECH"}); err != nil {
			return err
		}
		if _, err := p.until("ADDED_TO_WAITING_LIST"); err != nil {
			return err
		}
	}
	gameID := ""
	for _, p := range players {
		msg, err := p.until("MATCH_CREATED")
		if err != nil {
			return err
		}
		if msg.MetaData.GameID == "" {
			return fmt.Errorf("match has no game ID")
		}
		if gameID != "" && gameID != msg.MetaData.GameID {
			return fmt.Errorf("players were assigned different games")
		}
		gameID = msg.MetaData.GameID
	}
	fmt.Println("PASS: Kafka matchmaking and MongoDB game creation")
	// Questions are delivered asynchronously after MATCH_CREATED.
	time.Sleep(time.Second)
	for _, p := range players {
		if err := p.send(map[string]any{"command": "READY", "gameId": gameID}); err != nil {
			return err
		}
	}
	for round := 0; round < 10; round++ {
		for _, p := range players {
			msg, err := p.until("QUESTIONS_PUBLISHED")
			if err != nil {
				return err
			}
			if len(msg.MetaData.Questions) != 1 {
				return fmt.Errorf("expected one active question, got %d; run seed-demo first", len(msg.MetaData.Questions))
			}
			q := msg.MetaData.Questions[0]
			if len(q.Choices) < 2 {
				return fmt.Errorf("question has no choices")
			}
			if err := p.send(map[string]any{"command": "ANSWER", "gameId": gameID, "answer": map[string]string{"gameId": gameID, "questionId": q.ID, "choice": q.Choices[1]}}); err != nil {
				return err
			}
			if _, err := p.until("ANSWER_ACCEPTED"); err != nil {
				return err
			}
		}
	}
	fmt.Println("PASS: questions, readiness, and answer storage")
	for _, p := range players {
		msg, err := p.until("COMPLETED")
		if err != nil {
			return err
		}
		if len(msg.MetaData.LeaderBoard.Players) != 2 {
			return fmt.Errorf("final leaderboard is missing players")
		}
		for _, scored := range msg.MetaData.LeaderBoard.Players {
			if scored.Point <= 0 {
				return fmt.Errorf("seeded correct answer did not receive a positive score")
			}
		}
	}
	fmt.Println("PASS: Redis task scheduling and game completion")
	for _, p := range players {
		next, err := connectPlayer(p.token)
		if err != nil {
			return fmt.Errorf("reconnect after completion: %w", err)
		}
		next.conn.Close()
	}
	fmt.Println("PASS: finished players can reconnect for another game")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
}
