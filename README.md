# consul-editor

Utility for editing Consul KV storage via local editor. All folder 
hierachy is converted into YAML format, so editing deep KV trees is made 
more convenient with this utility than same done via UI.

## Installation
To compile and install consul-editor, run:
```console
  $ go install github.com/slmtnm/consul-editor@latest
```

To download static binary, go to [Releases](https://github.com/slmtnm/consul-editor/releases) page.

## Configuration

Application is configured via standard consul environment variables:

* **CONSUL_HTTP_ADDR** - HTTP address of consul host (default "localhost:9200")
* **CONSUL_HTTP_TOKEN** - ACL Token

## Example
Assume having following folder structure in Consul KV:
```
root/
  a/a_key
  b/
    c/c.json
```

After running `consul-editor /root` your local editor (specified by 
EDITOR environment variable) will be opened with this content:
```yaml
root:
  a:
    a_key: <content of a_key>
  b:
    c:
      c.json: <content of c.json>
```

After modifying and saving file all changes will be uploaded in consul. It
means you can create/update/delete keys and corresponding values.

## Testing checklist

- [x] Ability to change text to text
- [x] Ability to remove key completely
- [x] Ability to remove folder completely
- [x] Ability to change text key to folder key
- [x] Ability to change folder key to text key