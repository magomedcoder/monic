# Monic - лёгкая система мониторинга логов и событий

Monic состоит из двух частей: **Agent** и **Server**

- **Monic Agent** - маленький демон на каждом узле.  
  Подписывается на системные журналы, парсит события и отправляет их в серверную часть.  
  Поддерживаемые транспорты отправки:
    - HTTP
    - gRPC
    - режим отладки: вывод JSON в stdout

- **Monic Server** - центральный приёмник.  
  Принимает события по **HTTP** или **gRPC**, проверяет подпись/токен (если задан секрет) и сохраняет данные в
  ClickHouse.  
  Сервер может поднимать оба протокола одновременно.

---

## Статус

- [x] Сбор событий из journald (SSH: успешный вход, ошибка входа, невалидный пользователь, отключение)
- [x] Отправка событий агентом по HTTP
- [x] Приём событий сервером
- [x] Запись событий в ClickHouse
- [x] Отправка событий агентом по gRPC
- [ ] Детекция порт-скана:
    - [ ] через firewall/journal
    - [ ] через пассивный сниффинг
- [ ] Лёгкий веб-интерфейс (таблица событий и фильтры)

---

### Запуск сервера в Docker

```bash
docker compose up -d
```

### Проверить работу

```bash
docker compose logs -f monic-server
```

---

### Переменные окружения

- `MONIC_SERVER_HTTP_ADDR` - адрес HTTP-сервера, по умолчанию `:8000`.
- `MONIC_SERVER_GRPC_ADDR` - адрес gRPC-сервера, по умолчанию пусто (gRPC выключен). Пример: `:50051`.
- `MONIC_SERVER_TLS_CERT` - путь к `cert.pem` для gRPC TLS (опционально).
- `MONIC_SERVER_TLS_KEY` - путь к `key.pem` для gRPC TLS (опционально).
- `MONIC_SERVER_SHARED_SECRET` - общий секрет:
    - HTTP: используется для HMAC (заголовок `X-Signature: sha256=<hex>`).
    - gRPC: используется как Bearer-токен в metadata `authorization: Bearer <secret>`.
- `MONIC_SERVER_CLICKHOUSE_DSN` - DSN для подключения к ClickHouse, по
  умолчанию `tcp://127.0.0.1:9000?database=monic_db`
- `MONIC_SERVER_BATCH_SIZE` - размер батча перед вставкой, по умолчанию `500`
- `MONIC_SERVER_BATCH_WINDOW_MS` - окно времени в мс перед отправкой батча, по умолчанию `500`

### Запуск (HTTP)

```bash
MONIC_SERVER_HTTP_ADDR=:8000 \
MONIC_SERVER_CLICKHOUSE_DSN="tcp://127.0.0.1:9000?database=monic_db&username=default&password=default" \
MONIC_SERVER_SHARED_SECRET=secret \
monic-server
```

### Запуск (gRPC, без TLS)

```bash
MONIC_SERVER_GRPC_ADDR=:50051 \
MONIC_SERVER_CLICKHOUSE_DSN="tcp://127.0.0.1:9000?database=monic_db" \
MONIC_SERVER_SHARED_SECRET=secret \
monic-server
```

### Запуск (gRPC, с TLS)

```bash
MONIC_SERVER_GRPC_ADDR=:50051 \
MONIC_SERVER_TLS_CERT=/etc/monic/cert.pem \
MONIC_SERVER_TLS_KEY=/etc/monic/key.pem \
MONIC_SERVER_CLICKHOUSE_DSN="tcp://127.0.0.1:9000?database=monic_db" \
MONIC_SERVER_SHARED_SECRET=secret \
monic-server
```

Можно одновременно задать и `MONIC_SERVER_HTTP_ADDR`, и `MONIC_SERVER_GRPC_ADDR` - сервер запустит оба протокола.

---

# Агент

### Переменные окружения

- **HTTP режим**
    - `MONIC_HTTP_URL` - URL (например, `http://127.0.0.1:8000`)

- **gRPC режим (приоритетнее, если задан `MONIC_GRPC_ADDR`)**
    - `MONIC_GRPC_ADDR` - адрес gRPC-сервера (например, `127.0.0.1:50051`)
    - `MONIC_GRPC_INSECURE` - `true` для нешифрованного gRPC, по умолчанию `false`

**Общие параметры:**

- `MONIC_SHARED_SECRET` - общий секрет:
    - в HTTP режиме используется для HMAC (`X-Signature`)
    - в gRPC режиме используется как Bearer-токен (`authorization: Bearer <secret>`)

- `MONIC_JOURNAL_UNIT` - systemd unit для фильтрации, по умолчанию `sshd.service`.  
  Если пусто - используется `SYSLOG_IDENTIFIER=sshd`.

### Запуск (HTTP)

```bash
sudo MONIC_HTTP_URL=http://127.0.0.1:8000 \
MONIC_SHARED_SECRET=secret \
monic-agent
```

### Запуск (gRPC, без TLS)

```bash
sudo MONIC_GRPC_ADDR=127.0.0.1:50051 \
MONIC_GRPC_INSECURE=true \
MONIC_SHARED_SECRET=secret \
monic-agent
```

### Запуск (gRPC, с TLS)

```bash
sudo MONIC_GRPC_ADDR=127.0.0.1:50051 \
MONIC_GRPC_INSECURE=false \
MONIC_SHARED_SECRET=secret \
monic-agent
```

---

### Сборка

- **go** 1.24
- **make**
- Для агента: `libsystemd-dev` и `pkg-config`

> **protoc**, **protoc-gen-go** и **protoc-gen-go-grpc** нужны только в том случае,  
> если вы хотите **перегенерировать gRPC/protobuf код** из `.proto` файлов.
> - **protoc** 3.20
> - **protoc-gen-go** 1.36
> - **protoc-gen-go-grpc** 1.5

```bash
sudo apt install libsystemd-dev pkg-config make

pkg-config --cflags --libs libsystemd
```

```bash
make build
```

---

### Формат события

```json
{
  "dateTime": "2025-09-22T10:20:30Z",
  "server": "host01",
  "type": "ssh_accepted|ssh_failed|ssh_invalid_user|ssh_disconnect",
  "user": "root",
  "remoteIp": "1.2.3.4",
  "port": "22",
  "method": "publickey|password|...",
  "message": "accepted|failed|invalid_user|disconnected|connection_closed",
  "raw": "<исходная строка journald>"
}
```
