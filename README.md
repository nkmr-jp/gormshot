# GormShot

GormShot adds Snapshot Testing for [GORM](https://gorm.io/index.html). <br>
Inspired by [StoryShots](https://github.com/storybookjs/storybook/tree/main/addons/storyshots/storyshots-core).

## Install

```sh
go get -u github.com/nkmr-jp/gormshot
```
```sh
# If you want to use the latest feature.
go get -u github.com/nkmr-jp/gormshot@develop
```

## Usage

See: [gormshot_test.go](gormshot_test.go)


## Snapshot file
Snapshot file is saved as [JSON Lines](https://jsonlines.org/) format file, so you can use [jq](https://stedolan.github.io/jq/) command to show pretty output. like this.

### example 1 (compact and colored output)
```shell
cat .snapshot/TestAssert__value_is_match.jsonl | jq -c
```
```json lines
{"Name":"Carol","Age":31}
{"Name":"Bob","Age":45}
{"Name":"Alice","Age":20}
```

### example 2 (formatted output)
```shell
cat .snapshot/TestAssert__value_is_match.jsonl | jq 
```
```json lines
{
  "Name": "Carol",
  "Age": 31
}
{
  "Name": "Bob",
  "Age": 45
}
{
  "Name": "Alice",
  "Age": 20
}
```

### example 3 (select attribute)
```shell
cat .snapshot/TestAssert__value_is_match.jsonl | jq .Name 
```
```json lines
"Carol"
"Bob"
"Alice"
```