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
	"gmetrics/cmd/server/handlers/ping"
	"gmetrics/internal/contextkeys"
	"gmetrics/internal/database"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"gmetrics/internal/middlewares"
	"go.uber.org/zap"
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

	// стартуем приложение
	if err = runApplication(&wg); err != nil {
		logger.Log.Error(err)
	}

	wg.Wait() // Ожидаем завершения всех горутин перед завершением программы
	logger.Log.Info("End program")
}

// runApplication производим старт приложения
func runApplication(wg *sync.WaitGroup) error {
	ctx, cancel := context.WithCancel(context.Background()) // Контекст для правильной остановки синхронизации
	defer func() {
		logger.Log.Info("Cancel context")
		cancel()
	}()
	// Регистрируем прослушиватель для закрытия записи в файл и завершения сервера
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		logger.Log.Info("Stopping server")
		cancel()
	}()
	_, err := InitLogger()
	if err != nil {
		return err
	}

	// Вызываем функцию закрытия базы данных
	defer closeDB()
	// Инициализируем базу данных
	err = initDB(ctx, wg)
	if err != nil {
		return err
	}

	// Вызываем функцию закрытия
	defer closeStorage()
	// Инициализируем хранилище
	InitStore(ctx, wg)

	if err = run(ctx, wg); err != nil { // Запускаем сервер
		return err
	}

	return nil
}

// closeStorage функция закрытия хранилища
func closeStorage() {
	st, ok := metrics.MeStore.(*metrics.DurationFileStorage)
	if !ok {
		return
	}
	logger.Log.Info("Close storage")
	if dErr := st.FlushAndClose(); dErr != nil {
		logger.Log.Error(dErr)
	}
}

// run запуск сервера
func run(ctx context.Context, wg *sync.WaitGroup) (err error) {
	logger.Log.Infof("Running server on %s", config.Params.Address)
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
	defer wg.Done()
	<-ctx.Done()
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
	// Устанавилваем мидлваре
	router.Use(
		cMiddleware.StripSlashes,          // Убираем лишние слеши
		logger.LogRequests,                // Логируем данные запроса
		middlewares.GZIPCompressResponse,  // Сжимаем ответ TODO исключить для роутов, которые будут возвращать не application/json или text/html. Проверять в мидлваре или компрессоре может быть не эффективно,так как заголовок с контентом может быть поставлен позже записи контента
		middlewares.GZIPDecompressRequest, // Разжимаем тело ответа
	)
	// Сохранение метрики по URL
	router.Post("/update/{type}/{name}/{value}", handlemetric.URLHandler)
	// Получение всех метрик
	router.Get("/", getmetrics.Handler)
	// Получение отдельной метрики
	router.Get("/value/{type}/{name}", getmetric.URLHandler)

	// проверка состояния соединения с базой данных
	router.Get("/ping", ping.Handler)

	router.Group(func(r chi.Router) {
		// Устанавилваем мидлваре с логированием запросов
		r.Use(middlewares.JSONHeaders)
		// Сохранение метрики с помощью JSON тела
		r.Post("/update", handlemetric.JSONHandler)
		// Сохранение метрик с помощью JSON тела
		r.Post("/updates", handlemetric.JSONManyHandler)
		// Получение отдельной метрики
		r.Post("/value", getmetric.JSONHandler)
	})
	return router
}

// InitStore устанавливаем глобальное хранилище метрик.
func InitStore(ctx context.Context, wg *sync.WaitGroup) {
	// Если указан путь к файлу, то будет создано хранилище с сохранением в файл, иначе будет создано хранилище в памяти
	if config.Params.DatabaseDSN != "" {
		logger.Log.Info("Set database store")
		store, err := metrics.NewDBStorage(ctx, database.DB)
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
		ctx = context.WithValue(ctx, contextkeys.SyncInterval, config.Params.StoreInterval)
		// Запускаем синхронизацию в файл
		if !store.SyncMode {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store.Sync(ctx)
			}()
		}
	} else {
		logger.Log.Info("Set in-memory store")
		metrics.MeStore = metrics.NewMemStorage()
	}
}

// InitLogger инициализируем логер
func InitLogger() (*zap.SugaredLogger, error) {
	lgr, err := logger.New(config.Params.LogLevel)
	if err != nil {
		return nil, err
	}
	logger.Log = lgr
	// Показываем конфигурацию сервера
	logger.Log.Infow("Running server with configuration",
		"address", config.Params.Address,
		"logLevel", config.Params.LogLevel,
		"fileStorage", config.Params.FileStorage,
		"restore", config.Params.Restore,
		"storeInterval", config.Params.StoreInterval,
		"databaseDSN", config.Params.DatabaseDSN,
	)

	return lgr, nil
}

// initDB инициализация подключения к бд
func initDB(ctx context.Context, wg *sync.WaitGroup) error {
	// Создание пула подключений к базе данных для приложения
	var err error
	database.DB, err = database.NewPgDB(config.Params.DatabaseDSN)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		logger.Log.Info("Closing database connection for context")
		err = database.DB.Close()
		if err != nil {
			logger.Log.Error(err)
		}
	}()

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
