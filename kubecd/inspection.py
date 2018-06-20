import logging
import subprocess
from copy import copy
from os import path
from tempfile import TemporaryDirectory
from typing import List, TextIO

from ruamel.yaml import YAML

from .helm import generate_helm_template_argv
from .model import Environment, Release
from .utils import resolve_file_path


class ValidationError(Exception):
    pass


def load_yaml(data: TextIO):
    yaml = YAML(typ='safe')
    return yaml.load_all(data)


def run_command(cmd: List[str], dry_run=False):
    logging.debug('Executing: "%s"', ' '.join(cmd))
    return subprocess.run(cmd, check=True)


def expand_release(rel: Release, env: Environment):
    if rel.resourceFiles:
        return expand_resource_files(rel.resourceFiles, rel.from_file)
    if rel.chart:
        if rel.chart.dir:
            return expand_helm_chart_dir(rel, env)
        elif rel.chart.reference:
            with TemporaryDirectory() as tmpdir:
                rel.chart = copy(rel.chart)
                rel.chart.dir = path.join(tmpdir, path.basename(rel.chart.reference))
                subprocess.run([
                    'helm', 'fetch', '--untar', '--untardir', tmpdir,
                    rel.chart.reference, '--version', rel.chart.version
                ])
                rel.chart.reference = None
                rel.chart.version = None
                return expand_helm_chart_dir(rel, env)


def expand_helm_chart_dir(rel: Release, env: Environment) -> List[dict]:
    cmd = generate_helm_template_argv(rel, env, rel.from_file)
    logging.debug('Running %r', cmd)
    output = subprocess.check_output(cmd).decode('utf-8')
    resources = load_yaml(output)
    if isinstance(resources, dict):
        resources = [resources]
    return resources


def expand_resource_files(file_names: List[str], from_file: str) -> List[dict]:
    resources = []
    for file_name in file_names:
        file_name = resolve_file_path(file_name, relative_to_file=from_file)
        with open(file_name, 'r') as f:
            tmp = load_yaml(f)
            if isinstance(tmp, dict):
                resources.append(tmp)
            else:
                resources.extend(tmp)
    return resources


def index_resource_list(resources: List[dict]):
    index = {}
    for obj in resources:
        key = '{kind}.{name}'.format(kind=obj['kind'], name=obj['metadata']['name'])
        index[key] = obj
    return index


def find_ingress_conflicts(resources: List[dict]) -> List[ValidationError]:
    errors = []
    path_owner = {}
    for obj in resources:
        if obj is None:
            continue
        logging.debug('checking obj %r', obj)
        if obj['kind'] == 'Ingress':
            ingress_name = obj['metadata']['name']
            for rule in obj['spec']['rules']:
                for http_path in rule['http']['paths']:
                    host_path = rule['host'] + (http_path['path'] if 'path' in http_path else '/')
                    if host_path in path_owner:
                        errors.append(ValidationError('Ingress host+path "{host_path}" defined in both {a} and {b}'.format(
                            host_path=host_path, a=path_owner[host_path], b=ingress_name
                        )))
                        continue
                    path_owner[host_path] = ingress_name
    return errors


def validate_environment(env: Environment) -> List[ValidationError]:
    resources = []
    errors = []
    for release in env.all_releases:
        logging.info('expanding release: %s', release.name)
        resources.extend(expand_release(release, env))
    logging.info('checking for: ingress conflicts')
    errors.extend(find_ingress_conflicts(resources))
    return errors
