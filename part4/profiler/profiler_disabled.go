//go:build noprofiler

package profiler

type Profiler struct{}
type Result struct{}
type Block struct{}

func Begin(ProfileKind) (bl Block)                      { return }
func BeginWithBandwidth(ProfileKind, uint64) (bl Block) { return }
func End(Block)                                         {}
func PrintReport()                                      {}
