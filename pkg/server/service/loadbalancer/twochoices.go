package loadbalancer

import (
	crand "crypto/rand"
	"math/rand/v2"
)

type rnd interface {
	IntN(int) int
}

// strategyTwoRandomChoices implements "the power-of-two-random-choices" algorithm for load balancing.
// The idea of this is two take two of the backends at random from the available backends, and select
// the backend that has the fewest in-flight requests. This algorithm more effectively balances the
// load than a round-robin approach, while also being constant time when picking: The strategy also
// has more beneficial "herd" behaviour than the "fewest connections" algorithm, especially when the
// load balancer doesn't have perfect knowledge about the global number of connections to the backend,
// for example, when running in a distributed fashion.
type strategyTwoRandomChoices struct {
	handlers []*namedHandler
	rand     rnd
}

func newStrategyTRC() Strategy {
	return &strategyTwoRandomChoices{
		rand: newRand(),
	}
}

func (s *strategyTwoRandomChoices) nextServer(status map[string]struct{}) *namedHandler {
	// if there is only one healthy backend, we choose it unconditionally: this is O(n), but presumably
	// n will be small
	if len(status) == 1 {
		var healthy string
		for name := range status {
			healthy = name
		}
		for _, h := range s.handlers {
			if h.name == healthy {
				return h
			}
		}
	}

	for {
		n1, n2 := s.rand.IntN(len(s.handlers)), s.rand.IntN(len(s.handlers))
		if n1 == n2 {
			continue
		}

		h1, h2 := s.handlers[n1], s.handlers[n2]
		// ensure h1 has fewer inflight requests than h2
		if h2.inflight.Load() < h1.inflight.Load() {
			h1, h2 = h2, h1
		}

		if _, ok := status[h1.name]; !ok {
			continue
		}

		return h1
	}
}

func (s *strategyTwoRandomChoices) add(h *namedHandler) {
	s.handlers = append(s.handlers, h)
}

func (s *strategyTwoRandomChoices) name() string {
	return StrategyNameTwoRandomChoices
}

func (s *strategyTwoRandomChoices) len() int {
	return len(s.handlers)
}

func newRand() *rand.Rand {
	var seed [16]byte
	_, err := crand.Read(seed[:])
	if err != nil {
		panic(err)
	}
	var seed1, seed2 uint64
	for i := 0; i < 16; i += 8 {
		seed1 = seed1<<8 + uint64(seed[i])
		seed2 = seed2<<8 + uint64(seed[i+1])
	}
	return rand.New(rand.NewPCG(seed1, seed2))
}
