# Накопительная система лояльности «Гофермарт»

Система представляет собой HTTP API со следующими возможностями:

* регистрация, аутентификация и авторизация пользователей;
* приём номеров заказов от зарегистрированных пользователей;
* учёт и ведение списка переданных номеров заказов зарегистрированного пользователя;
* учёт и ведение накопительного счёта зарегистрированного пользователя;
* проверка принятых номеров заказов через систему расчёта баллов лояльности;
* начисление за каждый подходящий номер заказа положенного вознаграждения на счёт лояльности пользователя.

# About implementation

This implementation provides a complete loyalty system for Gophermart with the following features:

User Management:

* Registration and authentication
* JWT-based authorization


Order Management:

* Uploading orders with Luhn algorithm validation
* Tracking order status
* Retrieving user orders


Loyalty Points:

* Balance tracking
* Points accrual from orders
* Points withdrawal


Integration with Accrual System:

* Polling for order status updates
* Processing rewards



The system follows clean architecture principles with:

* Domain layer (entities, repositories interfaces, services)
* Infrastructure layer (database implementations)
* Delivery layer (HTTP handlers)

To run the application:

* Ensure PostgreSQL is running
* Run with go run cmd/api/main.go or build with go build -o gophermart cmd/api/main.go
* Use environment variables or flags to customize settings

The implemented API endpoints:

* POST /api/user/register - User registration
* POST /api/user/login - User login
* POST /api/user/orders - Upload new order
* GET /api/user/orders - Get user orders
* GET /api/user/balance - Get user balance
* POST /api/user/balance/withdraw - Withdraw points
* GET /api/user/withdrawals - Get withdrawal history