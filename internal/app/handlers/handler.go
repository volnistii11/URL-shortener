package handlers

import (
	"encoding/json"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/volnistii11/URL-shortener/internal/app/storage/database"
	"github.com/volnistii11/URL-shortener/internal/model"
	"net/http"

	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/storage/file"

	"github.com/gin-gonic/gin"
)

type HandlerProvider interface {
	CreateShortURL(ctx *gin.Context)
	GetFullURL(ctx *gin.Context)
	PingDatabaseServer(ctx *gin.Context)
	GetStorageType() string
}

func NewHandlerProvider(repository storage.Repository, cfg config.Flags) HandlerProvider {
	return &handlerURL{
		repo:  repository,
		flags: cfg,
	}
}

type handlerURL struct {
	repo  storage.Repository
	flags config.Flags
}

func (h *handlerURL) CreateShortURL(ctx *gin.Context) {
	ctx.Header("content-type", "text/plain; charset=utf-8")
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

	userID, _ := ctx.Get("user_id")

	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	respondingServerAddress := scheme + "://" + ctx.Request.Host + ctx.Request.RequestURI
	if h.flags.GetRespondingServer() != "" {
		respondingServerAddress = h.flags.GetRespondingServer() + "/"
	}

	var shortURL string
	switch h.GetStorageType() {
	case "database":
		db := database.NewInitializerReaderWriter(h.repo, h.flags)
		if err := db.CreateTableIfNotExists(); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		url := &model.URL{}
		err := json.Unmarshal(body, &url)
		if err != nil {
			url.OriginalURL = string(body)
		}
		url.UserID = userID.(int)
		shortURL, err = db.WriteURL(url)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
					ctx.String(http.StatusConflict, "%v%v", respondingServerAddress, shortURL)
					return
				}
			}
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	case "file":
		Producer, err := file.NewProducer(h.flags.GetFileStoragePath())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		defer Producer.Close()
		bufEvent := &model.URL{}
		err = json.Unmarshal(body, bufEvent)
		if err != nil {
			bufEvent.OriginalURL = string(body)
			shortURL, err = h.repo.WriteURL(bufEvent.OriginalURL)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
			bufEvent.ShortURL = shortURL
		} else {
			shortURL = bufEvent.ShortURL
			err = h.repo.WriteShortAndOriginalURL(bufEvent.ShortURL, bufEvent.OriginalURL)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
		}
		Producer.WriteEvent(bufEvent)
	case "memory":
		originalURL := string(body)
		shortURL, err = h.repo.WriteURL(originalURL)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	ctx.String(http.StatusCreated, "%v%v", respondingServerAddress, shortURL)
}

func (h *handlerURL) GetFullURL(ctx *gin.Context) {
	var (
		deletedFlag bool
		fullURL     string
		err         error
	)
	shortURL := ctx.Params.ByName("short_url")

	switch h.GetStorageType() {
	case "database":
		db := database.NewInitializerReaderWriter(h.repo, h.flags)

		deletedFlag, err = db.CheckRecordDeletedOrNot(shortURL)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		if deletedFlag {
			ctx.Status(http.StatusGone)
			return
		}
		fullURL, err = db.ReadURL(shortURL)
	default:
		fullURL, err = h.repo.ReadURL(shortURL)
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	ctx.Header("Location", fullURL)
	ctx.Status(http.StatusTemporaryRedirect)
}

func (h *handlerURL) PingDatabaseServer(ctx *gin.Context) {
	if err := h.repo.GetDatabase().Ping(); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *handlerURL) GetStorageType() string {
	if h.flags.GetDatabaseDSN() != "" {
		return "database"
	} else if h.flags.GetFileStoragePath() != "" {
		return "file"
	} else {
		return "memory"
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
