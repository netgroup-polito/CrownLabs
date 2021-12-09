# Datasheet

LogQL uses labels and operators for filtering the query. The query is composed of:
- a log stream selector, e.g., `{container="query-frontend",namespace="loki-dev"}` which targets the query-frontend container in the loki-dev namespace.
- a log pipeline, e.g., `|= "metrics.go" | logfmt | duration > 10s and throughput_mb < 500`, which filters out the logs that contain the word `metrics.go`, then parses each log line to extract more labels and filters through them.

## Pipeline

|  LINE FILTER OPERATOR | DESCRIPTION  | 
|:---:|---:|
| \|=  | exact equal  |
|  != |  not  equal|
|  \|~ | regex matches  |
|  !~ | regex does not match  |

### Example:
> {container="query-frontend",namespace="loki-dev"} |= "metrics.go" | logfmt | duration > 10s and throughput_mb < 500

## Stream Selector

| LABEL FILTER OPERATOR|  DESCRIPTION|
|:---:|---:|
|  = |  exact equal |
|  != |  not  equal |
|  =~ |  regex matches |
|  !~ |   regex does not match|
|  > >= |  greater than and greater than or equal |
|  < <= |  lesser than and lesser than or equal |

### Example:
> {instance=~"kafka-[23]",name="kafka"}
## Functions

|FUNCTION|DESCRIPTION|
|:---:|---:|
|rate(log-range)| calculates the number of entries per second|
|count_over_time(log-range)| counts the entries for each log stream within the given range|
|bytes_rate(log-range)|calculates the number of bytes per second for each stream|
|bytes_over_time(log-range)|counts the amount of bytes used by each log stream for a given range|

***Note**: This kind of functions are mainly used to create charts for dashboards*
 
 ### Examples:
- Count all the log lines within the last five minutes for the MySQL job. 
  > *count_over_time({job="mysql"}[5m])*
 
- This aggregation includes filters and parsers. It returns the per-second rate of all non-timeout errors within the last minutes per host for the MySQL job and only includes errors whose duration is above ten seconds.
  > *sum by (host) (rate({job="mysql"} |= "error" != "timeout" | json | duration > 10s [1m]))*
