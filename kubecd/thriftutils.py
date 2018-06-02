from thrift.Thrift import TType


class SchemaError(BaseException):
    pass


def load_yaml_with_schema(yaml_file: str, schema):
    with open(yaml_file, 'r') as fd:
        from ruamel import yaml
        try:
            return to_thrift_object(yaml.safe_load(fd), schema, '')
        except SchemaError as e:
            raise SchemaError('{}: {}'.format(yaml_file, str(e))) from e


def to_thrift_object(in_dict: dict, schema, obj_path: str):
    if not isinstance(in_dict, dict):
        raise SchemaError('unexpected type at "{}", expected dict, found {}'.format(obj_path, type(in_dict).__name__))

    ctor_args = {}
    for field_spec in schema.thrift_spec:
        if field_spec is None:
            continue
        t_type = field_spec[1]
        t_field = field_spec[2]
        if t_field in in_dict:
            new_path = obj_path
            if len(new_path) > 0:
                new_path += '.'
            new_path += t_field
            ctor_args[t_field] = to_thrift_type(in_dict[t_field], t_type, field_spec[3], new_path)
    extra_keys = set(in_dict.keys()) - set(ctor_args.keys())
    if len(extra_keys) > 0:
        raise SchemaError('extraneous keys for "{}": {}'.format(schema.__name__, ', '.join(extra_keys)))

    return schema(**ctor_args)


def to_thrift_type(value, t_type, t_subtype, obj_path: str):
    """
    Convert a Python value to a Thrift value
    :param value: python value
    :param t_type:
    :param t_subtype:
    :param obj_path:
    :return: Thrift-ified version of the input value
    :raises SchemaError: if the input value does not match the schema
    """
    if t_type == TType.I08 or t_type == TType.I16 or t_type == TType.I32 or t_type == TType.I64:
        return int(value)
    elif t_type == TType.BOOL:
        return bool(value)
    elif t_type == TType.DOUBLE:
        return float(value)
    elif t_type == TType.STRING:
        # t_subtype is encoding
        return str(value)
    elif t_type == TType.STRUCT:
        return to_thrift_object(value, t_subtype[0], obj_path)
    elif t_type == TType.LIST:
        if not isinstance(value, list):
            raise SchemaError('value is not a list at %s' % obj_path)
        return to_thrift_list(value, t_subtype, obj_path)
    elif t_type == TType.MAP:
        if not isinstance(value, dict):
            raise SchemaError('value is not a dict at %s' % obj_path)
        return to_thrift_map(value, t_subtype, obj_path)
    else:
        raise SchemaError('unknown or unsupported thrift type %d at %s' % (t_type, obj_path))


def to_thrift_list(in_list: list, spec: tuple, list_path: str):
    return [to_thrift_type(v, spec[0], spec[1], '{0}[{1}]'.format(list_path, i)) for i, v in enumerate(in_list)]


def to_thrift_map(in_map: dict, spec: tuple, map_path: str):
    # spec: (TType.STRING, 'UTF8', TType.STRING, 'UTF8', False)
    return {
        to_thrift_type(k, spec[0], spec[1], map_path):
            to_thrift_type(v, spec[2], spec[3], '{0}[{1}]'.format(map_path, k)) for k, v in in_map.items()
    }
