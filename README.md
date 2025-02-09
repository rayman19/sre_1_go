Компания разрабатывает микросервис, который взаимодействует с внешним API для получения данных из интергации (курс валют или о погоде). Однако API иногда отвечает с ошибками (например, код 500 или 502), и в таких случаях запрос следует повторять, но не бесконечно.

Твоя задача — реализовать на любом языке функцию `GetData(url string) (string, error)`, которая:
Должна выполнять HTTP-запрос к `url`.
В случае ошибок (код 500, 502, 503, 504) должна **повторять запрос до 3 раз**.
Использовать **экспоненциальную задержку** перед повторной попыткой (`time.Sleep`):
1-я попытка → сразу
2-я попытка → подождать 1 секунду
3-я попытка → подождать 2 секунды
Если после всех попыток API так и не ответило корректно, вернуть ошибку.
Если API отвечает с кодом 200, вернуть содержимое ответа.

В домашнем задании необходимо:
Предоставить исходный код в архиве, скриншоты
опционально можно приложить видео или скриншоты демонстрации как работает функция
