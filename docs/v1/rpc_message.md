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
This design allows different authentication and signing strategies without modifying the core RPC message format. Signatures are state-specific, not RPC-specific. State signatures are included within the `state` objects passed as parameters, where they logically belong.
