package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nlkli/raytrade/internal/broker/bybit/models"
	"nlkli/raytrade/internal/ws"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
	MAINNET             = "https://api.bybit.com"
	STREAM_MAINNET      = "wss://stream.bybit.com"
	DEFAULT_RECV_WINDOW = "5000"
)

type Client struct {
	baseURL string

	apiKey    string
	apiSecret string

	recvWindow string

	httpClient *http.Client
}

func NewClient(apiKey, apiSecret string, opts ...Option) *Client {
	client := &Client{
		baseURL:    MAINNET,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		recvWindow: DEFAULT_RECV_WINDOW,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func NewClientFromEnv(opts ...Option) *Client {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}

	apiKey := os.Getenv("BYBIT_API_KEY")
	if apiKey == "" {
		log.Fatal("env var BYBIT_API_KEY is missing")
	}

	apiSecret := os.Getenv("BYBIT_API_SECRET")
	if apiSecret == "" {
		log.Fatal("env var BYBIT_API_SECRET is missing")
	}

	return NewClient(apiKey, apiSecret, opts...)
}

type Option func(*Client)

func WithRecvWindow(recvWindow string) Option {
	return func(c *Client) {
		c.recvWindow = recvWindow
	}
}

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithHttpClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

type apiResponse struct {
	RetCode    int             `json:"retCode"`
	RetMsg     string          `json:"retMsg"`
	Result     json.RawMessage `json:"result"`
	RetExtInfo struct{}        `json:"retExtInfo"`
	Time       int64           `json:"time"`
}

func (c *Client) genSignature(s string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) callAPI(

	req *http.Request,
	queryString string,
	result any,

) error {

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	payload := timestamp + c.apiKey + c.recvWindow + queryString

	signature := c.genSignature(payload)

	req.Header.Set("X-BAPI-API-KEY", c.apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", c.recvWindow)
	req.Header.Set("X-BAPI-SIGN", signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResponse apiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return err
	}

	if apiResponse.RetCode != 0 {
		return fmt.Errorf("BybitApiError: %s", apiResponse.RetMsg)
	}

	if result != nil {
		if err := json.Unmarshal(apiResponse.Result, result); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) CreatePublicStreamV2(

	category models.Category,
	tx chan []byte,
	onConnected ws.OnConnectedFn,

	opts ...ws.PolicyOption,

) *StreamV2 {
	url := fmt.Sprintf("%s/v5/public/%s", STREAM_MAINNET, category)
	return NewStreamV2(url, tx, onConnected, nil, opts...)
}

func (c *Client) CreatePrivateStreamV2(

	tx chan []byte,
	onConnected ws.OnConnectedFn,

	opts ...ws.PolicyOption,

) *StreamV2 {
	url := fmt.Sprintf("%s/v5/private", STREAM_MAINNET)
	return NewStreamV2(
		url,
		tx,
		nil,
		func(conn *websocket.Conn, sendCh chan<- []byte, n int) error {
			recvWindow, err := strconv.Atoi(c.recvWindow)
			if err != nil {
				return err
			}

			expires := time.Now().Add(
				time.Duration(recvWindow) * time.Millisecond,
			).UnixMilli()

			payload := fmt.Sprintf("GET/realtime%d", expires)

			signature := c.genSignature(payload)

			opReq := &models.StreamOpRequest{
				ReqID: nextReqID(),
				Op:    models.StreamOpAuth,
				Args: []any{
					c.apiKey, expires, signature,
				},
			}

			b, err := json.Marshal(opReq)
			if err != nil {
				return err
			}

			sendCh <- b

			time.Sleep(time.Second * 5)

			if onConnected != nil {
				if err := onConnected(conn, sendCh, n); err != nil {
					return err
				}
			}

			return nil
		},
		opts...,
	)
}
