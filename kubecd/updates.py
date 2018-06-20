import json
import subprocess
from collections import defaultdict, OrderedDict
from typing import Tuple, Dict, List
import dateutil.parser
import logging

import requests
import semantic_version

from . import semver
from .model import Environment, Release
from .helm import key_is_in_values, lookup_value, get_resolved_values

logger = logging.getLogger(__name__)


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


def get_tags_for_gcr_image(registry: str, repo: str) -> OrderedDict:
    full_repo = registry + '/' + repo
    gcp_project = repo.split('/')[0]
    cmd = ['gcloud', 'container', 'images', 'list-tags', full_repo, '--project', gcp_project, '--format', 'json']
    gcr_response = json.loads(subprocess.check_output(cmd).decode('utf-8'))
    response = OrderedDict()
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


def get_tags_for_docker_hub_image(repo: str) -> OrderedDict:
    response = OrderedDict()
    if '/' not in repo:
        repo = 'library/' + repo
    api_url = 'https://registry.hub.docker.com/v2/repositories/{repo}/tags?page_size=256'.format(repo=repo)
    docker_response = requests.get(api_url).json()
    if 'results' in docker_response:
        for item in docker_response['results']:
            response[item['name']] = parse_docker_timestamp(item['last_updated'])
    return response


def get_tags_for_docker_v2_registry(registry: str, repo: str) -> OrderedDict:
    response = OrderedDict()
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


def get_tags_for_image(full_repo: str) -> OrderedDict:
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
        if semantic_version.validate(semver.normalize(tag)):
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
        parsed = semver.parse(v)
        versions[parsed] = v
    current = semver.parse(tag)
    best = semver.best_upgrade(current, list(versions.keys()), track)
    if best is None:
        return None
    return versions[best]


class ImageUpdate(object):
    """
    An object representing an image with an available update. Just type safety convenience.
    """
    def __init__(self, new_tag: str, tag_value: str, release: Release, old_tag: str=None, image_repo: str=None):
        self.old_tag = old_tag
        self.new_tag = new_tag
        self.release = release
        self.tag_value = tag_value
        self.image_repo = image_repo


def release_wants_tag_update(release: Release, new_tag: str) -> List[ImageUpdate]:
    updates = []
    for trigger in release.triggers:
        if trigger.image is None or trigger.image.tagValue is None:
            continue
        values = get_resolved_values(release, for_env=None, skip_value_from=True)
        tag_value = trigger.image.tagValue
        current_tag = lookup_value(tag_value, values)
        # if the current version is not semver, consider any value to be an update
        if not semver.is_semver(current_tag):
            updates.append(ImageUpdate(new_tag=new_tag, tag_value=tag_value, release=release))
            continue
        # if the new version is not semver, consider it an update only if track == Newest
        if not semver.is_semver(new_tag) and trigger.image.track == 'Newest':
            updates.append(ImageUpdate(new_tag=new_tag, tag_value=tag_value, release=release))
            continue
        if current_tag is not None and semver.is_wanted_upgrade(semver.parse(current_tag),
                                                                semver.parse(new_tag),
                                                                trigger.image.track):
            updates.append(ImageUpdate(new_tag=new_tag, tag_value=tag_value, release=release))
    return updates


def find_updates_for_release(release: Release, environment: Environment) -> Dict[str, List[ImageUpdate]]:
    updates = defaultdict(list)
    if release.triggers is None:
        return updates
    for trigger in release.triggers:
        if trigger.image is None:
            continue
        tag_value = trigger.image.tagValue
        repo_value = trigger.image.repoValue
        prefix_value = trigger.image.repoPrefixValue
        track = trigger.image.track
        values = get_resolved_values(release, for_env=environment, skip_value_from=True)
        logger.debug('found trigger for image "%s" from value "%s": %s final values: %s',
                     lookup_value(repo_value, values),
                     tag_value,
                     json.dumps(trigger.image.__dict__),
                     json.dumps(values))
        if key_is_in_values(repo_value, values):
            image_repo = lookup_value(repo_value, values)
            if key_is_in_values(prefix_value, values):
                image_repo = lookup_value(prefix_value, values) + image_repo
            image_tag = lookup_value(tag_value, values)
            all_tags = get_tags_for_image(image_repo)
            tag_timestamp = all_tags[image_tag] if image_tag in all_tags else 0
            updated_tag = get_newest_matching_tag(image_tag, all_tags, track, tag_timestamp)
            if updated_tag is not None:
                logger.debug('found update for "%s": "%s" -> "%s"', image_repo, image_tag, updated_tag)
                updates[release.from_file].append(ImageUpdate(old_tag=image_tag,
                                                              new_tag=updated_tag,
                                                              release=release,
                                                              tag_value=tag_value,
                                                              image_repo=image_repo))
    logger.debug('find_updates_for_release: returning: %r', updates)
    return updates


def find_updates_for_releases(releases: List[Release], environment: Environment) -> Dict[str, List[ImageUpdate]]:
    env_updates = defaultdict(list)
    for release in releases:
        logger.info('checking updates for release: {env}/{release}'.format(env=environment.name, release=release.name))
        image_updates = find_updates_for_release(release, environment)
        for file, updates in image_updates.items():
            env_updates[file].extend(updates)
    logger.debug('find_updates_for_env: returning: %r', env_updates)
    return env_updates


def find_updates_for_env(environment: Environment) -> Dict[str, List[ImageUpdate]]:
    env_updates = defaultdict(list)
    for release in environment.all_releases:
        logger.info('checking updates for release: {env}/{release}'.format(env=environment.name, release=release.name))
        image_updates = find_updates_for_release(release, environment)
        for file, updates in image_updates.items():
            env_updates[file].extend(updates)
    logger.debug('find_updates_for_env: returning: %r', env_updates)
    return env_updates
