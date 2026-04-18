package main

import (
	"avito-talk/internal/api"
	"avito-talk/internal/handler"
	"avito-talk/internal/job"
	"avito-talk/internal/repository"
	"avito-talk/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Не найден .env файл")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL не задан")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET не задан")
	}

	ctx := context.Background()
	db, err := repository.ConnectionDB(ctx, dbURL)
	if err != nil {
		log.Fatalf("DB connection: %v", err)
	}
	defer db.CloseDB()

	// репозитории
	roomRepos := repository.NewRoomRepository(db)
	scheduleRepos := repository.NewScheduleRepository(db)
	slotRepos := repository.NewSlotRepository(db)
	bookingRepos := repository.NewBookingRepository(db)

	// сервисы
	authService := service.NewAuthService(jwtSecret)
	roomService := service.NewRoomService(roomRepos)
	scheduleService := service.NewScheduleService(scheduleRepos, roomRepos)
	slotService := service.NewSlotService(slotRepos, roomRepos)
	bookingService := service.NewBookingService(bookingRepos, slotRepos)

	// джоба генерации слотов: запуск сразу + каждую ночь
	slotGen := job.NewSlotGenerator(roomRepos, scheduleRepos, slotRepos)
	jobCtx, jobCancel := context.WithCancel(ctx)
	defer jobCancel()
	go slotGen.Start(jobCtx)

	// хендлер, реализующий api.ServerInterface
	srv := handler.NewServer(authService, roomService, scheduleService, slotService, bookingService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/_info", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// обёртка из сгенерированного кода: парсит параметры и ставит BearerAuthScopes
	wrapper := api.ServerInterfaceWrapper{
		Handler:          srv,
		ErrorHandlerFunc: handler.OapiErrorHandler,
	}

	authMw := handler.AuthMiddleware(authService)

	// публичные эндпоинты — без авторизации
	r.Group(func(r chi.Router) {
		r.Post("/dummyLogin", wrapper.PostDummyLogin)
		r.Post("/login", wrapper.PostLogin)
		r.Post("/register", wrapper.PostRegister)
	})

	// эндпоинты, доступные любому авторизованному пользователю
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Get("/rooms/list", wrapper.GetRoomsList)
		r.Get("/rooms/{roomId}/slots/list", wrapper.GetRoomsRoomIdSlotsList)
	})

	// эндпоинты только для admin
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(handler.RequireRole("admin"))
		r.Post("/rooms/create", wrapper.PostRoomsCreate)
		r.Post("/rooms/{roomId}/schedule/create", wrapper.PostRoomsRoomIdScheduleCreate)
		r.Get("/bookings/list", wrapper.GetBookingsList)
	})

	// эндпоинты только для user
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(handler.RequireRole("user"))
		r.Post("/bookings/create", wrapper.PostBookingsCreate)
		r.Get("/bookings/my", wrapper.GetBookingsMy)
		r.Post("/bookings/{bookingId}/cancel", wrapper.PostBookingsBookingIdCancel)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	httpSrv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server started on port %s", port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Выключение...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Ошибка завершения работы: %v", err)
	}
}
