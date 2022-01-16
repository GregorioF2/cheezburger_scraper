package semaphore

type CustomSemaphore struct {
	sem chan int
}

func NewCustomSemaphore(capacity int) *CustomSemaphore {
	return &CustomSemaphore{
		sem: make(chan int, capacity),
	}
}

func (s *CustomSemaphore) CurrentlyRunning() int {
	return len(s.sem)
}

func (s *CustomSemaphore) Close() {
	close(s.sem)
}

func (s *CustomSemaphore) Take() {
	s.sem <- 1
}

func (s *CustomSemaphore) Signal() {
	if len(s.sem) <= 0 {
		return
	}
	<-s.sem
}
