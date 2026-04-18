package job

import (
	"avito-talk/internal/domain"
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/google/uuid"
)

const (
	slotDuration = 30 * time.Minute
	generateDays = 30
	nightHourUTC = 2 // запуск в 02:00 UTC
)

// SlotGenerator генерирует слоты для всех комнат.
type SlotGenerator struct {
	roomRepo     RoomRepo
	scheduleRepo ScheduleRepo
	slotRepo     SlotRepo
}

type RoomRepo interface {
	List(ctx context.Context) ([]domain.Room, error)
}

type ScheduleRepo interface {
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error)
}

type SlotRepo interface {
	UpsertSlots(ctx context.Context, slots []domain.Slot) error
	LastSlotDate(ctx context.Context, roomID uuid.UUID) (*time.Time, error)
}

func NewSlotGenerator(roomRepo RoomRepo, scheduleRepo ScheduleRepo, slotRepo SlotRepo) *SlotGenerator {
	return &SlotGenerator{
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
		slotRepo:     slotRepo,
	}
}

// GenerateAll генерирует слоты для всех комнат на generateDays дней вперёд.
// Для каждой комнаты начинает с дня после последнего существующего слота (или с сегодня, если слотов нет).
func (g *SlotGenerator) GenerateAll(ctx context.Context) error {
	rooms, err := g.roomRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list rooms: %w", err)
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := today.AddDate(0, 0, generateDays)

	for _, room := range rooms {
		schedule, err := g.scheduleRepo.GetByRoomID(ctx, room.ID)
		if err != nil {
			log.Printf("slot_generator: schedule for room %s: %v", room.ID, err)
			continue
		}
		if schedule == nil {
			continue
		}

		from := today
		lastSlotDate, err := g.slotRepo.LastSlotDate(ctx, room.ID)
		if err != nil {
			log.Printf("slot_generator: last slot date for room %s: %v", room.ID, err)
			continue
		}
		if lastSlotDate != nil {
			// начинаем со следующего дня после последнего слота
			nextDay := time.Date(lastSlotDate.Year(), lastSlotDate.Month(), lastSlotDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
			if nextDay.After(from) {
				from = nextDay
			}
		}

		if !from.Before(end) {
			continue
		}

		slots, err := g.generateForRoom(room.ID, schedule, from, end)
		if err != nil {
			log.Printf("slot_generator: generate for room %s: %v", room.ID, err)
			continue
		}
		if len(slots) == 0 {
			continue
		}
		if err := g.slotRepo.UpsertSlots(ctx, slots); err != nil {
			log.Printf("slot_generator: upsert for room %s: %v", room.ID, err)
		}
	}

	return nil
}

func (g *SlotGenerator) generateForRoom(roomID uuid.UUID, schedule *domain.Schedule, from, to time.Time) ([]domain.Slot, error) {
	startTime, err := parseTimeStr(schedule.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}
	endTime, err := parseTimeStr(schedule.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	var slots []domain.Slot
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		weekday := int(d.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if !slices.Contains(schedule.DaysOfWeek, weekday) {
			continue
		}

		slotStart := time.Date(d.Year(), d.Month(), d.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
		slotEnd := time.Date(d.Year(), d.Month(), d.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

		for ts := slotStart; ts.Add(slotDuration).Before(slotEnd) || ts.Add(slotDuration).Equal(slotEnd); ts = ts.Add(slotDuration) {
			slots = append(slots, domain.Slot{
				ID:     uuid.New(),
				RoomID: roomID,
				Start:  ts,
				End:    ts.Add(slotDuration),
			})
		}
	}

	return slots, nil
}

// Start запускает генерацию сразу и затем каждую ночь в nightHourUTC.
// Блокирует до отмены контекста.
func (g *SlotGenerator) Start(ctx context.Context) {
	log.Println("slot_generator: initial run")
	if err := g.GenerateAll(ctx); err != nil {
		log.Printf("slot_generator: initial run failed: %v", err)
	}

	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, nightHourUTC, 0, 0, 0, time.UTC)
		timer := time.NewTimer(next.Sub(now))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			log.Println("slot_generator: nightly run")
			if err := g.GenerateAll(ctx); err != nil {
				log.Printf("slot_generator: nightly run failed: %v", err)
			}
		}
	}
}

func parseTimeStr(s string) (time.Time, error) {
	for _, layout := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %q", s)
}
