package types

const (
	BalancingStrategyWRR = "wrr"
	BalancingStrategyP2C = "p2c"
)

func ValidBalancingStrategy(strategy string) bool {
	switch strategy {
	case BalancingStrategyWRR, BalancingStrategyP2C:
		return true
	}
	return false
}
