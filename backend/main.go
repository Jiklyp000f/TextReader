package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// TextAnalysis представляет результат анализа текста
type TextAnalysis struct {
	CharCount     int
	WordCount     int
	SentenceCount int
	FrequentWords []WordFrequency
	ReadingTime   string
}

// WordFrequency представляет слово и его частоту
type WordFrequency struct {
	Word  string
	Count int
}

// ==================== Use Case Layer ====================

// TextAnalyzer - интерфейс для анализа текста
type TextAnalyzer interface {
	Analyze(text string, delimiter string) TextAnalysis
}

// DefaultAnalyzer - реализация анализатора текста
type DefaultAnalyzer struct{}

// Analyze выполняет анализ текста
func (a *DefaultAnalyzer) Analyze(text string, delimiter string) TextAnalysis {
	charCount := countCharacters(text)
	words := extractWords(text)
	wordCount := len(words)
	sentenceCount := countSentences(text, delimiter)
	frequentWords := getFrequentWords(words, 2)
	readingTime := calculateReadingTimeSimple(wordCount, charCount)

	return TextAnalysis{
		CharCount:     charCount,
		WordCount:     wordCount,
		SentenceCount: sentenceCount,
		FrequentWords: frequentWords,
		ReadingTime:   readingTime,
	}
}

// ==================== Infrastructure Layer ====================

// HTTPHandler обрабатывает HTTP запросы
type HTTPHandler struct {
	analyzer TextAnalyzer
}

// NewHTTPHandler создает новый HTTP обработчик
func NewHTTPHandler(analyzer TextAnalyzer) *HTTPHandler {
	return &HTTPHandler{analyzer: analyzer}
}

// AnalyzeTextHandler обработчик для анализа текста
func (h *HTTPHandler) AnalyzeTextHandler(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		sendJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var request struct {
		Text      string `json:"text"`
		Delimiter string `json:"delimiter"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONError(w, http.StatusBadRequest, "Неверный JSON")
		return
	}

	// Используем use case
	analysis := h.analyzer.Analyze(request.Text, request.Delimiter)

	// Преобразуем в DTO для HTTP
	response := map[string]interface{}{
		"charCount":     analysis.CharCount,
		"wordCount":     analysis.WordCount,
		"sentenceCount": analysis.SentenceCount,
		"readingTime":   analysis.ReadingTime,
		"frequentWords": convertToMap(analysis.FrequentWords),
	}

	sendJSON(w, http.StatusOK, response)
}

// ==================== Business Logic (Pure Functions) ====================

// countCharacters подсчитывает количество символов
func countCharacters(text string) int {
	return len([]rune(text))
}

// extractWords извлекает слова из текста
func extractWords(text string) []string {
	return strings.Fields(text)
}

// countSentences подсчитывает количество предложений
// Если delimiter пустой, используется стандартная логика [.!?]+
// Если delimiter указан, используется указанный символ(ы) для разделения
func countSentences(text string, delimiter string) int {
	if delimiter == "" {
		// Стандартная логика: используем . ! ?
		sentenceRegex := regexp.MustCompile(`[.!?]+`)
		sentences := sentenceRegex.Split(text, -1)
		count := 0
		for _, s := range sentences {
			if strings.TrimSpace(s) != "" {
				count++
			}
		}
		return count
	}
	
	// Пользовательский разделитель: просто разбиваем по указанному символу
	parts := strings.Split(text, delimiter)
	count := 0
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			count++
		}
	}
	return count
}

// getFrequentWords возвращает самые частые слова
func getFrequentWords(words []string, topN int) []WordFrequency {
	// Подсчет частоты
	freqMap := make(map[string]int)
	for _, word := range words {
		cleanWord := cleanWord(word)
		if cleanWord != "" {
			freqMap[cleanWord]++
		}
	}

	// Преобразование в слайс для сортировки
	var frequencies []WordFrequency
	for word, count := range freqMap {
		frequencies = append(frequencies, WordFrequency{Word: word, Count: count})
	}

	// Сортировка по убыванию частоты, при равенстве - по алфавиту
	sort.Slice(frequencies, func(i, j int) bool {
		if frequencies[i].Count == frequencies[j].Count {
			return frequencies[i].Word < frequencies[j].Word
		}
		return frequencies[i].Count > frequencies[j].Count
	})

	// Возвращаем топ N
	if topN > len(frequencies) {
		topN = len(frequencies)
	}

	return frequencies[:topN]
}

// cleanWord очищает слово от знаков препинания
func cleanWord(word string) string {
	clean := strings.ToLower(word)
	trimChars := ".,!?;:\"'()[]{}"
	return strings.Trim(clean, trimChars)
}

// calculateReadingTimeSimple простой, но улучшенный расчёт
func calculateReadingTimeSimple(wordCount, charCount int) string {
	if wordCount == 0 {
		return "0 минут"
	}

	// Рассчитываем среднюю длину слова
	averageWordLength := float64(charCount) / float64(wordCount)

	// Базовая скорость чтения
	baseSpeed := 200.0 // слов в минуту

	// Корректируем скорость в зависимости от средней длины слова
	// Формула: чем длиннее слова, тем медленнее читаем
	// Эмпирическая формула: speed = 200 * (5 / averageWordLength)
	// Где 5 - средняя длина слова в русском языке
	if averageWordLength > 0 {
		adjustedSpeed := baseSpeed * (5.0 / averageWordLength)
		// Ограничиваем разумными пределами
		if adjustedSpeed < 100 {
			adjustedSpeed = 100
		}
		if adjustedSpeed > 300 {
			adjustedSpeed = 300
		}
		baseSpeed = adjustedSpeed
	}

	minutes := float64(wordCount) / baseSpeed

	// Форматирование результата
	if minutes < 1 {
		return "меньше минуты"
	}

	// Правильное склонение минут для русского языка
	lastDigit := int(minutes) % 10
	lastTwoDigits := int(minutes) % 100

	if lastTwoDigits >= 11 && lastTwoDigits <= 19 {
		return fmt.Sprintf("%.0f минут", minutes)
	}

	switch lastDigit {
	case 1:
		return fmt.Sprintf("%.0f минута", minutes)
	case 2, 3, 4:
		return fmt.Sprintf("%.0f минуты", minutes)
	default:
		return fmt.Sprintf("%.0f минут", minutes)
	}
}

// ==================== Utility Functions ====================

// convertToMap преобразует WordFrequency в []map[string]int для JSON
func convertToMap(words []WordFrequency) []map[string]int {
	result := make([]map[string]int, len(words))
	for i, wf := range words {
		result[i] = map[string]int{wf.Word: wf.Count}
	}
	return result
}

// sendJSON отправляет JSON ответ
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendJSONError отправляет ошибку в формате JSON
func sendJSONError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, map[string]string{"error": message})
}

// ==================== Main Application ====================

func main() {
	// Инициализация зависимостей
	analyzer := &DefaultAnalyzer{}
	handler := NewHTTPHandler(analyzer)

	// Настройка маршрутов
	mux := http.NewServeMux()
	mux.HandleFunc("/api/analyze", handler.AnalyzeTextHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Настройка сервера
	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	log.Println("🚀 Сервер запущен на порту 8082")
	log.Println("📌 Пример запроса:")
	log.Println(`curl -X POST http://localhost:8082/api/analyze \`)
	log.Println(`  -H "Content-Type: application/json" \`)
	log.Println(`  -d '{"text":"Привет мир!"}'`)

	log.Fatal(server.ListenAndServe())
}
