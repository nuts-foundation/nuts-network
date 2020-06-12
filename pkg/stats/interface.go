package stats

type Statistic interface {
	Name() string
	String() string
}

type StatsProvider interface {
	Statistics() []Statistic
}