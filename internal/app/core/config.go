package core

type Config struct {
	InitWindow   InitWindow        `json:"init_window"`
	Layout       *Component        `json:"layout"`
	InitCommands []string          `json:"init_commands"`
	Vars         map[string]string `json:"vars"`
	LeaderKey    string            `json:"leader_key"`
	Binds        [][2]string       `json:"binds"`
	TargetFPS    int32             `json:"target_fps"`
	LoadFont     string            `json:"load_font"`
	RowHeight    float32           `json:"row_height"`
	Theme        Theme             `json:"theme"`
}

type InitWindow struct {
	Width  int32  `json:"width"`
	Height int32  `json:"height"`
	Title  string `json:"title"`
}

type Component struct {
	Type   string         `json:"type"`
	Params map[string]any `json:"params,omitempty"`
	A      *Component     `json:"a,omitempty"`
	B      *Component     `json:"b,omitempty"`
}

type Theme struct {
	Name    string      `json:"name"`
	IsLight bool        `json:"is_light"`
	Colors  ColorScheme `json:"colors"`
}

type ColorScheme struct {
	Background    [5]string       `json:"background"`
	Foreground    [4]string       `json:"foreground"`
	Selection     SelectionColors `json:"selection"`
	Cursor        CursorColors    `json:"cursor"`
	Base          AnsiColors      `json:"base"`
	Bright        AnsiColors      `json:"bright"`
	Dim           AnsiColors      `json:"dim"`
	Diff          DiffColors      `json:"diff"`
	CodeSelection [2]string       `json:"code_selection"`
	Comment       string          `json:"comment"`
}

type SelectionColors struct {
	Bg string `json:"bg"`
	Fg string `json:"fg"`
}

type CursorColors struct {
	Bg string `json:"bg"`
	Fg string `json:"fg"`
}

type AnsiColors struct {
	Black   string `json:"black"`
	Red     string `json:"red"`
	Green   string `json:"green"`
	Yellow  string `json:"yellow"`
	Blue    string `json:"blue"`
	Magenta string `json:"magenta"`
	Cyan    string `json:"cyan"`
	White   string `json:"white"`
	Orange  string `json:"orange"`
	Pink    string `json:"pink"`
}

type DiffColors struct {
	Add    string `json:"add"`
	Delete string `json:"delete"`
	Change string `json:"change"`
	Text   string `json:"text"`
}
