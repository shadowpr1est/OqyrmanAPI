# Monitoring deployment notes

## Required external steps

1. Add an A record for `monitoring.oqyrman.app` pointing to the server IP.
2. Issue a TLS certificate for `monitoring.oqyrman.app` and place it under `/etc/letsencrypt/live/monitoring.oqyrman.app/`.
3. Fill the monitoring secrets in `.env`:
   - `GRAFANA_ADMIN_USER`
   - `GRAFANA_ADMIN_PASSWORD`

Grafana uses its own login page at `https://monitoring.oqyrman.app/login`.

## After that

Run the prod stack and reload Nginx. Grafana will be available at `https://monitoring.oqyrman.app` and Prometheus will stay internal to the Docker network.

## Logs

Container logs are collected by Promtail and stored in Loki with a 7-day retention window. In Grafana, use the Loki datasource to search logs by container or compose service.

## Dashboards

Provisioned dashboards:

1. `Oqyrman Server` - host CPU, RAM, disk, and load.
2. `Oqyrman API` - request rate, latency, in-flight requests, and 5xx rate.
3. `Oqyrman Database` - Postgres connections, size, and transaction activity.
