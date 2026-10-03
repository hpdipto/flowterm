package main

type config struct {
	focusSessionName string
	focusMinute      int
	breakSessionName string
	breakMinute      int
}

func main() {
	c := config{
		focusSessionName: "Focus",
		focusMinute:      25,
		breakSessionName: "Break",
		breakMinute:      5,
	}
	InitTUi(c)
}
