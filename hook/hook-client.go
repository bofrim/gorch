package hook

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

const HookClientBufferSize = 100
const HookClientIdleTimeout = HookListenIdleTimeout / 2
const HookClientReqTimeout = 3 * time.Second

type HookClient struct {
	Address    string
	UpdateChan chan []byte
	outputLog  [][]byte
	ctx        context.Context
	cancel     context.CancelFunc
	isRunning  bool
	client     *http.Client
}

func NewHookClient(addr string) *HookClient {
	return &HookClient{
		Address:    addr,
		UpdateChan: make(chan []byte, HookClientBufferSize),
		outputLog:  make([][]byte, 0),
		isRunning:  false,
		client: &http.Client{
			Timeout: HookClientReqTimeout,
		},
	}
}

func (h *HookClient) Send(body []byte) error {
	if h.isRunning {
		h.UpdateChan <- body
		return nil
	} else {
		return fmt.Errorf("HookClient is not running")
	}
}

func (h *HookClient) Start() error {
	c, cancel := context.WithCancel(context.Background())
	h.ctx = c
	h.cancel = cancel

	h.isRunning = true
	keepAliveTicker := time.NewTicker(HookClientIdleTimeout)
	go func() {
		for {
			select {
			case body := <-h.UpdateChan:
				h.outputLog = append(h.outputLog, body)
				if err := h.update(body); err != nil {
					fmt.Println(err)
				}
			case <-keepAliveTicker.C:
				if err := h.sendKeepAlive(); err != nil {
					fmt.Println(err)
				}
			case <-h.ctx.Done():
				keepAliveTicker.Stop()
				if err := h.finish(); err != nil {
					fmt.Println(err)
				}
				return
			}
		}
	}()
	return nil
}

func (h *HookClient) Stop() {
	if h.isRunning {
		h.cancel()
	}
	h.isRunning = false
}

func (h *HookClient) update(body []byte) error {
	url := fmt.Sprintf("http://%s/update", h.Address)
	return h.post(url, body)
}

func (h *HookClient) sendKeepAlive() error {
	url := fmt.Sprintf("http://%s/keepalive", h.Address)
	return h.post(url, []byte{})
}

func (h *HookClient) finish() error {
	url := fmt.Sprintf("http://%s/finish", h.Address)
	return h.post(url, []byte{})
}

func (h *HookClient) post(url string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := h.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Bad request")
		return fmt.Errorf("request not OK: %d", resp.StatusCode)
	}
	return nil
}
