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

	Slot atomic.Pointer[CommitFn]
}

func InitCMD(ctx context.Context, c *Config) *CMD {

	cmd := &CMD{
		Tx:   make(chan string, 32),
		Vars: c.Vars,
	}

	cmd.updateReplacer()

	go func() {
		for {
			select {
			case p := <-cmd.Tx:
				f := cmd.processing(p)
				cmd.Slot.Store(&f)
			case <-ctx.Done():
				return
			}
		}
	}()

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

func (c *CMD) processing(prompt string) CommitFn {

	var commit CommitFn

	for command := range strings.SplitSeq(prompt, "|") {

		command = c.Replacer.Replace(command)
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
	case "set":
		varName, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing argument for command '%s'", head)
		}

		varValue, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing argument for command '%s'", head)
		}

		c.Vars[varName] = varValue
		c.updateReplacer()

		output := fmt.Sprintf("%s=%s", varName, varValue)

		return func(s *State) {
			s.CommandLine.Prompt = output
		}, nil

	case "read":
		varName, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing argument for command '%s'", head)
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
			return nil, fmt.Errorf("missing argument for command '%s'", head)
		}

		return func(s *State) {

			rh := float64(s.RH)
			if err := parseFloatValue(newRH, &rh); err != nil {
				CommitCommandLineError(err.Error())(s)
			}

			s.RH = max(2, float32(rh))
			s.RH_Dirty = true

			const NUMBERS = "1234567890"

			s.TextNumSV = rl.MeasureTextEx(s.F, NUMBERS, float32(s.F.BaseSize), 0)
			s.TextNumSV.X = s.TextNumSV.X / float32(len(NUMBERS))

			s.TextDotW = rl.MeasureTextEx(
				s.F, ".", float32(s.F.BaseSize), 0,
			).X

		}, nil

	case "sub":
		compType, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing argument for command '%s'", head)
		}

		compIdxV, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing argument for command '%s'", head)
		}

		compIdx, err := strconv.Atoi(compIdxV)
		if err != nil {
			return nil, fmt.Errorf("type error")

		}

		paramsV, ok := next()
		if !ok {
			return nil, fmt.Errorf("missing argument for command '%s'", head)
		}

		paramsIter := strings.SplitSeq(paramsV, ".")
		nextP, stopP := iter.Pull(paramsIter)
		defer stopP()

		switch compType {

		case "chart":
			categoryV, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing argument for command '%s'", head)
			}

			category, err := broker.CategoryFromString(categoryV)
			if err != nil {
				return nil, fmt.Errorf("category type error")
			}

			symbol, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing argument for command '%s'", head)
			}

			intervalV, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing argument for command '%s'", head)
			}

			interval, err := cdl.IntervalFromString(intervalV)
			if err != nil {
				return nil, fmt.Errorf("interval type error")
			}

			limitV, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing argument for command '%s'", head)
			}

			limit, err := strconv.Atoi(limitV)
			if err != nil {
				return nil, fmt.Errorf("limit type error")
			}

			c.BTX <- SubChart{
				Idx:      compIdx,
				Category: category,
				Symbol:   symbol,
				Interval: interval,
				Limit:    limit,
			}
		case "orderbook":
			categoryV, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing argument for command '%s'", head)
			}

			category, err := broker.CategoryFromString(categoryV)
			if err != nil {
				return nil, fmt.Errorf("category type error")
			}

			symbol, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing argument for command '%s'", head)
			}

			depthV, ok := nextP()
			if !ok {
				return nil, fmt.Errorf("missing argument for command '%s'", head)
			}

			depth, err := strconv.Atoi(depthV)
			if err != nil {
				return nil, fmt.Errorf("limit type error")
			}

			c.BTX <- SubOrderBook{
				Idx:      compIdx,
				Category: category,
				Symbol:   symbol,
				Depth:    depth,
			}
		default:
			return nil, fmt.Errorf("unknown component type: %s", compType)
		}

		return func(s *State) {}, nil

	case "overlay":
		return func(s *State) {
			s.ShowOverlay = !s.ShowOverlay
		}, nil

	}

	return nil, fmt.Errorf("unknown command: %s", head)
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
