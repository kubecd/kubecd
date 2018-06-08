from os import path


def resolve_file_path(file_path: str, relative_to_file: str = None, relative_to_dir: str = None) -> str:
    """
    Resolve a relative file path (or return absolute ones as-is).
    :param file_path:         the relative path
    :param relative_to_file:  a file this path is relative to (note: this is a file, so we
                              will use the file's dirname, not the file itself)
    :param relative_to_dir:   a directory this path is relative to
    :return: a full path name
    """
    if path.isabs(file_path):
        return file_path
    if relative_to_file is not None:
        return path.join(path.dirname(relative_to_file), file_path)
    if relative_to_dir is None:
        raise ValueError('resolve_file_path requires either relative_to_file or relative_to_dir arg!')
    return path.join(relative_to_dir, file_path)


def kube_context(env_name: str):
    return 'env:' + env_name
