package wrr

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

const (
	Name       = "custom_weighted_round_robin"
	minWeight  = 10   // 最小权重
	maxWeight  = 5000 // 最大权重
	adjustStep = 1    // 调整步长
)

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		md, _ := sci.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		conns = append(conns, &weightConn{
			SubConn:       sc,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}

	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var total int
	var maxCC *weightConn
	for _, c := range p.conns {
		total += c.weight
		c.currentWeight = c.currentWeight + c.weight
		if maxCC == nil || maxCC.currentWeight < c.currentWeight {
			maxCC = c
		}
	}

	maxCC.currentWeight = maxCC.currentWeight - total

	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			// 要在这里进一步调整weight/currentWeight
			// failover 要在这里做文章
			// 根据调用结果的具体错误信息进行容错
			// 1. 如果要是触发了限流了，
			// 1.1 你可以考虑直接挪走这个节点，后面再挪回来
			// 1.2 你可以考虑直接将 weight/currentWeight 调整到极低
			// 2. 触发了熔断呢？
			// 3. 降级呢？
			p.adjustWeight(maxCC, info)
		},
	}, nil

}

// adjustWeight 根据RPC调用结果动态调整权重
func (p *Picker) adjustWeight(c *weightConn, doneInfo balancer.DoneInfo) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if doneInfo.Err != nil {
		// 调用失败，减少权重
		c.weight -= adjustStep
		if c.weight < minWeight {
			c.weight = minWeight // 保证权重不会低于最小值
		}
	} else {
		// 调用成功，增加权重
		c.weight += adjustStep
		if c.weight > maxWeight {
			c.weight = maxWeight // 保证权重不会超过最大值
		}
	}
}

type weightConn struct {
	balancer.SubConn
	weight        int
	currentWeight int

	available bool
}
