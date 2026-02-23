# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Сборка и запуск

### Простая сборка

```bash
go build -o shortener.exe ./cmd/shortener
```

### Сборка с информацией о версии

Приложение поддерживает встраивание информации о версии, дате сборки и коммите через ldflags:

**Linux/Mac:**
```bash
./build.sh
# или с кастомной версией
VERSION=v2.0.0 ./build.sh
```

**Makefile:**
```bash
make build-with-version
# или с кастомной версией
make build-with-version VERSION=v2.0.0
```

**Ручная сборка:**
```bash
go build -ldflags "\
  -X 'main.buildVersion=v1.0.0' \
  -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' \
  -X 'main.buildCommit=$(git rev-parse --short HEAD)'" \
  -o shortener.exe ./cmd/shortener
```

При запуске приложение выведет информацию о сборке:
```
Build version: v1.0.0
Build date: 2026/02/14 16:40:00
Build commit: 29ac65e
```

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Инструменты разработки

### Reset Code Generator

Автоматический генератор методов `Reset()` для структур Go. Позволяет эффективно сбрасывать состояние объектов к значениям по умолчанию.

Использование:

```bash
go run cmd/reset/main.go
```

Подробнее см. [cmd/reset/README.md](cmd/reset/README.md)

### Object Pool

Универсальный пул объектов с автоматическим вызовом `Reset()`. Использует `sync.Pool` для эффективного переиспользования объектов и снижения нагрузки на GC.

```go
// Создаём пул для типа с методом Reset()
pool := lib.New(func() *model.URLEnt {
    return &model.URLEnt{}
})

// Получаем объект
ent := pool.Get()
ent.ShortURL = "abc123"

// Возвращаем в пул - автоматически вызывается Reset()
pool.Put(ent)
```

**Производительность**: 0 allocs/op, ~18 ns/op  
Подробнее см. [internal/lib/pool.go](internal/lib/pool.go)

## Конфигурационный файл

Помимо флагов командной строки и переменных окружения, приложение поддерживает
загрузку настроек из JSON файла конфигурации.

**Приоритет источников (от низшего к высшему):**

```
значения по умолчанию < файл конфигурации < переменные окружения < флаги командной строки
```

### Указание пути к файлу

| Способ | Пример |
|--------|--------|
| Флаг | `./shortener.exe --config /etc/urlshrter/config.json` или `-c config.json` |
| Переменная окружения | `CONFIG=/etc/urlshrter/config.json ./shortener.exe` |

### Формат файла

```json
{
    "server_address": "localhost:8080",
    "base_url": "http://localhost",
    "file_storage_path": "/path/to/file.db",
    "database_dsn": "postgres://user:pass@localhost/db",
    "encryption_key": "supersecretkey",
    "audit_file": "/var/log/audit.log",
    "audit_url": "http://audit.internal/events",
    "enable_https": true,
    "cert_dir": "/etc/urlshrter/tls"
}
```

Все поля необязательны. Незаданные поля не перезаписывают значения из других источников.
Поле `enable_https` принимает `true` или `false`; если поле отсутствует — значение не изменяется.

### Пример использования

```bash
# Запуск с конфигурационным файлом
./shortener.exe -c config.json

# Переменная окружения перекрывает значение из файла
BASE_URL=https://my.domain ./shortener.exe -c config.json

# Флаг перекрывает и файл, и переменную окружения
./shortener.exe -c config.json -b https://override.domain
```

## Graceful Shutdown

Приложение поддерживает корректное завершение работы при получении сигналов `SIGINT` (Ctrl+C) или `SIGTERM`:

```bash
# Запуск
./shortener.exe

# Graceful остановка через Ctrl+C
```

**Что происходит при остановке:**
1. Сервер перестает принимать новые соединения
2. Завершаются все активные HTTP запросы (таймаут 30 сек)
3. Закрывается соединение с БД
4. Приложение корректно завершается

## HTTPS

Сервер поддерживает работу по HTTPS с автоматической генерацией самоподписанного TLS-сертификата.

### Включение

| Способ | Пример |
|--------|--------|
| Флаг | `./shortener.exe -s` или `./shortener.exe --https` |
| Переменная окружения | `ENABLE_HTTPS=true ./shortener.exe` |

### Директория сертификатов

По умолчанию сертификат (`cert.pem`) и ключ (`key.pem`) хранятся в директории `certs/` относительно рабочей директории. Расположение можно переопределить:

| Способ | Пример |
|--------|--------|
| Флаг | `--cert-dir /etc/urlshrter/tls` |
| Переменная окружения | `CERT_DIR=/etc/urlshrter/tls` |

### Поведение

- Если файлы `cert.pem` и `key.pem` уже существуют в директории — они переиспользуются без изменений.
- Если файлы отсутствуют — директория создаётся автоматически и генерируется новый самоподписанный ECDSA P-256 сертификат со сроком действия 10 лет.
- Генерация выполняется с помощью пакета `crypto/rand`.

### Пример запуска с HTTPS

```bash
# Запуск с HTTPS, сертификаты в директории по умолчанию (certs/)
./shortener.exe -s

# Запуск с HTTPS и явным указанием директории сертификатов
./shortener.exe -s --cert-dir /etc/urlshrter/tls

# Через переменные окружения
ENABLE_HTTPS=true CERT_DIR=/etc/urlshrter/tls ./shortener.exe
```

## Performance Profiling

### Анализ оптимизации памяти

Сравнение профилей памяти до и после оптимизаций:

```bash
go tool pprof -top -diff_base="profiles/base.pprof" "profiles/result.pprof"
```

#### Результаты профилирования

**Общая экономия памяти: -0.48MB (10.72% улучшение от 4.51MB)**

| Функция | Изменение | % от общего |
|---------|-----------|-------------|
| `math/rand.newSource` | **-0.50MB** | 11.15% |
| `runtime.allocm` | **-0.50MB** | 11.12% |
| `repository.NewURLRepository` | **-0.50MB** | 11.10% |
| `encoding/json.literalStore` | **-0.50MB** | 11.10% |
| `regexp/syntax.compiler.inst` | **-0.50MB** | 11.00% |

##### Новые аллокации

| Функция | Изменение | % от общего |
|---------|-----------|-------------|
| `pgx/pgtype.Map.buildReflectTypeToType` | +0.51MB | 11.39% |
| `go-cache.cache.Set` | +0.50MB | 11.15% |

**Примечание**: Новые аллокации связаны с улучшенным пулом соединений БД и кэшированием, что дает общий выигрыш в производительности.

<details>
<summary>Полный вывод pprof (развернуть)</summary>

```
File: shortener.exe
Type: inuse_space
Time: 2026-01-25 17:57:47 MSK
Showing nodes accounting for -0.48MB, 10.72% of 4.51MB total

      flat  flat%   sum%        cum   cum%
    0.51MB 11.39% 11.39%     0.51MB 11.39%  github.com/jackc/pgx/v5/pgtype.(*Map).buildReflectTypeToType (inline)
    0.50MB 11.15% 22.54%     0.50MB 11.15%  github.com/patrickmn/go-cache.(*cache).Set
   -0.50MB 11.15% 11.39%    -0.50MB 11.15%  math/rand.newSource (inline)
   -0.50MB 11.12%  0.27%    -0.50MB 11.12%  runtime.allocm
    0.50MB 11.11% 11.38%     0.50MB 11.11%  regexp.onePassCopy
   -0.50MB 11.10%  0.28%    -0.50MB 11.10%  runtime.malg
   -0.50MB 11.10% 10.81%    -0.50MB 11.10%  context.(*cancelCtx).Done
    0.50MB 11.10%  0.28%     0.50MB 11.10%  regexp/syntax.(*parser).newRegexp (inline)
   -0.50MB 11.10% 10.81%    -0.50MB 11.04%  github.com/rebaxis/urlshrter/internal/repository.NewURLRepository
   -0.50MB 11.10% 21.91%    -0.50MB 11.10%  encoding/json.(*decodeState).literalStore
    0.50MB 11.09% 10.81%     0.50MB 11.09%  syscall.LoadDLL
    0.50MB 11.09%  0.28%     0.50MB 11.09%  os.newFile
   -0.50MB 11.00% 10.72%    -0.50MB 11.00%  regexp/syntax.(*compiler).inst (inline)
         0     0% 10.72%     0.51MB 11.39%  database/sql.(*DB).Ping (inline)
         0     0% 10.72%     0.51MB 11.39%  database/sql.(*DB).PingContext
         0     0% 10.72%    -0.50MB 11.10%  database/sql.(*DB).connectionOpener
         0     0% 10.72%    -0.50MB 11.10%  encoding/json.Unmarshal
         0     0% 10.72%     1.01MB 22.31%  github.com/asaskevich/govalidator.init
         0     0% 10.72%     0.51MB 11.39%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 10.72%     0.51MB 11.39%  github.com/rebaxis/urlshrter/internal/db.NewDB
         0     0% 10.72%    -0.50MB 11.04%  main.NewRepo
         0     0% 10.72%    -0.49MB 10.81%  main.main
         0     0% 10.72%    -0.50MB 11.15%  math/rand.NewSource (inline)
```

</details>
