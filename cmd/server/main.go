package main

import (
	"context"
	"errors"
	"gmetrics/cmd/server/config"
	"gmetrics/cmd/server/handlers/getmetric"
	"gmetrics/cmd/server/handlers/getmetrics"
	"gmetrics/cmd/server/handlers/handlemetric"
	"gmetrics/cmd/server/handlers/ping"
	"gmetrics/internal/buildflags"
	"gmetrics/internal/contextkeys"
	"gmetrics/internal/database"
	"gmetrics/internal/encrypt"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"gmetrics/internal/middlewares"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"
)

func main() {
	buildflags.PrintBuildInformation()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// Устанавливаем настройки
	cnf, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	config.Params = cnf

	// стартуем приложение
	if err = runApplication(); err != nil {
		logger.Log.Error(err)
	}

	logger.Log.Info("End program")
}

// runApplication производим старт приложения
func runApplication() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	//ctx, cancel := context.WithCancel(context.Background()) // Контекст для правильной остановки синхронизации
	defer func() {
		logger.Log.Info("Cancel context")
		cancel()
	}()

	_, err := logger.InitLogger(config.Params.LogLevel)
	if err != nil {
		return err
	}

	// Показываем конфигурацию сервера
	logger.Log.Infow("Running server with configuration",
		"address", config.Params.Address,
		"logLevel", config.Params.LogLevel,
		"fileStorage", config.Params.FileStorage,
		"restore", config.Params.Restore,
		"storeInterval", config.Params.StoreInterval,
		"databaseDSN", config.Params.DatabaseDSN,
	)

	// Вызываем функцию закрытия базы данных
	defer closeDB()
	// Инициализируем базу данных
	//err = initDB(ctx, wg)
	err = initDB()
	if err != nil {
		return err
	}

	// Вызываем функцию закрытия
	defer closeStorage()
	wg := new(errgroup.Group)
	//wg := sync.WaitGroup{} // Группа для синхронизации
	// Инициализируем хранилище
	InitStore(ctx)
	// Запускаем синхронизацию хранилища, если оно это подразумевает
	if st, ok := metrics.MeStore.(metrics.ISynchronizationStorage); ok {
		ctx = context.WithValue(ctx, contextkeys.SyncInterval, config.Params.StoreInterval)
		// Запускаем синхронизацию в файл
		if !st.IsSyncMode() {
			wg.Go(func() error {
				return st.Sync(ctx)
			})
		}
	}

	server := initServer()
	// Запускаем сервер
	wg.Go(func() error {
		sErr := server.ListenAndServe()
		if sErr != nil && !errors.Is(sErr, http.ErrServerClosed) {
			return sErr
		}
		return nil
	})
	// Регистрируем прослушиватель для закрытия записи в файл и завершения сервера
	<-ctx.Done()
	logger.Log.Info("Stopping server")
	//cancel()
	if err = stopServer(server, ctx); err != nil { // Запускаем сервер
		return err
	}

	// Ожидаем завершения всех горутин перед завершением программы
	if err = wg.Wait(); err != nil {
		logger.Log.Error(err)
	}
	return nil
}

// closeStorage функция закрытия хранилища
func closeStorage() {
	st, ok := metrics.MeStore.(metrics.ISynchronizationStorage)
	if !ok {
		return
	}
	logger.Log.Info("Close storage")
	if dErr := st.FlushAndClose(); dErr != nil {
		logger.Log.Error(dErr)
	}
	logger.Log.Info("Storage closed")
}

// run запуск сервера
func initServer() *http.Server {
	logger.Log.Infof("Running server on %s", config.Params.Address)
	server := http.Server{
		Addr:    config.Params.Address,
		Handler: getRouter(),
	}

	return &server
}

// stopServer корректно завершает работу предоставленного HTTP-сервера, используя заданный контекст. Регистрирует ошибки в случае сбоя завершения работы.
func stopServer(server *http.Server, ctx context.Context) error {
	// Заставляем завершиться сервер и ждём его завершения
	err := server.Shutdown(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to shutdown server: %v", err)
	}
	logger.Log.Info("Server stop")

	return err
}

// getRouter конфигурация роутинга приложение
func getRouter() chi.Router {
	router := chi.NewRouter()
	decrypter := encrypt.NewDecrypter(config.Params.CryptoKey)
	// Устанавилваем мидлваре
	router.Use(
		cMiddleware.StripSlashes,         // Убираем лишние слеши
		logger.LogRequests,               // Логируем данные запроса
		middlewares.GZIPCompressResponse, // Сжимаем ответ TODO исключить для роутов, которые будут возвращать не application/json или text/html. Проверять в мидлваре или компрессоре может быть не эффективно,так как заголовок с контентом может быть поставлен позже записи контента
		middlewares.CheckSign,
		middlewares.GZIPDecompressRequest, // Разжимаем тело ответа
		decrypter.Middleware,
	)
	// Сохранение метрики по URL
	router.Post("/update/{type}/{name}/{value}", handlemetric.URLHandler)
	// Получение всех метрик
	router.Get("/", getmetrics.Handler)
	// Получение отдельной метрики
	router.Get("/value/{type}/{name}", getmetric.URLHandler)

	// проверка состояния соединения с базой данных
	router.Get("/ping", ping.NewController(database.DB).Handler)

	router.Group(func(r chi.Router) {
		// Устанавилваем мидлваре
		r.Use(middlewares.JSONHeaders)

		router.Group(func(r chi.Router) {
			// Сохранение метрики с помощью JSON тела
			r.Post("/update", handlemetric.JSONHandler)
			// Сохранение метрик с помощью JSON тела
			r.Post("/updates", handlemetric.JSONManyHandler)
		})
		// Получение отдельной метрики
		r.Post("/value", getmetric.JSONHandler)
	})
	return router
}

// InitStore устанавливаем глобальное хранилище метрик.
func InitStore(ctx context.Context) {
	// Если указан путь к файлу, то будет создано хранилище с сохранением в файл, иначе будет создано хранилище в памяти
	if config.Params.DatabaseDSN != "" {
		logger.Log.Info("Set database store")
		store, err := metrics.NewDBStorage(ctx, metrics.NewDBAdapter(database.DB), config.Params.Restore, config.Params.StoreInterval == 0)
		if err != nil {
			logger.Log.Fatal(err)
		}
		metrics.MeStore = store
	} else if config.Params.FileStorage != "" {
		logger.Log.Info("Set file store")
		store, err := metrics.NewFileStorage(config.Params.FileStorage, config.Params.Restore, config.Params.StoreInterval == 0)
		if err != nil {
			logger.Log.Fatal(err)
		}
		metrics.MeStore = store
	} else {
		logger.Log.Info("Set in-memory store")
		metrics.MeStore = metrics.NewMemStorage()
	}
}

// initDB инициализация подключения к бд
// func initDB(ctx context.Context, wg *errgroup.Group*) error {
func initDB() error {
	// Создание пула подключений к базе данных для приложения
	var err error
	database.DB, err = database.NewPgDB(config.Params.DatabaseDSN)
	if err != nil {
		return err
	}

	if config.Params.DatabaseDSN != "" {
		logger.Log.Info("Migrate migrations")
		// Применим миграции
		migrator, err := migrations()
		if err != nil {
			return err
		}
		if err = migrator.Migrate(database.DB); err != nil {
			return err
		}
	}

	return nil
}

// closeDB закрытие базы данных
func closeDB() {
	logger.Log.Info("Closing database connection for defer")
	if database.DB != nil {
		err := database.DB.Close()
		if err != nil {
			logger.Log.Error(err)
		}
	}
}
