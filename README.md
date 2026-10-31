# Ramekin advanced example

Releases of a full-stack [TodoMVC](https://todomvc.com) app (web page, JSON API and SQLite database in one small Go server), packaged with [Ramekin](https://github.com/ascension-association/ramekin) for a gokrazy device. It's the project from the Advanced Example in Ramekin's README, which walks through it step by step.

The example shows:

- A two-stage Apptainer definition file. The build stage cross-compiles on the workstation, so nothing runs under emulation, and the final stage is a 10 MiB image on `cgr.dev/chainguard/wolfi-base`.
- A database in `/data`, which survives updates and rollbacks.
- A health check that exercises the app (`--health-http http://127.0.0.1/api/todos`), so a release counts as healthy only once the API answers and the database is open.

## What's here

- The source: `main.go` (the whole server), `index.html` (the page it embeds), `example.def` (the definition file) and `ramekin.yaml` (the project, which publishes to this repository).
- [Releases](https://github.com/ascension-association/ramekin-advanced-example/releases), one per version, each with the signed, encrypted image (`todomvc.sif.age`), its SBOM and a signed manifest. `todomvc-1.0.0` titles the page "Ramekin todos", and `todomvc-1.1.0` (the source here) changes it to "Ramekin fleet todos".
- `todomvc/stable.json` on `main`, the signed channel pointer that devices poll. `ramekin publish` and `ramekin promote` write it, so don't edit it by hand. The project sets `releases.branch: main` because this repository holds only releases.

The images are encrypted with the publisher's fleet key, so they only install on devices provisioned from that project. To run the app yourself, build it with your own keys:

```sh
git clone https://github.com/ascension-association/ramekin-advanced-example && cd ramekin-advanced-example
ramekin keys generate
ramekin build
ramekin device init todo --no-updates --health-http http://127.0.0.1/api/todos
gok -i todo overwrite --full /dev/sdX
ramekin push todo
curl http://todo/api/todos
```

To publish your own releases, set `releases.github` in `ramekin.yaml` to a repository of yours.
