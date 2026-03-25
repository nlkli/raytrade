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

### Order and position
*9 марта 2026, 23:23*

Выставление и отмена лимитных заказов

Позиция
![order-and-pos](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/order-and-pos-o.gif)

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

## SCC

```text
Wed Mar 25 23:18:58 MSK 2026
───────────────────────────────────────────────────────────────────────────────
Language                 Files     Lines   Blanks  Comments     Code Complexity
───────────────────────────────────────────────────────────────────────────────
Go                          38      9516     1735       351     7430       1155
JSON                         2       373        0         0      373          0
License                      1        21        4         0       17          0
Markdown                     1       195       44         0      151          0
Shell                        1        15        4         3        8          1
───────────────────────────────────────────────────────────────────────────────
Total                       43     10120     1787       354     7979       1156
───────────────────────────────────────────────────────────────────────────────
Estimated Cost to Develop (organic) $239,133
Estimated Schedule Effort (organic) 7.99 months
Estimated People Required (organic) 2.66
───────────────────────────────────────────────────────────────────────────────
Processed 217642 bytes, 0.218 megabytes (SI)
───────────────────────────────────────────────────────────────────────────────
```

## Tree

```text
Wed Mar 25 23:18:58 MSK 2026
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
│   │   ├── binance
│   │   │   └── broker.go
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
│   │   ├── interval.go
│   │   └── ta
│   │       ├── ma.go
│   │       └── ta.go
│   ├── utils
│   │   └── utils.go
│   └── ws
│       └── ws.go
├── main.go
├── main_test.go
├── mylayout.json
├── recol.sh
└── resources
    ├── ChangaOne-Regular.ttf
    ├── ComicRelief-Bold.ttf
    ├── ComicRelief-Regular.ttf
    ├── ConcertOne-Regular.ttf
    ├── Jersey15-Regular.ttf
    ├── Micro5-Regular.ttf
    └── Ubuntu-Bold.ttf

14 directories, 52 files
```
