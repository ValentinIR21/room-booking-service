<h1>Room Booking Service (Тестовое задание Avito)</h1>

Сервис бронирования переговорок с JWT-авторизацией


<h2>Запуск:</h2>

- ```docker-compose up --build```

Сервис станет доступен по адресу: ```http://localhost:8080```

<h2>Эндпоинты:</h2>

- GET /_info - Health check (200)

- POST /dummyLogin (admin) - получить JWT токен. JSON: ```{"role":"admin"}```

Токен передаётся в заголовке: Authorization: Bearer <token>


- POST /dummyLogin (user) - получить JWT токен. JSON: ```{"role":"user"}```

- GET /rooms/list (admin/user) - Список переговорок

- POST /rooms/create (admin) - Создать переговорку. JSON: 

```{ "name":"Тестовый-зал2", "description":"тестовая вместительность", "capacity":20 }```


- POST /rooms/{roomId}/schedule/create (admin). Создать расписание (один раз) JSON:

```{ "daysOfWeek":[1,2,3,4,5,6,7], "startTime":"09:00:00", "endTime":"17:00:00" }```


- GET /rooms/{roomId}/slots/list?date=2025-04-10 (admin/user). Свободные слоты на дату + 7 дней

- POST /bookings/create (user). Создать бронь JSON: ```{"slotId":"roomId"}```

- POST /bookings/{bookingId}/cancel Отменить бронь (идемпотентно).

- GET /bookings/my (user). Список всех будущих бронирований.

- GET /bookings/list?page=1&pageSize=10 (admin). Получение всех броней с пагинацией