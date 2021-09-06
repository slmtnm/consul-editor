# consul-editor

Utility for editing Consul KV storage via local editor. All folder 
hierachy is nested in YAML format, so editing deep KV trees is made 
more convenient with this utility than same done via UI.

## Installation
To install consul-editor, run:
```console
  $ go install github.com/slmtnm/consul-editor@latest
```
Requires go 1.17 or higher.

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