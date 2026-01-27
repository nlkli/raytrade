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
	baseURL    string
	apiKey     string
	apiSecret  string
	recvWindow int
	httpClient *http.Client
	ctx        context.Context
}

func NewClient(apiKey, apiSecret string, ctx context.Context, opts ...Option) *Client {
	client := &Client{
		baseURL:    MAINNET,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		recvWindow: DEFAULT_RECV_WINDOW,
		httpClient: &http.Client{Timeout: 15 * time.Second},
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
	return NewClient(apiKey, apiSecret, ctx, opts...)
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
		c.httpClient.Timeout = timeout
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
	req = req.WithContext(c.ctx)
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

// func (c *Client) createPrivateStream(category models.Category) {
// 	url := fmt.Sprintf("%s/v5/public/%s", STREAM_MAINNET, category)
// 	ws.NewConn(c.ctx, url, nil)
// }
