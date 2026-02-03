package config

type Config struct {
	InitWin InitWin `json:"init_win"`
	Theme   Theme   `json:"theme"`
}

type InitWin struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
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
