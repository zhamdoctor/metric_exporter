package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"runtime"
	"sync"
)

var reqCount int32
var hostname string

type NodeCollector struct {
	//这里也可以用map[string]*prometheus.Desc和sync.Mutex
	requestDesc   *prometheus.Desc
	nodeMetrics   nodeStatsMetrics
	goroutineDesc *prometheus.Desc
	threadsDesc   *prometheus.Desc
	summaryDesc   *prometheus.Desc
	histogramDesc *prometheus.Desc
	mutex         sync.Mutex
}

func (n NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- n.requestDesc
	for _, metric := range n.nodeMetrics {
		ch <- metric.desc
	}
	ch <- n.goroutineDesc
	ch <- n.threadsDesc
	ch <- n.summaryDesc
	ch <- n.histogramDesc
}

//采集动作
func (n NodeCollector) Collect(ch chan<- prometheus.Metric) {
	n.mutex.Lock()
	ch <- prometheus.MustNewConstMetric(n.requestDesc, prometheus.CounterValue, 0, hostname)
	vm, err := mem.VirtualMemory()
	if err != nil {
		up := prometheus.MustNewConstMetric(prometheus.NewDesc(prometheus.BuildFQName("default", "", "up"), "", nil, nil), prometheus.GaugeValue, 1)
		ch <- up
	}
	for _, metric := range n.nodeMetrics {
		ch <- prometheus.MustNewConstMetric(metric.desc, metric.valType, metric.eval(vm))
	}
	ch <- prometheus.MustNewConstMetric(n.goroutineDesc, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
	num, _ := runtime.ThreadCreateProfile(nil)
	ch <- prometheus.MustNewConstMetric(n.threadsDesc, prometheus.GaugeValue, float64(num))
	//模拟数据
	ch <- prometheus.MustNewConstSummary(n.summaryDesc, 4711, 403.34, map[float64]float64{0.5: 42.3, 0.9: 323.3}, "200", "get")
	//模拟数据
	ch <- prometheus.MustNewConstHistogram(n.histogramDesc, 4722, 403.34, map[float64]uint64{25: 121, 50: 2403, 100: 3221, 200: 4233}, "200", "get")
	n.mutex.Unlock()
}

type nodeStatsMetrics []struct {
	desc    *prometheus.Desc
	eval    func(stat *mem.VirtualMemoryStat) float64
	valType prometheus.ValueType
}

//初始化采集器
//这是一个exporter，
//也可以传一个namespace参数进去。
func NewNodeCollector(namespace string) prometheus.Collector {
	host, _ := host.Info() //动态标签值，不是用于metric指标，指标在collect方法中定义
	hostname := host.Hostname
	return &NodeCollector{
		requestDesc: prometheus.NewDesc(
			"total_request_count",
			"请求数",
			[]string{"DYNAMIC_HOST_NAME"},
			prometheus.Labels{"STATIC_LABEL1": "静态标签", "HOST_NAME": hostname},
		),
		nodeMetrics: nodeStatsMetrics{
			{desc: prometheus.NewDesc("total_name", "内存总量", nil, nil), valType: prometheus.GaugeValue, eval: func(stat *mem.VirtualMemoryStat) float64 {
				return float64(stat.Total) / 1e9
			},
			},
			{
				desc:    prometheus.NewDesc("free_mem", "内存空闲", nil, nil),
				valType: prometheus.GaugeValue,
				eval: func(stat *mem.VirtualMemoryStat) float64 {
					return float64(stat.Free) / 1e9
				},
			},
		},
		goroutineDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, "goroutine", "count"), "携程数量", nil, nil),
		threadsDesc:   prometheus.NewDesc(prometheus.BuildFQName(namespace, "thread_num", "count"), "线程数", nil, nil),
		summaryDesc:   prometheus.NewDesc(prometheus.BuildFQName(namespace, "summary_http_request_duration_seconds", "count"), "summary类型", []string{"code", "method"}, prometheus.Labels{"owner": "example"}),
		histogramDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, "histogram_http_request_duration_seconds", "count"), "histogram类型", []string{"code", "method"}, prometheus.Labels{"owner": "example"}),
		mutex:         sync.Mutex{},
	}
}
