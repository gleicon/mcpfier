# MCPFier

MCPFier is a configurable wrapper that enables executing any command or script using a standardized command pattern, allowing for dynamic execution and integration with larger systems. It supports configuration via a `YAML` file, enabling users to define custom commands easily. The current implementation focuses on local execution through stdin/stdout, with planned features for server execution and containerization.

## Features
- Configurable command execution using YAML files
- Local stdin/stdout execution
- Extensible for server mode and containerized execution

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/gleicon/mcpfier.git
    cd mcpfier
    ```

2. Build the project:
    ```bash
    go build -o mcpfier
    ```

## Configuration

MCPFier uses a YAML configuration file named `config.yaml` to define custom commands. Each command includes a name, the script or executable to run, and any necessary arguments.

Example configuration file (`config.yaml`):
```yaml
commands:
  - name: my-pipeline-execution
    script: /path/to/my_pipeline.py
    args: ["--option1", "value1", "--option2", "value2"]
```

### Define a Command

- **name**: A unique identifier for the command.
- **script**: The path to the script or executable to be run.
- **args**: (Optional) A list of arguments to pass to the script.

## Usage

1. Add your commands to the `config.yaml` as described above.
2. Run the `mcpfier` executable with the command name you wish to execute:
   ```bash
   ./mcpfier my-pipeline-execution
   ```

This will execute the configured script and pass any specified arguments.

## Testing

MCPFier comes with unit tests to ensure its functionality. Run the tests using:
```bash
go test ./...
```

## Future Features

- Server mode to manage and execute commands from remote requests
- Containerized execution to provide isolated environments for each command run

## Contributing

We welcome contributions to enhance MCPFier. Please submit pull requests or create issues to discuss potential changes.
