package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
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
	maxCatalogBooks    = 30 // книги каталога в промпте для RAG (15 рекомендованных + 15 популярных)
	maxHistoryMessages = 20 // последние 20 сообщений идут в контекст LLM
	maxTitleLen        = 60 // обрезаем заголовок беседы по первому сообщению
)

var baseSystemPrompt = `Ты — книжный ассистент платформы Oqyrman. Отвечай на том языке, на котором пишет пользователь — на русском или казахском. Если пользователь пишет на казахском — отвечай только на казахском. Если на русском — только на русском.
Помогаешь пользователю с выбором книг, обсуждаешь литературу и отвечаешь на вопросы о чтении.
Будь дружелюбным, лаконичным и полезным.

ПРАВИЛО РЕКОМЕНДАЦИЙ: Рекомендуй ТОЛЬКО книги из каталога платформы Oqyrman, которые указаны в твоём контексте ниже.
Если пользователь просит книгу по теме — сначала используй [ACTION:search_books:запрос] чтобы найти её в каталоге, затем рекомендуй из результатов поиска.
Никогда не рекомендуй книги, которых нет в каталоге платформы. Если подходящей книги нет — честно скажи об этом.

ВАЖНО: Ты отвечаешь ТОЛЬКО на вопросы, связанные с книгами, чтением, литературой и функциями платформы Oqyrman (мероприятия, вишлист, история чтения и т.д.).
Если пользователь задаёт вопрос на любую другую тему — вежливо откажись и напомни, что ты книжный ассистент. Не давай советов по медицине, юриспруденции, финансам, программированию, политике и другим темам, не связанным с книгами и платформой.

Ты можешь выполнять действия на платформе. Для этого вставь тег действия В КОНЦЕ своего ответа (после текста для пользователя):
- [ACTION:search_books:запрос] — поиск книг в каталоге (используй когда пользователь просит рекомендацию по теме или жанру)
- [ACTION:add_wishlist:ID_книги] — добавить книгу в вишлист пользователя
- [ACTION:list_events] — показать ближайшие мероприятия
Одно действие на сообщение.`

var actionRe = regexp.MustCompile(`\[ACTION:(\w+)(?::([^\]]*))?\]`)

// catalogCache holds a pre-built catalog snippet for the system prompt.
// Rebuilt at most once per cacheTTL to avoid re-querying on every message.
type catalogCache struct {
	mu        sync.Mutex
	snippet   string
	builtAt   time.Time
}

const catalogCacheTTL = 5 * time.Minute

type aiUseCase struct {
	sessionRepo     repository.ReadingSessionRepository
	wishlistRepo    repository.WishlistRepository
	bookRepo        repository.BookRepository
	convRepo        repository.ConversationRepository
	reviewRepo      repository.ReviewRepository
	genreRepo       repository.GenreRepository
	authorRepo      repository.AuthorRepository
	eventRepo       repository.EventRepository
	libraryBookRepo repository.LibraryBookRepository
	llm             llm.LLMClient
	catalog         catalogCache
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
	libraryBookRepo repository.LibraryBookRepository,
	llm llm.LLMClient,
) domainUseCase.AIUseCase {
	return &aiUseCase{
		sessionRepo:     sessionRepo,
		wishlistRepo:    wishlistRepo,
		bookRepo:        bookRepo,
		convRepo:        convRepo,
		reviewRepo:      reviewRepo,
		genreRepo:       genreRepo,
		authorRepo:      authorRepo,
		eventRepo:       eventRepo,
		libraryBookRepo: libraryBookRepo,
		llm:             llm,
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
		fmt.Fprintf(&prompt, "Вишлист: %s.\n", strings.Join(wishBooks, ", "))
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

// ── RecommendBooks ────────────────────────────────────────────────────────────

func (u *aiUseCase) RecommendBooks(ctx context.Context, userID uuid.UUID) ([]*entity.BookView, error) {
	// 1. Параллельно получаем историю пользователя
	var (
		sessions []*entity.ReadingSession
		wishlist []*entity.Wishlist
		reviews  []*entity.Review
		wg       sync.WaitGroup
	)
	wg.Add(3)
	go func() { defer wg.Done(); sessions, _ = u.sessionRepo.ListByUser(ctx, userID) }()
	go func() { defer wg.Done(); wishlist, _ = u.wishlistRepo.ListByUser(ctx, userID) }()
	go func() { defer wg.Done(); reviews, _ = u.reviewRepo.ListByUser(ctx, userID) }()
	wg.Wait()

	// Нет истории — нечего рекомендовать
	if len(sessions) == 0 && len(wishlist) == 0 && len(reviews) == 0 {
		return []*entity.BookView{}, nil
	}

	// 2. Загружаем каталог
	allBooks, _, err := u.bookRepo.ListView(ctx, 200, 0)
	if err != nil {
		return nil, fmt.Errorf("RecommendBooks list books: %w", err)
	}
	if len(allBooks) == 0 {
		return []*entity.BookView{}, nil
	}

	// 3. Индексируем каталог и собираем книги пользователя (для исключения)
	catalogByID := make(map[uuid.UUID]*entity.BookView, len(allBooks))
	for _, b := range allBooks {
		catalogByID[b.ID] = b
	}
	userBookIDs := make(map[uuid.UUID]bool)
	for _, s := range sessions {
		userBookIDs[s.BookID] = true
	}
	for _, w := range wishlist {
		userBookIDs[w.BookID] = true
	}
	for _, r := range reviews {
		userBookIDs[r.BookID] = true
	}

	// 4. Определяем топ-жанры пользователя и фильтруем каталог
	genreCount := make(map[uuid.UUID]int)
	for _, s := range sessions {
		if b, ok := catalogByID[s.BookID]; ok {
			genreCount[b.GenreID]++
		}
	}
	for _, w := range wishlist {
		if b, ok := catalogByID[w.BookID]; ok {
			genreCount[b.GenreID]++
		}
	}

	filtered := make([]*entity.BookView, 0, 60)
	for _, b := range allBooks {
		if genreCount[b.GenreID] > 0 {
			filtered = append(filtered, b)
		}
	}
	if len(filtered) < 12 {
		// Недостаточно книг по жанрам — берём первые 50 из полного каталога
		if len(allBooks) > 50 {
			filtered = allBooks[:50]
		} else {
			filtered = allBooks
		}
	}

	// 5. Компактный каталог для промпта (только отфильтрованные книги)
	offlineIDs, _ := u.libraryBookRepo.BookIDsInLibraries(ctx)

	type catalogItem struct {
		ID              string  `json:"id"`
		Title           string  `json:"title"`
		Author          string  `json:"author"`
		Genre           string  `json:"genre"`
		Year            int     `json:"year"`
		Rating          float64 `json:"rating"`
		AvailableOnline bool    `json:"available_online"`
		AvailableOffline bool   `json:"available_offline"`
	}
	items := make([]catalogItem, 0, len(filtered))
	for _, b := range filtered {
		items = append(items, catalogItem{
			ID:               b.ID.String(),
			Title:            b.Title,
			Author:           b.AuthorName,
			Genre:            b.GenreName,
			Year:             b.Year,
			Rating:           b.AvgRating,
			AvailableOnline:  b.BookFileID != nil,
			AvailableOffline: offlineIDs != nil && offlineIDs[b.ID],
		})
	}
	catalogJSON, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("RecommendBooks marshal catalog: %w", err)
	}

	// 6. История пользователя в виде строк
	var historyLines []string
	const historyLimit = 10
	for i, s := range sessions {
		if i >= historyLimit {
			break
		}
		if b, ok := catalogByID[s.BookID]; ok {
			status := "читает"
			if s.Status == entity.StatusFinished {
				status = "прочитал"
			}
			historyLines = append(historyLines, fmt.Sprintf("«%s» — %s, прогресс %d%%", b.Title, status, s.Progress))
		}
	}
	for i, w := range wishlist {
		if i >= historyLimit {
			break
		}
		if b, ok := catalogByID[w.BookID]; ok {
			historyLines = append(historyLines, fmt.Sprintf("«%s» — в вишлисте", b.Title))
		}
	}
	for i, r := range reviews {
		if i >= historyLimit {
			break
		}
		if b, ok := catalogByID[r.BookID]; ok {
			historyLines = append(historyLines, fmt.Sprintf("«%s» — оценка %d/5", b.Title, r.Rating))
		}
	}

	// 7. Формируем промпт
	systemPrompt := "Ты — книжный рекомендательный движок. Отвечай ТОЛЬКО валидным JSON-массивом UUID строк, без пояснений, markdown и лишнего текста."

	var userPrompt strings.Builder
	userPrompt.WriteString("Каталог библиотеки (JSON):\n")
	userPrompt.Write(catalogJSON)
	userPrompt.WriteString("\n\nИстория пользователя:\n")
	if len(historyLines) > 0 {
		userPrompt.WriteString(strings.Join(historyLines, "\n"))
	} else {
		userPrompt.WriteString("(пустая)")
	}
	userPrompt.WriteString("\n\nВыбери ровно 6 книг из каталога для этого пользователя. ")
	userPrompt.WriteString("НЕ включай книги из истории пользователя. ")
	userPrompt.WriteString("Учитывай жанр, автора и рейтинг. ")
	userPrompt.WriteString(`Верни ТОЛЬКО JSON-массив из 6 UUID книг из каталога. Пример: ["uuid1","uuid2","uuid3","uuid4","uuid5","uuid6"]`)

	// 8. Вызываем LLM
	resp, err := u.llm.Complete(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: userPrompt.String()},
	})
	if err != nil {
		return nil, fmt.Errorf("RecommendBooks llm: %w", err)
	}

	// 9. Парсим JSON-массив из ответа (ИИ может обернуть в markdown)
	var ids []string
	start := strings.Index(resp, "[")
	end := strings.LastIndex(resp, "]")
	if start >= 0 && end > start {
		_ = json.Unmarshal([]byte(resp[start:end+1]), &ids)
	}

	// 10. Валидируем ID и собираем результат
	result := make([]*entity.BookView, 0, 6)
	seen := make(map[uuid.UUID]bool)
	for _, idStr := range ids {
		id, parseErr := uuid.Parse(strings.TrimSpace(idStr))
		if parseErr != nil {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		if bv, ok := catalogByID[id]; ok && !userBookIDs[id] {
			result = append(result, bv)
		}
		if len(result) >= 6 {
			break
		}
	}

	// 11. Fallback: если ИИ не вернул валидных ID — берём рекомендованные из БД
	if len(result) == 0 {
		return u.bookRepo.ListRecommendedView(ctx, userID, 6)
	}

	return result, nil
}

// ── Suggested Prompts ────────────────────────────────────────────────────────

func (u *aiUseCase) SuggestedPrompts(ctx context.Context, userID uuid.UUID, lang string) []string {
	isKK := lang == "kk"

	var prompts []string
	if isKK {
		prompts = []string{
			"Кітап ұсыныңыз",
			"Жақын арада қандай іс-шаралар бар?",
		}
	} else {
		prompts = []string{
			"Порекомендуй книгу",
			"Какие мероприятия скоро?",
		}
	}

	sessions, _ := u.sessionRepo.ListByUser(ctx, userID)
	hasReading := false
	for _, s := range sessions {
		if s.Status == entity.StatusReading {
			hasReading = true
			if book, err := u.bookRepo.GetByID(ctx, s.BookID); err == nil {
				if isKK {
					prompts = append(prompts, fmt.Sprintf("«%s» туралы айтып бер", book.Title))
				} else {
					prompts = append(prompts, fmt.Sprintf("Расскажи про «%s»", book.Title))
				}
			}
			break
		}
	}

	if hasReading {
		if isKK {
			prompts = append(prompts, "Алдымен нені оқып бітіру керек?")
		} else {
			prompts = append(prompts, "Что дочитать в первую очередь?")
		}
	}

	wishlist, _ := u.wishlistRepo.ListByUser(ctx, userID)
	if len(wishlist) > 0 {
		if isKK {
			prompts = append(prompts, "Тілектер тізімінен нені оқыған жөн?")
		} else {
			prompts = append(prompts, "Что из вишлиста стоит прочитать?")
		}
	}

	if len(sessions) == 0 && len(wishlist) == 0 {
		if isKK {
			prompts = append(prompts, "Неден бастап оқысам?")
			prompts = append(prompts, "Жаңадан бастаған оқырманға кітаптар ұсыныңыз")
		} else {
			prompts = append(prompts, "С чего начать читать?")
			prompts = append(prompts, "Посоветуй книги для начинающего читателя")
		}
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
	lang := "ru"
	if isKazakh(message) {
		lang = "kk"
	}
	reply, _ = u.processActions(ctx, reply, userID, lang)

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
	msgLang := "ru"
	if isKazakh(message) {
		msgLang = "kk"
	}
	finalContent, results := u.processActions(ctx, fullReply.String(), userID, msgLang)

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

// bookInfo собирает информацию о книге с жанром и автором (параллельно).
func (u *aiUseCase) bookInfo(ctx context.Context, book *entity.Book) string {
	var authorName, genreName string
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(2)
	go func() {
		defer wg.Done()
		if a, err := u.authorRepo.GetByID(ctx, book.AuthorID); err == nil {
			mu.Lock()
			authorName = a.Name
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		if g, err := u.genreRepo.GetByID(ctx, book.GenreID); err == nil {
			mu.Lock()
			genreName = g.Name
			mu.Unlock()
		}
	}()
	wg.Wait()

	info := fmt.Sprintf("«%s»", book.Title)
	if authorName != "" {
		info += fmt.Sprintf(" (%s)", authorName)
	}
	if genreName != "" {
		info += fmt.Sprintf(" [%s]", genreName)
	}
	return info
}

// buildSystemPrompt добавляет в базовый промпт контекст пользователя.
// Все DB-запросы выполняются параллельно для минимальной задержки.
func (u *aiUseCase) buildSystemPrompt(ctx context.Context, userID uuid.UUID) string {
	// 1. Параллельно получаем три списка активности пользователя.
	var (
		sessions []*entity.ReadingSession
		wishlist []*entity.Wishlist
		reviews  []*entity.Review
		wg       sync.WaitGroup
	)
	wg.Add(3)
	go func() { defer wg.Done(); sessions, _ = u.sessionRepo.ListByUser(ctx, userID) }()
	go func() { defer wg.Done(); wishlist, _ = u.wishlistRepo.ListByUser(ctx, userID) }()
	go func() { defer wg.Done(); reviews, _ = u.reviewRepo.ListByUser(ctx, userID) }()
	wg.Wait()

	if len(sessions) > maxBooksInPrompt {
		sessions = sessions[:maxBooksInPrompt]
	}
	if len(wishlist) > maxBooksInPrompt {
		wishlist = wishlist[:maxBooksInPrompt]
	}
	if len(reviews) > maxBooksInPrompt {
		reviews = reviews[:maxBooksInPrompt]
	}

	// 2. Параллельно загружаем книги по всем трём спискам.
	type sessionEntry struct {
		info     string
		status   entity.ReadingStatus
		progress int
	}
	type reviewEntry struct {
		title  string
		rating int
		body   string
	}

	sessionEntries := make([]sessionEntry, len(sessions))
	wishInfos := make([]string, len(wishlist))
	reviewEntries := make([]reviewEntry, len(reviews))

	wg.Add(len(sessions) + len(wishlist) + len(reviews))
	for i, s := range sessions {
		i, s := i, s
		go func() {
			defer wg.Done()
			book, err := u.bookRepo.GetByID(ctx, s.BookID)
			if err != nil {
				return
			}
			sessionEntries[i] = sessionEntry{info: u.bookInfo(ctx, book), status: s.Status, progress: s.Progress}
		}()
	}
	for i, w := range wishlist {
		i, w := i, w
		go func() {
			defer wg.Done()
			book, err := u.bookRepo.GetByID(ctx, w.BookID)
			if err != nil {
				return
			}
			wishInfos[i] = u.bookInfo(ctx, book)
		}()
	}
	for i, r := range reviews {
		i, r := i, r
		go func() {
			defer wg.Done()
			book, err := u.bookRepo.GetByID(ctx, r.BookID)
			if err != nil {
				return
			}
			body := r.Body
			if len([]rune(body)) > 100 {
				body = string([]rune(body)[:100]) + "…"
			}
			reviewEntries[i] = reviewEntry{title: book.Title, rating: r.Rating, body: body}
		}()
	}
	wg.Wait()

	// 3. Собираем промпт.
	var sb strings.Builder
	sb.WriteString(baseSystemPrompt)

	var reading, completed, wishTitles, reviewLines []string
	for _, e := range sessionEntries {
		if e.info == "" {
			continue
		}
		switch e.status {
		case entity.StatusReading:
			reading = append(reading, fmt.Sprintf("%s — прогресс %d%%", e.info, e.progress))
		case entity.StatusFinished:
			completed = append(completed, e.info)
		}
	}
	for _, info := range wishInfos {
		if info != "" {
			wishTitles = append(wishTitles, info)
		}
	}
	for _, e := range reviewEntries {
		if e.title == "" {
			continue
		}
		line := fmt.Sprintf("«%s» — %d/5", e.title, e.rating)
		if e.body != "" {
			line += fmt.Sprintf(": %s", e.body)
		}
		reviewLines = append(reviewLines, line)
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

	u.appendCatalog(ctx, &sb, userID)
	return sb.String()
}

// appendCatalog добавляет в промпт снимок каталога платформы.
// Снимок кешируется в памяти на catalogCacheTTL — при обычном чате каталог
// не меняется часто, поэтому повторные запросы в DB не нужны.
func (u *aiUseCase) appendCatalog(ctx context.Context, sb *strings.Builder, _ uuid.UUID) {
	snippet := u.catalogSnippet(ctx)
	sb.WriteString(snippet)
}

// catalogSnippet возвращает готовую строку каталога из кеша или перестраивает её.
func (u *aiUseCase) catalogSnippet(ctx context.Context) string {
	u.catalog.mu.Lock()
	if u.catalog.snippet != "" && time.Since(u.catalog.builtAt) < catalogCacheTTL {
		snippet := u.catalog.snippet
		u.catalog.mu.Unlock()
		return snippet
	}
	u.catalog.mu.Unlock()

	// Кеш пустой или устарел — строим заново.
	var (
		popular        []*entity.BookView
		offlineBookIDs map[uuid.UUID]bool
		wg             sync.WaitGroup
		mu             sync.Mutex
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		p, _, _ := u.bookRepo.ListPopularView(ctx, maxCatalogBooks, 0)
		mu.Lock()
		popular = p
		mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		ids, _ := u.libraryBookRepo.BookIDsInLibraries(ctx)
		mu.Lock()
		offlineBookIDs = ids
		mu.Unlock()
	}()
	wg.Wait()

	var snap strings.Builder
	snap.WriteString("\n\nКаталог Oqyrman (рекомендуй ТОЛЬКО книги отсюда):")
	snap.WriteString("\nДоступность: [онлайн] = читать в приложении, [оффлайн] = только в библиотеке, [онлайн+оффлайн] = оба варианта.")
	for _, b := range popular {
		fmt.Fprintf(&snap, "\n- «%s» — %s [%s], %d г., %.1f★ %s",
			b.Title, b.AuthorName, b.GenreName, b.Year, b.AvgRating,
			bookAvailability(b, offlineBookIDs))
	}
	snippet := snap.String()

	u.catalog.mu.Lock()
	u.catalog.snippet = snippet
	u.catalog.builtAt = time.Now()
	u.catalog.mu.Unlock()

	return snippet
}

// bookAvailability returns the availability marker for a book.
func bookAvailability(b *entity.BookView, offlineIDs map[uuid.UUID]bool) string {
	online := b.BookFileID != nil
	offline := offlineIDs != nil && offlineIDs[b.ID]
	switch {
	case online && offline:
		return "[онлайн + оффлайн]"
	case online:
		return "[онлайн]"
	case offline:
		return "[оффлайн]"
	default:
		return ""
	}
}

// actionResult хранит результат выполненного действия.
type actionResult struct {
	Action string
	Data   string
}

// isKazakh returns true if the string contains Kazakh-specific letters.
func isKazakh(s string) bool {
	for _, r := range s {
		switch r {
		case 'Ә', 'ә', 'Ғ', 'ғ', 'Қ', 'қ', 'Ң', 'ң', 'Ө', 'ө', 'Ұ', 'ұ', 'Ү', 'ү', 'Һ', 'һ', 'І', 'і':
			return true
		}
	}
	return false
}

// processActions ищет теги [ACTION:...] в ответе AI, выполняет действия и возвращает:
// - очищенный текст (без тегов)
// - результаты действий (для вставки в ответ)
func (u *aiUseCase) processActions(ctx context.Context, reply string, userID uuid.UUID, lang string) (string, []actionResult) {
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
			res = u.actionSearchBooks(ctx, arg, lang)
		case "add_wishlist":
			res = u.actionAddWishlist(ctx, userID, arg, lang)
		case "list_events":
			res = u.actionListEvents(ctx, lang)
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

func (u *aiUseCase) actionSearchBooks(ctx context.Context, query, lang string) actionResult {
	query = strings.TrimSpace(query)
	if query == "" {
		if lang == "kk" {
			return actionResult{Action: "search_books", Data: "🔍 Іздеу сұранысы көрсетілмеген."}
		}
		return actionResult{Action: "search_books", Data: "🔍 Не указан поисковый запрос."}
	}

	books, _, err := u.bookRepo.Search(ctx, query, 5, 0)
	if err != nil || len(books) == 0 {
		if lang == "kk" {
			return actionResult{Action: "search_books", Data: fmt.Sprintf("🔍 «%s» сұранысы бойынша кітаптар табылмады.", query)}
		}
		return actionResult{Action: "search_books", Data: fmt.Sprintf("🔍 По запросу «%s» книг не найдено.", query)}
	}

	var sb strings.Builder
	if lang == "kk" {
		fmt.Fprintf(&sb, "🔍 «%s» іздеу нәтижелері:", query)
	} else {
		fmt.Fprintf(&sb, "🔍 Результаты поиска «%s»:", query)
	}
	for _, b := range books {
		info := u.bookInfo(ctx, b)
		if b.AvgRating > 0 {
			fmt.Fprintf(&sb, "\n  • [%s] %s — рейтинг %.1f", b.ID, info, b.AvgRating)
		} else {
			fmt.Fprintf(&sb, "\n  • [%s] %s", b.ID, info)
		}
	}
	return actionResult{Action: "search_books", Data: sb.String()}
}

func (u *aiUseCase) actionAddWishlist(ctx context.Context, userID uuid.UUID, bookIDStr, lang string) actionResult {
	bookID, err := uuid.Parse(strings.TrimSpace(bookIDStr))
	if err != nil {
		if lang == "kk" {
			return actionResult{Action: "add_wishlist", Data: "Кітапты қосу мүмкін болмады — ID қате."}
		}
		return actionResult{Action: "add_wishlist", Data: "Не удалось добавить — неверный ID книги."}
	}

	exists, _ := u.wishlistRepo.ExistsByUserAndBook(ctx, userID, bookID)
	if exists {
		if lang == "kk" {
			return actionResult{Action: "add_wishlist", Data: "📚 Бұл кітап тілектер тізімінде бар!"}
		}
		return actionResult{Action: "add_wishlist", Data: "📚 Эта книга уже в вашем вишлисте!"}
	}

	_, err = u.wishlistRepo.Add(ctx, userID, bookID, entity.ShelfWantToRead)
	if err != nil {
		if lang == "kk" {
			return actionResult{Action: "add_wishlist", Data: "Кітапты тілектер тізіміне қосу сәтсіз аяқталды."}
		}
		return actionResult{Action: "add_wishlist", Data: "Не удалось добавить книгу в вишлист."}
	}

	book, _ := u.bookRepo.GetByID(ctx, bookID)
	if book != nil {
		if lang == "kk" {
			return actionResult{Action: "add_wishlist", Data: fmt.Sprintf("✅ «%s» кітабы тілектер тізіміне қосылды!", book.Title)}
		}
		return actionResult{Action: "add_wishlist", Data: fmt.Sprintf("✅ Книга «%s» добавлена в ваш вишлист!", book.Title)}
	}
	if lang == "kk" {
		return actionResult{Action: "add_wishlist", Data: "✅ Кітап тілектер тізіміне қосылды!"}
	}
	return actionResult{Action: "add_wishlist", Data: "✅ Книга добавлена в ваш вишлист!"}
}

func (u *aiUseCase) actionListEvents(ctx context.Context, lang string) actionResult {
	events, err := u.eventRepo.FindUpcoming(ctx, 7*24*time.Hour)
	if err != nil || len(events) == 0 {
		if lang == "kk" {
			return actionResult{Action: "list_events", Data: "📅 Жақын арада іс-шаралар жоқ."}
		}
		return actionResult{Action: "list_events", Data: "📅 Ближайших мероприятий не найдено."}
	}

	var sb strings.Builder
	if lang == "kk" {
		sb.WriteString("📅 Жақындағы іс-шаралар:")
	} else {
		sb.WriteString("📅 Ближайшие мероприятия:")
	}
	for _, e := range events {
		title := e.Title
		if lang == "kk" && e.TitleKK != "" {
			title = e.TitleKK
		}
		loc := "место не указано"
		if lang == "kk" {
			loc = "орын көрсетілмеген"
		}
		if e.Location != nil {
			loc = *e.Location
		}
		fmt.Fprintf(&sb, "\n  • %s — %s, %s",
			title,
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

// ── Reader selection actions ─────────────────────────────────────────────────

const (
	maxSelectionChars   = 2000
	maxDescriptionChars = 400
)

var readerSelectionSystemPrompt = `Ты помогаешь читателю понять выделенный фрагмент книги. Отвечай на том языке, на котором написан запрос пользователя — на русском или казахском. Будь кратким и по делу, без вводных фраз о себе и без рекламы платформы. Опирайся на контекст книги, если он дан.
Отвечай ТОЛЬКО по теме конкретного фрагмента или книги. Если запрос не связан с книгой или чтением — откажись.`

func (u *aiUseCase) ExplainSelection(
	ctx context.Context,
	userID, bookID uuid.UUID,
	action, selection, surrounding, targetLang string,
	cb domainUseCase.StreamCallback,
) error {
	trimmed := strings.TrimSpace(selection)
	if utf8.RuneCountInString(trimmed) < 2 {
		return fmt.Errorf("%w: selection must not be empty", entity.ErrValidation)
	}
	if !isReaderAction(action) {
		return fmt.Errorf("%w: unknown action %q", entity.ErrValidation, action)
	}

	// Authorization: каждый, кто использует AI по фрагменту, должен реально
	// читать эту книгу — то есть иметь reading_session. Это отсекает попытки
	// дёргать дорогой LLM-эндпойнт по произвольным book_id.
	if _, err := u.sessionRepo.GetByUserAndBook(ctx, userID, bookID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.ErrForbidden
		}
		return fmt.Errorf("aiUseCase.ExplainSelection check session: %w", err)
	}

	book, err := u.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.ErrNotFound
		}
		return fmt.Errorf("aiUseCase.ExplainSelection get book: %w", err)
	}

	var authorName, genreName string
	if author, err := u.authorRepo.GetByID(ctx, book.AuthorID); err == nil {
		authorName = author.Name
	}
	if genre, err := u.genreRepo.GetByID(ctx, book.GenreID); err == nil {
		genreName = genre.Name
	}

	prompt := buildSelectionPrompt(action, trimmed, surrounding, targetLang, book, authorName, genreName)

	return u.llm.CompleteStream(ctx, readerSelectionSystemPrompt, []llm.Message{
		{Role: "user", Content: prompt},
	}, func(chunk string) error {
		return cb(chunk)
	})
}

func isReaderAction(a string) bool {
	switch a {
	case "ask", "translate":
		return true
	}
	return false
}

const maxSurroundingChars = 1600

func buildSelectionPrompt(action, selection, surrounding, targetLang string, book *entity.Book, author, genre string) string {
	sel := truncate(selection, maxSelectionChars)

	var ctxLines []string
	ctxLines = append(ctxLines, fmt.Sprintf("Книга: «%s»", book.Title))
	if author != "" {
		ctxLines[len(ctxLines)-1] += fmt.Sprintf(" — %s", author)
	}
	if genre != "" {
		ctxLines[len(ctxLines)-1] += fmt.Sprintf(" [%s]", genre)
	}
	if book.Year > 0 {
		ctxLines[len(ctxLines)-1] += fmt.Sprintf(", %d г.", book.Year)
	}
	if book.Language != "" {
		ctxLines = append(ctxLines, fmt.Sprintf("Язык книги: %s", book.Language))
	}
	if desc := strings.TrimSpace(book.Description); desc != "" {
		ctxLines = append(ctxLines, fmt.Sprintf("Описание книги: %s", truncate(desc, maxDescriptionChars)))
	}
	bookCtx := strings.Join(ctxLines, "\n")

	var task string
	switch action {
	case "ask":
		task = "Помоги читателю разобраться с выделенным фрагментом. Сам реши, что уместнее по сути:\n" +
			"— если во фрагменте есть конкретный объект (человек, место, событие, термин, явление, произведение), про который читатель может не знать — дай краткую справку (3–5 предложений), при необходимости перечисли несколько объектов;\n" +
			"— если фрагмент содержит сложную мысль, метафору, отсылку или незнакомую концепцию — объясни её простыми словами (2–3 коротких абзаца), опираясь на контекст книги.\n" +
			"Не разъясняй очевидное, не пересказывай дословно. Без вводных фраз о себе."
	case "translate":
		switch targetLang {
		case "en":
			task = "Translate this passage into English, preserving meaning, tone, and style. Begin your answer with «(Translation → EN):». No extra commentary."
		case "kk":
			task = "Осы үзіндіні қазақ тіліне аудар, мағынасын, тонын және стилін сақта. Жауабыңды «(Аударма → KK):» белгісінен бастаңыз. Қосымша түсініктемесіз."
		default:
			task = "Переведи этот фрагмент на русский язык, сохранив смысл, тон и стиль. Если фрагмент уже на русском — переведи его на английский и начни ответ с пометки «(Перевод RU → EN):». Без лишних комментариев."
		}
	}

	parts := []string{bookCtx}
	if surr := strings.TrimSpace(surrounding); surr != "" {
		parts = append(parts, fmt.Sprintf("Окружающий текст (для контекста):\n«%s»", truncate(surr, maxSurroundingChars)))
	}
	parts = append(parts,
		fmt.Sprintf("Выделенный фрагмент:\n«%s»", sel),
		fmt.Sprintf("Задача: %s", task),
	)
	return strings.Join(parts, "\n\n")
}

// ── Seed conversation from a selection action ────────────────────────────────

func (u *aiUseCase) SeedConversationFromSelection(
	ctx context.Context,
	userID, bookID uuid.UUID,
	action, selection, answer string,
) (*entity.Conversation, error) {
	sel := strings.TrimSpace(selection)
	ans := strings.TrimSpace(answer)
	if utf8.RuneCountInString(sel) < 2 {
		return nil, fmt.Errorf("%w: selection must not be empty", entity.ErrValidation)
	}
	if ans == "" {
		return nil, fmt.Errorf("%w: answer must not be empty", entity.ErrValidation)
	}
	if !isReaderAction(action) {
		return nil, fmt.Errorf("%w: unknown action %q", entity.ErrValidation, action)
	}

	// Та же проверка, что и в ExplainSelection — пользователь должен реально читать книгу.
	if _, err := u.sessionRepo.GetByUserAndBook(ctx, userID, bookID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, entity.ErrForbidden
		}
		return nil, fmt.Errorf("aiUseCase.SeedConversationFromSelection check session: %w", err)
	}

	book, err := u.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("aiUseCase.SeedConversationFromSelection get book: %w", err)
	}

	now := time.Now()
	conv := &entity.Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     truncate(fmt.Sprintf("📖 «%s» — фрагмент", book.Title), maxTitleLen),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := u.convRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("aiUseCase.SeedConversationFromSelection create conv: %w", err)
	}

	actionLabel := "разобраться с"
	if action == "translate" {
		actionLabel = "перевести"
	}
	userText := fmt.Sprintf(
		"Из книги «%s» я выделил фрагмент и попросил тебя %s ним:\n\n«%s»",
		book.Title, actionLabel, truncate(sel, maxSelectionChars),
	)
	userMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		Role:           "user",
		Content:        userText,
		CreatedAt:      now,
	}
	if _, err := u.convRepo.SaveMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("aiUseCase.SeedConversationFromSelection save user msg: %w", err)
	}

	aiMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        ans,
		CreatedAt:      now.Add(time.Millisecond),
	}
	if _, err := u.convRepo.SaveMessage(ctx, aiMsg); err != nil {
		return nil, fmt.Errorf("aiUseCase.SeedConversationFromSelection save ai msg: %w", err)
	}

	_ = u.convRepo.Touch(ctx, conv.ID)
	return conv, nil
}
