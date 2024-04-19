<p align="center">
      <img src="./assets/e1s-label.png" alt="e1s" width="300" height="150" >
</p>


# E1S - Easily Manage AWS ECS Resources in Terminal 🐱

`e1s` is a terminal application to easily browse and manage AWS ECS resources, with a focus on [Fargate](https://aws.amazon.com/fargate). Inspired by [k9s](https://github.com/derailed/k9s).

![e1s-screenshot](./assets/e1s-screenshot.png)

<details>
  <summary>A quick video demo</summary>
  ![e1s-demo](./assets/e1s-top-demo.gif)
</details>

## AWS credentials and configuration

`e1s` uses the default [aws-cli configuration](https://github.com/aws/aws-cli/blob/develop/README.rst#configuration). It does not store or send your access and secret key anywhere. The access and secret key are used only to securely connect to AWS API via AWS SDK. Both profile and region are overridable via the `AWS_PROFILE` and `AWS_REGION`.

## Installation

`e1s` is available on Linux, macOS and Windows platforms.

- Binaries for Linux, Windows and Mac are available in the [release](https://github.com/keidarcy/e1s/releases) page.
- Via Homebrew for maxOS or Linux

```bash
brew install keidarcy/tap/e1s
# upgrade
# brew upgrade keidarcy/tap/e1s
```

## Features

### Basic

- [x] Read only mode
- [x] Auto refresh
- [x] Describe clusters
- [x] Describe services
- [x] Describe tasks
- [x] Describe containers
- [x] Describe task definitions
- [x] Describe service autoscaling
- [x] Show cloudwatch logs(Only support awslogs logDriver)
  - [x] Realtime log streaming(Only support one log group)
- [x] Show Metrics
  - [x] CPUUtilization
  - [x] MemoryUtilization
- [x] Show autoscaling target and policy
- [x] Open selected resource in browser(Support new UI(v2))
- [x] SSH into container
- [x] Edit service
  - [x] Desired count
  - [x] Force new deployment
  - [x] Task definition family
  - [x] Task definition revision
- [x] Register new task definition
- [x] Start port forwarding session
- [x] Start remote host port forwarding session
- [x] Transfer files to and from your local machine and a remote host like `aws s3 cp`

### SSH into container ([ECS Exec](https://docs.aws.amazon.com/AmazonECS/latest/userguide/ecs-exec.html))

Use [aws-ecs-exec-checker](https://github.com/aws-containers/amazon-ecs-exec-checker) to check for the pre-requisites to use ECS exec.

<details>
  <summary>ssh demo</summary>

  ![ssh-demo](./assets/e1s-ssh-demo.gif)
</details>

Use `ctrl` + `d` to exit ssh session.

#### Troubleshooting

*The execute command failed because execute command...* - check [service execute command](https://github.com/keidarcy/e1s/blob/c9587a0bd89eacc08a1fd392523f518309e2437f/tests/ecs.tf#L102), [task role policy](https://github.com/keidarcy/e1s/blob/c9587a0bd89eacc08a1fd392523f518309e2437f/tests/ecs.tf#L157-L168)

*Session Manager plugin not found* - [document](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-troubleshooting.html#plugin-not-found).

### Update service([Docs](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html))

<details>
  <summary>update service demo</summary>

  ![update-service-demo](./assets/e1s-update-service-demo.gif)
</details>

- Force new deployment
- Desired tasks
- Task definition family
- Task definition revision

### Register task definition([Docs](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_RegisterTaskDefinition.html))

<details>
  <summary>Register task definition</summary>

  ![register-task-definition-demo](./assets/e1s-register-task-definition-demo.gif)
</details>


### Start port forwarding session([Docs](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-sessions-start.html#sessions-start-port-forwarding))

With a specified task and container, to start port forwarding session you need to specify a port and a local port. The local port is the port on your local machine that you want to use to access the container port.

<details>
  <summary>Port forwarding session</summary>

  ![port-forwarding-session-demo](./assets/e1s-port-forwarding-session-demo.gif)
</details>

### Start remote host port forwarding session([Docs](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-sessions-start.html#sessions-remote-port-forwarding))

With a specified task and container, to start a remote host port forwarding session you need to specify a port, a host and a local port. The local port is the port on your local machine that you want to use to access the remote host port though container.

<details>
  <summary>Remote host port forwarding session</summary>

  ![remote-host-port-forwarding-session-demo](./assets/e1s-remote-host-port-forwarding-session-demo.gif)
</details>

### File transfer

Since file transfer though a S3 Bucket and aws-cli in container, you need a S3 bucket and add permissions S3 bucket permission to the task role and e1s role, and also need a aws-cli installed container.

<details>
  <summary>File transfer</summary>

  ![file-transfer-demo](./assets/e1s-file-transfer-demo.gif)
</details>

## Usage

### Run `e1s`


| Long | Short | Description | Default Value |
| ---- | ----- | ---- | ---- |
| `--debug` | `-d` | sets debug mode | false |
| `--json` | `-j` | log output json format | false |
| `--log-file-path` | `-l` | custom e1s log file path | ${TMPDIR}e1s.log |
| `--readonly` |  | sets read only mode | false |
| `--refresh` | `-r` | specify the default refresh rate as an integer (sec) (default 30, set -1 to stop auto refresh) | 30 |
| `--shell` | `-s` | specify ecs exec ssh shell | /bin/sh |


Make sure you have the AWS CLI installed and properly configured with the necessary permissions to access your ECS resources.

Using default profile

```bash
$ e1s
```

Using my-profile profile, us-east-1 region

```bash
$ AWS_PROFILE=my-profile AWS_REGION=us-east-1 e1s
```

Using read only, debug, stop auto refresh with a custom log path json output

```bash
$ e1s --readonly --debug --refresh -1 --log-file-path /tmp/e1s.log --json
```

### Key Bindings

Press `?` to check overall key bindings, top right corner to check current resource specific hot keys.

### Development

```bash
go run main.go --debug --log-file-path /tmp/e1s.log
```

```bash
tail -f /tmp/e1s.log
```

## Feature Requests & Bug Reports

If you have any feature requests or bug reports, please submit them through GitHub [Issues](https://github.com/keidarcy/e1s/issues).

## Publish new version

- Bump version
- `make tag`

## Thanks

- [tview](https://github.com/rivo/tview)
- [k9s](https://github.com/derailed/k9s)
- [ecsview](https://github.com/swartzrock/ecsview)

## License

MIT
