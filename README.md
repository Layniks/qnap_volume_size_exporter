# qnap_volume_size_exporter
QNAP Volume Size exporter for Prometheus. This exporter using QNAP API to get data about volumes

## Logic
1. Get SID from QNAP API using QNAP user and token
2. API requests to QNAP to get data about volumes sizes (capacity, free and used)
3. Parse gathered data
4. Expose data as web page

## Prerequisite
1. GOlang latest version
2. Installed golang packages: **prometheus** and **promhttp**
3. Get **qtoken** using username and password via QNAP API

## Installation
1. Download the latest version of **qnap_volume_size_exporter**
2. Place binary file where you want

## Usage
Run `./qnap_volume_size_exporter --hostname "qnap_hostname" --qnap_user "user" --token "qtoken" --port ":port"`
- **hostname**: QNAP hostname
- **qnap_user**: QNAP read-only user
- **token**: Token for authentication (read QNAP manual about how to get "**qtoken**")
- **port**: On which port run exporter. The colon in :port is necessary for GOlang http.ListenAndServe

## systemd
Create systemd service to autorun exporter:
```bash
[Unit]
Description=qnap volume size exporter
After=network.target

[Service]
Type=simple
ExecStart=/path/to/file/qnap_volume_size_exporter --hostname "qnap_hostname" --qnap_user "user" --token "qtoken" --port ":port"
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

## Metrics
- **qnap_volume_capacity**:
    - What it measures: Total storage capacity of QNAP volume
    - Unit: KB (kilobytes)
    - Labels: VolName, Unit, host
    - Type: gauge

- **qnap_volume_used_size**:
    - What it measures: Currently used storage space on QNAP volume
    - Unit: KB (kilobytes)
    - Labels: VolName, Unit, host
    - Type: gauge

- **qnap_volume_free_size**:
    - What it measures: Available free space on QNAP volume
    - Unit: KB (kilobytes)
    - Labels: VolName, Unit, host
    - Type: gauge
