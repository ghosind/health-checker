# health-checker

![build](https://github.com/ghosind/health-checker/workflows/build/badge.svg)

A simple server instances health status checker, and it will send alarm notifications to specified email via [AWS SES](https://aws.amazon.com/cn/ses/).

## Getting Start

1. Download this repo.
2. Run `go build` to build binary executable file.
    - Run `make` to make binary executable file for x86-64 Linux.
3. Create your config json file, see the [Configuration](#configuration) or [Example](#example) section for details.
4. Run `health-checker config.json` to test it. (replace `config.json` to your config file path)
5. Add it into crontab. (Optional)

## Configurations

| Name | Value Type | Description |
|:----:|:----------:|:------------|
| `uri` | String | The uri for health check |
| `timeout` | Number | The request timeout in seconds |
| `groups` | Array<[`Group`](#group)> | Instance groups config |
| `instances` | Array<[`Instance`](#instance)> | Instances config (without group) |
| `aws` | [`AWS`](#aws) | AWS credential and settings |
| `recipient` | String | The email address to receive alarm notification |

### Group

| Name | Value Type | Description |
|:----:|:----------:|:------------|
| `name` | String | Group name |
| `type` | `'all'` or `'any'` | See [Group Type](#group-type) section |
| `instances` | Array<[`Instance`](#instance)> | The instances of this group |

### Instance

| Name | Value Type | Description |
|:----:|:----------:|:------------|
| `addr` | String | Instance address |
| `uri` | String | The uri of health check for this instance |

### AWS

| Name | Value Type | Description |
|:----:|:----------:|:------------|
| `clientId` | String | Your AWS access key id |
| `clientSecret` | String | Your AWS secret access key |
| `region` | String | AWS region |
| `sender` | String | The sender email that must be verified with SES |

### Group Type

The group support the following types:

- `all`: The group will be failed if some instances are unreachable.
- `any`: The group will be failed if all instances are unreachable.

## Example

There are a configuration file example:

```json
{
  "group": [{
    "name": "Group 1",
    "type": "all",
    "instances": [{
      "addr": "192.168.1.100:8000"
    },{
      "addr": "192.168.1.100:8001"
    }]
  }, {
    "name": "Group 2",
    "type": "any",
    "instances": [{
      "addr": "192.168.1.101:8000"
    },{
      "addr": "192.168.1.101:8001"
    }]
  }],
  "instances": [{
    "addr": "192.168.1.110",
    "uri": "/health/special"
  }],
  "uri": "/health/check",
  "timeout": 5,
  "aws": {
    "clientId": "<your_aws_access_key_id>",
    "clientSecret": "<your_aws_secret_access_key>",
    "region": "<your_aws_region>",
    "sender": "<your_sender_email>"
  },
  "recipient" "<your_email>"
}
```

## License

This project was published under MIT license.
