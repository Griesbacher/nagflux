package target

//HasWorker is a interface to represent a struct which can start and stop workers.
type HasWorker interface {
	AddWorker()
	RemoveWorker()
}
