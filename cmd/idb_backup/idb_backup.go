package main

import (
	lib "devstats"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func copySeries(ch chan bool, ctx *lib.Ctx, from, to, seriesName string) {
	// Connect to InfluxDB
	ic := lib.IDBConn(ctx)
	defer func() { lib.FatalOnError(ic.Close()) }()

	// Get BatchPoints
	var pts lib.IDBBatchPointsN
	bp := lib.IDBBatchPointsWithDB(ctx, &ic, to)
	pts.NPoints = 0
	pts.Points = &bp

	// Get values from series
	res := lib.QueryIDBWithDB(ic, ctx, "select * from "+seriesName+" group by *", from)
	allSeries := res[0].Series
	for _, series := range allSeries {
		// Add batch point(s)
		dt := time.Now()
		columns := series.Columns
		values := series.Values
		tags := series.Tags
		for _, value := range values {
			fields := make(map[string]interface{})
			for i, column := range columns {
				if column == "time" {
					dt = lib.TimeParseIDB(value[i].(string))
				} else if value[i] != nil {
					switch interfaceValue := value[i].(type) {
					case json.Number:
						fVal, err := interfaceValue.Float64()
						lib.FatalOnError(err)
						fields[column] = fVal
					case string:
						fields[column] = interfaceValue
					default:
						lib.FatalOnError(fmt.Errorf("unknown type %T/%+v for field \"%s\"", interfaceValue, interfaceValue, column))
					}
				}
			}
			if ctx.Debug > 0 {
				fmt.Printf("%s: tags=%+v, fields=%+v, dt=%v\n", series.Name, tags, fields, dt)
			}
			pt := lib.IDBNewPointWithErr(series.Name, tags, fields, dt)
			lib.IDBAddPointNWithDB(ctx, &ic, &pts, pt, to)
		}
	}
	// Write the batch
	if !ctx.SkipIDB {
		lib.FatalOnError(lib.IDBWritePointsN(ctx, &ic, &pts))
	} else if ctx.Debug > 0 {
		lib.Printf("Skipping tags series write\n")
	}
	if ch != nil {
		ch <- true
	}
}

// Backup all series "from" --> "to"
func idbBackup(from, to string) {
	// Environment context parse
	var ctx lib.Ctx
	ctx.Init()

	// Get number of CPUs available
	thrN := lib.GetThreadsNum(&ctx)
	lib.Printf("idb_backup.go: Running (%v CPUs)\n", thrN)

	// Connect to InfluxDB
	ic := lib.IDBConn(&ctx)

	// Get all series names from input database
	res := lib.QueryIDBWithDB(ic, &ctx, "show series", from)
	if len(res[0].Series) < 1 {
		lib.Printf("Nothing to copy\n")
		return
	}
	iSeries := res[0].Series[0].Values

	// Get unique series name (without tags)
	uniSeries := make(map[string]struct{})
	for _, ser := range iSeries {
		split := strings.Split(ser[0].(string), ",")
		uniSeries[split[0]] = struct{}{}
	}
	series := []string{}
	for ser := range uniSeries {
		series = append(series, ser)
	}
	nSeries := len(series)

	// Close connection
	lib.FatalOnError(ic.Close())

	//series[0] = "company_multi_cluster_issues_y"
	//nSeries = 1
	dtStart := time.Now()
	lastTime := dtStart
	checked := 0
	lib.Printf("Processing %d series\n", nSeries)

	// Copy series
	if thrN > 1 {
		ch := make(chan bool)
		nThreads := 0
		for i := 0; i < nSeries; i++ {
			go copySeries(ch, &ctx, from, to, series[i])
			nThreads++
			if nThreads == thrN {
				<-ch
				nThreads--
				checked++
				lib.ProgressInfo(checked, nSeries, dtStart, &lastTime, time.Duration(10)*time.Second, "")
			}
		}
		lib.Printf("Final threads join\n")
		for nThreads > 0 {
			<-ch
			nThreads--
			checked++
			lib.ProgressInfo(checked, nSeries, dtStart, &lastTime, time.Duration(10)*time.Second, "final join...")
		}
	} else {
		lib.Printf("Using single threaded version\n")
		for i := 0; i < nSeries; i++ {
			copySeries(nil, &ctx, from, to, series[i])
			lib.ProgressInfo(i, nSeries, dtStart, &lastTime, time.Duration(10)*time.Second, "")
		}
	}
	// Finished
	lib.Printf("All done.\n")
}

func main() {
	dtStart := time.Now()
	if len(os.Args) < 3 {
		lib.Printf("%s: Required args: source_database destination_database\n", os.Args[0])
		os.Exit(1)
	}
	fmt.Printf(
		"Consider fresh restart of `influxd` service, this program temporarily doubles influxd memory usage.\n",
	)
	idbBackup(os.Args[1], os.Args[2])
	dtEnd := time.Now()
	lib.Printf("Time: %v\n", dtEnd.Sub(dtStart))
}
