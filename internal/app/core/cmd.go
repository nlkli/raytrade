package core

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func CommitCommandLineError(text string) CommitFn {
	return func(s *State) {
		s.CommandLine.Prompt = text
		s.CommandLine.Color = s.P.Base.Red
	}
}

func CommitCommandLineErrorAnd(text string, f CommitFn) CommitFn {
	return func(s *State) {
		CommitCommandLineError(text)(s)
		f(s)
	}
}

func MergeCommits(a CommitFn, b CommitFn) CommitFn {
	return func(s *State) {
		a(s)
		b(s)
	}
}

type CMD struct {
	Tx chan string

	BTX chan<- Task // Backgorund tx

	Vars     map[string]string
	Replacer *strings.Replacer

	// GlobalVars   map[string]string
	// GlobReplacer *strings.Replacer

	InitCommands []string

	Slot atomic.Pointer[CommitFn]
}

func InitCMD(ctx context.Context, c *Config) *CMD {

	cmd := &CMD{
		Tx:           make(chan string, 32),
		Vars:         c.Vars,
		InitCommands: c.InitCommands,
	}

	cmd.updateReplacer()
	// cmd.updateGlobReplacer()

	go func() {
		for {
			select {
			case p := <-cmd.Tx:
				if len(p) == 0 {
					continue
				}

				f := cmd.processing(p)
				if f != nil {
					cmd.Slot.Store(&f)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	if len(c.InitCommands) > 0 {
		go func() {
			time.Sleep(200 * time.Millisecond)
			cmd.Tx <- "init"
		}()
	}

	return cmd
}

func (c *CMD) Update(s *State) {
	if ptr := c.Slot.Swap(nil); ptr != nil {
		(*ptr)(s)
	}
}

func (c *CMD) updateReplacer() {
	vars := make([]string, 0, len(c.Vars)*2)
	for k, v := range c.Vars {
		vars = append(vars, "$"+k, v)
	}

	c.Replacer = strings.NewReplacer(vars...)
}

// func (c *CMD) updateGlobReplacer() {
// 	vars := make([]string, 0, len(c.GlobalVars)*2)
// 	for k, v := range c.GlobalVars {
// 		vars = append(vars, "$"+k, v)
// 	}
//
// 	c.GlobReplacer = strings.NewReplacer(vars...)
// }

func (c *CMD) processing(prompt string) CommitFn {

	var commit CommitFn

	for command := range strings.SplitSeq(prompt, "|") {

		// command = c.GlobReplacer.Replace(command)

		command = c.Replacer.Replace(command)
		if strings.ContainsRune(command, '$') {
			command = c.Replacer.Replace(command)
		}

		command = strings.TrimSpace(command)

		args := strings.SplitSeq(command, " ")

		c, err := c.translateV2(args)
		if err != nil {
			return CommitCommandLineError(err.Error())
		}

		if c == nil {
			continue
		}

		if commit == nil {
			commit = c
			continue
		}

		MergeCommits(commit, c)
	}

	return commit
}

func (c *CMD) translateV2(args iter.Seq[string]) (CommitFn, error) {
	next, stop := iter.Pull(args)
	defer stop()

	head, ok := next()
	if !ok {
		return nil, errors.New("empty command")
	}

	switch head {
	case "init":
		if len(c.InitCommands) > 0 {
			c.Tx <- strings.Join(c.InitCommands, "|")
		}

	// case "gset":

	// 	varName, ok := next()
	// 	if !ok {
	// 		return nil, fmt.Errorf("missing variable name for 'gset' command")
	// 	}

	// 	if _, ok := c.GlobalVars[varName]; !ok {
	// 		return nil, fmt.Errorf("unknown global variable: %s", varName)
	// 	}

	// 	varValue, ok := next()
	// 	if !ok {
	// 		return nil, fmt.Errorf("missing value for variable '%s'", varName)
	// 	}

	// 	c.GlobalVars[varName] = varValue
	// 	c.updateGlobReplacer()

	// 	return nil, nil

	case "set":
		varName, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing variable name for 'set' command")
		}

		varValue, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for variable '%s'", varName)
		}

		c.Vars[varName] = varValue
		c.updateReplacer()

		return nil, nil

	case "read":
		varName, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing variable name for 'read' command")
		}

		varValue, ok := c.Vars[varName]
		if !ok {
			varValue = "nil"
		}

		output := fmt.Sprintf("%s=%s", varName, varValue)

		return func(s *State) {
			s.CommandLine.Prompt = output
		}, nil

	case "rh":
		newRH, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for 'rh' command")
		}

		return func(s *State) {

			rh := float64(s.RH)
			if err := parseFloatValue(newRH, &rh); err != nil {
				CommitCommandLineError(err.Error())(s)
			}

			s.ApplyNewRH(float32(rh))

		}, nil

	case "prompt":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for 'prompt' command")
		}

		var prompt strings.Builder
		prompt.WriteRune(':')
		for ok {
			prompt.WriteString(value)
			value, ok = next()
			if ok {
				prompt.WriteString(" ")
			}
		}

		return func(s *State) {
			s.M = Input
			s.CommandLine.Prompt = prompt.String()
			s.CommandLine.PromptW = rl.MeasureTextEx(
				s.F,
				s.CommandLine.Prompt,
				s.RH,
				0,
			).X
			s.CommandLine.Color = s.P.Fg[1]
		}, nil

	case "porder":

		return c.translatePlaceOrder(next)

	case "corder":

		return c.translateCancelOrder(next)

	case "sub":

		return c.translateSubCommand(next)

	case "chart":

		return c.translateChartCommands(next)

	case "orderbook":

		return c.translateOrderBookCommands(next)

	case "overlay":
		return func(s *State) {
			s.ShowOverlay = !s.ShowOverlay
		}, nil

	}

	return nil, fmt.Errorf("unknown command: %s", head)
}

func (c *CMD) translateCancelOrder(next func() (string, bool)) (CommitFn, error) {
	paramsV, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing cancel order parameters")
	}

	paramsIter := strings.SplitSeq(paramsV, ",")
	nextP, stopP := iter.Pull(paramsIter)
	defer stopP()

	orderByV, ok := nextP()
	if !ok {
		return nil, fmt.Errorf("missing order by")
	}

	var orderBy OrderBy
	var orderByValue any

	switch orderByV {
	case "0", "I", "Idx", "Index":
		orderBy = OrderIndex
		idxV, ok := nextP()
		if !ok {
			return nil, fmt.Errorf("missing order index")
		}
		idx, err := strconv.Atoi(idxV)
		if err != nil {
			return nil, fmt.Errorf("invalid order index: %s", idxV)
		}
		orderByValue = idx

	case "1", "S", "Selected":
		orderBy = SelectedOrder

	case "2", "F", "First":
		orderBy = FirstOrder

	case "3", "L", "Last":
		orderBy = LastOrder
	}

	c.BTX <- &CancelOrder{
		OrderBy:      orderBy,
		OrderByValue: orderByValue,
	}

	return nil, nil
}

// porder L F,BRCUSDT,0,10,1,0,70000
func (c *CMD) translatePlaceOrder(next func() (string, bool)) (CommitFn, error) {
	orderType, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing order type")
	}

	paramsV, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing order parameters")
	}

	paramsIter := strings.SplitSeq(paramsV, ",")
	nextP, stopP := iter.Pull(paramsIter)
	defer stopP()

	switch orderType {
	case "0", "L", "Limit":
		category, symbol, err := parseInstrumentFromParams(nextP)
		if err != nil {
			return nil, err
		}

		sideV, ok := nextP()
		if !ok {
			return nil, fmt.Errorf("missing side")
		}

		var side broker.Side
		switch sideV {
		case "0", "L", "Long", "Buy":
			side = broker.Long
		case "1", "S", "Short", "Sell":
			side = broker.Short
		default:
			return nil, fmt.Errorf("invalid side: %s", sideV)
		}

		qtyV, ok := nextP()
		if !ok {
			return nil, fmt.Errorf("missing quantity")
		}

		qty, err := strconv.ParseFloat(qtyV, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid quantity: %s", qtyV)
		}

		marketUnitV, ok := nextP()
		if !ok {
			return nil, fmt.Errorf("missing market unit")
		}

		var marketUnit broker.MarketUnit
		switch marketUnitV {
		case "0", "BC", "BaseCoin":
			marketUnit = broker.BaseCoin
		case "1", "QC", "QuoteCoin":
			marketUnit = broker.QuoteCoin
		default:
			return nil, fmt.Errorf("invalid market unit: %s", marketUnitV)
		}

		orderPriceDataV, ok := nextP()
		if !ok {
			return nil, fmt.Errorf("missing price data type")
		}

		var priceBy OrderPriceBy
		var priceByValue any

		switch orderPriceDataV {
		case "0", "P", "Price":
			priceBy = Price
			priceStr, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing price value")
			}
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid price: %s", priceStr)
			}
			priceByValue = price

		case "1", "SP", "SelectedPrice":
			priceBy = SelectedPrice

		case "2", "BP", "BidPrice":
			priceBy = BidPrice
			idxV, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing bid index")
			}
			idx, err := strconv.Atoi(idxV)
			if err != nil {
				return nil, fmt.Errorf("invalid bid index: %s", idxV)
			}
			priceByValue = idx

		case "3", "AP", "AskPrice":
			priceBy = AskPrice
			idxV, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing ask index")
			}
			idx, err := strconv.Atoi(idxV)
			if err != nil {
				return nil, fmt.Errorf("invalid ask index: %s", idxV)
			}
			priceByValue = idx

		default:
			return nil, fmt.Errorf("invalid price data type: %s", orderPriceDataV)
		}

		c.BTX <- &PlaceLimitOrder{
			Category:     category,
			Symbol:       symbol,
			Side:         side,
			Qty:          qty,
			MarketUnit:   marketUnit,
			PriceBy:      priceBy,
			PriceByValue: priceByValue,
		}

		return nil, nil

	default:
		return nil, fmt.Errorf("invalid order type: %s", orderType)
	}
}

func (c *CMD) translateSubCommand(next func() (string, bool)) (CommitFn, error) {
	compType, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing component type for 'sub' command")
	}

	if compType == "position" ||
		compType == "order" ||
		compType == "execution" {

		var filter []InstrumentFilter

		filterV, ok := next()
		if ok {
			for f := range strings.SplitSeq(filterV, ",") {
				parts := strings.SplitN(f, ".", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid position filter")
				}

				category, err := broker.CategoryFromString(parts[0])
				if err != nil {
					return nil, err
				}

				symbol := parts[1]

				filter = append(filter, InstrumentFilter{
					Category: category,
					Symbol:   symbol,
				})
			}
		}

		if compType == "position" {
			c.BTX <- &SubPosition{
				Filter: filter,
			}
		}

		if compType == "order" {
			c.BTX <- &SubOrder{
				Filter: filter,
			}
		}

		if compType == "execution" {
			c.BTX <- &SubExecution{
				Filter: filter,
				Limit:  200,
			}
		}

		return nil, nil
	}

	compIdxV, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing component index for 'sub' command")
	}

	compIdx, err := strconv.Atoi(compIdxV)
	if err != nil {
		return nil, fmt.Errorf("invalid component index: %s", compIdxV)

	}

	paramsV, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing parameters for '%s' component", compType)
	}

	paramsIter := strings.SplitSeq(paramsV, ",")
	nextP, stopP := iter.Pull(paramsIter)
	defer stopP()

	switch compType {

	case "chart":
		category, symbol, err := parseInstrumentFromParams(nextP)
		if err != nil {
			return nil, err
		}

		intervalV, ok := nextP()
		if !ok {
			return nil, fmt.Errorf("missing interval for chart component")
		}

		interval, err := cdl.IntervalFromString(intervalV)
		if err != nil {
			return nil, fmt.Errorf("invalid interval: %s", intervalV)
		}

		c.BTX <- &SubChart{
			Idx:      compIdx,
			Category: category,
			Symbol:   symbol,
			Interval: interval,
		}
	case "orderbook":

		category, symbol, err := parseInstrumentFromParams(nextP)
		if err != nil {
			return nil, err
		}

		depthV, ok := nextP()
		if !ok {
			return nil, fmt.Errorf("missing depth for orderbook component")
		}

		depth, err := strconv.Atoi(depthV)
		if err != nil {
			return nil, fmt.Errorf("invalid depth value: %s", depthV)
		}

		c.BTX <- &SubOrderBook{
			Idx:      compIdx,
			Category: category,
			Symbol:   symbol,
			Depth:    depth,
		}
	default:
		return nil, fmt.Errorf("unknown component type: %s", compType)
	}

	return nil, nil
}

func (c *CMD) translateOrderBookCommands(next func() (string, bool)) (CommitFn, error) {

	idxV, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing orderbook index for 'chart' command")
	}

	idx, err := strconv.Atoi(idxV)
	if err != nil {
		return nil, fmt.Errorf("invalid chart orderbook: %s", idxV)
	}

	param, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing parameter for orderbook %d", idx)
	}

	switch param {
	case "rhd":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for orderbook %d parameter '%s'", idx, param)
		}

		return func(s *State) {
			obS := s.OrderBook[idx]
			rhd := float64(obS.RHD)
			if err := parseFloatValue(value, &rhd); err != nil {
				CommitCommandLineError(err.Error())(s)
			}
			obS.RHD = float32(rhd)
			obS.Forced = true
		}, nil

	case "vm":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for orderbook %d parameter '%s'", idx, param)
		}
		return func(s *State) {
			obS := s.OrderBook[idx]
			switch value {
			case "1":
				obS.VM = 1
			default:
				obS.VM = 0
			}
			obS.Forced = true
		}, nil

	default:
		return nil, fmt.Errorf("unknown orderbook parameter: %s", param)
	}
}

func (c *CMD) translateChartCommands(next func() (string, bool)) (CommitFn, error) {

	idxV, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing chart index for 'chart' command")
	}

	idx, err := strconv.Atoi(idxV)
	if err != nil {
		return nil, fmt.Errorf("invalid chart index: %s", idxV)
	}

	param, ok := next()
	if !ok {
		return nil, fmt.Errorf("missing parameter for chart %d", idx)
	}

	switch param {
	case "rhd":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for chart %d parameter '%s'", idx, param)
		}

		return func(s *State) {
			cs := s.Chart[idx]
			rhd := float64(cs.RHD)
			if err := parseFloatValue(value, &rhd); err != nil {
				CommitCommandLineError(err.Error())(s)
			}
			cs.RHD = float32(rhd)
			cs.Forced = true
		}, nil

	case "sx":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for chart %d parameter '%s'", idx, param)
		}

		return func(s *State) {
			cs := s.Chart[idx]
			sx := float64(cs.Scale.X)
			if err := parseFloatValue(value, &sx); err != nil {
				CommitCommandLineError(err.Error())(s)
			}
			cs.Scale.X = max(0.00001, float32(sx))
			cs.Forced = true
		}, nil

	case "sy":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for chart %d parameter '%s'", idx, param)
		}

		return func(s *State) {
			cs := s.Chart[idx]
			sy := float64(cs.Scale.Y)
			if err := parseFloatValue(value, &sy); err != nil {
				CommitCommandLineError(err.Error())(s)
			}
			cs.Scale.Y = max(0.00001, float32(sy))
			cs.Forced = true
		}, nil
	case "tx":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for chart %d parameter '%s'", idx, param)
		}

		return func(s *State) {
			cs := s.Chart[idx]
			tx := float64(cs.Shift.X)
			if err := parseFloatValue(value, &tx); err != nil {
				CommitCommandLineError(err.Error())(s)
			}
			cs.Shift.X = float32(tx)
			cs.Forced = true
		}, nil
	case "ty":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for chart %d parameter '%s'", idx, param)
		}

		return func(s *State) {
			cs := s.Chart[idx]
			ty := float64(cs.Shift.Y)
			if err := parseFloatValue(value, &ty); err != nil {
				CommitCommandLineError(err.Error())(s)
			}
			cs.Shift.Y = float32(ty)
			cs.Forced = true
		}, nil
	case "cg": // candle gap
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for chart %d parameter '%s'", idx, param)
		}

		return func(s *State) {
			cs := s.Chart[idx]
			cg := float64(cs.CG)
			if err := parseFloatValue(value, &cg); err != nil {
				CommitCommandLineError(err.Error())(s)
			}
			cs.CG = float32(cg)
		}, nil
	case "uline":
		return func(s *State) {
			cs := s.Chart[idx]
			n := len(cs.Lines)
			if n > 0 {
				cs.Lines = cs.Lines[:n-1]
			}
		}, nil
	case "clines":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.Lines = cs.Lines[:0]
		}, nil
	case "alevel":
		value, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing value for chart %d parameter '%s'", idx, param)
		}

		levelPrice, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid level price %s", value)
		}

		return func(s *State) {
			cs := s.Chart[idx]
			cs.Levels = append(cs.Levels, levelPrice)
		}, nil
	case "ulevel":
		return func(s *State) {
			cs := s.Chart[idx]
			n := len(cs.Levels)
			if n > 0 {
				cs.Levels = cs.Levels[:n-1]
			}
		}, nil
	case "clevels":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.Levels = cs.Levels[:0]
		}, nil
	case "show_execution":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.ShowExecution = !cs.ShowExecution
		}, nil
	case "show_position":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.ShowPosition = !cs.ShowPosition
		}, nil
	case "show_order":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.ShowOrder = !cs.ShowOrder
		}, nil
	case "show_lable":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.ShowLable = !cs.ShowLable
		}, nil
	case "show_grid":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.ShowGrid = !cs.ShowGrid
		}, nil
	case "show_price_bar":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.ShowPriceBar = !cs.ShowPriceBar
		}, nil
	case "show_time_line":
		return func(s *State) {
			cs := s.Chart[idx]
			cs.ShowTimeLine = !cs.ShowTimeLine
		}, nil

	default:
		return nil, fmt.Errorf("unknown chart parameter: %s", param)
	}
}

func parseInstrumentFromParams(nextP func() (string, bool)) (broker.Category, string, error) {
	categoryV, ok := nextP()
	if !ok {
		return -1, "", fmt.Errorf("missing category")
	}

	category, err := broker.CategoryFromString(categoryV)
	if err != nil {
		return -1, "", fmt.Errorf("invalid category: %s", categoryV)
	}

	symbol, ok := nextP()
	if !ok {
		return -1, "", fmt.Errorf("missing symbol")
	}

	return category, symbol, nil
}

func parseFloatValue(v string, f *float64) (err error) {
	switch rune(v[0]) {
	case '+':
		pf, err := strconv.ParseFloat(v[1:], 64)
		if err != nil {
			return err
		}
		*f += pf
	case '-':
		pf, err := strconv.ParseFloat(v[1:], 64)
		if err != nil {
			return err
		}
		*f -= pf
	case '=':
		*f, err = strconv.ParseFloat(v[1:], 64)
	default:
		*f, err = strconv.ParseFloat(v, 64)
	}
	return err
}
