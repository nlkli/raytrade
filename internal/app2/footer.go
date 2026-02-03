package app2

type Footer struct {
	*Rect
	parent *Rect

	cl *CommandLine
	sl *StatusLine
}

func (f *Footer) Render(s *State) {
}

type CommandLine struct {
	*Rect
	parent *Rect
}

func (cl *CommandLine) Render(s *State) {
}

type StatusLine struct {
	*Rect
	parent *Rect
}

func (sl *StatusLine) Render(s *State) {
}
