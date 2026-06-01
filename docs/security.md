# Security

myHarness prioritizes autonomous coding workflows over isolation.

Mounted directories are fully available to the agent. The container is not a
hostile-code sandbox, and non-root execution does not make mounted files safe.

The default runtime maps the container `harness` user to the invoking host UID
and GID. This prevents common root-owned workspace file problems, but any
read-write mount remains available to the agent process.

Use narrow mounts, private state, and separate workspaces when handling
sensitive code or credentials.

Vaka can reduce network blast radius. It does not protect mounted files,
credentials, shell history, or any other data intentionally mounted into the
container.
