package ai

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm"
)

const (
	maxBooksInPrompt   = 10
	maxCatalogBooks    = 20   // книги каталога в промпте для RAG
	maxHistoryMessages = 20   // последние 20 сообщений идут в контекст LLM
	maxTitleLen        = 60   // обрезаем заголовок беседы по первому сообщению
)

var baseSystemPrompt = `Ты — книжный ассистент платформы Oqyrman. Отвечай на русском языке.
Помогаешь пользователю с выбором книг, обсуждаешь литературу и отвечаешь на вопросы о чтении.
Будь дружелюбным, лаконичным и полезным.
Когда рекомендуешь книги — отдавай предпочтение книгам из каталога платформы (если они подходят).
Если книги нет в каталоге, можешь рекомендовать и внешнюю, но отмечай это.

Ты можешь выполнять действия на платформе. Для этого вставь тег действия В КОНЦЕ своего ответа (после текста для пользователя):
- [ACTION:search_books:запрос] — поиск книг в каталоге
- [ACTION:add_wishlist:ID_книги] — добавить книгу в вишлист пользователя
- [ACTION:list_events] — показать ближайшие мероприятия
Используй действия только когда пользователь явно просит. Одно действие на сообщение.`

var actionRe = regexp.MustCompile(`\[ACTION:(\w+)(?::([^\]]*))?\]`)

type aiUseCase struct {
	sessionRepo  repository.ReadingSessionRepository
	wishlistRepo repository.WishlistRepository
	bookRepo     repository.BookRepository
	convRepo     repository.ConversationRepository
	reviewRepo   repository.ReviewRepository
	genreRepo    repository.GenreRepository
	authorRepo   repository.AuthorRepository
	eventRepo    repository.EventRepository
	llm          llm.LLMClient
}

func NewAIUseCase(
	sessionRepo repository.ReadingSessionRepository,
	wishlistRepo repository.WishlistRepository,
	bookRepo repository.BookRepository,
	convRepo repository.ConversationRepository,
	reviewRepo repository.ReviewRepository,
	genreRepo repository.GenreRepository,
	authorRepo repository.AuthorRepository,
	eventRepo repository.EventRepository,
	llm llm.LLMClient,
) domainUseCase.AIUseCase {
	return &aiUseCase{
		sessionRepo:  sessionRepo,
		wishlistRepo: wishlistRepo,
		bookRepo:     bookRepo,
		convRepo:     convRepo,
		reviewRepo:   reviewRepo,
		genreRepo:    genreRepo,
		authorRepo:   authorRepo,
		eventRepo:    eventRepo,
		llm:          llm,
	}
}

// ── Recommend ─────────────────────────────────────────────────────────────────

func (u *aiUseCase) Recommend(ctx context.Context, userIDStr string) (string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend invalid userID: %w", err)
	}

	sessions, _ := u.sessionRepo.ListByUser(ctx, userID)
	wishlist, _ := u.wishlistRepo.ListByUser(ctx, userID)
	reviews, _ := u.reviewRepo.ListByUser(ctx, userID)

	var readBooks, wishBooks, reviewLines []string
	for i, s := range sessions {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, s.BookID); err == nil {
			readBooks = append(readBooks, u.bookInfo(ctx, book))
		}
	}
	for i, w := range wishlist {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, w.BookID); err == nil {
			wishBooks = append(wishBooks, u.bookInfo(ctx, book))
		}
	}
	for i, r := range reviews {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, r.BookID); err == nil {
			reviewLines = append(reviewLines, fmt.Sprintf("«%s» — %d/5", book.Title, r.Rating))
		}
	}

	var prompt strings.Builder
	prompt.WriteString("Дай 5 персональных рекомендаций книг для чтения.\n\n")
	if len(readBooks) > 0 {
		fmt.Fprintf(&prompt, "История чтения: %s.\n", strings.Join(readBooks, ", "))
	} else {
		prompt.WriteString("История чтения пуста.\n")
	}
	if len(wishBooks) > 0 {
		fmt.Fprintf(&prompt, "Виш��ист: %s.\n", strings.Join(wishBooks, ", "))
	} else {
		prompt.WriteString("Вишлист пуст.\n")
	}
	if len(reviewLines) > 0 {
		fmt.Fprintf(&prompt, "Оценки пользователя: %s.\n", strings.Join(reviewLines, "; "))
	}
	prompt.WriteString("\nПорекомендуй 5 книг, объясни кратко почему каждая.")

	return u.llm.Complete(ctx, baseSystemPrompt, []llm.Message{
		{Role: "user", Content: prompt.String()},
	})
}

// ── Suggested Prompts ────────────────────────────────────────────────────────

func (u *aiUseCase) SuggestedPrompts(ctx context.Context, userID uuid.UUID) []string {
	prompts := []string{
		"Порекомендуй книгу",
		"Какие мероп��иятия скоро?",
	}

	sessions, _ := u.sessionRepo.ListByUser(ctx, userID)
	hasReading := false
	for _, s := range sessions {
		if s.Status == entity.StatusReading {
			hasReading = true
			if book, err := u.bookRepo.GetByID(ctx, s.BookID); err == nil {
				prompts = append(prompts, fmt.Sprintf("Расскажи про «%s»", book.Title))
			}
			break
		}
	}

	if hasReading {
		prompts = append(prompts, "Что дочитать в первую очередь?")
	}

	wishlist, _ := u.wishlistRepo.ListByUser(ctx, userID)
	if len(wishlist) > 0 {
		prompts = append(prompts, "Что из вишлиста стоит прочитать?")
	}

	if len(sessions) == 0 && len(wishlist) == 0 {
		prompts = append(prompts, "С чего начать читать?")
		prompts = append(prompts, "Посоветуй книги для начинающего читателя")
	}

	return prompts
}

// ── Conversations ─────────────────────────────────────────────────────────────

func (u *aiUseCase) CreateConversation(ctx context.Context, userID uuid.UUID) (*entity.Conversation, error) {
	now := time.Now()
	conv := &entity.Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     "Новый чат",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return u.convRepo.Create(ctx, conv)
}

func (u *aiUseCase) SendMessage(ctx context.Context, convID, userID uuid.UUID, message string) (*entity.ChatMessage, *entity.ChatMessage, error) {
	if len([]rune(strings.TrimSpace(message))) == 0 {
		return nil, nil, fmt.Errorf("%w: message must not be empty", entity.ErrValidation)
	}

	// Проверяем владение беседой
	conv, err := u.convRepo.GetByID(ctx, convID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, entity.ErrConversationNotFound
		}
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage get conv: %w", err)
	}
	if conv.UserID != userID {
		return nil, nil, entity.ErrForbidden
	}

	// Сохраняем сообщение пользователя
	userMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "user",
		Content:        message,
		CreatedAt:      time.Now(),
	}
	if _, err := u.convRepo.SaveMessage(ctx, userMsg); err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage save user msg: %w", err)
	}

	// Если первое сообщение — ставим заголовок из текста
	if conv.Title == "Новый чат" {
		title := truncate(message, maxTitleLen)
		_ = u.convRepo.UpdateTitle(ctx, convID, title)
	}

	// Загружаем историю (последние N сообщений, уже включая только что сохранённое)
	history, err := u.convRepo.ListMessages(ctx, convID, maxHistoryMessages)
	if err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage load history: %w", err)
	}

	// Формируем системный промпт с контекстом пользователя
	systemPrompt := u.buildSystemPrompt(ctx, userID)

	// Переводим историю в формат LLM
	msgs := make([]llm.Message, len(history))
	for i, h := range history {
		msgs[i] = llm.Message{Role: h.Role, Content: h.Content}
	}

	// Вызываем LLM
	reply, err := u.llm.Complete(ctx, systemPrompt, msgs)
	if err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage llm: %w", err)
	}

	// Обрабатываем действия из ответа
	reply, _ = u.processActions(ctx, reply, userID)

	// Сохраняем ответ ИИ
	aiMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "assistant",
		Content:        reply,
		CreatedAt:      time.Now(),
	}
	if _, err := u.convRepo.SaveMessage(ctx, aiMsg); err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage save ai msg: %w", err)
	}

	_ = u.convRepo.Touch(ctx, convID)

	return userMsg, aiMsg, nil
}

func (u *aiUseCase) SendMessageStream(ctx context.Context, convID, userID uuid.UUID, message string, cb domainUseCase.StreamCallback) (*entity.ChatMessage, *entity.ChatMessage, error) {
	if len([]rune(strings.TrimSpace(message))) == 0 {
		return nil, nil, fmt.Errorf("%w: message must not be empty", entity.ErrValidation)
	}

	conv, err := u.convRepo.GetByID(ctx, convID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, entity.ErrConversationNotFound
		}
		return nil, nil, fmt.Errorf("aiUseCase.SendMessageStream get conv: %w", err)
	}
	if conv.UserID != userID {
		return nil, nil, entity.ErrForbidden
	}

	userMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "user",
		Content:        message,
		CreatedAt:      time.Now(),
	}
	if _, err := u.convRepo.SaveMessage(ctx, userMsg); err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessageStream save user msg: %w", err)
	}

	if conv.Title == "Новый чат" {
		title := truncate(message, maxTitleLen)
		_ = u.convRepo.UpdateTitle(ctx, convID, title)
	}

	history, err := u.convRepo.ListMessages(ctx, convID, maxHistoryMessages)
	if err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessageStream load history: %w", err)
	}

	systemPrompt := u.buildSystemPrompt(ctx, userID)
	msgs := make([]llm.Message, len(history))
	for i, h := range history {
		msgs[i] = llm.Message{Role: h.Role, Content: h.Content}
	}

	// Стримим ответ, собирая полный текст для сохранения
	var fullReply strings.Builder
	if err := u.llm.CompleteStream(ctx, systemPrompt, msgs, func(chunk string) error {
		fullReply.WriteString(chunk)
		return cb(chunk)
	}); err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessageStream llm: %w", err)
	}

	// Обрабатываем действия из полного ответа
	finalContent, results := u.processActions(ctx, fullReply.String(), userID)

	// Если были действия — отправляем результат как дополнительный chunk
	if len(results) > 0 {
		// Отправляем результат действия клиенту
		actionText := "\n"
		for _, r := range results {
			actionText += "\n" + r.Data
		}
		_ = cb(actionText)
	}

	aiMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "assistant",
		Content:        finalContent,
		CreatedAt:      time.Now(),
	}
	if _, err := u.convRepo.SaveMessage(ctx, aiMsg); err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessageStream save ai msg: %w", err)
	}

	_ = u.convRepo.Touch(ctx, convID)

	return userMsg, aiMsg, nil
}

func (u *aiUseCase) ListConversations(ctx context.Context, userID uuid.UUID) ([]*entity.Conversation, error) {
	return u.convRepo.ListByUser(ctx, userID)
}

func (u *aiUseCase) GetConversation(ctx context.Context, id, userID uuid.UUID) (*entity.Conversation, []*entity.ChatMessage, error) {
	conv, err := u.convRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, entity.ErrConversationNotFound
		}
		return nil, nil, fmt.Errorf("aiUseCase.GetConversation: %w", err)
	}
	if conv.UserID != userID {
		return nil, nil, entity.ErrForbidden
	}
	messages, err := u.convRepo.ListMessages(ctx, id, 200)
	if err != nil {
		return nil, nil, err
	}
	return conv, messages, nil
}

func (u *aiUseCase) DeleteConversation(ctx context.Context, id, userID uuid.UUID) error {
	conv, err := u.convRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.ErrConversationNotFound
		}
		return fmt.Errorf("aiUseCase.DeleteConversation: %w", err)
	}
	if conv.UserID != userID {
		return entity.ErrForbidden
	}
	return u.convRepo.Delete(ctx, id)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// bookInfo собирает информацию о книге с жанром и автором.
func (u *aiUseCase) bookInfo(ctx context.Context, book *entity.Book) string {
	info := fmt.Sprintf("«%s»", book.Title)
	if author, err := u.authorRepo.GetByID(ctx, book.AuthorID); err == nil {
		info += fmt.Sprintf(" (%s)", author.Name)
	}
	if genre, err := u.genreRepo.GetByID(ctx, book.GenreID); err == nil {
		info += fmt.Sprintf(" [%s]", genre.Name)
	}
	return info
}

// buildSystemPrompt добавляет в базовый промпт контекст пользователя.
func (u *aiUseCase) buildSystemPrompt(ctx context.Context, userID uuid.UUID) string {
	var sb strings.Builder
	sb.WriteString(baseSystemPrompt)

	sessions, _ := u.sessionRepo.ListByUser(ctx, userID)
	wishlist, _ := u.wishlistRepo.ListByUser(ctx, userID)
	reviews, _ := u.reviewRepo.ListByUser(ctx, userID)

	var reading, completed, wishTitles, reviewLines []string

	for i, s := range sessions {
		if i >= maxBooksInPrompt {
			break
		}
		book, err := u.bookRepo.GetByID(ctx, s.BookID)
		if err != nil {
			continue
		}
		info := u.bookInfo(ctx, book)
		switch s.Status {
		case entity.StatusReading:
			reading = append(reading, fmt.Sprintf("%s — прогресс %d%%", info, s.Progress))
		case entity.StatusFinished:
			completed = append(completed, info)
		}
	}

	for i, w := range wishlist {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, w.BookID); err == nil {
			wishTitles = append(wishTitles, u.bookInfo(ctx, book))
		}
	}

	for i, r := range reviews {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, r.BookID); err == nil {
			line := fmt.Sprintf("«%s» — %d/5", book.Title, r.Rating)
			if r.Body != "" {
				body := r.Body
				if len([]rune(body)) > 100 {
					body = string([]rune(body)[:100]) + "…"
				}
				line += fmt.Sprintf(": %s", body)
			}
			reviewLines = append(reviewLines, line)
		}
	}

	if len(reading) > 0 || len(completed) > 0 || len(wishTitles) > 0 || len(reviewLines) > 0 {
		sb.WriteString("\n\nКонтекст пользователя:")
		if len(reading) > 0 {
			fmt.Fprintf(&sb, "\n- Читает сейчас: %s", strings.Join(reading, "; "))
		}
		if len(completed) > 0 {
			fmt.Fprintf(&sb, "\n- Прочитал: %s", strings.Join(completed, ", "))
		}
		if len(wishTitles) > 0 {
			fmt.Fprintf(&sb, "\n- В вишлисте: %s", strings.Join(wishTitles, ", "))
		}
		if len(reviewLines) > 0 {
			fmt.Fprintf(&sb, "\n- Оценки: %s", strings.Join(reviewLines, "; "))
		}
	}

	// RAG: добавляем каталог книг платформы
	u.appendCatalog(ctx, &sb, userID)

	return sb.String()
}

// appendCatalog добавляет в промпт популярные и рекомендованные книги с платформы.
func (u *aiUseCase) appendCatalog(ctx context.Context, sb *strings.Builder, userID uuid.UUID) {
	sb.WriteString("\n\nКаталог книг на платформе Oqyrman (рекомендуй в первую очередь из этого списка):")

	// Персональные рекомендации
	recommended, _ := u.bookRepo.ListRecommendedView(ctx, userID, maxCatalogBooks/2)
	if len(recommended) > 0 {
		sb.WriteString("\n\nРекомендованные для пользователя:")
		for _, b := range recommended {
			fmt.Fprintf(sb, "\n- «%s» — %s [%s], %d г., рейтинг %.1f",
				b.Title, b.AuthorName, b.GenreName, b.Year, b.AvgRating)
			if b.Description != "" {
				desc := b.Description
				if len([]rune(desc)) > 120 {
					desc = string([]rune(desc)[:120]) + "…"
				}
				fmt.Fprintf(sb, " — %s", desc)
			}
		}
	}

	// Популярные книги
	popular, _, _ := u.bookRepo.ListPopularView(ctx, maxCatalogBooks/2, 0)
	if len(popular) > 0 {
		sb.WriteString("\n\nПопулярные на платформе:")
		for _, b := range popular {
			fmt.Fprintf(sb, "\n- «%s» — %s [%s], %d г., рейтинг %.1f",
				b.Title, b.AuthorName, b.GenreName, b.Year, b.AvgRating)
		}
	}
}

// actionResult хранит результат выполненного действия.
type actionResult struct {
	Action string
	Data   string
}

// processActions ищет теги [ACTION:...] в ответе AI, выполняет действия и возвращает:
// - очищенный текст (без тегов)
// - результаты действий (для вставки в ответ)
func (u *aiUseCase) processActions(ctx context.Context, reply string, userID uuid.UUID) (string, []actionResult) {
	matches := actionRe.FindAllStringSubmatch(reply, 1) // макс 1 действие
	if len(matches) == 0 {
		return reply, nil
	}

	var results []actionResult
	for _, m := range matches {
		action := m[1]
		arg := ""
		if len(m) > 2 {
			arg = m[2]
		}

		var res actionResult
		switch action {
		case "search_books":
			res = u.actionSearchBooks(ctx, arg)
		case "add_wishlist":
			res = u.actionAddWishlist(ctx, userID, arg)
		case "list_events":
			res = u.actionListEvents(ctx)
		default:
			continue
		}
		results = append(results, res)
	}

	// Убираем теги из ответа
	cleaned := actionRe.ReplaceAllString(reply, "")
	cleaned = strings.TrimSpace(cleaned)

	// Добавляем результаты действий к ответу
	if len(results) > 0 {
		for _, r := range results {
			cleaned += "\n\n" + r.Data
		}
	}

	return cleaned, results
}

func (u *aiUseCase) actionSearchBooks(ctx context.Context, query string) actionResult {
	query = strings.TrimSpace(query)
	if query == "" {
		return actionResult{Action: "search_books", Data: "🔍 Не указан поисковый запрос."}
	}

	books, _, err := u.bookRepo.Search(ctx, query, 5, 0)
	if err != nil || len(books) == 0 {
		return actionResult{Action: "search_books", Data: fmt.Sprintf("🔍 По запросу «%s» книг не найдено.", query)}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "🔍 Результаты поиска «%s»:", query)
	for _, b := range books {
		info := u.bookInfo(ctx, b)
		if b.AvgRating > 0 {
			fmt.Fprintf(&sb, "\n  • %s — рейтинг %.1f", info, b.AvgRating)
		} else {
			fmt.Fprintf(&sb, "\n  • %s", info)
		}
	}
	return actionResult{Action: "search_books", Data: sb.String()}
}

func (u *aiUseCase) actionAddWishlist(ctx context.Context, userID uuid.UUID, bookIDStr string) actionResult {
	bookID, err := uuid.Parse(strings.TrimSpace(bookIDStr))
	if err != nil {
		return actionResult{Action: "add_wishlist", Data: "Не удалось добавить — неверный ID книги."}
	}

	exists, _ := u.wishlistRepo.ExistsByUserAndBook(ctx, userID, bookID)
	if exists {
		return actionResult{Action: "add_wishlist", Data: "📚 Эта книга уже в вашем вишлисте!"}
	}

	_, err = u.wishlistRepo.Add(ctx, userID, bookID, entity.ShelfWantToRead)
	if err != nil {
		return actionResult{Action: "add_wishlist", Data: "Не удалось добавить книгу в вишлист."}
	}

	book, _ := u.bookRepo.GetByID(ctx, bookID)
	if book != nil {
		return actionResult{Action: "add_wishlist", Data: fmt.Sprintf("✅ Книга «%s» добавлена в ваш вишлист!", book.Title)}
	}
	return actionResult{Action: "add_wishlist", Data: "✅ Книга добавлена в ваш вишлист!"}
}

func (u *aiUseCase) actionListEvents(ctx context.Context) actionResult {
	events, err := u.eventRepo.FindUpcoming(ctx, 7*24*time.Hour) // ближайшие 7 дней
	if err != nil || len(events) == 0 {
		return actionResult{Action: "list_events", Data: "📅 Ближайших мероприятий не найдено."}
	}

	var sb strings.Builder
	sb.WriteString("📅 Ближайшие мероприятия:")
	for _, e := range events {
		loc := "место не указано"
		if e.Location != nil {
			loc = *e.Location
		}
		fmt.Fprintf(&sb, "\n  • %s — %s, %s",
			e.Title,
			e.StartsAt.Format("02.01.2006 15:04"),
			loc)
	}
	return actionResult{Action: "list_events", Data: sb.String()}
}

func truncate(s string, maxChars int) string {
	if utf8.RuneCountInString(s) <= maxChars {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxChars]) + "…"
}
