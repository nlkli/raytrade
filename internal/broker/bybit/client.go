package bybit

import (
	"context"
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

	"github.com/joho/godotenv"
)

const (
	MAINNET             = "https://api.bybit.com"
	STREAM_MAINNET      = "wss://stream.bybit.com"
	DEFAULT_RECV_WINDOW = 5000
)

type Client struct {
	baseURL string

	apiKey    string
	apiSecret string

	recvWindow int

	httpClient *http.Client
	timeout    time.Duration

	ctx context.Context
}

func NewClient(ctx context.Context, apiKey, apiSecret string, opts ...Option) *Client {
	client := &Client{
		baseURL:    MAINNET,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		recvWindow: DEFAULT_RECV_WINDOW,
		httpClient: http.DefaultClient,
		timeout:    7 * time.Second,
		ctx:        ctx,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func NewClientFromEnv(ctx context.Context, opts ...Option) *Client {
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

	return NewClient(ctx, apiKey, apiSecret, opts...)
}

type Option func(*Client)

func WithRecvWindow(recvWindow int) Option {
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

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

type apiResponse struct {
	RetCode    int             `json:"retCode"`
	RetMsg     string          `json:"retMsg"`
	Result     json.RawMessage `json:"result"`
	RetExtInfo struct{}        `json:"retExtInfo"`
	Time       int64           `json:"time"`
}

func (c *Client) signature(s string) (string, error) {
	hmac256 := hmac.New(sha256.New, []byte(c.apiSecret))
	if _, err := hmac256.Write([]byte(s)); err != nil {
		return "", err
	}

	return hex.EncodeToString(hmac256.Sum(nil)), nil
}

func (c *Client) callAPI(req *http.Request, queryString string, result any) error {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	signatureString := fmt.Sprintf("%s%s%d%s", timestamp, c.apiKey, c.recvWindow, queryString)
	signature, err := c.signature(signatureString)
	if err != nil {
		return err
	}

	req.Header = map[string][]string{
		"X-BAPI-API-KEY":     {c.apiKey},
		"X-BAPI-TIMESTAMP":   {timestamp},
		"X-BAPI-SIGN":        {signature},
		"X-BAPI-RECV-WINDOW": {strconv.Itoa(c.recvWindow)},
		"Content-Type":       {"application/json"},
		"Accept":             {"application/json"},
	}

	if err := c.ctx.Err(); err != nil {
		// TODO
		log.Println("PARENT CONTEXT DEAD:", err)
	}

	if c.timeout > 0 {
		ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

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

func (c *Client) CreatePublicStream(category models.Category) *Stream {
	url := fmt.Sprintf("%s/v5/public/%s", STREAM_MAINNET, category)
	return NewStream(c.ctx, url, nil)
}

func (c *Client) CreatePrivateStream(category models.Category) *Stream {
	url := fmt.Sprintf("%s/v5/private/%s", STREAM_MAINNET, category)
	return NewStream(c.ctx, url, func(conn *ws.Conn) error {
		expires := time.Now().UnixNano()/1e6 + 10000
		val := fmt.Sprintf("GET/realtime%d", expires)
		signature, err := c.signature(val)
		if err != nil {
			return err
		}
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

		return conn.Send(b)
	})
}
