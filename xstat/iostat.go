/*
 * benchyou
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package xstat

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"github.com/xelabs/benchyou/xcommon"

	"golang.org/x/crypto/ssh"
)

// IOStat tuple.
type IOStat struct {
	RRQM_S float64
	WRQM_S float64
	RS    float64
	WS    float64
	RKB_S  float64
	WKB_S  float64
	AWAIT  float64
	UTIL   float64
}

// IOS tuple.
type IOS struct {
	cmd    string
	Stat   *IOStat
	All    *IOStat
	t      *time.Ticker
	client *ssh.Client
}

// NewIOS creates new IOS.
func NewIOS(conf *xcommon.Conf) *IOS {
	client, err := sshConnect(
		conf.SSHUser,
		conf.SSHPassword,
		conf.SSHHost,
		conf.SSHPort)
	if err != nil {
		fmt.Printf("WARNING: ssh error: %+v\n", err)
	}

	return &IOS{
		cmd:    "iostat -x -g ALL 1 2",
		Stat:   &IOStat{},
		All:    &IOStat{},
		t:      time.NewTicker(time.Second),
		client: client,
	}
}

/*
Linux 4.4.0-42-generic (i-5i85yss9)     11/07/2016      _x86_64_        (16 CPU)

avg-cpu:  %user   %nice %system %iowait  %steal   %idle
           0.01    0.00    0.00    0.00   99.64    0.34

Device:         rrqm/s   wrqm/s     r/s     w/s    rkB/s    wkB/s avgrq-sz avgqu-sz   await r_await w_await  svctm  %util
sdb               0.00   355.01  175.59  472.40  5396.65  9657.07    46.46     0.57    0.89    0.36    1.08   0.20  13.15
sdc               0.04     0.31    0.04    0.05     0.31     1.45    38.90     0.00    2.60    0.22    4.33   2.24   0.02
sda               0.00     6.09    0.24    1.78     4.26    55.69    59.36     0.01    6.26    0.38    7.07   0.21   0.04
 ALL              0.04   361.41  175.87  474.23  5401.22  9714.22    46.50     0.59    0.90    0.36    1.10   0.20   4.40
*/
func (v *IOS) parse(outs string) (err error) {
	var line string

	lines := strings.Split(outs, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, " ALL") {
			line = l
			//break
		}
	}

	cols := splitColumns(line)

	if v.Stat.RRQM_S, err = strconv.ParseFloat(cols[1], 64); err != nil {
		return
	}
	v.All.RRQM_S += v.Stat.RRQM_S

	if v.Stat.WRQM_S, err = strconv.ParseFloat(cols[2], 64); err != nil {
		return
	}
	v.All.WRQM_S += v.Stat.WRQM_S

	if v.Stat.RS, err = strconv.ParseFloat(cols[3], 64); err != nil {
		return
	}
	v.All.RS += v.Stat.RS

	if v.Stat.WS, err = strconv.ParseFloat(cols[4], 64); err != nil {
		return
	}
	v.All.WS += v.Stat.WS

	if v.Stat.RKB_S, err = strconv.ParseFloat(cols[5], 64); err != nil {
		return
	}
	v.All.RKB_S += v.Stat.RKB_S

	if v.Stat.WKB_S, err = strconv.ParseFloat(cols[6], 64); err != nil {
		return
	}
	v.All.WKB_S += v.Stat.WKB_S

	if v.Stat.AWAIT, err = strconv.ParseFloat(cols[9], 64); err != nil {
		return
	}
	v.All.AWAIT += v.Stat.AWAIT

	if v.Stat.UTIL, err = strconv.ParseFloat(cols[13], 64); err != nil {
		return
	}
	v.All.UTIL += v.Stat.UTIL
	return
}

func (v *IOS) fetch() error {
	if v.client == nil {
		return nil
	}
	session, err := v.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	outs, err := session.CombinedOutput(v.cmd)
	if err != nil {
		return err
	}
	return v.parse(string(outs))
}

// Start used to start the ios.
func (v *IOS) Start() {
	go func() {
		for _ = range v.t.C {
			if err := v.fetch(); err != nil {
				log.Printf("iostat.fetch.error[%v]\n", err)
			}
		}
	}()
}

// Stop used to stop the ios.
func (v *IOS) Stop() {
	v.t.Stop()
	v.client.Close()
}
