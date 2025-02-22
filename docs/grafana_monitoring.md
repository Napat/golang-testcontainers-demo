# Grafana Monitoring

This document describes the Grafana monitoring setup in the project, including dashboard configurations, metrics, and usage instructions.

## Overview

The project uses Grafana for visualization of metrics collected by Prometheus. The setup includes:

- Pre-configured dashboards for different metrics categories
- Prometheus data source integration
- Automated dashboard provisioning

## Setup Configuration

### Docker Compose Setup

The monitoring stack is defined in `docker-compose.monitoring.yml` with the following components:

- **Grafana**: Running on port 13000
  - Default credentials:
    - Username: admin
    - Password: admin
  - Configured with automatic dashboard and data source provisioning

- **Prometheus**: Running on port 9090
  - Configured as the default data source for Grafana
  - Collects metrics from the application

## Available Dashboards

### Repository Metrics Dashboard

The main dashboard includes several metric categories:

1. **HTTP Metrics**
   - Request Rate: Shows the rate of HTTP requests by endpoint
   - Average Response Time: Displays response time trends

2. **Database Metrics**
   - Open Database Connections: Tracks connection pool usage
   - Average Query Duration: Shows query performance metrics

3. **Cache Metrics**
   - Cache Hit/Miss Rate: Monitors cache effectiveness
   - Cache Operation Duration: Tracks cache performance

4. **Message Metrics**
   - Message Publishing Rate
   - Message Processing Duration

5. **Search Metrics**
   - Search Operation Duration
   - Search Results Statistics

## Metrics Collection

The application exports the following metric types:

### HTTP Metrics

- `http_requests_total`: Total number of HTTP requests
- `http_request_duration_seconds`: Duration of HTTP requests

### Database Metrics

- `*_db_connections_open`: Number of open database connections
- `*_db_query_duration_seconds`: Duration of database queries
- `*_db_queries_total`: Total number of database queries

### Cache Metrics

- `cache_hits_total`: Total number of cache hits
- `cache_misses_total`: Total number of cache misses
- `cache_operation_duration_seconds`: Duration of cache operations

### Message Metrics

- `messages_published_total`: Total number of messages published
- `message_publish_duration_seconds`: Duration of message publish operations

### Search Metrics

- `search_duration_seconds`: Duration of search operations
- `searches_total`: Total number of searches performed
- `search_results_returned`: Number of results returned by search operations

## Accessing Grafana

1. Start the monitoring stack:

   ```bash
   docker-compose -f docker-compose.monitoring.yml up -d
   ```

2. Access Grafana UI:
   - Open your browser and navigate to: <http://localhost:13000>
   - Log in with the default credentials (admin/admin)

3. View Dashboards:
   - Click on the Dashboards icon in the left sidebar
   - Select "Repository Metrics" dashboard

## Dashboard Customization

The dashboards are provisioned automatically but can be customized:

1. Dashboard configurations are stored in:
   - `/configs/grafana/dashboards/`

2. Data source configurations are in:
   - `/configs/grafana/provisioning/datasources/`

3. Dashboard provisioning settings are in:
   - `/configs/grafana/provisioning/dashboards/`

## Best Practices

1. **Monitoring**:
   - Regularly check the dashboards for anomalies
   - Set up alerts for critical metrics
   - Review historical trends for capacity planning

2. **Dashboard Management**:
   - Make dashboard changes through version control
   - Document any custom modifications
   - Use consistent naming conventions for metrics

3. **Troubleshooting**:
   - Use the query explorer to investigate specific metrics
   - Compare metrics across different time ranges
   - Correlate different metrics for root cause analysis
