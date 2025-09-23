# Monic

### Переменные окружения

- `MONIC_WEBHOOK_URL` - URL вебхука
- `MONIC_SHARED_SECRET` - секрет для HMAC
- `MONIC_JOURNAL_UNIT` - systemd unit для фильтрации, по умолчанию sshd.service. Если пусто - используется SYSLOG_IDENTIFIER=sshd
- `MONIC_STATE_DIR` - каталог для хранения курсора, по умолчанию /var/lib/monic

## Сборка

### Debian / Ubuntu

```bash
sudo apt update && sudo apt install -y libsystemd-dev pkg-config make
pkg-config --cflags --libs libsystemd
make build
```

### Запуск

```bash
sudo MONIC_WEBHOOK_URL=http://127.0.0.1:8000/webhook MONIC_SHARED_SECRET=secret ./build/monic
```

### Формат события

```json
{
  "ts": "2025-09-22T10:20:30Z",
  "server": "host01",
  "type": "ssh_accepted | ssh_failed | ssh_invalid_user | ssh_disconnect",
  "user": "root",
  "rhost": "1.2.3.4",
  "port": "51432",
  "method": "publickey|password|...",
  "message": "accepted|failed|invalid_user|disconnected|connection_closed",
  "raw": "<исходная строка journald>"
}
```