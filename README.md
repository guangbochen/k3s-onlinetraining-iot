# k3s-online-training-iot

This project contains the online demo of rancher k3s online-training-IoT

## About the projects
1. `charts` contains a curated MQTT Broker and InfluxDB chart.
2. `db-metrics` helps to collect the data upon MQTT topics and store it to the InfluxDB
3. `device-temp-demo` using bluetooth to connect with the BLE device(e.g XiaoMi Temp Sensor) and push message to the MQTT broker.

## How to run it
1. Create a k3s cluster and import it to the Rancher2.0
2. Deploy the EMQX and InfluxDB chart
    ```
    using helm or Rancher Catalog UI
    1. default username and password of EMQX char is admin/public
    2. default username and password of InfluxDB is admin/passwd
    3. crate a new database named `mydb` of InfluxDB through pod console
    (https://docs.influxdata.com/influxdb/v1.7/introduction/getting-started/)
    ```
3. Add label to the k3s node(the one support BLE device connection, e.g. RaspberryPi)
    ```
    $kubectl label node pi prtocol:bluetooth

    Notes: install RaspberryPi BLE package with `$sudo apt-get install pi-bluetooth`
    ```
4. Deploy the `device-temp-demo` and `db-metrics` app via
    ```
   $ kubectl apply -f device-temp-demo/deploy/device-temp-demo.yaml
   $ kubectl apply -f device-temp-demo/deploy/influxdb-writer.yaml
   ```
5. Create the XiaoMi-Temp Grafana dashboard with json file in `grafana` directory.

## License

Copyright (c) 2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
