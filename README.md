[![Maintainability](https://api.codeclimate.com/v1/badges/09d87b74d2f9cd86fa5e/maintainability)](https://codeclimate.com/github/k-kinzal/postfix-prometheus-exporter/maintainability)

# Postfix Prometheus Exporter

**NOTE: You should use [kumina/postfix_exporter](https://github.com/kumina/postfix_exporter), which is the de facto standard. This repository exists for study by [@k-kinzal](https://github.com/k-kinzal).**

Use poostfix prometheus exporter to monitor ostfix in prometheus.

## Overview

It converts the messages present in Postfix's showq into prometheus-formatted metrics, and finally exposes them via an HTTP server, where they are collected by prometheus.

## Get started

In this section, we show how to quickly run postfix prometheus exporter for postfix.

### Prerequisites

We assume that you have already installed prometheus and postfix. Additionally, you need to:

- If you have strict access control in postfix, you need to grant access rights to postfix prometheus exporter users.
    - e.g. `postconf -e "authorized_mailq_users = [your postfix prometheus export user]"`

### Running the Exporter in a Docker Container

To export postfix metrics, run:
```
$ postfix-prometheus-exporter
```

## Usage

### Command-line Arguments

```
usage: postfix-prometheus-exporter [<flags>]

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9154"  
                             Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"  
                             Path under which to expose metrics.
      --web.disable-exporter-metrics  
                             Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).
      --postfix.showq-path="/var/spool/postfix/public/showq"  
                             Path to showq in postfix.
      --postfix.interval=60  Postfix queue in the background to collect statistics on the interval (seconds).
      --log.level=info       Only log messages with the given severity or above. One of: [debug, info, warn,
                             error]
      --log.format=logfmt    Output format of log messages. One of: [logfmt, json]
      --version              Show application version.
```

### Exported Metrics

- `postfix_queue_age_seconds` -- Age of messages in the queue, in seconds
- `postfix_queue_size_bytes` -- Total message size in the queue
- `postfix_scope_collector_duration_seconds` -- Duration of a collector scrap
- `postfix_scope_collector_success` -- Whether a collector succeeded