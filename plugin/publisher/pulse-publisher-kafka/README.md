## Pulse Plugin - Kakfa Publisher


### Description

Allows publishing of data to [Apache Kafka](http://kafka.apache.org)

### Dependencies

Uses [sarama](http://shopify.github.io/sarama/) golang client for Kafka by Shopify

## Configuration details

| key      | value type | required | default  | description  |
|----------|------------|----------|----------|--------------|
| topic | string | yes | | The topic to send messages |         
| brokers  | string | yes | | Semicolon delimited list of "server:port" brokers |
