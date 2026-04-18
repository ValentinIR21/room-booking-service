package service

import "avito-talk/internal/testutil"

type mockRoomRepo = testutil.MockRoomRepo
type mockScheduleRepo = testutil.MockScheduleRepo
type mockSlotRepo = testutil.MockSlotRepo
type mockBookingRepo = testutil.MockBookingRepo

var (
	newMockRoomRepo     = testutil.NewMockRoomRepo
	newMockScheduleRepo = testutil.NewMockScheduleRepo
	newMockSlotRepo     = testutil.NewMockSlotRepo
	newMockBookingRepo  = testutil.NewMockBookingRepo
)
