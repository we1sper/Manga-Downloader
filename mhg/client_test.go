package mhg

import (
	"testing"

	"github.com/we1sper/Manga-Downloader/mhg/config"
)

func TestManHuaGuiClient_QueryManga(t *testing.T) {
	cfg := config.Default()
	cfg.Proxy = "http://localhost:10808"

	client, err := NewManHuaGuiClient(cfg)
	if err != nil {
		t.Fatalf("create client error: %v", err)
	}

	manga, err := client.QueryManga(35652)
	if err != nil {
		t.Fatalf("query manga error: %v", err)
	}

	t.Logf("id: %v\n", manga.Id)
	t.Logf("name: %v\n", manga.Name)
	t.Logf("author: %v\n", manga.Author)
	t.Logf("introduction: %v\n", manga.Introduction)
	t.Logf("date: %v\n", manga.Date)
	t.Logf("status: %v\n", manga.Status)
	t.Logf("cover: %v\n", manga.Cover)
	t.Logf("contents: %v\n", manga.Contents)
}

func TestManHuaGuiClient_QueryChapter(t *testing.T) {
	cfg := config.Default()
	cfg.Proxy = "http://localhost:10808"

	client, err := NewManHuaGuiClient(cfg)
	if err != nil {
		t.Fatalf("create client error: %v", err)
	}

	chapter, err := client.QueryChapter("https://www.manhuagui.com/comic/35652/772538.html")
	if err != nil {
		t.Fatalf("query chapter error: %v", err)
	}

	t.Logf("id: %v\n", chapter.Id)
	t.Logf("name: %v\n", chapter.Name)
	t.Logf("files: %v\n", chapter.Files)
}

func TestManHuaGuiClient_DownloadChapter(t *testing.T) {
	cfg := config.Default()
	cfg.SaveDir = "./data"
	cfg.Proxy = "http://localhost:10808"

	client, err := NewManHuaGuiClient(cfg)
	if err != nil {
		t.Fatalf("create client error: %v", err)
	}

	chapter, err := client.QueryChapter("https://www.manhuagui.com/comic/35652/772538.html")
	if err != nil {
		t.Fatalf("query chapter error: %v", err)
	}

	err = client.DownloadChapter(chapter, func(count, total int) {
		t.Logf("progress: %.2f%%", float64(count)/float64(total)*100.0)
	})
	if err != nil {
		t.Fatalf("download chapter error: %v", err)
	}
}
