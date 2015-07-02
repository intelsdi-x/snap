# psutil - influxdb - grafana

This example will

1.  start & configures grafana
2.  start & configures influxdb
3.  start pulse
4.  adds a pulse task to collect (psutil) metrics from your host and publishes them to influxdb
5. starts pulse task

### How to run the example

- ./run.sh *\<docker-machine name\>*
- open your browswer to *(your docker-machine IP)* at port 3000 to view the Grafana pulse dashboard {'user':'admin', 'password':'admin'}
- open your browser to *(your docker-machine IP)* at port 8083 to inspect the influxdb data through the web UI {'user':'admin', 'password':'admin'}

### Requirements
- docker-machine 
    + with a machine created

- docker-compose
    + installed

### Issues/Warning

- Make sure the time on your docker-machine vm is syncd with the time on your host 


   

