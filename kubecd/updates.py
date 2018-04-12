import json
import re
import subprocess
import time
from collections import defaultdict
from typing import Tuple, Dict, List

import requests
import semver

from .environments import Environment, key_is_in_values, lookup_value


def parse_docker_timestamp(timestamp: str) -> int:
    """
    Parse a timestamp of the format "2015-02-13T10:04:55.62062787Z" and
    return unix time
    :param timestamp:
    :return: unix timestamp (seconds since 1970-01-01 00:00:00 UTC)
    """
    return int(time.mktime(time.strptime(timestamp[:19], '%Y-%m-%dT%H:%M:%S')))


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
            datetime = re.sub(r'\+(\d\d):(\d\d)$', r'+\1\2', img_tag['timestamp']['datetime'])
            timestamp = int(time.mktime(time.strptime(datetime, '%Y-%m-%d %H:%M:%S%z')))
        else:
            timestamp = 0
        if 'tags' in img_tag:
            print(img_tag)
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
        try:
            semver.parse(tag)
            output.append(tag)
        except ValueError:
            pass
    return output


def filter_candidate_tags(tag: str, tags: Dict[str, int], track: str):
    if track == 'Newest':
        return tags.keys()
    candidates = filter_semver_tags(list(tags.keys()))
    if track == 'PatchLevel':
        ver_le = '<' + semver.bump_minor(tag)
        return filter(lambda x: semver.match(x, ver_le), candidates)
    if track == 'MinorVersion':
        ver_le = '<' + semver.bump_major(tag)
        return filter(lambda x: semver.match(x, ver_le), candidates)
    if track == 'MajorVersion':
        return candidates
    raise ValueError('unsupported "track": {track}'.format(track=track))


def find_updates_for_env(environment: Environment):
    updates = defaultdict(list)
    for release in environment.all_releases:
        if release.trigger and release.trigger.image:
            tag_value = release.trigger.image.tagValue
            repo_value = release.trigger.image.repoValue
            prefix_value = release.trigger.image.repoPrefixValue
            track = release.trigger.image.track
            values = release.get_resolved_values(for_env=environment)
            if key_is_in_values(repo_value, values):
                image_repo = lookup_value(repo_value, values)
                if key_is_in_values(prefix_value, values):
                    image_repo = lookup_value(prefix_value, values) + image_repo
                image_tag = lookup_value(tag_value, values)
                all_tags = get_tags_for_image(image_repo)
                tag_timestamp = all_tags[image_tag]
                updated_tag = None
                for candidate in filter_candidate_tags(image_tag, all_tags, track):
                    if all_tags[candidate] > tag_timestamp:
                        tag_timestamp = all_tags[candidate]
                        updated_tag = candidate
                if updated_tag is not None:
                    updates[release.from_file].append({
                        'old_tag': image_tag,
                        'new_tag': updated_tag,
                        'release': release.name,
                        'tag_value': tag_value,
                        'image_repo': image_repo,
                    })
    return updates
