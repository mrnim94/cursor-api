## cursor-api

A minimal HTTP bridge that shells out to `cursor-agent` to generate responses. Ships with a Docker image that installs the Cursor Agent CLI inside the container.

### Requirements
- **Docker** 20+
- A **Cursor API key** (set via env or header)

### Build the Docker image
```bash
docker build -t cursor-api .
```

### Run the container
- Option 1: set the API key for all requests (no need to send it per request)
```bash
docker run --rm -p 1994:1994 -e CURSOR_API_KEY=YOUR_KEY cursor-api
```

- Option 2: run without a container-level key and send it per request via header
```bash
docker run --rm -p 1994:1994 cursor-api
```

### Make a request (Linux/macOS)
```bash
curl -sS \
  -X POST http://localhost:1994/v1beta/models/cursor \
  -H 'Content-Type: application/json' \
  -H 'x-cursor-api-key: YOUR_KEY' \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{ "text": "Say hi in one short sentence." }]
      }
    ]
  }'
```

### Make a request (Windows PowerShell)
```powershell
curl.exe -sS \
  -X POST http://localhost:1994/v1beta/models/cursor \
  -H "Content-Type: application/json" \
  -H "x-cursor-api-key: YOUR_KEY" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{ "text": "Say hi in one short sentence." }]
      }
    ]
  }'
```

### Example response
```json
{
  "model": "cursor",
  "candidates": [
    {
      "content": {
        "parts": [
          { "text": "Hello!" }
        ]
      }
    }
  ]
}
```

### Endpoint
- `POST /v1beta/models/:model`
- Request body shape (subset of Gemini-style schema used by this bridge):
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [ { "text": "your prompt" } ]
    }
  ]
}
```

### Auth
- Per-request: add header `x-cursor-api-key: YOUR_KEY`
- Or container-wide: run with `-e CURSOR_API_KEY=YOUR_KEY`

### Configuration
- `CURSOR_API_KEY`: API key for Cursor Agent
- `CURSOR_AGENT_CMD` (optional): override the `cursor-agent` binary path, e.g. `/home/appuser/.local/bin/cursor-agent`

### Notes
- The Docker image installs the Cursor Agent CLI during build and symlinks it into `/usr/local/bin`.
- If you see `Authentication required`, provide the API key via env or header as shown above.


