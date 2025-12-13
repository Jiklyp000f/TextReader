package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Структуры для JSON
type Request struct {
	Text string `json:"text"`
}

type WordFrequency struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type Response struct {
	CharCount     int             `json:"charCount"`
	WordCount     int             `json:"wordCount"`
	SentenceCount int             `json:"sentenceCount"`
	FrequentWords []WordFrequency `json:"frequentWords"`
	ReadingTime   string          `json:"readingTime"`
}

// Middleware для CORS
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Обработчик для /api/analyze
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS заголовки
	enableCORS(w)

	// Обрабатываем OPTIONS запрос (preflight CORS)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Проверяем метод
	if r.Method != "POST" {
		http.Error(w, "Только POST-запросы поддерживаются", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type должен быть application/json", http.StatusBadRequest)
		return
	}

	// Декодируем JSON
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON формат", http.StatusBadRequest)
		return
	}

	// Проверяем, что текст не пустой
	if strings.TrimSpace(req.Text) == "" {
		http.Error(w, "Текст не может быть пустым", http.StatusBadRequest)
		return
	}

	// Анализируем текст
	result := analyzeText(req.Text)

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Функция анализа текста
func analyzeText(text string) Response {
	// 1. Подсчет символов
	charCount := len([]rune(text))

	// 2. Подсчет слов
	words := strings.Fields(text)
	wordCount := len(words)

	// 3. Подсчет предложений
	sentenceCount := 0
	sentenceEndings := regexp.MustCompile(`[.!?]+`)
	sentences := sentenceEndings.Split(text, -1)
	for _, s := range sentences {
		if strings.TrimSpace(s) != "" {
			sentenceCount++
		}
	}

	// 4. Частотность слов
	wordFreq := make(map[string]int)
	for _, word := range words {
		// Очищаем слово от знаков препинания
		cleanWord := strings.ToLower(word)
		cleanWord = strings.Trim(cleanWord, ".,!?;:\"'()[]{}")
		if cleanWord != "" {
			wordFreq[cleanWord]++
		}
	}

	// 5. Топ-2 самых частых слов
	frequentWords := make([]WordFrequency, 0)
	for word, count := range wordFreq {
		frequentWords = append(frequentWords, WordFrequency{Word: word, Count: count})
	}

	// Сортируем по убыванию частоты
	sort.Slice(frequentWords, func(i, j int) bool {
		if frequentWords[i].Count == frequentWords[j].Count {
			return frequentWords[i].Word < frequentWords[j].Word
		}
		return frequentWords[i].Count > frequentWords[j].Count
	})

	// Берем только топ-2
	topCount := 2
	if len(frequentWords) < topCount {
		topCount = len(frequentWords)
	}
	topWords := frequentWords[:topCount]

	// 6. Время чтения (200 слов в минуту)
	readingTime := ""
	if wordCount == 0 {
		readingTime = "0 минут"
	} else {
		minutes := float64(wordCount) / 200.0
		if minutes < 1 {
			readingTime = "меньше минуты"
		} else if minutes < 2 {
			readingTime = "1 минута"
		} else if minutes < 5 {
			readingTime = fmt.Sprintf("%.0f минуты", minutes)
		} else {
			readingTime = fmt.Sprintf("%.0f минут", minutes)
		}
	}

	return Response{
		CharCount:     charCount,
		WordCount:     wordCount,
		SentenceCount: sentenceCount,
		FrequentWords: topWords,
		ReadingTime:   readingTime,
	}
}

// Простой обработчик для корня
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, `Анализатор текста API

Эндпоинт: POST /api/analyze

Пример использования через curl:
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"text":"Привет, мир! Это тестовый текст."}'

Или через PowerShell:
Invoke-RestMethod -Uri "http://localhost:8080/api/analyze" -Method Post 
  -ContentType "application/json" 
  -Body '{\"text\":\"Привет, мир! Это тестовый текст.\"}'
`)
}

func main() {
	// Настройка маршрутов
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/analyze", analyzeHandler)

	// Настройка сервера
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запуск сервера
	log.Println("Сервер запущен на http://localhost:8080")
	log.Println("Для тестирования используйте команды ниже:")
	log.Println()
	log.Println("curl:")
	log.Println(`curl -X POST http://localhost:8080/api/analyze \`)
	log.Println(`  -H "Content-Type: application/json" \`)
	log.Println(`  -d '{"text":"Привет, мир! Это тестовый текст."}'`)
	log.Println()
	log.Println("PowerShell:")
	log.Println(`Invoke-RestMethod -Uri "http://localhost:8080/api/analyze" -Method Post \`)
	log.Println(`  -ContentType "application/json" \`)
	log.Println(`  -Body '{\"text\":\"Привет, мир! Это тестовый текст.\"}'`)
	log.Println()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
