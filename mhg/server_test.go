package mhg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/we1sper/Manga-Downloader/mhg/config"
	"github.com/we1sper/go-map-retriever"
)

func TestApiServer_Run(t *testing.T) {
	cfg := config.Default()
	cfg.ApiServerPort = 8080

	apiServer, err := NewApiServer(cfg)
	if err != nil {
		t.Fatalf("create api server error: %v", err)
	}

	go func() {
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()

		<-timer.C

		apiServer.Stop()
	}()

	apiServer.Run()
}

func TestApiServer_queryManga(t *testing.T) {
	cfg := config.Default()
	cfg.ApiServerPort = 8080
	cfg.Proxy = "http://localhost:10808"

	apiServer, err := NewApiServer(cfg)
	if err != nil {
		t.Fatalf("create api server error: %v", err)
	}

	go apiServer.Run()

	time.Sleep(1 * time.Second)

	client := &http.Client{}

	resp, err := client.Get("http://localhost:8080/query/manga?mid=35652")
	if err != nil {
		t.Fatalf("query manga error: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("query manga error: status code %v", resp.StatusCode)
	}

	bytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	fmt.Println(string(bytes))

	apiServer.Stop()
}

func TestApiServer_downloadChapters(t *testing.T) {
	cfg := config.Default()
	cfg.ApiServerPort = 8080
	cfg.SaveDir = "./data"
	cfg.Proxy = "http://localhost:10808"

	apiServer, err := NewApiServer(cfg)
	if err != nil {
		t.Fatalf("create api server error: %v", err)
	}

	go apiServer.Run()

	time.Sleep(1 * time.Second)

	client := &http.Client{}

	data := []map[string]interface{}{
		{
			"mid": 35652,
			"cid": 487852,
		},
	}

	payload, _ := json.Marshal(data)
	request, _ := http.NewRequest("POST", "http://localhost:8080/download/chapters", bytes.NewReader(payload))

	resp, err := client.Do(request)
	if err != nil {
		t.Fatalf("submit task error: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("submit task error: status code %v", resp.StatusCode)
	}

	bytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	fmt.Println(string(bytes))

	for {
		resp, err = client.Get("http://localhost:8080/query/records")
		if err != nil {
			break
		}

		var raw []interface{}
		bytes, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		_ = json.Unmarshal(bytes, &raw)

		v := mapretriever.NewMapRetriever(raw)

		status := v.At(0).Get("status").Unsafe().String()
		if status == "success" {
			break
		}
	}

	apiServer.Stop()
}
