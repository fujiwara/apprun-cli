# apprun-cli

apprun-cli is a command line interface for AppRun β Sakura Cloud.

This is an unofficial tool.

See also https://manual.sakura.ad.jp/cloud/manual-sakura-apprun.html

## Usage

```
Usage: apprun-cli <command> [flags]

Flags:
  -h, --help     Show context-sensitive help.
      --debug    Enable debug mode ($DEBUG)

Commands:
  init --name=STRING [flags]
    Initialize files from existing application

  deploy <application> [flags]
    Deploy an application

  list [flags]
    List applications

  diff <application> [flags]
    Show diff of applications

  render <application> [flags]
    Render application

  status <application> [flags]
    Show status of applications

  delete <application> [flags]
    Delete the application

  versions <application> [flags]
    Manage versions of the application

  traffics <application> [flags]
    Manage traffics of application

  user <operation> [flags]
    Manage apprun user

Run "apprun-cli <command> --help" for more information on a command.
```

## Installation

```
$ go install github.com/fujiwara/apprun-cli/cmd/apprun-cli@latest
```

or `brew install fujiwara/tap/apprun-cli`

or download from [Releases](https://github.com/fujiwara/apprun-cli/releases)


## Configuration

apprun-cli reads configuration from environment variables.

- `SAKURACLOUD_ACCESS_TOKEN`
- `SAKURACLOUD_ACCESS_TOKEN_SECRET`

## Examples

### List

`apprun-cli list` lists all applications.

```console
$ apprun-cli list
{
  "created_at": "2024-12-31T14:07:36.90833+09:00",
  "id": "ff7abff9-eacc-44ef-b92d-ab0f65fe8ed4",
  "name": "example",
  "public_url": "https://app-ff7abff9-eacc-44ef-b92d-ab0f65fe8ed4.ingress.apprun.sakura.ne.jp",
  "status": "Success"
}
{
  "created_at": "2024-12-31T14:27:39.918056+09:00",
  "id": "e8e32d63-2260-4093-aef2-c277b375de4e",
  "name": "example2",
  "public_url": "https://app-e8e32d63-2260-4093-aef2-c277b375de4e.ingress.apprun.sakura.ne.jp",
  "status": "Success"
}
```

### Initialize

`apprun-cli init --name=example` shows the application definition to stdout.

You can save the output to a file and edit it to deploy the application.

Note: The password is not set in the output.

```
Usage: apprun-cli init --name=STRING [flags]

Initialize files from existing application
Flags:
      --name=STRING    name of the application to init
      --jsonnet        Use jsonnet to generate files
```

```console
$ apprun-cli init --name example
2024/12/31 14:18:05 INFO initializing app=example
2024/12/31 14:18:05 INFO found id=04946938-e424-463d-b6ad-9ce5e03de55d
{
  "components": [
    {
      "deploy_source": {
        "container_registry": {
          "image": "example.sakuracr.jp/example:latest",
          "password": null,
          "server": "example.sakuracr.jp",
          "username": "apprun"
        }
      },
      "env": [
        {
          "key": "FOO",
          "value": "baz"
        }
      ],
      "max_cpu": "1",
      "max_memory": "2Gi",
      "name": "example",
      "probe": {
        "http_get": {
          "path": "/",
          "port": 8080
        }
      }
    }
  ],
  "max_scale": 2,
  "min_scale": 0,
  "name": "example",
  "port": 8080,
  "timeout_seconds": 60
}
```

### Deploy

`apprun-cli deploy app.json` deploys the application.

app.json or app.jsonnet is the output of `apprun-cli init`.

We recommend to use Jsonnet format to read environment variables.

When the application is not found, `apprun-cli deploy` creates a new application.

```jsonnet
local must_env = std.native('must_env');
{
  components: [
    {
      deploy_source: {
        container_registry: {
          image: 'example.sakuracr.jp/example:latest',
          password: must_env('REGISTRY_PASSWORD'),
          server: 'example.sakuracr.jp',
          username: 'apprun',
        },
      },
      env: [
        {
          key: 'FOO',
          value: 'bar',
        },
      ],
      max_cpu: '1',
      max_memory: '2Gi',
      name: 'example',
      probe: {
        http_get: {
          path: '/',
          port: 8080,
        },
      },
    },
  ],
  max_scale: 2,
  min_scale: 0,
  name: 'example',
  port: 8080,
  timeout_seconds: 60,
}
```

### Diff

`apprun-cli diff app.jsonnet` shows the difference between the current application and the definition file.

```diff
--- 0b523faa-b0de-4c26-bc42-a3ff500b9367
+++ app.jsonnet
@@ -12,7 +12,7 @@
       "env": [
         {
           "key": "FOO",
-          "value": "bar"
+          "value": "baz"
         }
       ],
       "max_cpu": "1",
```

### Render

`apprun-cli render app.jsonnet` shows the rendered application.

### Status

`apprun-cli status app.jsonnet` shows the status of the application.

```console
{
  "created_at": "2024-12-31T14:27:39.918056+09:00",
  "id": "e8e32d63-2260-4093-aef2-c277b375de4e",
  "name": "example2",
  "public_url": "https://app-e8e32d63-2260-4093-aef2-c277b375de4e.ingress.apprun.sakura.ne.jp",
  "status": "Success"
}
```

### Delete

`apprun-cli delete app.jsonnet` deletes the application.

### Versions

`apprun-cli versions app.jsonnet` manages the versions of the application.

```
Flags:
      --id=STRING    Show the detailed information of the specified version id
      --delete       Delete the specified version id
      --force        Force delete without confirmation
```

If no flags are specified, it shows the list of versions.

`--id` shows the detailed information of the specified version id.

`--delete` deletes the specified version id. Requires `--id` flag.

`--force` forces delete without confirmation. default is false.

### Traffics

`apprun-cli traffics app.jsonnet` manages the traffics of the application.

```
Flags:
      --versions=KEY=VALUE,...    Traffic percentage for each version
```

If no flags are specified, it shows the current list of traffics.

`--versions` sets the traffic percentage for each version.

The value is a comma-separated list of `version_name=percentage`.
The version name is shown in the list of `versions` command.

For example, the following command sets the traffic to 50% for each version.

```console
$ apprun-cli traffics app.jsonnet \
  --versions app-0b523faa-b0de-4c26-bc42-a3ff500b9367=50,app-ff7abff9-eacc-44ef-b92d-ab0f65fe8ed4=50
```

The sum of the traffic percentage must be 100.

### User

`apprun-cli user` manages the user for AppRun.

Before using `apprun-cli`, you need to create a user only at once.

- `apprun-cli user create` creates a new user.
- `apprun-cli user read` confirm the user is created.

## LICENSE

MIT

## Author

Fujiwara Shunichiro (@fujiwara)
