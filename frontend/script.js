document.addEventListener('DOMContentLoaded', function() {
    const analyzeBtn = document.getElementById('analyze-btn');
    const textInput = document.getElementById('text-input');
    const resultsDiv = document.getElementById('results');
    const loader = document.getElementById('loader');
    const errorMessageDiv = document.getElementById('error-message');
    const useCustomDelimiterCheckbox = document.getElementById('use-custom-delimiter');
    const delimiterWrapper = document.getElementById('delimiter-wrapper');
    const delimiterInput = document.getElementById('delimiter-input');
    const tooltipTrigger = document.querySelector('.tooltip-trigger');
    const tooltip = document.getElementById('delimiter-tooltip');
    
    // Устанавливаем таймаут для запроса (3 секунды)
    const REQUEST_TIMEOUT = 10000; // 3 секунды
    
    // Пример текста по умолчанию
    textInput.value = "Привет! Это пример текста для анализа.";
    
    // Обработчик чекбокса для показа/скрытия поля ввода разделителя
    useCustomDelimiterCheckbox.addEventListener('change', function() {
        if (this.checked) {
            delimiterWrapper.style.display = 'block';
        } else {
            delimiterWrapper.style.display = 'none';
            delimiterInput.value = '';
        }
    });
    
    // Обработчик для показа tooltip при наведении
    if (tooltipTrigger && tooltip) {
        tooltipTrigger.addEventListener('mouseenter', function() {
            tooltip.style.display = 'block';
        });
        
        tooltipTrigger.addEventListener('mouseleave', function() {
            tooltip.style.display = 'none';
        });
    }
    
    analyzeBtn.addEventListener('click', async function() {
        const text = textInput.value.trim();
        
        if (!text) {
            alert('Пожалуйста, введите текст для анализа');
            return;
        }
        
        // Получаем разделитель: если чекбокс отмечен, берем значение из поля, иначе пустую строку
        const delimiter = useCustomDelimiterCheckbox.checked ? delimiterInput.value.trim() : '';
        
        // Скрываем предыдущие результаты и ошибки
        resultsDiv.style.display = 'none';
        errorMessageDiv.style.display = 'none';
        
        // Показываем загрузку
        loader.style.display = 'block';
        
        // Создаем AbortController для возможности отмены запроса
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT);
        
        try {
            // Отправляем запрос на бэкенд с таймаутом
            const response = await fetch('http://localhost:8082/api/analyze', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    text: text,
                    delimiter: delimiter
                }),
                signal: controller.signal // Добавляем возможность прерывания
            });
            
            clearTimeout(timeoutId); // Очищаем таймаут, если запрос успешен
            
            if (!response.ok) {
                throw new Error(`Ошибка сервера: ${response.status}`);
            }
            
            const data = await response.json();
            
            // Обновляем интерфейс с полученными данными
            document.getElementById('char-count').textContent = data.charCount || 0;
            document.getElementById('word-count').textContent = data.wordCount || 0;
            document.getElementById('sentence-count').textContent = data.sentenceCount || 0;
            document.getElementById('message').textContent = data.message || 'Анализ завершен';
            document.getElementById('reading-time').textContent = data.readingTime || '0 минут';
            
            // Очищаем и заполняем список частоты слов
            const wordFrequencyDiv = document.getElementById('word-frequency');
            wordFrequencyDiv.innerHTML = '';

            // Преобразуем массив объектов в плоский объект для фронтенда
            let wordFrequencyObj = {};
            if (Array.isArray(data.frequentWords)) {
                data.frequentWords.forEach(item => {
                    const word = Object.keys(item)[0];
                    const count = item[word];
                    wordFrequencyObj[word] = count;
                });
            } else if (typeof data.frequentWords === 'object') {
                wordFrequencyObj = data.frequentWords;
            }

            // Преобразуем объект в массив и сортируем по убыванию частоты
            const wordFrequencyArray = Object.entries(wordFrequencyObj)
                .sort((a, b) => b[1] - a[1]);

            if (wordFrequencyArray.length > 0) {
                wordFrequencyArray.forEach(([word, count]) => {
                    const wordItem = document.createElement('div');
                    wordItem.className = 'word-item';
                    wordItem.innerHTML = `
                        <span class="word">${word}</span>
                        <span class="count">${count}</span>
                    `;
                    wordFrequencyDiv.appendChild(wordItem);
                });
            } else {
                wordFrequencyDiv.innerHTML = '<p style="text-align: center; color: #666;">Нет данных о частоте слов</p>';
            }
            
            // Показываем результаты, скрываем загрузку
            loader.style.display = 'none';
            resultsDiv.style.display = 'block';
            
        } catch (error) {
            console.error('Ошибка:', error);
            
            // Очищаем таймаут в случае ошибки
            clearTimeout(timeoutId);
            
            // Скрываем загрузку
            loader.style.display = 'none';
            
            // Показываем сообщение об ошибке
            errorMessageDiv.style.display = 'block';
            
            // Меняем текст ошибки в зависимости от типа
            const errorTitle = errorMessageDiv.querySelector('.error-title');
            const errorText = errorMessageDiv.querySelector('p');
            
            if (error.name === 'AbortError') {
                errorTitle.textContent = '⚠️ Превышено время ожидания';
                errorText.innerHTML = `Сервер не ответил за ${REQUEST_TIMEOUT/1000} секунд. Пожалуйста:<br>
                    1. Проверьте, запущен ли сервер на localhost:8082<br>
                    2. Убедитесь, что сервер обрабатывает запросы<br>
                    3. Попробуйте позже`;
            } else {
                errorTitle.textContent = '⚠️ Ошибка соединения';
                errorText.innerHTML = `Не удалось подключиться к серверу:<br>${error.message}`;
            }
        }
    });
});