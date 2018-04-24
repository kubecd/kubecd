import json
import subprocess
from collections import defaultdict
from typing import Tuple, Dict, List
import dateutil.parser

import requests
import semantic_version

from .environments import Environment, key_is_in_values, lookup_value


def semver_normalize(version: str):
    return version[1:] if version.startswith('v') else version


def semver_parse(version: str) -> semantic_version.Version:
    version = semver_normalize(version)
    return semantic_version.Version.coerce(version)


def parse_docker_timestamp(timestamp: str) -> int:
    """
    Parse a timestamp of the format "2015-02-13T10:04:55.62062787Z" and
    return unix time
    :param timestamp:
    :return: unix timestamp (seconds since 1970-01-01 00:00:00 UTC)
    """
    return int(dateutil.parser.parse(timestamp).timestamp())


def parse_image_repo(repo: str) -> Tuple[str, str]:
    """
    If 'repo' contains a host prefix, return a tuple of that and the rest.
    If no host is found, we use "docker.io" instead.
    :param repo: full repository string
    :return: a tuple of (registry host, repository)
    """
    elem = repo.split('/')
    if len(elem) > 1 and '.' in elem[0]:
        return elem[0], '/'.join(elem[1:])
    return 'docker.io', repo


def get_tags_for_gcr_image(registry: str, repo: str) -> Dict[str, int]:
    full_repo = registry + '/' + repo
    gcp_project = repo.split('/')[0]
    cmd = ['gcloud', 'container', 'images', 'list-tags', full_repo, '--project', gcp_project, '--format', 'json']
    gcr_response = json.loads(subprocess.check_output(cmd).decode('utf-8'))
    response = {}
    for img_tag in gcr_response:
        if 'timestamp' in img_tag:
            timestamp = parse_docker_timestamp(img_tag['timestamp']['datetime'])
        else:
            timestamp = 0
        if 'tags' in img_tag:
            # print(img_tag)
            for tag in img_tag['tags']:
                response[tag] = timestamp
    return response


def get_tags_for_docker_hub_image(repo: str) -> Dict[str, int]:
    response = {}
    if '/' not in repo:
        repo = 'library/' + repo
    api_url = 'https://registry.hub.docker.com/v2/repositories/{repo}/tags?page_size=256'.format(repo=repo)
    docker_response = requests.get(api_url).json()
    if 'results' in docker_response:
        for item in docker_response['results']:
            response[item['name']] = parse_docker_timestamp(item['last_updated'])
    return response


def get_tags_for_docker_v2_registry(registry: str, repo: str) -> Dict[str, int]:
    response = {}
    api_url = 'https://{registry}/v2/{repo}/tags/list'.format(registry=registry, repo=repo)
    registry_response = requests.get(api_url).json()
    if 'tags' in registry_response:
        for tag in registry_response['tags']:
            mf_url = 'https://{registry}/v2/{repo}/manifests/{tag}'.format(registry=registry, repo=repo, tag=tag)
            mf_response = requests.get(mf_url).json()
            if 'history' in mf_response and len(mf_response['history']) > 0:
                v1_compat = json.loads(mf_response['history'][0]['v1Compatibility'])
                response[tag] = parse_docker_timestamp(v1_compat['created'])
    return response


def get_tags_for_image(full_repo: str) -> Dict[str, int]:
    registry, repo = parse_image_repo(full_repo)
    if registry.endswith('gcr.io'):
        tags = get_tags_for_gcr_image(registry, repo)
    elif registry == 'docker.io':
        tags = get_tags_for_docker_hub_image(repo)
    else:
        tags = get_tags_for_docker_v2_registry(registry, repo)
    return tags


def filter_semver_tags(tags: List[str]) -> List[str]:
    """
    Filter a list of tags to only the tags that follow a semantic versioning scheme.
    :param tags:  input tags
    :return:  those input tags that are semver
    """
    output = []
    for tag in tags:
        if semantic_version.validate(semver_normalize(tag)):
            output.append(tag)
    return output


def get_newest_matching_tag(tag: str, tags: Dict[str, int], track: str, tag_timestamp: int = 0):
    if track == 'Newest':
        # filter out tags that are not newer than the current one, and ignore "latest"
        tags = {k: v for k, v in tags.items() if k != 'latest' and v > tag_timestamp}
        sorted_tags = sorted(tags, key=tags.__getitem__, reverse=True)  # sort by descending timestamp
        return sorted_tags[0] if len(sorted_tags) > 0 else None
    # If not "Newest", it is one of the semver variants
    candidates = filter_semver_tags(list(tags.keys()))  # filter out non-semver tags
    versions = {}
    for v in candidates:
        parsed = semver_parse(v)
        versions[parsed] = v
    current = semver_parse(tag)
    norm_tag = semver_normalize(tag)
    if track == 'PatchLevel':
        spec = semantic_version.Spec('>{current},<{next_minor}'.format(current=norm_tag,
                                                                       next_minor=str(current.next_minor())))
    elif track == 'MinorVersion':
        spec = semantic_version.Spec('>{current},<{next_major}'.format(current=norm_tag,
                                                                       next_major=str(current.next_major())))
    elif track == 'MajorVersion':
        spec = semantic_version.Spec('>{current}'.format(current=norm_tag))
    else:
        raise ValueError('unsupported "track": {track}'.format(track=track))
    best = spec.select(versions.keys())
    if best is None:
        return None
    return versions[best]


def find_updates_for_env(environment: Environment):
    updates = defaultdict(list)
    for release in environment.all_releases:
        if release.trigger and release.trigger.image:
            tag_value = release.trigger.image.tagValue
            repo_value = release.trigger.image.repoValue
            prefix_value = release.trigger.image.repoPrefixValue
            track = release.trigger.image.track
            values = release.get_resolved_values(for_env=environment)
            # print('found trigger for image "{image}" from value "{value}": {trigger}\nvalues: {values}'.format(
            #     image=lookup_value(repo_value, values),
            #     trigger=release.trigger,
            #     value=tag_value,
            #     values=values,
            # ))
            if key_is_in_values(repo_value, values):
                image_repo = lookup_value(repo_value, values)
                if key_is_in_values(prefix_value, values):
                    image_repo = lookup_value(prefix_value, values) + image_repo
                image_tag = lookup_value(tag_value, values)
                all_tags = get_tags_for_image(image_repo)
                tag_timestamp = all_tags[image_tag] if image_tag in all_tags else 0
                updated_tag = get_newest_matching_tag(image_tag, all_tags, track, tag_timestamp)
                if updated_tag is not None:
                    updates[release.from_file].append({
                        'old_tag': image_tag,
                        'new_tag': updated_tag,
                        'release': release.name,
                        'tag_value': tag_value,
                        'image_repo': image_repo,
                    })
    return updates
