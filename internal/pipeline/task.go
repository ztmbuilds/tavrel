package pipeline

type Task interface {
	Process()
}

type TaskFunc func()

func (f TaskFunc) Process() {
	f()
}
