## Pulse Plugin - Kakfa Publisher


### Description

Allows publishing of data to [Apache Kafka](http://kafka.apache.org)

### Dependencies

Uses [sarama](http://shopify.github.io/sarama/) golang client for Kafka by Shopify

## Configuration details

| key      | value type | required | default  |
|----------|------------|----------|----------|
| username | string     | no       | admin    |
| password | string     | no       | password |
| port     | integer    | no       | 9092     |
