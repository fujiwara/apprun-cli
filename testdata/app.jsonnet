local must_env = std.native('must_env');
local env = std.native('env');
{
  components: [
    {
      deploy_source: {
        container_registry: {
          image: 'debian:latest',
          password: must_env('REGISTRY_PASSWORD'),
          server: 'registry.example.com',
          username: 'user',
        },
      },
      env: [
        {
          key: 'FOO',
          value: 'BAR',
        },
      ],
      max_cpu: '0.1',
      max_memory: '1Gi',
      name: 'test',
      probe: {
        http_get: {
          headers: [
            {
              name: 'X-Test',
              value: 'test',
            },
          ],
          path: '/',
          port: 80,
        },
      },
    },
  ],
  max_scale: 2,
  min_scale: 1,
  name: env('APPLICATION_NAME', 'test'),
  port: 80,
  timeout_seconds: 10,
}
