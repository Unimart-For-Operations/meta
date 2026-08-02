# ${{ values.name }}

${{ values.description }}

Elixir/Phoenix microservice scaffolded from the `phoenix-service` template.

## Development

All tooling comes from the Nix devShell:

```bash
nix develop
mix deps.get
mix phx.server     # http://localhost:${{ values.port }}
mix test
```

## Container

```bash
docker build -t ${{ values.name }}:latest .
```

Runtime env: `SECRET_KEY_BASE` (required), `PHX_HOST`, `PORT`.
