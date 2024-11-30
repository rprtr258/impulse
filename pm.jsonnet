[
  {
    name: "impulse-backend",
    command: "go",
    args: ["run", "cmd/main.go"],
    autorestart: false,
    watch: "\\.go$",
    cwd: "backend",
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
