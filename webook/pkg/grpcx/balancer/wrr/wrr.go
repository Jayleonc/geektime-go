package wrr

import (
	"errors"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

const (
	Name            = "custom_weighted_round_robin"
	minWeight       = 10              // 最小权重
	maxWeight       = 5000            // 最大权重
	adjustStep      = 1               // 调整步长
	healthCheckFreq = 5 * time.Second // 健康检查频率

)

var (
	ErrRateLimited  = errors.New("rate limited")
	ErrCircuitBreak = errors.New("circuit break")
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
			p.adjustWeightV1(maxCC, info)
		},
	}, nil

}

// adjustWeightV1 根据返回的错误码进行判断
func (p *Picker) adjustWeightV1(c *weightConn, doneInfo balancer.DoneInfo) {
	p.lock.Lock()
	defer p.lock.Unlock()

	s, ok := status.FromError(doneInfo.Err)

	if !ok {
		switch s.Code() {
		case codes.ResourceExhausted:
			// 触发限流，降低权重
			c.weight -= adjustStep * 10
			if c.weight < minWeight {
				c.weight = minWeight
			}
		case codes.Unavailable:
			// 触发熔断，移除节点，并启动健康检查
			c.available = false
			go p.healthCheck(c, doneInfo)
		default:
			// 其他错误，轻微减少权重
			c.weight -= adjustStep
			if c.weight < minWeight {
				c.weight = minWeight
			}
		}
	} else {
		// 调用成功，增加权重
		c.weight += adjustStep
		if c.weight > maxWeight {
			c.weight = maxWeight
		}
	}
}

func (p *Picker) adjustWeight(c *weightConn, doneInfo balancer.DoneInfo) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if doneInfo.Err != nil {
		switch doneInfo.Err {
		case ErrRateLimited:
			// 触发限流，降低权重
			c.weight -= adjustStep * 10
			if c.weight < minWeight {
				c.weight = minWeight
			}
		case ErrCircuitBreak:
			// 触发熔断，移除节点，并启动健康检查
			c.available = false
			go p.healthCheck(c, doneInfo)
		default:
			// 其他错误，轻微减少权重
			c.weight -= adjustStep
			if c.weight < minWeight {
				c.weight = minWeight
			}
		}
	} else {
		// 调用成功，增加权重
		c.weight += adjustStep
		if c.weight > maxWeight {
			c.weight = maxWeight
		}
	}
}

// healthCheck 定时对熔断的节点进行健康检查
func (p *Picker) healthCheck(c *weightConn, info balancer.DoneInfo) {
	ticker := time.NewTicker(healthCheckFreq)
	defer ticker.Stop()
	for range ticker.C {
		if p.checkNodeHealth(c, info) {
			p.lock.Lock()
			c.available = true
			p.lock.Unlock()
			return
		}
	}
}

// checkNodeHealth 对节点进行健康检查
func (p *Picker) checkNodeHealth(c *weightConn, info balancer.DoneInfo) bool {
	// 这个应该怎么实现？？？
	return true
}

type weightConn struct {
	balancer.SubConn
	weight        int
	currentWeight int

	available bool
}
