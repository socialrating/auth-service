# Auth Service
Сервис аутентификации, реализованный на Go, с использованием JWT и MongoDB.

# Используемые технологии
-Go
-JWT
-MongoDB
-bcrypt
-Gin

# Задание
Реализованы два REST-маршрута:

1. POST /generate-tokens
2. POST /refresh-tokens

# Требования по безопасности

Access токен не сохраняется в базе данных.
Refresh токен** нельзя использовать повторно.
Все токены связаны по `jti`.
Ошибки возвращаются в явном виде пользователю.

# Принцип работы

1. http://localhost:8080/generate-tokens?user_id=_
