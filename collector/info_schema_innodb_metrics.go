// Scrape `information_schema.innodb_metrics`.

package collector

import (
	"database/sql"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const infoSchemaInnodbMetricsQuery = `
		SELECT
		  name, subsystem, type, comment,
		  count
		  FROM information_schema.innodb_metrics
		  WHERE status = 'enabled'
		`

// Metrics descriptors.
var (
	infoSchemaBufferPageReadTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "innodb_metrics_buffer_page_read_total"),
		"Total number of buffer pages read total.",
		[]string{"type"}, nil,
	)
	infoSchemaBufferPageWrittenTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "innodb_metrics_buffer_page_written_total"),
		"Total number of buffer pages written total.",
		[]string{"type"}, nil,
	)
)

// Regexp for matching metric aggregations.
var bufferPageRE = regexp.MustCompile(`^buffer_page_(read|written)_(.*)$`)

// ScrapeInnodbMetrics collects from `information_schema.innodb_metrics`.
func ScrapeInnodbMetrics(db *sql.DB, ch chan<- prometheus.Metric) error {
	innodbMetricsRows, err := db.Query(infoSchemaInnodbMetricsQuery)
	if err != nil {
		return err
	}
	defer innodbMetricsRows.Close()

	var (
		name, subsystem, metricType, comment string
		value                                float64
	)

	for innodbMetricsRows.Next() {
		if err := innodbMetricsRows.Scan(
			&name, &subsystem, &metricType, &comment, &value,
		); err != nil {
			return err
		}
		// Special handling of the "buffer_page_io" subsystem.
		if subsystem == "buffer_page_io" {
			match := bufferPageRE.FindStringSubmatch(name)
			if len(match) != 3 {
				log.Warnln("innodb_metrics subsystem buffer_page_io returned an invalid name:", name)
				continue
			}
			switch match[1] {
			case "read":
				ch <- prometheus.MustNewConstMetric(
					infoSchemaBufferPageReadTotalDesc, prometheus.CounterValue, value, match[2],
				)
			case "written":
				ch <- prometheus.MustNewConstMetric(
					infoSchemaBufferPageWrittenTotalDesc, prometheus.CounterValue, value, match[2],
				)
			}
			continue
		}
		metricName := "innodb_metrics_" + subsystem + "_" + name
		// MySQL returns counters named two different ways. "counter" and "status_counter"
		// value >= 0 is necessary due to upstream bugs: http://bugs.mysql.com/bug.php?id=75966
		if (metricType == "counter" || metricType == "status_counter") && value >= 0 {
			description := prometheus.NewDesc(
				prometheus.BuildFQName(namespace, informationSchema, metricName+"_total"),
				comment, nil, nil,
			)
			ch <- prometheus.MustNewConstMetric(
				description,
				prometheus.CounterValue,
				value,
			)
		} else {
			description := prometheus.NewDesc(
				prometheus.BuildFQName(namespace, informationSchema, metricName),
				comment, nil, nil,
			)
			ch <- prometheus.MustNewConstMetric(
				description,
				prometheus.GaugeValue,
				value,
			)
		}
	}
	return nil
}