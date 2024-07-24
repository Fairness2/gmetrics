package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	"gmetrics/cmd/server/config"
	"gmetrics/cmd/server/handlers/getmetric"
	"gmetrics/cmd/server/handlers/getmetrics"
	"gmetrics/cmd/server/handlers/handlemetric"
	"gmetrics/internal/contextkeys"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"gmetrics/internal/middlewares"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// Устанавливаем настройки
	cnf, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	config.Params = cnf

	wg := sync.WaitGroup{} // Группа для синхронизации
	defer func() {
		wg.Wait() // Ожидаем завершения всех горутин перед завершением программы
		logger.G.Info("End program")
	}()

	ctx, cancel := context.WithCancel(context.Background()) // Контекст для правильной остановки синхронизации
	defer func() {
		logger.G.Info("Cancel context")
		cancel()
	}()
	// Регистрируем прослушиватель для закрытия записи в файл и завершения сервера
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		logger.G.Info("Stopping server")
		cancel()
	}()

	InitLogger(func() {
		InitStore(func() {
			if err := run(ctx, &wg); err != nil { // Запускаем сервер
				logger.G.Error(err)
			}
		}, ctx, &wg)
	}, ctx)
}

// run запуск сервера
func run(ctx context.Context, wg *sync.WaitGroup) (err error) {
	logger.G.Infof("Running server on %s", config.Params.Address)
	server := http.Server{
		Addr:    config.Params.Address,
		Handler: getRouter(),
	}
	wg.Add(1)
	go func() {
		err = stopServer(&server, ctx, wg)
	}()
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func stopServer(server *http.Server, ctx context.Context, wg *sync.WaitGroup) error {
	<-ctx.Done()
	// Заставляем завершиться сервер и ждём его завершения
	err := server.Shutdown(ctx)
	if err != nil {
		logger.G.Errorf("Failed to shutdown server: %v", err)
	}
	logger.G.Info("Server stop")
	wg.Done()
	return err
}

// getRouter конфигурация роутинга приложение
func getRouter() chi.Router {
	router := chi.NewRouter()
	// Устанавилваем мидлваре
	router.Use(
		cMiddleware.StripSlashes,          // Убираем лишние слеши
		logger.LogRequests,                // Логируем данные запроса
		middlewares.GZIPCompressResponse,  // Сжимаем ответ TODO исключить для роутов, которые будут возвращать не application/json или text/html. Проверять в мидлваре или компрессоре может быть не эффективно,так как заголовок с контентом может быть поставлен позже записи контента
		logger.LogResponse,                // Логируем данные ответа
		middlewares.GZIPDecompressRequest, // Разжимаем тело ответа
	)
	// Сохранение метрики по URL
	router.Post("/update/{type}/{name}/{value}", handlemetric.URLHandler)
	// Получение всех метрик
	router.Get("/", getmetrics.Handler)
	// Получение отдельной метрики
	router.Get("/value/{type}/{name}", getmetric.URLHandler)

	router.Group(func(r chi.Router) {
		// Устанавилваем мидлваре с логированием запросов
		r.Use(middlewares.JSONHeaders)
		// Сохранение метрики с помощью JSON тела
		r.Post("/update", handlemetric.JSONHandler)
		// Получение отдельной метрики
		r.Post("/value", getmetric.JSONHandler)
	})
	return router
}

type next func()

// InitStore устанавливаем глобальное хранилище метрик.
func InitStore(n next, ctx context.Context, wg *sync.WaitGroup) {
	//Если указан путь к файлу, то будет создано хранилище с сохранением в файл, иначе будет создано хранилище в памяти
	if config.Params.FileStorage != "" {
		logger.G.Info("Set file store")
		store, err := metrics.NewFileStorage(config.Params.FileStorage, config.Params.Restore, config.Params.StoreInterval == 0)
		if err != nil {
			logger.G.Fatal(err)
		}
		defer func() {
			logger.G.Info("Close storage")
			if dErr := store.FlushAndClose(); dErr != nil {
				logger.G.Error(dErr)
			}
		}()
		metrics.MeStore = store
		ctx = context.WithValue(ctx, contextkeys.SyncInterval, config.Params.StoreInterval)
		// Запускаем синхронизацию в файл
		if !store.SyncMode {
			wg.Add(1)
			go func() {
				store.Sync(ctx)
				defer wg.Done()
			}()
		}
	} else {
		logger.G.Info("Set in-memory store")
		metrics.MeStore = metrics.NewMemStorage()
	}

	n()
}

// InitLogger инициализируем логер
func InitLogger(n next, ctx context.Context) {
	lgr, err := logger.New(config.Params.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.G = lgr
	// Показываем конфигурацию сервера
	logger.G.Infow("Running server with configuration",
		"address", config.Params.Address,
		"logLevel", config.Params.LogLevel,
		"fileStorage", config.Params.FileStorage,
		"restore", config.Params.Restore,
		"storeInterval", config.Params.StoreInterval,
	)

	n()
}
