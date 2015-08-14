## Pulse Plugin - Mail Publisher


### Description

Allows publishing of data by e-mail

### Dependencies

There is no external dependencies

## Configuration details

| key      | value type | required | default  | description  |
|----------|------------|----------|----------|--------------|
| username | string | yes | | User name used in smtp connection |
| password | string | yes | | Password used in smtp connection |
| sender address | string | yes | | Mail address to set as sender address |
| mail addresses | string | yes | | Comma separated list of email recipients addresses |
| server address | string | yes | smtp.gmail.com | SMTP server address to use |
| server port | int | yes | 587 | SMTP server port to use |
| subject | string | yes | \[Pulse metrics\] | Mail subject |
