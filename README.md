# protoc-gen-graphql

`protoc-gen-graphql` is a highly customizable Protobuf compiler plugin to generate GraphQL schema definition language (SDL) files.

## Usage

First, download the plugin and place the executable in your path.
Then, run the Protobuf compiler with the `--graphql_out` flag to enable this plugin.
Set the value of this flag to the directory you want the SDL files to be generated into.

```shell script
protoc -I . --graphql_out=output_dir path/to/file.proto ...
```

### Parameters

The SDL generation can be customized by passing optional parameters to the plugin.
Parameters are specified using a comma separated list of parameters before the output directory, separated by a colon.
For example:

```shell script
protoc -I . --graphql_out=root_type_prefix=GRPC,null_wrappers:output_dir path/to/file.proto ...
```

Note that parameter settings apply to all the generated GraphQL types and files.
For settings that apply to single Protobuf objects, like messages and fields, refer to the Protobuf options section below.

Available parameters are:

| Key | Values | Default | Description |
| --- | --- | --- | --- |
| `field_name` | `lower_camel_case`, `preserve` | `lower_camel_case` | Transformation from Protobuf field names to GraphQL field names. Default is lowerCamelCase. Use `preserve` to use the Protobuf name as-is. |
| `trim_prefix` | string | | Trims the provided prefix from all generated GraphQL type names. Useful if your Protobuf package names have a common prefix you want to omit. |
| `root_type_prefix` | string | | If set, a gRPC service's mapped query and mutation types will extend some custom root type with name given by the provided prefix plus `Query` or `Mutation`. Set to empty string to extend the root `Query` and `Mutation` types. |
| `input_mode` | `all`, `service`, `none` | `none` | The input mode determines what GraphQL input objects will be generated. `all` will generate an input object for each Protobuf message. `service` will only generate inputs for messages that are transitively used in each gRPC methods' request messages. `none` will not generate any input objects. |
| `null_wrappers` | bool | `false` | If true, well known wrapper types (e.g. `google.protobuf.StringValue`) will be mapped to nullable GraphQL scalar types instead of the corresponding object type. |
| `js_64bit_type` | `string`, `number` | `number` | Whether to use a `String` or `Float` scalar type when mapping 64bit Protobuf types (`int64`, `uint64`, `sint64`, `fixed64`, `sfixed64`). |
| `timestamp` | string | | GraphQL type name to use for the well known `google.protobuf.Timestamp` type. |
| `duration` | string | | GraphQL type name to use for the well known `google.protobuf.Duration` type. |

### Protobuf options

[Protobuf options file](protobuf/graphql/options.proto)

TODO

## Protobuf to GraphQL mapping

TODO

### Messages

#### Maps

#### Oneofs

### Enums

### Services
