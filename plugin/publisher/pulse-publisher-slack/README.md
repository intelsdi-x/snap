## Pulse Plugin - Slack Publisher


### Description

Allows publishing of data by sending it onto [Slack](http://slack.com) channel

### Dependencies

Uses [Bluele Slack](https://github.com/bluele/slack) library to connect to Slack api

## Configuration details

| key      | value type | required | default  | description  |
|----------|------------|----------|----------|--------------|
| token | string | yes | | Api token used in connection |
| channel | string | yes | | Channel name |
| name | string | yes | Pulse metrics | Name of message originator |

### Notes

You can generate api token on [this](https://api.slack.com/web) site
