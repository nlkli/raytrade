package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func SubstituteVarsIntoPrompt(prompt string) string {
	for k, v := range vars {
		prompt = strings.ReplaceAll(prompt, "$"+k, v)
	}
	return prompt
}

func SubstituteVarsIntoPromptV2(prompt string, vars []string) string {
	r := strings.NewReplacer(vars...)
	return r.Replace(prompt)
}

func buildReplacerArgs(m map[string]string) []string {
	args := make([]string, 0, len(m)*2)
	for k, v := range m {
		args = append(args, "$"+k, v)
	}
	return args
}

func TestS(t *testing.T) {
	start := time.Now()
	res := SubstituteVarsIntoPrompt(prompt)
	elapsed := time.Since(start)

	// В наносекундах
	fmt.Printf("Время выполнения: %d нс\n", elapsed.Nanoseconds())
	fmt.Printf("Время выполнения: %.6f мс\n", float64(elapsed.Nanoseconds())/1e6)

	if res != promptRes {
		t.Error("Test dont pass 000000")
	}

	println("-------------------------------------")

	vvars := buildReplacerArgs(vars)
	start = time.Now()
	res = SubstituteVarsIntoPromptV2(prompt, vvars)
	elapsed = time.Since(start)

	// В наносекундах
	fmt.Printf("Время выполнения: %d нс\n", elapsed.Nanoseconds())
	fmt.Printf("Время выполнения: %.6f мс\n", float64(elapsed.Nanoseconds())/1e6)

	println(res)

	if res != promptRes {
		t.Error("Test dont pass")
	}

	println("Test passed!")
}

var prompt = `
========================================
        ТОРГОВЫЙ ОТЧЕТ $user_name
========================================

Дата и время: $datetime ($timezone)
Пользователь: $user_name (ID: $user_id)
Уровень: $user_level | Рейтинг: $user_rating
Страна: $user_country, $user_city

----------------------------------------
ТЕКУЩАЯ ТОРГОВАЯ ПАРА: $symbol
----------------------------------------

Текущая цена: $price USDT
Изменения: 1ч: $change_1h% | 24ч: $change_24h% | 7д: $change_7d% | 30д: $change_30d%
24ч максимум: $high_24h | 24ч минимум: $low_24h
Объем 24ч: $volume_24h | Рыночная капитализация: $market_cap

----------------------------------------
ТЕХНИЧЕСКИЕ ИНДИКАТОРЫ
----------------------------------------

EMA (9/20/50/100/200): $ema_9 / $ema_20 / $ema_50 / $ema_100 / $ema_200
SMA (9/20/50/100/200): $sma_9 / $sma_20 / $sma_50 / $sma_100 / $sma_200
Полосы Боллинджера (верх/сред/ниж): $bollinger_upper / $bollinger_mid / $bollinger_lower
RSI (14): $rsi_14
Stochastic (K/D): $stoch_k / $stoch_d
MACD (линия/сигнал/гист): $macd_line / $macd_signal / $macd_hist
ADX: $adx | DMI (+/-): $dmi_plus / $dmi_minus
Ишимоку (тенкан/киджун/сенкоу): $ichimoku_tenkan / $ichimoku_kijun / $ichimoku_senkou
ATR: $atr | VWAP: $vwap | OBV: $obv

----------------------------------------
ДРУГИЕ ТОРГОВЫЕ ПАРЫ
----------------------------------------

ETH: $symbol_eth - цена $price_eth USDT
SOL: $symbol_sol - цена $price_sol USDT
ADA: $symbol_ada - цена $price_ada USDT
DOT: $symbol_dot - цена $price_dot USDT
MATIC: $symbol_matic - цена $price_matic USDT
LINK: $symbol_link - цена $price_link USDT
AVAX: $symbol_avax - цена $price_avax USDT
XRP: $symbol_xrp - цена $price_xrp USDT
DOGE: $symbol_doge - цена $price_doge USDT
SHIB: $symbol_shib - цена $price_shib USDT

----------------------------------------
СТАТИСТИКА ПОЛЬЗОВАТЕЛЯ
----------------------------------------

Баланс: $user_balance USDT
Всего сделок: $user_trades | Винрейт: $user_winrate%
Дата регистрации: $user_created | Последний вход: $user_last_login
Статус: $user_status | Риск-профиль: $user_risk
Стратегия: $user_strategy | Любимая пара: $user_fav_pair

----------------------------------------
ФИНАНСОВЫЕ ПОКАЗАТЕЛИ
----------------------------------------

Прибыль сегодня: $profit USDT ($profit_percent%)
Убыток сегодня: $loss USDT ($loss_percent%)
Общая прибыль: $total_profit USDT
Общий убыток: $total_loss USDT
Чистая прибыль: $net_profit USDT
ROI: $roi% | Коэффициент Шарпа: $sharpe_ratio
Макс. просадка: $max_drawdown% ($max_drawdown_usd USDT)
Фактор прибыли: $profit_factor
Средний выигрыш: $avg_win | Средний проигрыш: $avg_loss
Всего сделок: $trades_total (выигрышных: $trades_win, проигрышных: $trades_loss)
Общий объем: $volume_total USDT
Комиссия: $commission USDT ($commission_rate%)
Проскальзывание: $slippage

----------------------------------------
ПОЗИЦИИ И ОРДЕРА
----------------------------------------

Текущая позиция: $position_side | Размер: $position_size
P&L позиции: $position_pnl
Последний ордер: $order_id | Тип: $order_type | Сторона: $order_side
Статус: $order_status | Количество: $order_qty | Цена: $order_price
Итого: $order_total USDT

----------------------------------------
ИНФОРМАЦИЯ О БИРЖЕ
----------------------------------------

Биржа: $exchange ($exchange_full)
Комиссия биржи: $exchange_fee% | Статус: $exchange_status
Ставка финансирования: $funding_rate% | Открытый интерес: $open_interest
Long/Short соотношение: $long_ratio% / $short_ratio%

----------------------------------------
СИСТЕМНАЯ ИНФОРМАЦИЯ
----------------------------------------

API версия: $api_version | Endpoint: $api_endpoint
WebSocket: $ws_endpoint
Timeout: $timeout сек | Retry: $retry_count (задержка $retry_delay сек)
Rate limit: $rate_limit запросов / $rate_window сек
Лог-файл: $log_file | Уровень логирования: $log_level

База данных: $db_host:$db_port/$db_name
Пользователь БД: $db_user
Redis: $redis_host:$redis_port (БД: $redis_db)
Kafka: $kafka_broker (топик: $kafka_topic)
HTTP порт: $http_port | HTTPS порт: $https_port | gRPC порт: $grpc_port

----------------------------------------
УВЕДОМЛЕНИЯ И АЛЕРТЫ
----------------------------------------

Алерт: $alert_id | Тип: $alert_type
Условие: $alert_condition
Уведомление: $notification_id | Тип: $notification_type
Шаблон: $template_id ($template_name)

----------------------------------------
КОНТАКТНАЯ ИНФОРМАЦИЯ
----------------------------------------

Email: $user_email | Телефон: $user_phone
Telegram: $user_telegram | Discord: $user_discord
Twitter: $user_twitter | Instagram: $user_instagram

----------------------------------------
ДОПОЛНИТЕЛЬНАЯ ИНФОРМАЦИЯ
----------------------------------------

Дата создания отчета: $date $time
Временная метка: $timestamp ($unix_timestamp)
Текущий квартал: $quarter | Неделя: $week_number
День года: $day_of_year
Начало года: $date_start | Конец года: $date_end
Вчера: $date_prev | Завтра: $date_next

========================================
        КОНЕЦ ОТЧЕТА
========================================

P.S. Это тестовый отчет для демонстрации замены множества переменных.
Переменные могут использоваться многократно, например: $symbol, $symbol, $symbol
Или в разных контекстах: $price, $price, $price
А также с другими парами: $symbol_eth, $symbol_sol, $symbol_ada

Проверка повторного использования переменных:
$user_name (первый раз)
$user_name (второй раз) 
$user_name (третий раз)
$user_email (первый раз)
$user_email (второй раз)
$symbol (первый раз)
$symbol (второй раз)
$price (первый раз)
$price (второй раз)

Общее количество переменных в этом тексте: более 200 уникальных переменных,
каждая из которых будет заменена соответствующим значением.
Функция SubstituteVarsIntoPrompt должна обработать все эти замены корректно.
`

var vars = map[string]string{
	// Криптовалюты (30+)
	"symbol":         "BTCUSDT",
	"symbol_eth":     "ETHUSDT",
	"symbol_sol":     "SOLUSDT",
	"symbol_ada":     "ADAUSDT",
	"symbol_dot":     "DOTUSDT",
	"symbol_matic":   "MATICUSDT",
	"symbol_link":    "LINKUSDT",
	"symbol_avax":    "AVAXUSDT",
	"symbol_uni":     "UNIUSDT",
	"symbol_xrp":     "XRPUSDT",
	"symbol_doge":    "DOGEUSDT",
	"symbol_shib":    "SHIBUSDT",
	"symbol_ltc":     "LTCUSDT",
	"symbol_bch":     "BCHUSDT",
	"symbol_atom":    "ATOMUSDT",
	"symbol_etc":     "ETCUSDT",
	"symbol_fil":     "FILUSDT",
	"symbol_theta":   "THETAUSDT",
	"symbol_ftm":     "FTMUSDT",
	"symbol_vechain": "VETUSDT",
	"symbol_xlm":     "XLMUSDT",
	"symbol_algo":    "ALGOUSDT",
	"symbol_egld":    "EGLDUSDT",
	"symbol_axs":     "AXSUSDT",
	"symbol_sand":    "SANDUSDT",
	"symbol_mana":    "MANAUSDT",
	"symbol_gala":    "GALAUSDT",
	"symbol_ape":     "APEUSDT",
	"symbol_icp":     "ICPUSDT",
	"symbol_near":    "NEARUSDT",
	"symbol_flow":    "FLOWUSDT",
	"symbol_ftt":     "FTTUSDT",
	"symbol_cro":     "CROUSDT",
	"symbol_klay":    "KLAYUSDT",

	// Числовые значения (30+)
	"price":          "45000.50",
	"price_btc":      "42350.75",
	"price_eth":      "2850.25",
	"price_sol":      "98.45",
	"price_ada":      "0.52",
	"price_dot":      "7.89",
	"price_matic":    "0.95",
	"price_link":     "15.67",
	"price_avax":     "35.23",
	"price_uni":      "6.12",
	"price_xrp":      "0.62",
	"price_doge":     "0.082",
	"price_shib":     "0.000023",
	"price_ltc":      "82.34",
	"price_bch":      "345.67",
	"price_atom":     "11.23",
	"volume":         "1250000",
	"volume_24h":     "45000000",
	"market_cap":     "890000000",
	"market_cap_btc": "820000000000",
	"market_cap_eth": "340000000000",
	"market_cap_sol": "42000000000",
	"high_24h":       "46000.00",
	"low_24h":        "44500.00",
	"change_1h":      "0.5",
	"change_24h":     "-2.3",
	"change_7d":      "5.6",
	"change_30d":     "15.8",
	"rsi":            "65.4",
	"macd":           "125.6",
	"volume_profile": "Высокий",

	// Даты и время (20+)
	"date":            "2026-02-22",
	"time":            "15:30:45",
	"datetime":        "2026-02-22 15:30:45",
	"timestamp":       "1740231045",
	"year":            "2026",
	"month":           "02",
	"day":             "22",
	"hour":            "15",
	"minute":          "30",
	"second":          "45",
	"weekday":         "Воскресенье",
	"month_name":      "Февраль",
	"quarter":         "Q1",
	"week_number":     "8",
	"day_of_year":     "53",
	"unix_timestamp":  "1740231045",
	"date_next":       "2026-02-23",
	"date_prev":       "2026-02-21",
	"date_start":      "2026-01-01",
	"date_end":        "2026-12-31",
	"timezone":        "UTC+3",
	"timezone_offset": "+0300",

	// Пользователи и трейдеры (25+)
	"user_id":         "12345",
	"user_name":       "crypto_trader_99",
	"user_email":      "trader@example.com",
	"user_balance":    "100000",
	"user_level":      "VIP",
	"user_rating":     "5.0",
	"user_trades":     "1250",
	"user_winrate":    "67.5",
	"user_country":    "Россия",
	"user_city":       "Москва",
	"user_age":        "28",
	"user_created":    "2024-05-15",
	"user_last_login": "2026-02-21",
	"user_status":     "Активный",
	"user_risk":       "Средний",
	"user_strategy":   "Долгосрочный",
	"user_fav_pair":   "BTC/USDT",
	"user_referrer":   "friend123",
	"user_avatar":     "avatar_123.png",
	"user_bio":        "Профессиональный трейдер",
	"user_phone":      "+7-999-123-45-67",
	"user_telegram":   "@crypto_trader",
	"user_discord":    "trader#1234",
	"user_twitter":    "@crypto_trader",
	"user_instagram":  "crypto_trader_inst",

	// Технические индикаторы (25+)
	"ema_9":           "45123.45",
	"ema_20":          "44987.23",
	"ema_50":          "44123.67",
	"ema_100":         "43245.89",
	"ema_200":         "41234.56",
	"sma_9":           "45098.76",
	"sma_20":          "44965.43",
	"sma_50":          "44112.34",
	"sma_100":         "43234.78",
	"sma_200":         "41245.67",
	"bollinger_upper": "46234.56",
	"bollinger_mid":   "44987.23",
	"bollinger_lower": "43739.90",
	"rsi_14":          "62.5",
	"stoch_k":         "78.9",
	"stoch_d":         "75.6",
	"macd_line":       "125.6",
	"macd_signal":     "120.3",
	"macd_hist":       "5.3",
	"volume_sma":      "1200000",
	"vwap":            "45100.00",
	"atr":             "450.25",
	"obv":             "12500000",
	"adx":             "35.6",
	"dmi_plus":        "25.4",
	"dmi_minus":       "18.7",
	"ichimoku_tenkan": "45000",
	"ichimoku_kijun":  "44800",
	"ichimoku_senkou": "45200",

	// Финансовые показатели (30+)
	"profit":           "12500.50",
	"profit_percent":   "15.5",
	"loss":             "-2500.75",
	"loss_percent":     "-3.2",
	"total_profit":     "150000.00",
	"total_loss":       "-45000.00",
	"net_profit":       "105000.00",
	"roi":              "23.4",
	"sharpe_ratio":     "1.85",
	"sortino_ratio":    "2.34",
	"calmar_ratio":     "1.56",
	"max_drawdown":     "-12.5",
	"max_drawdown_usd": "-15000",
	"win_rate":         "68.5",
	"avg_win":          "850.25",
	"avg_loss":         "-425.50",
	"profit_factor":    "2.35",
	"trades_total":     "1250",
	"trades_win":       "856",
	"trades_loss":      "394",
	"volume_total":     "25000000",
	"commission":       "1250.75",
	"commission_rate":  "0.1",
	"slippage":         "5.25",
	"leverage":         "5",
	"margin":           "20000",
	"liquidation":      "42350.00",
	"funding_rate":     "0.01",
	"open_interest":    "450000000",
	"long_ratio":       "65.5",
	"short_ratio":      "34.5",

	// Системные параметры (25+)
	"api_version":     "v2",
	"api_endpoint":    "https://api.binance.com",
	"ws_endpoint":     "wss://stream.binance.com",
	"timeout":         "30",
	"retry_count":     "3",
	"retry_delay":     "5",
	"max_connections": "100",
	"rate_limit":      "1200",
	"rate_window":     "60",
	"cache_ttl":       "300",
	"log_level":       "INFO",
	"log_file":        "/var/log/trader.log",
	"config_file":     "config.yaml",
	"db_host":         "localhost",
	"db_port":         "5432",
	"db_name":         "trading_db",
	"db_user":         "trader",
	"db_password":     "secure_password_123",
	"redis_host":      "localhost",
	"redis_port":      "6379",
	"redis_db":        "0",
	"kafka_broker":    "localhost:9092",
	"kafka_topic":     "trades",
	"grpc_port":       "50051",
	"http_port":       "8080",
	"https_port":      "8443",

	// Произвольные переменные (20+)
	"exchange":          "Binance",
	"exchange_full":     "Binance Exchange",
	"exchange_fee":      "0.075",
	"exchange_status":   "online",
	"order_id":          "ORD-12345-67890",
	"order_type":        "LIMIT",
	"order_side":        "BUY",
	"order_status":      "FILLED",
	"order_qty":         "0.5",
	"order_price":       "45100.00",
	"order_total":       "22550.00",
	"position_id":       "POS-98765",
	"position_size":     "1.5",
	"position_side":     "LONG",
	"position_pnl":      "+1250.50",
	"alert_id":          "ALERT-001",
	"alert_type":        "PRICE",
	"alert_condition":   "price > 50000",
	"notification_id":   "NOTIF-123",
	"notification_type": "email",
	"template_id":       "TMPL-456",
	"template_name":     "price_alert",
}

var promptRes = `
========================================
        ТОРГОВЫЙ ОТЧЕТ crypto_trader_99
========================================

Дата и время: 2026-02-22 15:30:45 (UTC+3)
Пользователь: crypto_trader_99 (ID: 12345)
Уровень: VIP | Рейтинг: 5.0
Страна: Россия, Москва

----------------------------------------
ТЕКУЩАЯ ТОРГОВАЯ ПАРА: BTCUSDT
----------------------------------------

Текущая цена: 45000.50 USDT
Изменения: 1ч: 0.5% | 24ч: -2.3% | 7д: 5.6% | 30д: 15.8%
24ч максимум: 46000.00 | 24ч минимум: 44500.00
Объем 24ч: 1250000_24h | Рыночная капитализация: 890000000

----------------------------------------
ТЕХНИЧЕСКИЕ ИНДИКАТОРЫ
----------------------------------------

EMA (9/20/50/100/200): 45123.45 / 44987.23 / 44123.67 / 43245.89 / 44987.230
SMA (9/20/50/100/200): 45098.76 / 44965.43 / 44112.34 / 43234.78 / 44965.430
Полосы Боллинджера (верх/сред/ниж): 46234.56 / 44987.23 / 43739.90
RSI (14): 62.5
Stochastic (K/D): 78.9 / 75.6
MACD (линия/сигнал/гист): 125.6_line / 125.6_signal / 125.6_hist
ADX: 35.6 | DMI (+/-): 25.4 / 18.7
Ишимоку (тенкан/киджун/сенкоу): 45000 / 44800 / 45200
ATR: 450.25 | VWAP: 45100.00 | OBV: 12500000

----------------------------------------
ДРУГИЕ ТОРГОВЫЕ ПАРЫ
----------------------------------------

ETH: ETHUSDT - цена 45000.50_eth USDT
SOL: BTCUSDT_sol - цена 98.45 USDT
ADA: BTCUSDT_ada - цена 45000.50_ada USDT
DOT: BTCUSDT_dot - цена 45000.50_dot USDT
MATIC: BTCUSDT_matic - цена 0.95 USDT
LINK: BTCUSDT_link - цена 45000.50_link USDT
AVAX: AVAXUSDT - цена 45000.50_avax USDT
XRP: XRPUSDT - цена 0.62 USDT
DOGE: DOGEUSDT - цена 45000.50_doge USDT
SHIB: BTCUSDT_shib - цена 45000.50_shib USDT

----------------------------------------
СТАТИСТИКА ПОЛЬЗОВАТЕЛЯ
----------------------------------------

Баланс: 100000 USDT
Всего сделок: 1250 | Винрейт: 67.5%
Дата регистрации: 2024-05-15 | Последний вход: 2026-02-21
Статус: Активный | Риск-профиль: Средний
Стратегия: Долгосрочный | Любимая пара: BTC/USDT

----------------------------------------
ФИНАНСОВЫЕ ПОКАЗАТЕЛИ
----------------------------------------

Прибыль сегодня: 12500.50 USDT (12500.50_percent%)
Убыток сегодня: -2500.75 USDT (-2500.75_percent%)
Общая прибыль: 150000.00 USDT
Общий убыток: -45000.00 USDT
Чистая прибыль: 105000.00 USDT
ROI: 23.4% | Коэффициент Шарпа: 1.85
Макс. просадка: -12.5% (-12.5_usd USDT)
Фактор прибыли: 2.35
Средний выигрыш: 850.25 | Средний проигрыш: -425.50
Всего сделок: 1250 (выигрышных: 856, проигрышных: 394)
Общий объем: 1250000_total USDT
Комиссия: 1250.75 USDT (1250.75_rate%)
Проскальзывание: 5.25

----------------------------------------
ПОЗИЦИИ И ОРДЕРА
----------------------------------------

Текущая позиция: LONG | Размер: 1.5
P&L позиции: +1250.50
Последний ордер: ORD-12345-67890 | Тип: LIMIT | Сторона: BUY
Статус: FILLED | Количество: 0.5 | Цена: 45100.00
Итого: 22550.00 USDT

----------------------------------------
ИНФОРМАЦИЯ О БИРЖЕ
----------------------------------------

Биржа: Binance (Binance_full)
Комиссия биржи: 0.075% | Статус: Binance_status
Ставка финансирования: 0.01% | Открытый интерес: 450000000
Long/Short соотношение: 65.5% / 34.5%

----------------------------------------
СИСТЕМНАЯ ИНФОРМАЦИЯ
----------------------------------------

API версия: v2 | Endpoint: https://api.binance.com
WebSocket: wss://stream.binance.com
Timeout: 30 сек | Retry: 3 (задержка 5 сек)
Rate limit: 1200 запросов / 60 сек
Лог-файл: /var/log/trader.log | Уровень логирования: INFO

База данных: localhost:5432/trading_db
Пользователь БД: trader
Redis: localhost:6379 (БД: 0)
Kafka: localhost:9092 (топик: trades)
HTTP порт: 8080 | HTTPS порт: 8443 | gRPC порт: 50051

----------------------------------------
УВЕДОМЛЕНИЯ И АЛЕРТЫ
----------------------------------------

Алерт: ALERT-001 | Тип: PRICE
Условие: price > 50000
Уведомление: NOTIF-123 | Тип: email
Шаблон: TMPL-456 (price_alert)

----------------------------------------
КОНТАКТНАЯ ИНФОРМАЦИЯ
----------------------------------------

Email: trader@example.com | Телефон: +7-999-123-45-67
Telegram: @crypto_trader | Discord: trader#1234
Twitter: @crypto_trader | Instagram: crypto_trader_inst

----------------------------------------
ДОПОЛНИТЕЛЬНАЯ ИНФОРМАЦИЯ
----------------------------------------

Дата создания отчета: 2026-02-22 15:30:45
Временная метка: 15:30:45stamp (1740231045)
Текущий квартал: Q1 | Неделя: 8
День года: 22_of_year
Начало года: 2026-01-01 | Конец года: 2026-02-22_end
Вчера: 2026-02-22_prev | Завтра: 2026-02-22_next

========================================
        КОНЕЦ ОТЧЕТА
========================================

P.S. Это тестовый отчет для демонстрации замены множества переменных.
Переменные могут использоваться многократно, например: BTCUSDT, BTCUSDT, BTCUSDT
Или в разных контекстах: 45000.50, 45000.50, 45000.50
А также с другими парами: ETHUSDT, BTCUSDT_sol, BTCUSDT_ada

Проверка повторного использования переменных:
crypto_trader_99 (первый раз)
crypto_trader_99 (второй раз) 
crypto_trader_99 (третий раз)
trader@example.com (первый раз)
trader@example.com (второй раз)
BTCUSDT (первый раз)
BTCUSDT (второй раз)
45000.50 (первый раз)
45000.50 (второй раз)

Общее количество переменных в этом тексте: более 200 уникальных переменных,
каждая из которых будет заменена соответствующим значением.
Функция SubstituteVarsIntoPrompt должна обработать все эти замены корректно.
`
