# Redis host memory overcommit

Redis requires `vm.overcommit_memory = 1` on the Linux kernel running it so that
background saves, append-only file (AOF) rewrites, and replication can fork
reliably. This also applies to the current Compose image, `redis:7.4-alpine`,
which has AOF persistence enabled. Changing the Redis image does not configure
the host kernel. See [Redis administration](https://redis.io/docs/latest/operate/oss_and_stack/management/admin/).

This is the host requirement reported in
[issue #20](https://github.com/nima-abdpoor/brain-blitz/issues/20).

## Linux Docker Engine host

Run these commands on the Linux machine running the Docker daemon. If using a
remote Docker context, configure the remote machine.

Check the current value:

```sh
sysctl vm.overcommit_memory
```

Install the provided configuration to keep the setting across host reboots,
then load that file to apply it immediately without rebooting:

```sh
sudo install -m 0644 infra/sysctl/99-brain-blitz-redis.conf /etc/sysctl.d/99-brain-blitz-redis.conf
sudo sysctl -p /etc/sysctl.d/99-brain-blitz-redis.conf
sysctl vm.overcommit_memory
```

The expected result is `vm.overcommit_memory = 1`. Repeating these commands
updates the same file without appending duplicate settings. Check the value
again after a reboot; another host provisioning rule can override it.

The setting affects the whole Linux host. It changes memory allocation policy;
it does not increase available RAM or remove container memory limits. Continue
to allow memory headroom for Redis persistence.

## Docker Desktop on macOS or Windows

Redis uses Docker Desktop's Linux VM kernel. Changing a macOS sysctl or the
configuration of an unrelated WSL distribution does not change that kernel.
Start Docker Desktop and Redis, then check the value from the container:

```sh
docker compose exec -T bb-redis cat /proc/sys/vm/overcommit_memory
```

If it prints `1`, no change is needed. If it prints `0` or `2`, the following
one-time administrative command sets the value on Docker's Linux VM:

```sh
docker run --rm --privileged --network none --entrypoint sysctl redis:7.4-alpine -w vm.overcommit_memory=1
```

This helper changes the VM-wide setting and exits; Redis itself continues to
run with its normal permissions. The change is temporary. Recheck it after
Docker Desktop or its VM restarts or resets, and reapply it if needed.

## Kubernetes nodes

Apply the Linux host configuration above to every node eligible to run the
Redis pod, including replacement nodes, using your node provisioning process.
Verify the effective value from the running pod:

```sh
kubectl exec deployment/redis-deployment -- cat /proc/sys/vm/overcommit_memory
```

It must print `1`. This setting is not namespaced, so it cannot be configured
through Compose `sysctls` or a pod's `securityContext.sysctls`. See
[Docker's sysctl restrictions](https://docs.docker.com/reference/compose-file/services/#sysctls)
and [Kubernetes node-level sysctls](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/).

## Verify Redis

After starting the Compose deployment, check the effective setting and Redis
connectivity:

```sh
docker compose exec -T bb-redis cat /proc/sys/vm/overcommit_memory
docker compose exec -T bb-redis redis-cli ping
```

The expected results are `1` and `PONG`. Future Redis startups should no longer
log `Memory overcommit must be enabled`. Applying the setting takes effect
immediately; Redis does not need a restart. An older warning remains in the
container's historical logs.
