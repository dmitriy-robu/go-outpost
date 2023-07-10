package start

type RouletteStartJob struct {
	RouletteStart *RouletteStart
	RouletteID    int64
}

func (job *RouletteStartJob) Execute() {
	job.RouletteStart.UpdateRouletteUpdateAt(job.RouletteID)
}
