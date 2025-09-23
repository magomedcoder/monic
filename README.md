# Monic - лёгкая система мониторинга логов и событий

Monic состоит из двух частей: **Agent** и **Server**

### Архитектура

- **Monic Agent** - маленький демон на каждом узле.
  Подписывается на системные журналы, парсит события и отправляет их в серверную часть.
  Может работать и в режиме вывода JSON в stdout.

- **Monic Server** - центральный приёмник.
  Принимает события по HTTP, проверяет подпись HMAC (если задан секрет) и сохраняет данные в ClickHouse.

---

## Статус

- [x] Сбор событий из journald
- [x] Отправка событий агентом по HTTP
- [x] Приём событий сервером
- [ ] Запись событий в ClickHouse
- [ ] Отправка событий агентом по GRPC
- [ ] Лёгкий веб-интерфейс с таблицей событий и фильтрами

---

### Переменные окружения сервера

- `MONIC_SERVER_ADDR` - адрес для HTTP-сервера, по умолчанию `:8080`
- `MONIC_SERVER_SHARED_SECRET` - секрет для проверки HMAC-подписи (заголовок `X-Signature`)

### Сборка сервера

### Debian / Ubuntu

```bash
make build
```

### Запуск

```bash
MONIC_SERVER_ADDR=:8000 MONIC_SERVER_SHARED_SECRET=secret ./build/monic-server
```

---

### Переменные окружения агента

- `MONIC_WEBHOOK_URL` - URL вебхука
- `MONIC_SHARED_SECRET` - секрет для HMAC
- `MONIC_JOURNAL_UNIT` - systemd unit для фильтрации, по умолчанию `sshd.service`.
  Если пусто - используется `SYSLOG_IDENTIFIER=sshd`
- `MONIC_STATE_DIR` - каталог для хранения курсора, по умолчанию `/var/lib/monic-agent`

## Сборка агента

### Debian / Ubuntu

```bash
sudo apt update && sudo apt install -y libsystemd-dev pkg-config make
pkg-config --cflags --libs libsystemd
make build
```

### Запуск

```bash
sudo MONIC_WEBHOOK_URL=http://127.0.0.1:8000/webhook MONIC_SHARED_SECRET=secret ./build/monic-agent
```

---

### Формат события

```json
{
  "ts": "2025-09-22T10:20:30Z",
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