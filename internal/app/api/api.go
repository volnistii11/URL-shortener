package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/volnistii11/URL-shortener/internal/app/storage/database"
	"github.com/volnistii11/URL-shortener/internal/app/storage/file"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
	"github.com/volnistii11/URL-shortener/internal/model"
	"math/rand"
	"net/http"

	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"

	"github.com/gin-gonic/gin"
)

type Provider interface {
	CreateShortURL(ctx *gin.Context)
	CreateShortURLBatch(ctx *gin.Context)
	GetAllUserURLS(ctx *gin.Context)
	DeleteUserURLS(ctx *gin.Context)
}

func NewAPIServiceServer(repository storage.Repository, cfg config.Flags) Provider {
	return &api{
		repo:  repository,
		flags: cfg,
	}
}

type api struct {
	repo  storage.Repository
	flags config.Flags
}

type request struct {
	URL string `json:"url,omitempty"`
}

type response struct {
	Result string `json:"result,omitempty"`
}

func (a *api) CreateShortURL(ctx *gin.Context) {
	fmt.Println(a.GetStorageType())
	ctx.Header("content-type", "application/json")
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if len(body) == 0 {
		err = errors.New("body is empty")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	respondingServerAddress := fmt.Sprintf("%v://%v/", scheme, ctx.Request.Host)
	if a.flags.GetRespondingServer() != "" {
		respondingServerAddress = fmt.Sprintf("%v/", a.flags.GetRespondingServer())
	}

	userID, _ := ctx.Get("user_id")

	bufRequest := request{}
	if err = json.Unmarshal(body, &bufRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var shortURL string
	switch a.GetStorageType() {
	case "database":
		db := database.NewInitializerReaderWriter(a.repo, a.flags)
		if err := db.CreateTableIfNotExists(); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		shortURL, err = db.WriteURL(&model.URL{
			OriginalURL: bufRequest.URL,
			UserID:      userID.(int),
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
					buffResponse := response{
						Result: fmt.Sprintf("%v%v", respondingServerAddress, shortURL),
					}
					ctx.JSON(http.StatusConflict, buffResponse)
					return
				}
			}
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	default:
		shortURL, err = a.repo.WriteURL(bufRequest.URL)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	buffResponse := response{
		Result: fmt.Sprintf("%v%v", respondingServerAddress, shortURL),
	}
	fmt.Println("CreateShortURL", buffResponse)
	ctx.JSON(http.StatusCreated, buffResponse)
}

func (a *api) CreateShortURLBatch(ctx *gin.Context) {
	ctx.Header("content-type", "application/json")
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if len(body) == 0 {
		err = errors.New("body is empty")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var urls []model.URL
	if err = json.Unmarshal(body, &urls); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	respondingServerAddress := fmt.Sprintf("%v://%v/", scheme, ctx.Request.Host)
	if a.flags.GetRespondingServer() != "" {
		respondingServerAddress = fmt.Sprintf("%v/", a.flags.GetRespondingServer())
	}

	userID, ok := ctx.Get("user_id")
	if ok {
		for i := range urls {
			urls[i].UserID = userID.(int)
		}
	}

	switch a.GetStorageType() {
	case "database":
		db := database.NewInitializerReaderWriter(a.repo, a.flags)
		if err := db.CreateTableIfNotExists(); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}

		urls, err = db.WriteBatchURL(urls, respondingServerAddress)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusCreated, urls)
	case "file":
		Producer, err := file.NewProducer(a.flags.GetFileStoragePath())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		defer Producer.Close()

		response := make([]model.URL, 0, len(urls))
		for _, url := range urls {
			if url.ShortURL == "" {
				url.ShortURL = utils.RandString(10)
			}
			if url.ID == 0 {
				url.ID = uint(rand.Int())
			}
			Producer.WriteEvent(&url)
			err = a.repo.WriteShortAndOriginalURL(url.ShortURL, url.OriginalURL)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
			shortURL := fmt.Sprintf("%v%v", respondingServerAddress, url.ShortURL)
			response = append(response, model.URL{CorrelationID: url.CorrelationID, ShortURL: shortURL})
		}
		ctx.JSON(http.StatusCreated, response)
	case "memory":
		response := make([]model.URL, 0, len(urls))
		for _, url := range urls {
			if url.ShortURL == "" {
				url.ShortURL = utils.RandString(10)
			}
			err = a.repo.WriteShortAndOriginalURL(url.ShortURL, url.OriginalURL)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
			shortURL := fmt.Sprintf("%v%v", respondingServerAddress, url.ShortURL)
			response = append(response, model.URL{CorrelationID: url.CorrelationID, ShortURL: shortURL})
		}
		ctx.JSON(http.StatusCreated, response)
	}
}

func (a *api) GetAllUserURLS(ctx *gin.Context) {
	var (
		urls []model.URL
		err  error
	)
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.JSON(http.StatusNoContent, "user_id is empty")
		return
	}

	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	respondingServerAddress := fmt.Sprintf("%v://%v/", scheme, ctx.Request.Host)
	if a.flags.GetRespondingServer() != "" {
		respondingServerAddress = fmt.Sprintf("%v/", a.flags.GetRespondingServer())
	}

	db := database.NewInitializerReaderWriter(a.repo, a.flags)
	urls, err = db.ReadBatchURLByUserID(userID.(int), respondingServerAddress)
	if err != nil {
		ctx.JSON(http.StatusNoContent, "something wrong")
		return
	}
	if len(urls) == 0 {
		ctx.JSON(http.StatusNoContent, "this user has no records")
		return
	}
	ctx.JSON(http.StatusOK, urls)
}

func (a *api) DeleteUserURLS(ctx *gin.Context) {
	ctx.Status(http.StatusAccepted)

	userID, ok := ctx.Get("user_id")
	if !ok {
		return
	}

	body, err := ctx.GetRawData()
	if err != nil {
		return
	}

	var urls []string
	if err = json.Unmarshal(body, &urls); err != nil {
		return
	}

	db := database.NewInitializerReaderWriter(a.repo, a.flags)
	if err = db.UpdateDeletionStatusOfBatchURL(urls, userID.(int)); err != nil {
		fmt.Println(err)
		return
	}
}

func (a *api) GetStorageType() string {
	if a.flags.GetDatabaseDSN() != "" {
		return "database"
	} else if a.flags.GetFileStoragePath() != "" {
		return "file"
	} else {
		return "memory"
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
