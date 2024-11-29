[
  {
    name: "impulse-backend",
    command: "go",
    args: ["run", "cmd/main.go"],
    autorestart: false,
    watch: "\\.go$",
    tags: ["impulse"],
  },
  {
    name: "impulse-frontend",
    command: "bun",
    args: ["dev"],
    cwd: "frontend",
    tags: ["impulse"],
  },
]
