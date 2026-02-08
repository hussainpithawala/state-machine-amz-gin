# API Documentation

This document describes how to use the API documentation files generated for the State Machine API.

## Files Generated

1. **postman_collection.json** - Postman collection with all API endpoints
2. **openapi.json** - OpenAPI 3.0 specification
3. **API_DOCUMENTATION.md** - This file

## Using the Postman Collection

### Import into Postman

1. Open Postman
2. Click on "Import" button
3. Select the `postman_collection.json` file
4. The collection "State Machine API" will be imported with all endpoints organized into folders

### Collection Variables

The collection includes the following variables that you can customize:

- `baseUrl`: Default is `http://localhost:8080/state-machines/api/v1`
- `stateMachineId`: Placeholder for state machine IDs
- `executionId`: Placeholder for execution IDs

To set these variables:
1. Click on the collection name
2. Go to "Variables" tab
3. Update the "Current Value" for each variable

### Example Usage Flow

1. **Health Check** - Verify the service is running
2. **Create State Machine** - Create a new state machine definition
3. **Start Execution** - Start an execution for your state machine
4. **Get Execution** - Check the execution status
5. **Get Execution History** - View the state transition history

## Using the OpenAPI Specification

### Accessing via API

The OpenAPI specification is automatically served by the application at:

```
GET http://localhost:8080/state-machines/api/v1/openapi.json
```

### Using with Swagger UI

You can use the OpenAPI specification with Swagger UI:

1. Go to https://editor.swagger.io/
2. Click "File" → "Import URL"
3. Enter: `http://localhost:8080/state-machines/api/v1/openapi.json` (when your server is running)

Or import the file directly:
1. Go to https://editor.swagger.io/
2. Click "File" → "Import file"
3. Select the `openapi.json` file

### Using with Postman

Postman can also import OpenAPI specifications:

1. Open Postman
2. Click "Import"
3. Select the `openapi.json` file
4. Postman will generate a collection from the OpenAPI spec

### Using with API Clients

Many API client libraries can generate code from OpenAPI specifications:

- **JavaScript/TypeScript**: Use `openapi-generator-cli` or `swagger-codegen`
- **Python**: Use `openapi-python-client`
- **Java**: Use `swagger-codegen` or `openapi-generator`
- **Go**: Use `oapi-codegen`

Example with openapi-generator:
```bash
openapi-generator-cli generate -i openapi.json -g typescript-axios -o ./generated-client
```

## API Endpoints Overview

### Health & Monitoring
- `GET /health` - Health check

### State Machines
- `POST /state-machines` - Create state machine
- `GET /state-machines` - List state machines
- `GET /state-machines/:stateMachineId` - Get state machine

### Executions
- `POST /state-machines/:stateMachineId/executions` - Start execution
- `GET /state-machines/:stateMachineId/executions` - List executions
- `GET /state-machines/:stateMachineId/executions/count` - Count executions
- `GET /executions/:executionId` - Get execution
- `DELETE /executions/:executionId` - Stop execution
- `GET /executions/:executionId/history` - Get execution history

### Batch Operations
- `POST /state-machines/:stateMachineId/executions/batch` - Execute batch

### Queue Operations
- `POST /queue/enqueue` - Enqueue execution

### Message & Resume Operations
- `POST /executions/:executionId/resume` - Resume execution
- `POST /state-machines/:stateMachineId/resume-by-correlation` - Resume by correlation
- `GET /state-machines/:stateMachineId/waiting` - Find waiting executions

## Testing the API

### Using curl

```bash
# Health check
curl http://localhost:8080/state-machines/api/v1/health

# Get OpenAPI spec
curl http://localhost:8080/state-machines/api/v1/openapi.json

# Create a state machine
curl -X POST http://localhost:8080/state-machines/api/v1/state-machines \
  -H "Content-Type: application/json" \
  -d '{
    "id": "hello-world",
    "name": "Hello World",
    "definition": {
      "Comment": "A Hello World example",
      "StartAt": "HelloWorld",
      "States": {
        "HelloWorld": {
          "Type": "Pass",
          "Result": "Hello World!",
          "End": true
        }
      }
    },
    "version": "1.0"
  }'

# Start an execution
curl -X POST http://localhost:8080/state-machines/api/v1/state-machines/hello-world/executions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "execution-1",
    "input": {}
  }'
```

## Authentication

Currently, the API does not require authentication. If you add authentication in the future, update:
1. The Postman collection with auth settings
2. The OpenAPI spec with security schemes
3. This documentation with authentication instructions

## Support

For issues or questions about the API, please refer to the main project documentation.
