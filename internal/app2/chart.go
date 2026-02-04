package app2

type Chart struct {
	*Rect
	parent *Rect

	c  *Canvas
	tl *TimeLine
	pb *PriceBar
	cr *Crossing
}

func (ch *Chart) Render(s *State) {
	if s.WRF {
		ch.MoveTo(ch.parent.p.X, ch.parent.p.Y)
		ch.SetSize(ch.parent.s.X-OBW, ch.parent.s.Y)
	}

	ch.tl.Render(s)
	ch.pb.Render(s)
	ch.cr.Render(s)
	ch.c.Render(s)

	// ch.Outline(1, s.P.Base.Orange)
}

type Canvas struct {
	*Rect
	parent *Rect
}

func (c *Canvas) Render(s *State) {
	if s.WRF {
		c.MoveTo(c.parent.p.X, c.parent.p.Y)
		c.SetSize(c.parent.s.X-PBW-RPD, c.parent.s.Y-TLH)
	}

	// c.Fill(s.P.Bg[1])

	c.Outline(1, s.P.Base.Blue)
}

type TimeLine struct {
	*Rect
	parent *Rect
}

func (tl *TimeLine) Render(s *State) {
	if s.WRF {
		tl.MoveTo(tl.parent.p.X, tl.parent.s.Y-TLH+RPD)
		tl.SetSize(tl.parent.s.X-PBW-RPD, TLH)
	}

	tl.Fill(s.P.Bg[2])

	// tl.Outline(1, s.P.Base.Blue)
}

type PriceBar struct {
	*Rect
	parent *Rect
}

func (pb *PriceBar) Render(s *State) {
	if s.WRF {
		pb.MoveTo(pb.parent.s.X-PBW, pb.parent.p.Y)
		pb.SetSize(PBW+RPD, pb.parent.s.Y-TLH)
	}

	pb.Fill(s.P.Bg[0])

	// pb.Outline(1, s.P.Base.Green)
}

type Crossing struct {
	*Rect
	parent *Rect
}

func (cr *Crossing) Render(s *State) {
	if s.WRF {
		cr.MoveTo(cr.parent.s.X-PBW, cr.parent.s.Y-TLH+RPD)
		cr.SetSize(PBW+RPD, TLH)
	}

	cr.Fill(s.P.Bg[0])

	// cr.Outline(1, s.P.Base.Red)
}
