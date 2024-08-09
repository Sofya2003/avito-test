# Тестовое задание

Чтобы запустить сервис, в корне проекта надо выполнить ```docker compose up -d.```
Сервис запускается на порту 8080.

## Примеры запросов
Для отправки запросов использовался Postman.
Запросы отправлялись на http://127.0.0.1:8080/

### Регистрация
Запрос:
```
{
    "email": "test@mail.ru",
    "password": "test",
    "type": "moderator"
}
```
Ответ:
```
{"message":"user created"}
```

### Авторизация
Запрос:
```
{
    "email": "test@mail.ru",
    "password": "test"
}
```
Ответ:
```
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJSb2xlIjoibW9kZXJhdG9yIiwiZXhwIjoxNzIzNDYyODg2fQ.dih7A4PtEhUM4gvxApcIJjEB4offuuSwipvX2o25k8o"}
```

### Создание дома (модератор)
Запрос:
```
{
    "house_number": 1,
    "address": "test address",
    "year_built": 2020,
    "developer": "test-dev"
}
```
Ответ:
```
{"house_number":1,"address":"test address","year_built":2020,"developer":"test-dev","created_at":"2024-08-09T14:44:57.922089954+03:00","last_flat_added_at":"0001-01-01T00:00:00Z"}
```

### Создание квартиры (любой пользователь)
Запрос:
```
{
    "house_number": 1,
    "flat_number": 2,
    "price": 14000,
    "rooms": 2
}
```
Ответ:
```
{"ID":1,"house_number":1,"flat_number":2,"price":14000,"rooms":2,"status":"created","Moderator":""}
```

### Обновление статуса квартиры (модератор)
Запрос:
```
{
    "house_number": 1,
    "flat_number": 2,
    "status": "on moderation"
}
```
Ответ:
```
{"ID":1,"house_number":1,"flat_number":2,"price":14000,"rooms":2,"status":"on moderation","Moderator":""}
```
Ответ, если квартиру уже взял на проверку другой модератор:
```
another moderator has already been assigned to this flat
```

### GET запрос для получения списка квартир в доме
Ответ:
```
[{"ID":1,"house_number":1,"flat_number":2,"price":14000,"rooms":2,"status":"on moderation","Moderator":""}]
```

### GET запрос для подписку на дом
Ответ:
```
{"message":"Subscribed successfully"}
```