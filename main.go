package main

import (
	"exporter/exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func init() {
	//自身采集器
	prometheus.MustRegister(collector.NewNodeCollector("host"))
	//通用采集器
	//prometheus.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	//prometheus.MustRegister(collectors.NewGoCollector())
}

func main() {
	http.Handle("/metrics", promhttp.Handler()) //默认有NewGoCollector采集go运行时数据
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
