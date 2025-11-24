package bootstrap

import (
	"os"
	"strings"
)

type Topology struct {
	Exchanges map[string]string // name -> kind
	Queues    []string
	Bindings  [][3]string // [queue, exchange, routingKey]
}

// LoadTopologyFromEnv parses simple env-based topology configuration.
// RABBITMQ_EXCHANGES=name:kind,name2:kind2
// RABBITMQ_QUEUES=q1,q2
// RABBITMQ_BINDINGS=queue:exchange:key,queue2:exchange2:key2
func LoadTopologyFromEnv() Topology {
	top := Topology{
		Exchanges: map[string]string{},
		Queues:    []string{},
		Bindings:  [][3]string{},
	}
	if v := os.Getenv("RABBITMQ_EXCHANGES"); v != "" {
		for _, part := range strings.Split(v, ",") {
			parts := strings.SplitN(strings.TrimSpace(part), ":", 2)
			if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
				top.Exchanges[parts[0]] = parts[1]
			}
		}
	}
	if v := os.Getenv("RABBITMQ_QUEUES"); v != "" {
		for _, q := range strings.Split(v, ",") {
			q = strings.TrimSpace(q)
			if q != "" {
				top.Queues = append(top.Queues, q)
			}
		}
	}
	if v := os.Getenv("RABBITMQ_BINDINGS"); v != "" {
		for _, b := range strings.Split(v, ",") {
			parts := strings.SplitN(strings.TrimSpace(b), ":", 3)
			if len(parts) == 3 && parts[0] != "" && parts[1] != "" {
				top.Bindings = append(top.Bindings, [3]string{parts[0], parts[1], parts[2]})
			}
		}
	}
	return top
}


