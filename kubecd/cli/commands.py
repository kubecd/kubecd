import os
import subprocess
import sys
from typing import List, Dict

import logging
import ruamel.yaml
import ruamel.yaml.error
from blessings import Terminal
from ruamel.yaml import YAMLError

from .. import helm, updates, github
from .. import model
from ..inspection import validate_environment
from ..providers import get_cluster_provider
from ..thriftutils import SchemaError
from ..updates import find_updates_for_env, find_updates_for_releases

t = Terminal()
logger = logging.getLogger(__name__)


class CliError(Exception):
    pass


def lint_environment(environments_file: str, cluster: str=None, env_name: str=None, **kwargs):
    if not cluster and not env_name:
        raise CliError('please specify either --cluster or ENV')
    env_map = cluster_env_map(environments_file, cluster, env_name)
    errors = []
    for c, el in env_map.items():
        for env in el:
            errors.extend(validate_environment(env))
    for error in errors:
        print('{t.red}ERROR{t.normal}: {msg}'.format(msg=str(error), t=t), file=sys.stderr)

    if len(errors) > 0:
        raise CliError('validation failed')


def cluster_env_map(environments_file: str,
                    cluster: str=None,
                    env_name: str=None) -> Dict[model.Cluster, List[model.Environment]]:
    kcd_model = _load_model(environments_file)
    if cluster:
        try:
            return {kcd_model.get_cluster(cluster): kcd_model.environments_in_cluster(cluster)}
        except KeyError as e:
            raise CliError('no such cluster: {}'.format(e))
    try:
        e = kcd_model.get_environment(env_name)
        try:
            return {kcd_model.get_cluster(e.clusterName): [e]}
        except KeyError as e:
            raise CliError('no such cluster: {}'.format(e))
    except KeyError as e:
        raise CliError('no such environment: {}'.format(e))


def use_env_context(environments_file: str, env: str, **kwargs):
    kcd_model = _load_model(environments_file)
    try:
        kcd_model.get_environment(env)
    except KeyError as e:
        raise CliError('no such environment: {}'.format(e))
    run_command(helm.use_context_command(env))


def init_contexts(environments_file: str, cluster: str, env_name: str, contexts_only: bool=False, gitlab=False, dry_run=False, **kwargs):
    if not cluster and not env_name:
        raise CliError('please specify either --cluster or ENV')
    env_map = cluster_env_map(environments_file, cluster, env_name)
    commands_to_run = []
    commands_to_run.extend(model.config().init_commands())
    for c, el in env_map.items():
        cp = get_cluster_provider(c, gitlab)
        if not contexts_only and not gitlab:
            commands_to_run.extend(cp.cluster_init_commands())
        for e in el:
            commands_to_run.extend(cp.context_init_commands(e))
    run_commands(commands_to_run, dry_run=dry_run)


def run_command(cmd: List[str], dry_run=False):
    if dry_run:
        print(' '.join(cmd))
        return
    try:
        logger.debug('Executing: "%s"', ' '.join(cmd))
        return subprocess.run(cmd, check=True)
    except subprocess.CalledProcessError as e:
        raise CliError(e)


def run_commands(cmds: List[List[str]], dry_run=False):
    for cmd in cmds:
        run_command(cmd, dry_run)


def json2yaml(**kwargs):
    import json
    from ruamel import yaml
    obj = json.load(sys.stdin)
    yaml.safe_dump(obj, sys.stdout)


def diff_release(environments_file: str, releases: List[str], env: str, **kwargs):
    kcd_config = _load_model(environments_file)
    e = kcd_config.get_environment(env)
    for release in releases:
        r = e.named_release(release)
        cmd = helm.generate_helm_diff_argv(r, e, r.from_file)
        run_command(cmd)


def observe_new_image(environments_file: str, image: str, patch: bool, submit_pr: bool, **kwargs):
    kcd_config = _load_model(environments_file)
    image_repo, image_tag = image.split(':')
    image_index = kcd_config.image_index
    patched_files = []
    if image_repo in image_index:
        for release in image_index[image_repo]:
            update_list = updates.release_wants_tag_update(release, image_tag)
            for update in update_list:
                print('release: {release} tagValue: {tag}'.format(release=update.release.name, tag=update.tag_value))
            if len(update_list) > 0 and patch:
                patch_releases_file(release.from_file, update_list)
                patched_files.append(release.from_file)
    if submit_pr and len(patched_files) > 0:
        github.pr_from_files(patched_files, 'Observed image {image}'.format(image=image))


def apply_env(environments_file: str,
              init: bool=False,
              cluster: str=None,
              dry_run: bool=False,
              debug: bool=False,
              env_name: str=None,
              releases=None,
              **kwargs):
    commands_to_run = []
    for cluster, envs in cluster_env_map(environments_file, cluster, env_name).items():
        cp = get_cluster_provider(cluster)
        if init:
            commands_to_run.extend(cp.cluster_init_commands())
        for environment in envs:
            logger.info('Collecting commands for environment "%s"', environment.name)
            try:
                deploy_cmds = helm.deploy_commands(environment,
                                                   dry_run=dry_run,
                                                   limit_to_releases=releases,
                                                   debug=debug)
            except ValueError as e:
                raise CliError(str(e))
            if init and len(deploy_cmds) > 0 and init:
                commands_to_run.extend(cp.context_init_commands(environment))
            commands_to_run.extend(deploy_cmds)
    for cmd in commands_to_run:
        print('{t.yellow}{cmd}{t.normal}'.format(cmd=' '.join(cmd), t=t))
        run_command(cmd)


def dump_env(environments_file, cluster: str=None, env=None, **kwargs):
    target_envs = resolve_envs(env, cluster, file_name=environments_file)
    for cmd in model.config().init_commands():
        print('{t.yellow}{cmd}{t.normal}'.format(cmd=' '.join(cmd), t=t))
    for environment in target_envs:
        print('{t.green}Environment:{t.normal} {env_name}'.format(env_name=environment.name, t=t))
        for cmd in environment.init_commands():
            print('{t.yellow}{cmd}{t.normal}'.format(cmd=' '.join(cmd), t=t))
        for cmd in helm.deploy_commands(environment, dry_run=False):
            print(' '.join(cmd))
        print('')


def list_kind(environments_file, kind, **kwargs):
    kcd_model = _load_model(environments_file)
    if kind == 'env' or kind == 'environment':
        for environment in kcd_model:
            print(environment.name)
    elif kind == 'release' or kind == 'releases' or kind == 'rel':
        for environment in kcd_model:
            for release in environment.all_releases:
                print('{env} -r {release}'.format(env=environment.name, release=release.name))
    elif kind == 'cluster' or kind == 'clusters':
        for cluster in kcd_model.all_clusters():
            print('{cluster.name}'.format(cluster=cluster))
    else:
        raise CliError('unknown kind "{}"'.format(kind))


def poll_registries(environments_file, cluster=None, env=None, releases=None, patch=False, **kwargs):
    target_envs = resolve_envs(env, cluster, file_name=environments_file)
    for te in target_envs:
        if releases is None:
            logger.info('polling environment: "%s"', te.name)
            file_updates = find_updates_for_env(te)
        else:
            release_objs = [te.named_release(r) for r in releases]
            logger.info('polling env:%s releases: %s', te.name, ' '.join(releases))
            file_updates = find_updates_for_releases(release_objs, te)
        for release_file, update_list in file_updates.items():
            for update in update_list:
                print(
                    '{env}/{release}:\n\tfile: {file}\n\timage: {image}\n\ttag: {tag_from} -> {tag_to}'.format(
                        env=te.name,
                        file=release_file,
                        release=update.release.name,
                        image=update.image_repo,
                        tag_from=update.old_tag,
                        tag_to=update.new_tag,
                    ))
            if patch:
                patch_releases_file(release_file, update_list)


def patch_releases_file(releases_file: str, updates_list: List[updates.ImageUpdate]):
    logger.debug('loading releases file: "%s"', releases_file)
    mod_yaml = _load_yaml(releases_file)
    for update in updates_list:
        for mod_rel in mod_yaml['releases']:
            if mod_rel['name'] == update.release.name:
                if 'values' not in mod_rel:
                    mod_rel['values'] = []
                found_val = False
                for yv in mod_rel['values']:
                    if yv['key'] == update.tag_value:
                        logger.debug('patching value "{}"'.format(update.tag_value))
                        yv['value'] = update.new_tag
                        found_val = True
                        break
                if not found_val:
                    mod_rel['values'].append({'key': update.tag_value, 'value': update.new_tag})
    logger.debug('saving patched file: {file}'.format(file=releases_file))
    _save_yaml(mod_yaml, releases_file)


def indent_file(files, **kwargs):
    for filename in files:
        try:
            _save_yaml(_load_yaml(filename), filename)
        except YAMLError as e:
            raise CliError('invalid YAML file "{}": {}'.format(filename, str(e)))


def resolve_envs(env_name: str, cluster: str, file_name: str) -> List[model.Environment]:
    if not env_name and not cluster:
        raise CliError('please specify either --cluster or an environment name')
    kcd_model = _load_model(file_name)
    if env_name is not None:
        try:
            return [kcd_model.get_environment(env_name)]
        except KeyError:
            raise CliError('no such environment: {}'.format(env_name))
    if kcd_model.has_cluster(cluster):
        return kcd_model.environments_in_cluster(cluster)
    raise CliError('no such cluster: {}'.format(cluster))


def _load_model(file_name: str):
    try:
        logger.debug('loading model from file %s', file_name)
        return model.load(file_name)
    except FileNotFoundError as e:
        raise CliError('file not found: {}'.format(e.filename))
    except SchemaError as e:
        raise CliError(e)


def _load_yaml(file_name: str):
    yaml = ruamel.yaml.YAML(typ='rt')
    yaml.indent(sequence=4, offset=2)
    yaml.preserve_quotes = True
    with open(file_name, 'r') as rf:
        return yaml.load(rf)


def _save_yaml(mod_yaml, file_name: str):
    yaml = ruamel.yaml.YAML(typ='rt')
    yaml.indent(sequence=4, offset=2)
    yaml.preserve_quotes = True
    tmp_file = file_name + '.new'
    with open(tmp_file, 'w') as wf:
        yaml.dump(mod_yaml, wf)
    os.rename(tmp_file, file_name)
