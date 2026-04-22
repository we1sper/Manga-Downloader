package mhg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/we1sper/Manga-Downloader/mhg/config"
	"github.com/we1sper/Manga-Downloader/pkg/log"
	"github.com/we1sper/Manga-Downloader/pkg/server"
)

var (
	routines     = 16
	chanSize     = 16384
	recordFields = []string{"manga_id", "manga_name", "chapter_id", "chapter_name", "progress", "status", "count", "total"}
)

type DownloadRequest struct {
	MangaId   uint64 `json:"mid"`
	ChapterId uint64 `json:"cid"`
}

type ApiServer struct {
	server   *server.HttpServer
	client   *ManHuaGuiClient
	cfg      *config.Config
	lock     *sync.RWMutex
	records  map[string]map[string]interface{}
	taskChan chan *Chapter
	context  context.Context
	cancel   context.CancelFunc
}

func NewApiServer(cfg *config.Config) (*ApiServer, error) {
	client, err := NewManHuaGuiClient(cfg)
	if err != nil {
		return nil, err
	}

	apiServer := &ApiServer{
		client:   client,
		cfg:      cfg,
		lock:     &sync.RWMutex{},
		records:  make(map[string]map[string]interface{}),
		taskChan: make(chan *Chapter, chanSize),
	}

	// Register APIs.
	apiServer.server = server.NewHttpServer("", int(cfg.ApiServerPort)).AllowCORS().
		RegisterHandler("/query/manga", apiServer.queryManga).
		RegisterHandler("/query/records", apiServer.queryRecords).
		RegisterHandler("/download/chapters", apiServer.downloadChapters).
		RegisterOnShutdownHook(apiServer.destroy)

	apiServer.context, apiServer.cancel = context.WithCancel(context.Background())

	return apiServer, nil
}

func (s *ApiServer) Run() {
	if s.cfg.Concurrency == 0 {
		log.Warnf("provided concurrency is 0, fallbacks to default 3")
		s.cfg.Concurrency = 3
	}

	for id := 0; id < int(s.cfg.Concurrency); id++ {
		go s.downloader(id)
	}

	s.server.Start()
}

func (s *ApiServer) Stop() {
	s.server.Stop()
}

func (s *ApiServer) destroy() {
	s.cancel()

	// Wait routines stop.
	time.Sleep(1 * time.Second)

	drain := func() {
		for {
			select {
			case _ = <-s.taskChan:
			default:
				return
			}
		}
	}

	// Clear residual tasks.
	drain()

	close(s.taskChan)
}

func (s *ApiServer) queryManga(writer http.ResponseWriter, request *http.Request) {
	// Get query parameter 'mid'.
	mid, err := s.getIntParameter(request, "mid")
	if err != nil {
		s.response(writer, http.StatusBadRequest, err.Error())
		return
	}

	manga, err := s.client.QueryManga(uint64(mid))
	if err != nil {
		s.response(writer, http.StatusInternalServerError, fmt.Sprintf("failed to query manga %d: %v", mid, err))
		return
	}

	s.response(writer, http.StatusOK, manga)
}

func (s *ApiServer) queryRecords(writer http.ResponseWriter, request *http.Request) {
	records := make([]map[string]interface{}, 0)

	s.lock.RLock()

	keys := make([]string, 0, len(s.records))
	for key := range s.records {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		record := map[string]interface{}{}
		for _, field := range recordFields {
			record[field] = s.records[key][field]
		}
		records = append(records, record)
	}

	s.lock.RUnlock()

	s.response(writer, http.StatusOK, records)
}

func (s *ApiServer) downloadChapters(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		s.response(writer, http.StatusMethodNotAllowed, "only POST method is allowed")
		return
	}

	bytes, _ := io.ReadAll(request.Body)

	var requests []*DownloadRequest
	if err := json.Unmarshal(bytes, &requests); err != nil {
		s.response(writer, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	response := make([]map[string]interface{}, 0)

	var wg sync.WaitGroup
	var lock sync.RWMutex

	submitter := func(id int) {
		for i := id; i < len(requests); i += routines {
			// Construct URL of queried chapter by manga ID and chapter ID.
			url := fmt.Sprintf("https://www.manhuagui.com/comic/%d/%d.html", requests[i].MangaId, requests[i].ChapterId)

			result := map[string]interface{}{
				"mid": requests[i].MangaId,
				"cid": requests[i].ChapterId,
			}

			chapter, err := s.client.QueryChapter(url)
			if err != nil {
				result["status"] = fmt.Sprintf("error: %v", err)
			} else {
				s.submitTask(chapter)
				result["status"] = "ok"
			}

			lock.Lock()
			response = append(response, result)
			lock.Unlock()
		}
		wg.Done()
	}

	for id := 0; id < routines; id++ {
		wg.Add(1)
		go submitter(id)
	}

	wg.Wait()

	s.response(writer, http.StatusOK, response)
}

func (s *ApiServer) downloader(id int) {
	log.Infof("[downloader][%d] starts", id)
	defer log.Infof("[downloader][%d] exits", id)

	for {
		select {
		case <-s.context.Done():
			return
		case chapter := <-s.taskChan:
			key := s.getRecordKey(chapter)

			s.updateRecord(key, func(record map[string]interface{}) {
				record["status"] = "downloading"
			})

			start := time.Now()

			err := s.client.DownloadChapter(chapter, func(count, total int) {
				progress := float64(count) / float64(total) * 100.0
				s.updateRecord(key, func(record map[string]interface{}) {
					record["count"] = count
					record["progress"] = progress
				})
				log.Infof("[downloader][%d][%s][%s] %.1f%%", id, chapter.MangaName, chapter.Name, progress)
			})

			if err != nil {
				s.updateRecord(key, func(record map[string]interface{}) {
					record["status"] = "error: " + err.Error()
				})
				log.Infof("[downloader][%d][%s][%s] failed: %v", id, chapter.MangaName, chapter.Name, err)
			} else {
				elapsed := time.Since(start)

				s.updateRecord(key, func(record map[string]interface{}) {
					record["status"] = "success"
					record["elapsed"] = fmt.Sprintf("%vm", elapsed.Minutes())
				})

				log.Infof("[downloader][%d][%s][%s] succeeded using %.2fm", id, chapter.MangaName, chapter.Name, elapsed.Minutes())
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *ApiServer) getIntParameter(request *http.Request, parameter string) (int, error) {
	value := request.URL.Query().Get(parameter)

	if len(value) == 0 {
		return 0, fmt.Errorf("query parameter '%s' is required", parameter)
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("query parameter '%s' should be an integer", parameter)
	}

	return result, nil
}

func (s *ApiServer) response(writer http.ResponseWriter, code int, data interface{}) {
	writer.WriteHeader(code)

	var bytes []byte

	if message, ok := data.(string); ok {
		bytes = []byte(message)
	} else {
		bytes, _ = json.Marshal(data)
	}

	_, _ = writer.Write(bytes)
}

func (s *ApiServer) getRecordKey(chapter *Chapter) string {
	return fmt.Sprintf("%s-%s", chapter.MangaName, chapter.Name)
}

func (s *ApiServer) submitTask(chapter *Chapter) {
	s.lock.Lock()

	key := s.getRecordKey(chapter)
	if _, ok := s.records[key]; !ok {
		s.records[key] = make(map[string]interface{})
	}

	for _, field := range recordFields {
		switch field {
		case "manga_id":
			s.records[key][field] = chapter.MangaId
		case "manga_name":
			s.records[key][field] = chapter.MangaName
		case "chapter_id":
			s.records[key][field] = chapter.Id
		case "chapter_name":
			s.records[key][field] = chapter.Name
		case "progress":
			s.records[key][field] = 0.0
		case "status":
			s.records[key][field] = "waiting for download"
		case "count":
			s.records[key][field] = 0
		case "total":
			s.records[key][field] = len(chapter.Files)
		}
	}

	s.lock.Unlock()

	s.taskChan <- chapter
}

func (s *ApiServer) updateRecord(key string, updater func(map[string]interface{})) {
	s.lock.Lock()
	defer s.lock.Unlock()

	updater(s.records[key])
}
