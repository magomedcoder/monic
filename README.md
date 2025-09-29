# Monic - лёгкая система мониторинга логов и событий

Monic состоит из двух частей: **Agent** и **Server**

- **Monic Agent** - маленький демон на каждом узле.  
  Подписывается на системные журналы, парсит события и отправляет их в серверную часть.  
  Поддерживаемые транспорты отправки:
    - HTTP
    - gRPC

- **Monic Server** - центральный приёмник.  
  Принимает события по **HTTP** или **gRPC**, проверяет подпись/токен (если задан секрет) и сохраняет данные в
  ClickHouse.  
  Сервер может поднимать оба протокола одновременно.

---

### Статус

- [x] Сбор событий из journald (SSH: успешный вход, ошибка входа, невалидный пользователь, отключение)
- [x] Отправка событий агентом по HTTP
- [x] Приём событий сервером
- [x] Запись событий в ClickHouse
- [x] Отправка событий агентом по gRPC
- [x] Детекция порт-скана:
    - [x] через firewall/journal
    - [x] через пассивный сниффинг
- [ ] Лёгкий веб-интерфейс (таблица событий и фильтры)

---

##### ВАЖНО: перед запуском измените секрет и включите нужный режим в файлах monic-agent.service и monic-server.service

---

### Запуск сервера

#### Вариант 1: На хосте через systemd

#### Установка ClickHouse

```bash
# Следуйте официальной документации:
# https://clickhouse.com/docs/install/debian_ubuntu

# После установки ClickHouse, запустите клиент и создайте базу данных:
clickhouse-client --query "CREATE DATABASE monic_db;" --user **** --password ****
```

```bash
sudo cp /init/monic-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable monic-server
sudo systemctl start monic-server
```

#### Проверить работу

```bash
sudo systemctl status monic-server
sudo journalctl -u monic-server -f
```

#### Вариант 2: Docker

```bash
docker compose up -d
```

#### Проверить работу

```bash
docker compose logs -f monic-server
```

---

### Переменные окружения

- `MONIC_SERVER_HTTP_ADDR` - адрес HTTP-сервера, по умолчанию `:8000`.
- `MONIC_SERVER_GRPC_ADDR` - адрес gRPC-сервера, по умолчанию пусто (gRPC выключен). Пример: `:50051`.
- `MONIC_SERVER_TLS_CERT` - путь к `cert.pem` для gRPC TLS (опционально).
- `MONIC_SERVER_TLS_KEY` - путь к `key.pem` для gRPC TLS (опционально).
- `MONIC_SECRET` - секрет:
    - HTTP: используется для HMAC (заголовок `X-Signature: sha256=<hex>`).
    - gRPC: используется как Bearer-токен в metadata `authorization: Bearer <secret>`.
- `MONIC_SERVER_CLICKHOUSE_DSN` - DSN для подключения к ClickHouse, по
  умолчанию `tcp://127.0.0.1:9000?database=monic_db`
- `MONIC_SERVER_BATCH_SIZE` - размер батча перед вставкой, по умолчанию `500`
- `MONIC_SERVER_BATCH_WINDOW_MS` - окно времени в мс перед отправкой батча, по умолчанию `500`

Можно одновременно задать и `MONIC_SERVER_HTTP_ADDR`, и `MONIC_SERVER_GRPC_ADDR` - сервер запустит оба протокола.

---

# Агент

### Запуск агента

```bash
sudo cp /init/monic-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable monic-agent
sudo systemctl start monic-agent
```

#### Проверить работу

```bash
sudo systemctl status monic-agent
sudo journalctl -u monic-agent -f
```

### Переменные окружения

- **HTTP режим**
    - `MONIC_HTTP_URL` - URL (например, `http://127.0.0.1:8000`)

- **gRPC режим (приоритетнее, если задан `MONIC_GRPC_ADDR`)**
    - `MONIC_GRPC_ADDR` - адрес gRPC-сервера (например, `127.0.0.1:50051`)
    - `MONIC_GRPC_INSECURE` - `true` для нешифрованного gRPC, по умолчанию `false`

- **Общие параметры:**
    - `MONIC_SECRET` - секрет:
        - в HTTP режиме используется для HMAC (`X-Signature`)
        - в gRPC режиме используется как Bearer-токен (`authorization: Bearer <secret>`)

    - `MONIC_JOURNAL_UNIT` - systemd unit для фильтрации, по умолчанию `sshd.service`.  
      Если пусто - используется `SYSLOG_IDENTIFIER=sshd`.

        - `MONIC_ENABLE_PORTSCAN` - включить детектор/сниффинг, по умолчанию `false`
            - **Детектор порт-скана**
                - `MONIC_PORTSCAN_WINDOW_SECONDS` - окно T секунд (по умолчанию `10`)
                - `MONIC_PORTSCAN_DISTINCT_PORTS` - порог N разных портов (по умолчанию `12`)

            - **Сниффинг**
                - `MONIC_SNIFFER_IFACES` - запятая-разделённый список интерфейсов: `ens160,ens192`
                - `MONIC_SNIFFER_PROMISC` - `true` для promiscuous, по умолчанию `false`
                - `MONIC_SNIFFER_BPF` - кастомный BPF-фильтр.
                  По умолчанию: `(tcp[tcpflags] & (tcp-syn) != 0 and (tcp[tcpflags] & tcp-ack) = 0) or udp`

---

### Включение детектора порт-скана

#### Вариант А: через firewall/journal (nftables/iptables -> kernel-logs)

- Добавьте правила логирования входящих TCP SYN/UDP на «неразрешённые» порты.
- Агент уже слушает kernel-журнал и парсит строки вида ... PROTO=TCP ... SRC=1.2.3.4 ... DPT=3389 ....

#### iptables

```bash
iptables -N MONIC-AGENT
iptables -A INPUT -p tcp --syn -j MONIC-AGENT
iptables -A MONIC-AGENT -m multiport ! --dports 22,80,443 -j LOG --log-prefix "PORTSCAN "
iptables -A MONIC-AGENT -m multiport ! --dports 22,80,443 -j REJECT
```

#### nftables

```bash
nft add table inet monic-agent
nft 'add chain inet monic-agent input { type filter hook input priority 0; }'
nft add rule inet monic-agent input ct state established,related accept
nft add rule inet monic-agent input iif lo accept
nft add rule inet monic-agent input tcp dport { 22,80,443 } accept
nft add rule inet monic-agent input tcp flags syn log prefix "PORTSCAN "
nft add rule inet monic-agent input tcp flags syn reject
```

```bash
sudo MONIC_ENABLE_PORTSCAN=true monic-agent
```

#### Вариант Б: через сниффинг трафика

```bash
sudo MONIC_SNIFFER_IFACES=ens160,ens192 \
  MONIC_ENABLE_PORTSCAN=true \
  MONIC_SNIFFER_PROMISC=true \
  monic-agent
```

Агент будет генерировать «сырые» net_probe и агрегированные port_scan.

---

### Сборка

- **go** 1.24
- **make**
- Для агента: `pkg-config`, `libsystemd-dev`, `libpcap-dev`

> **protoc**, **protoc-gen-go** и **protoc-gen-go-grpc** требуются только для
> **генерации gRPC/protobuf кода** из **.proto** файлов.
> - **protoc** 3.20
> - **protoc-gen-go** 1.36
> - **protoc-gen-go-grpc** 1.5

```bash
sudo apt install pkg-config libsystemd-dev libpcap-dev make

pkg-config --cflags --libs libsystemd, libpcap
```

```bash
make build
```
