package mhg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/we1sper/Manga-Downloader/mhg/config"
	"github.com/we1sper/Manga-Downloader/mhg/script"
	"github.com/we1sper/Manga-Downloader/pkg/client"
	"github.com/we1sper/Manga-Downloader/pkg/util"
	"github.com/we1sper/go-html-retriever"
)

type Manga struct {
	Id           uint64              `json:"id"`
	Name         string              `json:"name"`
	Author       string              `json:"author"`
	Introduction string              `json:"introduction"`
	Date         string              `json:"date"`
	Status       string              `json:"status"`
	Contents     []map[string]string `json:"contents"`
}

type Chapter struct {
	MangaId   uint64   `json:"bid"`
	MangaName string   `json:"bname"`
	Id        uint64   `json:"cid"`
	Name      string   `json:"cname"`
	Files     []string `json:"files"`
	Path      string   `json:"path"`
	Sl        *Sl      `json:"sl"`
}

// Sl used for downloading, currently not used.
type Sl struct {
	E uint64 `json:"e"`
	M string `json:"m"`
}

type ManHuaGuiClient struct {
	decoder *Decoder
	client  *client.HttpClient
	cfg     *config.Config
}

func NewManHuaGuiClient(cfg *config.Config) (*ManHuaGuiClient, error) {
	// Load script from the provided location. If not found, use the built-in script.
	if len(cfg.ScriptLocation) > 0 {
		if err := script.LoadScript(cfg.ScriptLocation); err != nil {
			return nil, err
		}
	}

	httpClient, err := newHttpClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ManHuaGuiClient{
		decoder: NewDecoder(),
		client:  httpClient,
		cfg:     cfg,
	}, nil
}

func (m *ManHuaGuiClient) QueryManga(mangaId uint64) (*Manga, error) {
	resp, err := m.client.Get(fmt.Sprintf("%s/%d", "https://www.manhuagui.com/comic", mangaId))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	parent, err := htmlretriever.From(resp.Body)
	if err != nil {
		return nil, err
	}

	manga := &Manga{
		Id:           mangaId,
		Name:         parent.GetElementsByClassName("book-title").First().GetElementsByTagName("h1").First().Unsafe().GetText(),
		Introduction: strings.Join(parent.GetElementById("intro-all").ExtractTexts(), "\n"),
	}

	detail := parent.GetElementsByClassName("detail-list", "cf").First().ExtractTexts()
	for i := 0; i < len(detail); i++ {
		switch detail[i] {
		case "出品年代：":
			manga.Date = detail[i+1]
		case "漫画作者：":
			cursor := i
			for {
				if detail[cursor] == "漫画别名：" {
					break
				}
				cursor += 1
			}
			// Split by comma ',' ?
			manga.Author = strings.Join(detail[i+1:cursor], "")
		case "漫画状态：":
			manga.Status = detail[i+1]
		}
	}

	extractor := func(p *htmlretriever.EnhancedNodes) {
		p.GetElementsByTagName("a").ForEach(func(n *htmlretriever.EnhancedNode) {
			info := make(map[string]string)
			info["title"] = n.Unsafe().GetAttr("title")
			info["href"] = "https://www.manhuagui.com" + n.Unsafe().GetAttr("href")
			info["page"] = n.GetElementsByTagName("i").First().Unsafe().GetText()
			manga.Contents = append(manga.Contents, info)
		})
	}

	value := ""

	parent.GetElementsByTagName("input").ForEach(func(n *htmlretriever.EnhancedNode) {
		if len(value) == 0 && n.Unsafe().GetAttr("id") == "__VIEWSTATE" {
			value = n.Unsafe().GetAttr("value")
		}
	})

	if len(value) == 0 {
		// Decoding is not required.
		extractor(parent.GetElementsByClassName("chapter-list", "cf", "mt10"))
	} else {
		// Process hidden chapters.
		result, err := m.decoder.Decompress(value)
		if err != nil {
			return manga, err
		}

		node, err := htmlretriever.From(strings.NewReader(result))
		if err != nil {
			return manga, err
		}

		extractor(node.GetElementsByClassName("chapter-list", "cf", "mt10"))
	}

	return manga, nil
}

func (m *ManHuaGuiClient) QueryChapter(chapterUrl string) (*Chapter, error) {
	// The URL typically follows the format: https://www.manhuagui.com/comic/{manga_id}/{chapter_id}.html
	resp, err := m.client.Get(chapterUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	parent, err := htmlretriever.From(resp.Body)
	if err != nil {
		return nil, err
	}

	// Locate the target script at a fixed position.
	_script := parent.GetElementsByTagName("script").At(-6).Unsafe().GetText()

	// Locate parameters.
	start := strings.Index(_script, "return p;}(")
	end := strings.Index(_script, "['\\x73\\x70\\x6c\\x69\\x63']")
	parameters := _script[start+11 : end]

	result, err := m.decoder.Decrypt(parameters)
	if err != nil {
		return nil, err
	}

	// Extract data wrapped by 'SMH.imgData(data).preInit();'.
	data := result[12 : len(result)-12]

	chapter := &Chapter{}
	// Deserialize JSON data into Chapter struct (ignoring unused fields).
	err = json.Unmarshal([]byte(data), chapter)
	if err != nil {
		return nil, err
	}

	return chapter, nil
}

func (m *ManHuaGuiClient) DownloadChapter(chapter *Chapter, informer func(count, total int)) error {
	savePath := path.Join(m.cfg.SaveDir, chapter.MangaName, chapter.Name)

	if err := os.MkdirAll(savePath, 0755); err != nil {
		return err
	}

	download := func(file string) error {
		filePath := path.Join(savePath, file)
		if !m.cfg.Overwrite && util.IsFileExist(filePath) {
			return nil
		}
		// Ignore query parameters 'e' and 'm'.
		resp, err := m.client.Get("https://eu1.hamreus.com" + chapter.Path + file)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if err = util.SaveFromStream(filePath, resp.Body); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid status code: %d", resp.StatusCode)
		}

		return nil
	}

	for idx, file := range chapter.Files {
		if err := download(file); err != nil {
			return fmt.Errorf("download file %s error: %v", file, err)
		}
		if informer != nil {
			informer(idx+1, len(chapter.Files))
		}
	}

	return nil
}

func newHttpClient(cfg *config.Config) (*client.HttpClient, error) {
	httpClient, err := client.NewHttpClient(&client.Option{
		Proxy:                 cfg.Proxy,
		Retry:                 cfg.Retry,
		TimeoutInMilliseconds: cfg.Timeout,
	})
	if err != nil {
		return nil, err
	}

	if len(cfg.Cookie) > 0 {
		httpClient.SetCookie(cfg.Cookie)
	}

	httpClient.AddHeader("Referer", "https://www.manhuagui.com/")
	httpClient.AddHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36")

	return httpClient, nil
}
