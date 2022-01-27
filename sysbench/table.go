/*
 * benchyou
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package sysbench

import (
	"fmt"
	"log"

	"github.com/xelabs/benchyou/xworker"
)

// Table tuple.
type Table struct {
	workers []xworker.Worker
}

// NewTable creates the new table.
func NewTable(workers []xworker.Worker) *Table {
	return &Table{workers}
}

// Prepare used to prepare the tables.
func (t *Table) Prepare() {
	session := t.workers[0].S
	count := t.workers[0].N
	engine := t.workers[0].E
	for i := 0; i < count; i++ {
		sql := fmt.Sprintf(`create table benchyou%d (
							id varchar(1024) not null,
							k int not null,
							c varchar(120) not null,
							pad varchar(60) not null,
							primary key (id)
							)`, i)

		if err := session.Exec(sql); err != nil {
			log.Panicf("creata.table.error[%v]", err)
		}
		log.Printf("create table benchyou%d(engine=%v) finished...\n", i, engine)
	}
}

// Cleanup used to cleanup the tables.
func (t *Table) Cleanup() {
	session := t.workers[0].S
	count := t.workers[0].N
	for i := 0; i < count; i++ {
		sql := fmt.Sprintf(`drop table benchyou%d;`, i)

		if err := session.Exec(sql); err != nil {
			log.Panicf("drop.table.error[%v]", err)
		}
		log.Printf("drop table benchyou%d finished...\n", i)
	}
}
