/*
 * benchyou
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package xcmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
	"github.com/xelabs/benchyou/xcommon"
	"github.com/xelabs/benchyou/xstat"
	"github.com/xelabs/benchyou/xworker"
)

// Stats tuple.
type Stats struct {
	SystemCS uint64
	IdleCPU  uint64
	MemFree  uint64
	MemCache uint64
	SwapSi   uint64
	SwapSo   uint64
	RRQM_S   float64
	WRQM_S   float64
	RS      float64
	WS      float64
	RKB_S    float64
	WKB_S    float64
	AWAIT    float64
	UTIL     float64
}

// Monitor tuple.
type Monitor struct {
	conf    *xcommon.Conf
	workers []xworker.Worker
	ticker  *time.Ticker
	vms     *xstat.VMS
	ios     *xstat.IOS
	stats   *Stats
	all     *Stats
	seconds uint64
}

// NewMonitor creates the new monitor.
func NewMonitor(conf *xcommon.Conf, workers []xworker.Worker) *Monitor {
	return &Monitor{
		conf:    conf,
		workers: workers,
		ticker:  time.NewTicker(time.Second),
		vms:     xstat.NewVMS(conf),
		ios:     xstat.NewIOS(conf),
		stats:   &Stats{},
		all:     &Stats{},
	}
}

// Start used to start the monitor.
func (m *Monitor) Start() {
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	m.vms.Start()
	m.ios.Start()
	go func() {
		newm := &xworker.Metric{}
		oldm := &xworker.Metric{}
		for _ = range m.ticker.C {
			m.seconds++
			m.stats.SystemCS = m.vms.Stat.SystemCS
			m.stats.IdleCPU = m.vms.Stat.IdleCPU
			m.stats.MemFree = m.vms.Stat.MemFree
			m.stats.MemCache = m.vms.Stat.MemCache
			m.stats.SwapSi = m.vms.Stat.SwapSi
			m.stats.SwapSo = m.vms.Stat.SwapSo
			m.stats.RRQM_S = m.ios.Stat.RRQM_S
			m.stats.WRQM_S = m.ios.Stat.WRQM_S
			m.stats.RS = m.ios.Stat.RS
			m.stats.WS = m.ios.Stat.WS
			m.stats.RKB_S = m.ios.Stat.RKB_S
			m.stats.WKB_S = m.ios.Stat.WKB_S
			m.stats.AWAIT = m.ios.Stat.AWAIT
			m.stats.UTIL = m.ios.Stat.UTIL

			m.all.SystemCS += m.stats.SystemCS
			m.all.IdleCPU += m.stats.IdleCPU
			m.all.RRQM_S += m.stats.RRQM_S
			m.all.WRQM_S += m.stats.WRQM_S
			m.all.RS += m.stats.RS
			m.all.WS += m.stats.WS
			m.all.RKB_S += m.stats.RKB_S
			m.all.WKB_S += m.stats.WKB_S
			m.all.AWAIT += m.stats.AWAIT
			m.all.UTIL += m.stats.UTIL

			newm = xworker.AllWorkersMetric(m.workers)
			wtps := float64(newm.WNums - oldm.WNums)
			wcosts := float64(newm.WCosts - oldm.WCosts)
			rtps := float64(newm.QNums - oldm.QNums)
			rcosts := float64(newm.QCosts - oldm.QCosts)
			tps := wtps + rtps

			fmt.Fprintln(w, "time   \t\t   thds  \t tps   \twtps  \trtps  \trio  \trio/op \twio  \twio/op  \trMB   \trKB/op  \twMB   \twKB/op \tcpu/op\tfreeMB\tcacheMB\t w-rsp(ms)\tr-rsp(ms)\t  total-number")
			line := fmt.Sprintf("[%ds]\t\t[r:%d,w:%d,u:%d,d:%d]\t%d\t%d\t%d\t%d\t%.2f\t%d\t%0.2f\t%2.2f\t%.2f\t%2.2f\t%.2f\t%.2f\t%d\t%d\t %.2f\t%.2f\t  %v\n",
				m.seconds,
				m.conf.ReadThreads,
				m.conf.WriteThreads,
				m.conf.UpdateThreads,
				m.conf.DeleteThreads,
				int(tps),
				int(wtps),
				int(rtps),
				int(m.stats.RS),
				m.stats.RS/tps,
				int(m.stats.WS),
				m.stats.WS/tps,
				m.stats.RKB_S/1024,
				m.stats.RKB_S/tps,
				m.stats.WKB_S/1024,
				m.stats.WKB_S/tps,
				float64(m.stats.SystemCS)/tps,
				int(m.stats.MemFree),
				int(m.stats.MemCache),
				float64(wcosts)/1e6/wtps,
				float64(rcosts)/1e6/rtps,
				(newm.WNums + newm.QNums),
			)
			fmt.Fprintln(w, line)

			w.Flush()
			*oldm = *newm
		}
	}()
}

// Stop used to stop the monitor.
func (m *Monitor) Stop() {
	m.ticker.Stop()
	xworker.StopWorkers(m.workers)

	// avg results at the end
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	seconds := float64(m.seconds)
	all := xworker.AllWorkersMetric(m.workers)
	writes := float64(all.WNums)
	reads := float64(all.QNums)
	events := writes + reads

	fmt.Fprintln(w, "----------------------------------------------------------------------------------------------avg---------------------------------------------------------------------------------------------")
	fmt.Fprintln(w, "time   \t\t tps   \twtps  \trtps  \trio  \trio/op \twio  \twio/op  \trMB   \trKB/op  \twMB   \twKB/op \tcpu/op\t          w-rsp(ms)\t          r-rsp(ms)              total-number")
	line := fmt.Sprintf("[%ds]\t\t%d\t%d\t%d\t%d\t%.2f\t%d\t%0.2f\t%2.2f\t%.2f\t%2.2f\t%.2f\t%.2f\t[avg:%.2f,min:%.2f,max:%.2f]\t[avg:%.2f,min:%.2f,max:%.2f]\t    %v\n",
		m.seconds,
		int(events/seconds),
		int(writes/seconds),
		int(reads/seconds),
		int(m.stats.RS/seconds),
		m.stats.RS/events,
		int(m.stats.WS/seconds),
		m.stats.WS/events/seconds,
		m.stats.RKB_S/1024/seconds,
		m.stats.RKB_S/events/seconds,
		m.stats.WKB_S/1024/seconds,
		m.stats.WKB_S/events/seconds,
		float64(m.stats.SystemCS)/events/seconds,
		float64(all.WCosts)/1e6/writes/seconds,
		float64(all.WMin)/1e6,
		float64(all.WMax)/1e6,
		float64(all.QCosts)/1e6/reads/seconds,
		float64(all.QMin)/1e6,
		float64(all.QMax)/1e6,
		(all.WNums + all.QNums),
	)
	fmt.Fprintln(w, line)
	w.Flush()
}
