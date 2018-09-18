from typing import List

from semantic_version import Version, validate, Spec


def is_semver(version: str) -> bool:
    try:
        return validate(normalize(version))
    except ValueError:
        return False


def normalize(version: str) -> str:
    return version[1:] if version.startswith('v') else version


def parse(version: str) -> Version:
    version = normalize(version)
    return Version.coerce(version)


def filter_semver_tags(tags: List[str]) -> List[str]:
    """
    Filter a list of tags to only the tags that follow a semantic versioning scheme.
    :param tags:  input tags
    :return:  those input tags that are semver
    """
    output = []
    for tag in tags:
        if validate(normalize(tag)):
            output.append(tag)
    return output


def best_upgrade(current: Version, candidates: List[Version], track: str='MajorVersion'):
    if track == 'PatchLevel':
        spec = Spec('>{current},<{next_minor}'.format(current=str(current), next_minor=str(current.next_minor())))
    elif track == 'MinorVersion':
        spec = Spec('>{current},<{next_minor}'.format(current=str(current), next_minor=str(current.next_major())))
    elif track == 'MajorVersion':
        spec = Spec('>{current}'.format(current=str(current)))
    else:
        raise ValueError('unsupported "track": {track}'.format(track=track))
    return spec.select(candidates)


def is_wanted_upgrade(current: Version, candidate: Version, track: str='MajorVersion'):
    best = best_upgrade(current, [candidate], track)
    return best == candidate
