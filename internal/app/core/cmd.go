package core

import (
	"context"
	"fmt"
	"nlkli/raytrade/internal/broker"
	"nlkli/raytrade/internal/cdl"
	"strconv"
	"strings"
	"sync/atomic"
)

type CMD struct {
	Tx   chan string
	slot atomic.Pointer[CommitFn]
}

func InitCMD(ctx context.Context) *CMD {
	c := &CMD{
		Tx: make(chan string, 32),
	}

	go func() {
		for {
			select {
			case p := <-c.Tx:
				f := c.translate(p)
				c.slot.Store(&f)
			case <-ctx.Done():
				return
			}
		}
	}()

	return c
}

func (c *CMD) Update(s *State) {
	if ptr := c.slot.Swap(nil); ptr != nil {
		(*ptr)(s)
	}
}

func (c *CMD) translate(prompt string) CommitFn {

	parseFloatValue := func(v string, f *float64) (err error) {
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

	if len(prompt) == 0 {
		return CommitCommandLineError("empty command")
	}

	var commands []CommitFn

	parts := strings.SplitSeq(prompt, "|")

	// var intervalArg *cdl.Interval

	for part := range parts {

		part = strings.TrimSpace(part)

		args := strings.Split(part, " ")
		n := len(args)

		var command CommitFn

		switch args[0] {

		case "sub":

			if n <= 3 {
				break
			}

			comp := args[1]
			idx, err := strconv.Atoi(args[2])
			if err != nil {
				break
			}
			params := strings.Split(args[3], ".")

			switch comp {
			case "chart":
				if len(params) < 3 {
					break
				}

				category, err := broker.CategoryFromString(params[0])

				if err != nil {
					break
				}

				symbol := params[1]

				interval, err := cdl.IntervalFromString(params[2])

				if err != nil {
					break
				}

				command = func(s *State) {
					s.BTX <- SubChart{
						Idx:      idx,
						Category: category,
						Symbol:   symbol,
						Interval: interval,
						Limit:    200,
					}
				}
			case "orderbook":
				if len(params) < 3 {
					break
				}

				category, err := broker.CategoryFromString(params[0])

				if err != nil {
					break
				}

				symbol := params[1]

				depth, err := strconv.Atoi(params[2])

				if err != nil {
					break
				}

				command = func(s *State) {
					s.BTX <- SubOrderBook{
						Idx:      idx,
						Category: category,
						Symbol:   symbol,
						Depth:    depth,
					}
				}
			}

			// 	case "symbol", "s":

			// 		var symbol string
			// 		if n > 1 {
			// 			symbol = strings.ToUpper(args[1])
			// 		}

			// 		command = func(s *State) {
			// 			if s.StatusLine.Symbol == "..." {
			// 				CommitCommandLineError("TODO")(s)
			// 			}

			// 			interval := intervalArg
			// 			if interval == nil {
			// 				res, err := cdl.IntervalFromString(s.StatusLine.Interval)
			// 				if err != nil {
			// 					CommitCommandLineError(err.Error())(s)
			// 				}
			// 				interval = &res
			// 			}

			// 			if len(symbol) == 0 {
			// 				symbol = s.StatusLine.Symbol
			// 			}

			// 			if len(symbol) == 0 {
			// 				CommitCommandLineError("TODO")(s)
			// 			}

			// 			s.StatusLine.Symbol = "..."

			// 			limit := s.Chart[0].Cap * 2
			// 			if limit == 0 {
			// 				limit = 200
			// 			}

			// 			s.BTX <- InstrumentObserverT{
			// 				Category:         broker.Futures,
			// 				Symbol:           symbol,
			// 				Interval:         *interval,
			// 				InitCandlesLimit: limit,
			// 			}
			// 		}

		// case "interval", "i":

		// 	if n == 1 {
		// 		break
		// 	}

		// 	res, err := cdl.IntervalFromString(args[1])
		// 	if err != nil {
		// 		return CommitCommandLineError(err.Error())
		// 	}

		// 	// intervalArg = &res

		case "reset", "r":

			if n == 1 {
				break
			}

			switch args[1] {

			// case "ut":
			// 	tn := time.Now()

			// 	command = func(s *State) error {
			// 		s.ST = tn
			// 		return nil
			// 	}

			case "scalex", "sx":

				if n == 1 {
					break
				}

				command = func(s *State) {
					s.Chart[0].Forced = true
					s.Chart[0].Scale.X = 1
				}

			case "scaley", "sy":

				if n == 1 {
					break
				}

				command = func(s *State) {
					s.Chart[0].Forced = true
					s.Chart[0].Scale.Y = 1
				}

			case "targetx", "tx":

				if n == 1 {
					break
				}

				command = func(s *State) {
					s.Chart[0].Forced = true
					s.Chart[0].Shift.X = 1
				}

			case "targety", "ty":

				if n == 1 {
					break
				}

				command = func(s *State) {
					s.Chart[0].Forced = true
					s.Chart[0].Shift.Y = 1
				}

			case "rowheight", "rh":

				if n == 1 {
					break
				}

				command = func(s *State) {
					s.SetRH(20) // TODO
				}

			default:
			}

		case "scalex", "sx":

			if n == 1 {
				break
			}

			command = func(s *State) {
				f := float64(s.Chart[0].Scale.X)
				if err := parseFloatValue(args[1], &f); err != nil {
					CommitCommandLineError(err.Error())(s)
				}
				s.Chart[0].Scale.X = float32(f)
				s.Chart[0].Forced = true
			}

		case "scaley", "sy":

			if n == 1 {
				break
			}

			command = func(s *State) {
				f := float64(s.Chart[0].Scale.Y)
				if err := parseFloatValue(args[1], &f); err != nil {
					CommitCommandLineError(err.Error())(s)
				}
				s.Chart[0].Scale.Y = float32(f)
				s.Chart[0].Forced = true
			}

		case "targetx", "tx":

			if n == 1 {
				break
			}

			command = func(s *State) {
				f := float64(s.Chart[0].Shift.X)
				if err := parseFloatValue(args[1], &f); err != nil {
					CommitCommandLineError(err.Error())(s)
				}
				s.Chart[0].Shift.X = float32(f)
				s.Chart[0].Forced = true
			}

		case "targety", "ty":

			if n == 1 {
				break
			}

			command = func(s *State) {
				f := float64(s.Chart[0].Shift.Y)
				if err := parseFloatValue(args[1], &f); err != nil {
					CommitCommandLineError(err.Error())(s)
				}
				s.Chart[0].Shift.Y = float32(f)
				s.Chart[0].Forced = true
			}

		case "rowheight", "rh":

			if n == 1 {
				break
			}

			command = func(s *State) {
				f := float64(s.RH)
				if err := parseFloatValue(args[1], &f); err != nil {
					CommitCommandLineError(err.Error())(s)
				}

				s.SetRH(float32(f))
			}

		case "fps":

			command = func(s *State) {
				s.ShowFPS = !s.ShowFPS
			}

		case "overlay":

			command = func(s *State) {
				s.ShowOverlay = !s.ShowOverlay
			}

		default:
		}

		if command != nil {
			commands = append(commands, command)
		}
	}

	if len(commands) == 1 {
		return commands[0]
	}
	if len(commands) > 1 {
		return func(s *State) {
			for _, c := range commands {
				c(s)
			}
		}
	}

	return CommitCommandLineError(fmt.Sprintf("unknown command: %s", prompt))
}
