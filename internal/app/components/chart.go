package components

import "nlkli/raytrade/internal/app/state"

type Chart struct {
	*Rect

	pb *PriceBar
	tl *TimeLine
}

func (c *Chart) Render(s *state.State) state.Action {
	return nil
}

type PriceBar struct {
	*Rect
}

func (pb *PriceBar) Render(s *state.State) state.Action {
	return nil
}

type TimeLine struct {
	*Rect
}

func (tl *TimeLine) Render(s *state.State) state.Action {
	return nil
}
