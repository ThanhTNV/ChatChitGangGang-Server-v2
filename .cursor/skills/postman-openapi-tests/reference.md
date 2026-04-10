# Postman Collection v2.1 — reference for this skill

## Minimal collection shape

Use **Collection v2.1** (`info.schema` should be Postman’s v2.1 URL). Top-level fields commonly include:

- `info`: `name`, `_postman_id`, `description`, `schema`
- `variable`: collection variables (`base_url`, `access_token`)
- `auth`: optional `type: bearer`, `bearer: [{ key: token, value: {{access_token}}, type: string }]`
- `item`: array of folders or requests

## Folder item

```json
{
  "name": "Chat",
  "item": [
    {
      "name": "List channels",
      "request": {
        "method": "GET",
        "header": [],
        "url": "{{base_url}}/v1/channels"
      },
      "event": [
        {
          "listen": "test",
          "script": {
            "exec": [
              "pm.test('status 200', function () { pm.response.to.have.status(200); });",
              "const j = pm.response.json();",
              "pm.test('has channels array', function () { pm.expect(j).to.have.property('channels'); });"
            ],
            "type": "text/javascript"
          }
        }
      ]
    }
  ]
}
```

## Environment file

```json
{
  "name": "internal-comm local",
  "values": [
    { "key": "base_url", "value": "http://localhost:8080", "enabled": true },
    { "key": "access_token", "value": "", "enabled": true }
  ],
  "_postman_variable_scope": "environment"
}
```

## Naming

- Collection file: `internal-comm.postman_collection.json`
- Environment: `local.postman_environment.json`
- Doc: `README.md` (test plan + workflows)

## Mapping OpenAPI → Postman

- `operationId` → optional request `name` suffix for clarity.
- `tags[0]` → folder name (e.g. `Health`, `Chat`, `Profile`).
- `security: bearerAuth` → inherit collection bearer or set per-request **Authorization** header.
- Request body `$ref` → example JSON in raw body (from schema `example` or sensible default).
