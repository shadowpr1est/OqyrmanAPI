# Задачи

## Флоу выдачи книги через QR читательского билета

### Проблема
Сотрудник сканирует QR из «Читательского билета» пользователя, но получает ошибку.
Причина: эндпоинт `/staff/reservations/scan` ищет по `reservation.qr_token`,
а читательский билет содержит `user.qr_code` — это разные поля в разных таблицах.

### Желаемый флоу
1. Пользователь показывает QR из «Читательского билета» (один QR на всё)
2. Сотрудник сканирует его → видит список pending-броней этого пользователя в своей библиотеке
3. Сотрудник выбирает нужную книгу → нажимает «Выдать»
4. Бронь переходит в статус `active`

---

## Что нужно сделать

### 1. OqyrmanAPI — новый эндпоинт

Добавить `GET /staff/reservations/by-user-qr?qr=<user_qr_code>`:
- Middleware проверяет роль Staff и наличие `library_id`
- По `qr` находим пользователя: `SELECT id FROM users WHERE qr_code = $1`
- Возвращаем его `pending`-брони в библиотеке стаффа
- Ответ: такой же формат как у `GET /staff/reservations` (массив `reservationViewResponse`)

Файлы для изменения:
- `internal/delivery/http/handler/reservation/handler.go` — добавить хендлер `GetPendingByUserQR`
- `internal/domain/usecase/reservation_usecase.go` — добавить метод в интерфейс
- `internal/domain/repository/reservation_repository.go` — добавить метод в интерфейс
- `internal/repository/postgres/reservation_repo.go` — реализация SQL
- `internal/delivery/http/router.go` — зарегистрировать маршрут в группе `staff`

### 2. oqyrman-admin — двухшаговый флоу сканирования

Изменить `app/(staff)/staff/reservations/page.tsx`:

**Шаг 1 — сканирование:**
- Сканируем QR → вызываем `GET /staff/reservations/by-user-qr?qr=<token>`
- Если pending-броней нет → показываем сообщение «Нет активных бронирований»
- Если есть → переходим к шагу 2

**Шаг 2 — выбор книги:**
- Показываем модалку со списком pending-броней пользователя
- Каждая строка: обложка книги, название, срок брони
- Кнопка «Выдать» на каждой строке

**Шаг 3 — подтверждение:**
- При нажатии «Выдать» вызываем `PATCH /staff/reservations/{id}/status` с `{ status: 'active' }`
- Закрываем модалку, обновляем список броней