## Запуск

Для подключения PostgreSQL БД необходимо ввести конфигурацию в файл configuration.json:

```bash
{
    "host": "localhost",
	"port": "5432",
	"user": "postgres",
    "password": "postgres",
    "db_name": "avito_tech"
}
```

Представление SQL таблиц можно найти в файле tables.sql.

По умолчанию сервер слушает 5000 порт, но при помощи флага -port его можно изменить.

## Пример работы

[Документация](https://app.swaggerhub.com/apis/rdnply/safedeal-backend-trainee/1.0.0#/) 

(Может некорректно отображаться кириллица. Это происходит из-за того, что по умолчанию в консоли нет поддержки UTF-8.)

### Рассчитать стоимость доставки

Запрос:

```bash
curl -is --request POST http://localhost:5000/api/v1/products/1/cost-of-delivery \
	--data '{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50"}'
```

Ответ:

```bash
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Ratelimit-Limit: 10
X-Ratelimit-Remaining: 9
X-Ratelimit-Reset: 1592305860
Date: Tue, 16 Jun 2020 11:10:13 GMT
Content-Length: 183

{"destination":"Большая Садовая, 302-бис, пятый этаж, кв. № 50","from":"Большой Патриарший пер., 7, строение 1","price":2000}
```

### Создать заказ

Запрос:

```bash
curl -is --request POST http://localhost:5000/api/v1/products/1/order \ 
	--data '{"destination" : "Большая Садовая, 302-бис, пятый этаж, кв. № 50", \ 
	"time" : "2020-06-15T15:30:00Z"}'
```

Ответ:

```bash
HTTP/1.1 201 Created
X-Ratelimit-Limit: 10
X-Ratelimit-Remaining: 10
X-Ratelimit-Reset: 1592311380
Date: Tue, 16 Jun 2020 12:42:58 GMT
Content-Length: 0
```

### Получить информацию о заказе

Запрос:

```bash
curl -is --request GET http://localhost:5000/api/v1/orders/3
```

Ответ:

```bash
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Ratelimit-Limit: 10
X-Ratelimit-Remaining: 8
X-Ratelimit-Reset: 1592306340
Date: Tue, 16 Jun 2020 11:18:54 GMT
Content-Length: 380

{
  "id": 3,
  "product": {
    "id": 1,
    "name": "Сноуборд",
    "width": 40.5,
    "length": 143,
    "height": 20,
    "weight": 3.3,
    "place": "Большой Патриарший пер., 7, строение 1"
  },
  "from": "Большой Патриарший пер., 7, строение 1",
  "destination": "Большая Садовая, 302-бис, пятый этаж, кв. № 50",
  "time": "2020-06-15T15:30:00Z"
}
```

### Получить список заказов

Запрос:

```bash
curl -is --request GET http://localhost:5000/api/v1/orders
```

Ответ:

```bash
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Ratelimit-Limit: 10
X-Ratelimit-Remaining: 9
X-Ratelimit-Reset: 1592306340
Date: Tue, 16 Jun 2020 11:18:21 GMT
Content-Length: 151

[
  {
    "id": 1,
    "product_id": 1,
    "name": "Сноуборд"
  },
  {
    "id": 2,
    "product_id": 1,
    "name": "Сноуборд"
  }
]
```

## Тестовое задание

Необходимо разработать прототип API сервиса курьерской доставки на GoLang/PHP
 
### Что должен включать в себя сервис:
 
- API методы
  - Метод расчета стоимости доставки
  - Метод создания заказа
  - Метод получения информации о заказе
  - Метод получение списка заказов
- Сервис должен уметь взаимодействовать с клиентом при помощи REST API или JSON RPC запросов
- В сервисе должен быть реализован RateLimit с ограничением 10 RPM
 
### Дополнительная информация:
 
- Логику логистики реализовывать не обязательно. На усмотрение разработчика можно использовать mock ответов.  Цель разработать API сервис, а не полноценный сервис курьерской доставки.
-	Сервис разрабатывается для “Внутригородской доставки”
- Приветствуется покрытие кода тестами
- Приветствуется наличие документации с описанием работы API сервиса
- Приветствуется использования систем хранения данных Redis, PostgreSQL, mongoDB
 
### Контекст задачи:
 
В процессе работы с API будут участвовать 3 основных лица:
 
- Покупатель. Для него важно рассчитать стоимость доставки от адреса отправки до адреса получения (при этом адрес назначения у нас занесен в систему продавцом), для этого он использует метод "Расчет стоимостей доставки". После расчета стоимости, покупатель принимает решение об оформлении заказа с доставкой. Когда пользователь оформляет заказ, вызывается метод "Создать заказ". После чего, покупатель ожидает доставки своего товара.
- Продавец. Должен иметь возможность видеть весь список заказов, оформленных покупателями. Для этого доступен метод "Получить список заказов". Также для продавца важно иметь возможность посмотреть детальную информацию по заказу, чтобы передать заказ курьеру для доставки. Для этого доступен метод “Получить информацию о заказе”.
- Курьер с мобильным приложением. Должен иметь возможность просматривать информацию о заказе: какой товар, куда и когда нужно доставить. Для этого мобильное приложение будет вызывать метод "Получить информацию о заказе".
