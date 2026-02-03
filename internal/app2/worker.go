package app2

type Worker struct {
	tx chan<- int
	rx <-chan int
}

