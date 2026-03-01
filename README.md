# raytrade

Cобственная реализация торгового терминала на базе [raylib](https://github.com/raysan5/raylib).

Проект находится в разработке. До релеиза как до луны. Код не является безопасным. Использовать на свой страх и риск.

### Установка
```bash
git clone --depth 1 https://github.com/nlkli/raytrade
cd raytrade
go run .
```

### Авторизация API

Ключи API указываются в файле `.env`:

```bash
BYBIT_API_KEY=your_api_key_here
BYBIT_API_SECRET=your_api_secret_here
```

На данный момент поддерживается только биржа Bybit.  
Для добавления других бирж необходимо реализовать интерфейс [`internal/broker/broker.go`](https://github.com/nlkli/raytrade/blob/main/internal/broker/broker.go)

### Конфигурация
- Описание структуры: [`internal/app/core/config.go`](https://github.com/nlkli/raytrade/blob/main/internal/app/core/config.go)
- Пример файла: [`config.json`](https://github.com/nlkli/raytrade/blob/main/config.json)
- Логика построения дерева компонентов из конфига: [`internal/app/comps/root.go`](https://github.com/nlkli/raytrade/blob/main/internal/app/comps/root.go)

### Команды
Вход в режим ввода команд — клавиша `:`  
Доступные команды и их обработка: [`internal/app/core/cmd.go`](https://github.com/nlkli/raytrade/blob/main/internal/app/core/cmd.go)

Команды можно объединять в последовательность через `|`
В команды подставляются переменные из конфига

#### Переменные конфигурации
Используются для подстановки в команды (например, `sub chart 0 F.$symbol.1` → `$symbol` заменяется на `BTCUSDT`).  
Переменная может содержать команду (до двух уровней вложенности).  
Переменные можно назначать и менять через команды.

### Бинды клавиш
По умолчанию задан только вход в режим команд. Все остальные бинды настраиваются в конфиге.  
Поддерживаются последовательности клавиш (как в vim) и сочетания с `Ctrl` (например, `<C-a>`). 

*bind: char(s) - command prompt*

Обработка нажатий: [`internal/app/core/controller.go`](https://github.com/nlkli/raytrade/blob/main/internal/app/core/controller.go)

Сочетание биндов, переменных и команд даёт гибкую систему для настройки любого действия.

## Demo

### Watch Config
*1 марта 2026, 18:09*

Динамическое изменение конфигурации без перезагрузки.
![watch-config-o](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/watch-config-o.gif)

### Chart Increase
*1 марта 2026, 4:30*

Добавлена прокрутка orderbook, автоматическое расширение начала графика (мелкий шаг для демонстрации) и регенерация графика после падения стрима.

Реализована слежка за изменениями файла конфигурации во время работы — параметры (тема, чувствительность мыши и др.) подгружаются на лету без перезагрузки.

Фоновые задачи теперь выполняются в отдельных потоках и не блокируют отправку новых.
![chart-increase-o](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/chart-increase-o.gif)

### Matrix Demo
*27 февраля 2026, 3:14*

![matrix-demo-o](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/matrix-demo-o.gif)

### Lines and Levels
*26 февраля 2026, 4:32*

Добавлена возможность рисовать на графике и выставлять уровни цены.
![lines-and-levels-o](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/lines-and-levels-o.gif)

### Binds and Move
*24 февраля 2026, 22:04*

Добавлен вертикальный orderbook, перемещение графика мышкой и управление с клавиатуры.

Бинды и команды настраиваются в конфиге. Поддерживаются последовательности клавиш (как в vim), переменные из конфига (`$symbol` → BTCUSDT) и пайплайн команд (`|`). Переменные могут содержать команды (до 2 уровней, без рекурсии).
![binds-and-move](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/binds-and-move-o.gif)

### Custom Layout
*18 февраля 2026, 15:25*

Первые шаги в кастомизации разметки — возможность собирать интерфейс из доступных компонентов.
![custom-layout-o](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/custom-layout-o.gif)

### First Orderbook
*18 февраля 2026, 5:04*

Стрим и первая отрисовка Orderbook.
![first-orderbook-o](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/first-orderbook-o.gif)

### First Chart
*12 февраля 2026, 17:20*

Стрим свечей и первая отрисовка графика.
![first-chart-o](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/first-chart-o.gif)

---

## ![recol](https://github.com/nlkli/recol.git) (change colorscheme)

```bash
git clone https://github.com/nlkli/recol.git
cd recol
cargo build --release
cp target/release/recol /usr/local/bin/
```

## SCC

```text
Mon Mar  2 00:05:55 MSK 2026
───────────────────────────────────────────────────────────────────────────────
Language                 Files     Lines   Blanks  Comments     Code Complexity
───────────────────────────────────────────────────────────────────────────────
Go                          34      7317     1327       355     5635        774
JSON                         1       179        0         0      179          0
License                      1        21        4         0       17          0
Makefile                     1        15        5         0       10          0
Markdown                     1       107       36         0       71          0
Shell                        1        12        3         1        8          1
───────────────────────────────────────────────────────────────────────────────
Total                       39      7651     1375       356     5920        775
───────────────────────────────────────────────────────────────────────────────
Estimated Cost to Develop (organic) $174,796
Estimated Schedule Effort (organic) 7.09 months
Estimated People Required (organic) 2.19
───────────────────────────────────────────────────────────────────────────────
Processed 168588 bytes, 0.169 megabytes (SI)
───────────────────────────────────────────────────────────────────────────────
```

## Tree

```text
Mon Mar  2 00:07:11 MSK 2026
.
├── LICENSE
├── Makefile
├── README.md
├── config.json
├── go.mod
├── go.sum
├── internal
│   ├── app
│   │   ├── app.go
│   │   ├── comps
│   │   │   ├── chart.go
│   │   │   ├── footer.go
│   │   │   ├── order.go
│   │   │   ├── orderbook.go
│   │   │   ├── position.go
│   │   │   ├── rect.go
│   │   │   └── root.go
│   │   └── core
│   │       ├── background.go
│   │       ├── cmd.go
│   │       ├── config.go
│   │       ├── controller.go
│   │       ├── palette.go
│   │       └── state.go
│   ├── broker
│   │   ├── broker.go
│   │   ├── bybit
│   │   │   ├── broker.go
│   │   │   ├── client.go
│   │   │   ├── market.go
│   │   │   ├── models
│   │   │   │   ├── enum.go
│   │   │   │   ├── market.go
│   │   │   │   ├── position.go
│   │   │   │   ├── stream.go
│   │   │   │   └── trade.go
│   │   │   ├── position.go
│   │   │   ├── stream.go
│   │   │   └── trade.go
│   │   └── models.go
│   ├── cdl
│   │   ├── cdl.go
│   │   ├── csv.go
│   │   └── interval.go
│   ├── utils
│   │   └── utils.go
│   └── ws
│       └── ws.go
├── main.go
├── main_test.go
├── recol.sh
└── resources
    ├── ComicRelief-Bold.ttf
    └── ComicRelief-Regular.ttf

12 directories, 43 files
```
