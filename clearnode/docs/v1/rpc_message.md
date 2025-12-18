## RPC Message Format

All messages exchanged between clients and clearnodes follow this standardized format:

### Request Message

```json
{
  "req": [REQUEST_ID, METHOD, PARAMETERS, TIMESTAMP],
}
```

### Response Message

```json
{
  "res": [REQUEST_ID, METHOD, RESPONSE_DATA, TIMESTAMP],
}
```

The structure breakdown:

- `REQUEST_ID`: A unique identifier for the request/response pair (`uint64`)
- `METHOD`: The name of the method being called (`string`)
- `PARAMETERS`/`RESPONSE_DATA`: An object of parameters/response data (`map[string]any`)
- `TIMESTAMP`: Unix timestamp of the request/response in milliseconds (`uint64`)
