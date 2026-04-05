package handler

import (
	"avito-talk/internal/api"
	"avito-talk/internal/domain"
	"avito-talk/internal/service"
	"avito-talk/internal/testutil"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// --- test setup ---

type testEnv struct {
	router      http.Handler
	authService *service.AuthService
	roomRepo    *testutil.MockRoomRepo
	slotRepo    *testutil.MockSlotRepo
}

func setupTestEnv() *testEnv {
	roomRepo := testutil.NewMockRoomRepo()
	scheduleRepo := testutil.NewMockScheduleRepo()
	slotRepo := testutil.NewMockSlotRepo()
	bookingRepo := testutil.NewMockBookingRepo()

	authService := service.NewAuthService("test-secret")
	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo, roomRepo)
	slotService := service.NewSlotService(slotRepo, roomRepo)
	bookingService := service.NewBookingService(bookingRepo, slotRepo)

	srv := NewServer(authService, roomService, scheduleService, slotService, bookingService)

	r := chi.NewRouter()
	r.Get("/_info", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	wrapper := api.ServerInterfaceWrapper{
		Handler:          srv,
		ErrorHandlerFunc: OapiErrorHandler,
	}

	authMw := AuthMiddleware(authService)

	// публичные
	r.Post("/dummyLogin", wrapper.PostDummyLogin)
	r.Post("/login", wrapper.PostLogin)
	r.Post("/register", wrapper.PostRegister)

	// любой авторизованный
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Get("/rooms/list", wrapper.GetRoomsList)
		r.Get("/rooms/{roomId}/slots/list", wrapper.GetRoomsRoomIdSlotsList)
	})

	// admin
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(RequireRole("admin"))
		r.Post("/rooms/create", wrapper.PostRoomsCreate)
		r.Post("/rooms/{roomId}/schedule/create", wrapper.PostRoomsRoomIdScheduleCreate)
		r.Get("/bookings/list", wrapper.GetBookingsList)
	})

	// user
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(RequireRole("user"))
		r.Post("/bookings/create", wrapper.PostBookingsCreate)
		r.Get("/bookings/my", wrapper.GetBookingsMy)
		r.Post("/bookings/{bookingId}/cancel", wrapper.PostBookingsBookingIdCancel)
	})

	return &testEnv{
		router:      r,
		authService: authService,
		roomRepo:    roomRepo,
		slotRepo:    slotRepo,
	}
}

func getToken(env *testEnv, role string) string {
	var userID string
	if role == "admin" {
		userID = "11111111-1111-1111-1111-111111111111"
	} else {
		userID = "22222222-2222-2222-2222-222222222222"
	}
	token, _ := env.authService.GenerateToken(userID, role)
	return token
}

func doRequest(env *testEnv, method, path string, body any, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	return w
}

// seedSlots добавляет слоты напрямую в мок (эмулирует работу джобы генерации).
func seedSlots(env *testEnv, roomID string, date time.Time, startHour, endHour int) {
	rid := uuid.MustParse(roomID)
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	for h := startHour; h < endHour; h++ {
		for m := 0; m < 60; m += 30 {
			s := &domain.Slot{
				ID:     uuid.New(),
				RoomID: rid,
				Start:  d.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute),
				End:    d.Add(time.Duration(h)*time.Hour + time.Duration(m+30)*time.Minute),
			}
			env.slotRepo.Slots[s.ID] = s
		}
	}
}

// --- Tests ---

func TestInfoEndpoint(t *testing.T) {
	env := setupTestEnv()
	w := doRequest(env, "GET", "/_info", nil, "")
	if w.Code != 200 {
		t.Errorf("/_info status = %d, want 200", w.Code)
	}
}

func TestDummyLogin(t *testing.T) {
	env := setupTestEnv()
	w := doRequest(env, "POST", "/dummyLogin", map[string]string{"role": "admin"}, "")
	if w.Code != 200 {
		t.Fatalf("/dummyLogin status = %d, want 200", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("token is empty")
	}
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	env := setupTestEnv()
	w := doRequest(env, "POST", "/dummyLogin", map[string]string{"role": "superadmin"}, "")
	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreateRoom_AdminOnly(t *testing.T) {
	env := setupTestEnv()
	userToken := getToken(env, "user")
	w := doRequest(env, "POST", "/rooms/create", map[string]string{"name": "Room"}, userToken)
	if w.Code != 403 {
		t.Errorf("user creating room: status = %d, want 403", w.Code)
	}
}

func TestCreateRoom_Success(t *testing.T) {
	env := setupTestEnv()
	adminToken := getToken(env, "admin")
	w := doRequest(env, "POST", "/rooms/create", map[string]any{"name": "Room A", "capacity": 10}, adminToken)
	if w.Code != 201 {
		t.Fatalf("create room: status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestListRooms(t *testing.T) {
	env := setupTestEnv()
	adminToken := getToken(env, "admin")

	doRequest(env, "POST", "/rooms/create", map[string]string{"name": "Room 1"}, adminToken)
	doRequest(env, "POST", "/rooms/create", map[string]string{"name": "Room 2"}, adminToken)

	w := doRequest(env, "GET", "/rooms/list", nil, adminToken)
	if w.Code != 200 {
		t.Fatalf("list rooms: status = %d", w.Code)
	}
	var resp map[string][]any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp["rooms"]) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(resp["rooms"]))
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	env := setupTestEnv()
	w := doRequest(env, "GET", "/rooms/list", nil, "")
	if w.Code != 401 {
		t.Errorf("unauthenticated: status = %d, want 401", w.Code)
	}
}

func TestCreateBooking_AdminForbidden(t *testing.T) {
	env := setupTestEnv()
	adminToken := getToken(env, "admin")
	w := doRequest(env, "POST", "/bookings/create", map[string]string{"slotId": uuid.New().String()}, adminToken)
	if w.Code != 403 {
		t.Errorf("admin booking: status = %d, want 403", w.Code)
	}
}

// E2E test: создание переговорки → расписание → бронь
func TestE2E_CreateRoomScheduleBooking(t *testing.T) {
	env := setupTestEnv()
	adminToken := getToken(env, "admin")
	userToken := getToken(env, "user")

	// 1. Создаём переговорку
	w := doRequest(env, "POST", "/rooms/create", map[string]string{"name": "Meeting Room"}, adminToken)
	if w.Code != 201 {
		t.Fatalf("create room: %d %s", w.Code, w.Body.String())
	}
	var roomResp map[string]map[string]any
	json.NewDecoder(w.Body).Decode(&roomResp)
	roomID := roomResp["room"]["id"].(string)

	// 2. Создаём расписание (все дни недели, 09:00-18:00)
	w = doRequest(env, "POST", fmt.Sprintf("/rooms/%s/schedule/create", roomID),
		map[string]any{
			"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
			"startTime":  "09:00",
			"endTime":    "18:00",
		}, adminToken)
	if w.Code != 201 {
		t.Fatalf("create schedule: %d %s", w.Code, w.Body.String())
	}

	// 3. Генерируем слоты (эмуляция ночной джобы) и получаем на завтра
	tomorrow := time.Now().AddDate(0, 0, 1)
	seedSlots(env, roomID, tomorrow, 9, 18)

	w = doRequest(env, "GET", fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, tomorrow.Format("2006-01-02")), nil, userToken)
	if w.Code != 200 {
		t.Fatalf("get slots: %d %s", w.Code, w.Body.String())
	}
	var slotsResp map[string][]map[string]any
	json.NewDecoder(w.Body).Decode(&slotsResp)
	if len(slotsResp["slots"]) == 0 {
		t.Fatal("expected slots, got 0")
	}
	slotID := slotsResp["slots"][0]["id"].(string)

	// 4. Бронируем слот
	w = doRequest(env, "POST", "/bookings/create", map[string]string{"slotId": slotID}, userToken)
	if w.Code != 201 {
		t.Fatalf("create booking: %d %s", w.Code, w.Body.String())
	}
	var bookingResp map[string]map[string]any
	json.NewDecoder(w.Body).Decode(&bookingResp)
	if bookingResp["booking"]["status"] != "active" {
		t.Errorf("booking status = %v, want active", bookingResp["booking"]["status"])
	}

	// 5. Проверяем, что слот теперь занят (повторная бронь = 409)
	w = doRequest(env, "POST", "/bookings/create", map[string]string{"slotId": slotID}, userToken)
	if w.Code != 409 {
		t.Errorf("double booking: status = %d, want 409", w.Code)
	}

	// 6. Проверяем список моих броней
	w = doRequest(env, "GET", "/bookings/my", nil, userToken)
	if w.Code != 200 {
		t.Fatalf("my bookings: %d %s", w.Code, w.Body.String())
	}
}

// E2E test: отмена бронирования
func TestE2E_CancelBooking(t *testing.T) {
	env := setupTestEnv()
	adminToken := getToken(env, "admin")
	userToken := getToken(env, "user")

	// создаём комнату + расписание
	w := doRequest(env, "POST", "/rooms/create", map[string]string{"name": "Room X"}, adminToken)
	var roomResp map[string]map[string]any
	json.NewDecoder(w.Body).Decode(&roomResp)
	roomID := roomResp["room"]["id"].(string)

	// генерируем слоты (эмуляция ночной джобы)
	tomorrow := time.Now().AddDate(0, 0, 1)
	seedSlots(env, roomID, tomorrow, 10, 11)

	// получаем слоты
	w = doRequest(env, "GET", fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, tomorrow.Format("2006-01-02")), nil, userToken)
	var slotsResp map[string][]map[string]any
	json.NewDecoder(w.Body).Decode(&slotsResp)
	slotID := slotsResp["slots"][0]["id"].(string)

	// бронируем
	w = doRequest(env, "POST", "/bookings/create", map[string]string{"slotId": slotID}, userToken)
	var bookResp map[string]map[string]any
	json.NewDecoder(w.Body).Decode(&bookResp)
	bookingID := bookResp["booking"]["id"].(string)

	// отменяем
	w = doRequest(env, "POST", fmt.Sprintf("/bookings/%s/cancel", bookingID), nil, userToken)
	if w.Code != 200 {
		t.Fatalf("cancel: %d %s", w.Code, w.Body.String())
	}
	var cancelResp map[string]map[string]any
	json.NewDecoder(w.Body).Decode(&cancelResp)
	if cancelResp["booking"]["status"] != "cancelled" {
		t.Errorf("status = %v, want cancelled", cancelResp["booking"]["status"])
	}

	// повторная отмена — идемпотентна
	w = doRequest(env, "POST", fmt.Sprintf("/bookings/%s/cancel", bookingID), nil, userToken)
	if w.Code != 200 {
		t.Errorf("idempotent cancel: %d, want 200", w.Code)
	}

	// слот снова доступен для бронирования
	w = doRequest(env, "POST", "/bookings/create", map[string]string{"slotId": slotID}, userToken)
	if w.Code != 201 {
		t.Errorf("re-booking after cancel: %d, want 201", w.Code)
	}
}

// Test: отмена чужой брони — 403
func TestCancelBooking_OtherUser(t *testing.T) {
	env := setupTestEnv()
	adminToken := getToken(env, "admin")
	userToken := getToken(env, "user")

	w := doRequest(env, "POST", "/rooms/create", map[string]string{"name": "Room Z"}, adminToken)
	var roomResp map[string]map[string]any
	json.NewDecoder(w.Body).Decode(&roomResp)
	roomID := roomResp["room"]["id"].(string)

	// генерируем слоты
	tomorrow := time.Now().AddDate(0, 0, 1)
	seedSlots(env, roomID, tomorrow, 10, 11)

	w = doRequest(env, "GET", fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, tomorrow.Format("2006-01-02")), nil, userToken)
	var slotsResp map[string][]map[string]any
	json.NewDecoder(w.Body).Decode(&slotsResp)
	slotID := slotsResp["slots"][0]["id"].(string)

	w = doRequest(env, "POST", "/bookings/create", map[string]string{"slotId": slotID}, userToken)
	var bookResp map[string]map[string]any
	json.NewDecoder(w.Body).Decode(&bookResp)
	bookingID := bookResp["booking"]["id"].(string)

	// admin пытается отменить — 403 (не user)
	w = doRequest(env, "POST", fmt.Sprintf("/bookings/%s/cancel", bookingID), nil, adminToken)
	if w.Code != 403 {
		t.Errorf("admin cancel: status = %d, want 403", w.Code)
	}
}
