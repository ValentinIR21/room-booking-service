package main

import (
	"avito-talk/internal/handler"
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
	scheduleService := service.NewScheduleService(scheduleRepos)
	slotService := service.NewSlotService(slotRepos, scheduleRepos)
	bookingService := service.NewBookingService(bookingRepos, slotRepos)

	// хендлеры
	authHandler := handler.NewAuthHandler(authService)
	roomHandler := handler.NewRoomHandler(roomService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	slotHandler := handler.NewSlotHandler(slotService)
	bookingHandler := handler.NewBookingHandler(bookingService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// публичные
	r.Post("/dummyLogin", authHandler.DummyLogin)
	r.Get("/_info", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// защищённые
	r.Group(func(r chi.Router) {
		r.Use(handler.AuthMiddleware(authService))

		r.Get("/rooms/list", roomHandler.ListRooms)
		r.Post("/rooms/create", roomHandler.CreateRoom)

		r.Post("/rooms/{roomId}/schedule/create", scheduleHandler.CreateSchedule)

		r.Get("/rooms/{roomId}/slots/list", slotHandler.GetAvailableSlots)

		r.Post("/bookings/create", bookingHandler.CreateBooking)
		r.Get("/bookings/my", bookingHandler.MyBookings)
		r.Get("/bookings/list", bookingHandler.ListAllBookings)
		r.Post("/bookings/{bookingId}/cancel", bookingHandler.CancelBooking)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Выключение...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Ошибка завершения работы: %v", err)
	}
}
