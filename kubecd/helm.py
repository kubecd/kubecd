import re
import subprocess
from os import path
from typing import List, Union

import ruamel.yaml

from . import model
from .gen_py import ttypes
from .utils import resolve_file_path

import logging

logger = logging.getLogger(__name__)


def inspect(chart_reference: str, chart_version: str) -> str:
    cmd = ['helm', 'inspect', chart_reference, '--version', chart_version]
    output = subprocess.check_output(cmd).decode('utf-8').split("\n---\n")[1]
    return output


def deploy_commands(env: model.Environment, dry_run=False, debug=False, limit_to_release=None) -> List[List[str]]:
    commands = []
    if limit_to_release is None:
        for resource_file in env.all_resource_files:
            cmd = ['kubectl', '--context', 'env:{}'.format(env.name), 'apply']
            if dry_run:
                cmd.append('--dry-run')
            cmd.extend(['-f', resource_file])
            commands.append(cmd)
    for release in env._all_releases:
        if not limit_to_release or release.name == limit_to_release:
            rel_file = release.from_file
            commands.append(generate_helm_install_argv(
                release, env, release_file=rel_file, dry_run=dry_run, debug=debug))
    return commands


def generate_helm_base_argv(env: model.Environment) -> List[str]:
    return ['helm', '--kube-context', 'env:{}'.format(env.name)]


def generate_helm_values_argv(rel: model.Release, env: model.Environment, release_file: str) -> List[str]:
    argv = []
    if not rel.skipDefaultValues:
        if env.defaultValuesFile:
            def_val_file = resolve_file_path(env.defaultValuesFile, release_file)
            argv.extend(['--values', def_val_file])
        if env.defaultValues:
            argv.append('--set')
            argv.append(','.join(['='.join(resolve_value(x, env)) for x in env.defaultValues]))
    if rel.valuesFile:
        val_file = resolve_file_path(rel.valuesFile, release_file)
        argv.extend(['--values', val_file])
    if rel.values:
        argv.append('--set')
        argv.append(','.join(['='.join(resolve_value(x, env)) for x in rel.values]))
    return argv


def generate_helm_chart_arg(rel: model.Release, release_file: str) -> str:
    chart_arg = rel.chart.reference
    if chart_arg is None:
        chart_arg = resolve_file_path(rel.chart.dir, release_file)
        if not path.exists(chart_arg):
            raise ValueError('{}: release "{}" chart.dir "{}" does not exist'.format(
                release_file, rel.name, chart_arg))
    return chart_arg


def generate_helm_diff_argv(rel: model.Release, env: model.Environment, release_file: str) -> List[str]:
    argv = generate_helm_base_argv(env)
    argv.extend(['diff', 'upgrade', rel.name, generate_helm_chart_arg(rel, release_file)])
    argv.extend(generate_helm_values_argv(rel, env, release_file))
    return argv


def generate_helm_install_argv(rel: model.Release,
                               env: model.Environment,
                               release_file: str,
                               dry_run: bool = False,
                               debug: bool = False) -> List[str]:
    chart_arg = generate_helm_chart_arg(rel, release_file)
    argv = generate_helm_base_argv(env)
    argv.extend(['upgrade', rel.name, chart_arg, '-i', '--namespace', env.kubeNamespace])
    argv.extend(generate_helm_values_argv(rel, env, release_file))
    if dry_run:
        argv.append('--dry-run')
    if debug:
        argv.append('--debug')
    return argv


def resolve_gce_address_value(address: ttypes.GceAddressValueRef, env: model.Environment):
    cmd = ['gcloud', 'compute', 'addresses', 'describe', address.name, '--format', 'value(address)']
    provider = env.cluster.provider
    cmd.extend(['--project', provider.gke.project])
    if address.isGlobal:
        cmd.append('--global')
    else:
        cmd.extend(['--region', re.sub(r'-[a-z]$', '', provider.gke.zone)])
    return subprocess.check_output(cmd).decode('utf-8').strip()


def resolve_value(value: ttypes.ChartValue, env: model.Environment, skip_value_from=False) -> tuple:
    v = value.value
    if value.valueFrom and not skip_value_from:
        if value.valueFrom.gceResource:
            if value.valueFrom.gceResource.address:
                v = resolve_gce_address_value(value.valueFrom.gceResource.address, env)
    return value.key, v


def merge_values(from_dict: dict, onto_dict: dict) -> dict:
    for key, value in from_dict.items():
        if isinstance(value, dict):
            # get node or create one
            node = onto_dict.setdefault(key, {})
            if isinstance(node, dict):
                merge_values(value, node)
        else:
            onto_dict[key] = value
    return onto_dict


def load_values_file(file_name: str) -> dict:
    with open(file_name) as f:
        return ruamel.yaml.safe_load(f)


def values_list_to_dict(values: List[ttypes.ChartValue], env: model.Environment, skip_value_from=False) -> dict:
    """
    Convert a list of ChartValue objects with keys on the form ``"foo.bar"``
    to a nested dictionary like ``{"foo": {"bar": ...}}``.
    :param values: list of ChartValue objects
    :param env: current environment
    :param skip_value_from: whether to skip resolving "valueFrom" entries (which will invoke
                            commands or API calls)
    :return:
    """
    def val_to_dict(key: List[str], value) -> dict:
        if len(key) == 1:
            return {key[0]: value}
        return {key[0]: val_to_dict(key[1:], value)}

    output = {}
    for value_obj in values:
        k, v = resolve_value(value_obj, env, skip_value_from=skip_value_from)
        output = merge_values(from_dict=val_to_dict(k.split('.'), v), onto_dict=output)
    return output


def lookup_value(key: Union[List[str], str], values: dict):
    if type(key) == str:
        key = key.split('.')
    if len(key) > 0 and len(values) > 0:
        if key[0] in values:
            if len(key) == 1:
                return values[key[0]]
            return lookup_value(key[1:], values[key[0]])
    return None


def key_is_in_values(key: Union[List[str], str], values: dict) -> bool:
    if type(key) == str:
        key = key.split('.')
    if len(key) > 0 and len(values) > 0:
        if len(key) == 1:
            return key[0] in values
        if key[0] in values:
            return key_is_in_values(key[1:], values[key[0]])
    return False


def get_resolved_values(release: model.Release, for_env: Union[model.Environment, None], skip_value_from=False) -> dict:
    # 1. get values from chart if in a local dir
    # 2. merge with valuesFile:
    # 3. merge with values:
    values = {}
    if release.chart.dir:
        values_file = resolve_file_path(path.join(release.chart.dir, 'values.yaml'), relative_to_file=release._from_file)
        if path.exists(values_file):
            values = merge_values(from_dict=load_values_file(values_file), onto_dict=values)
    elif release.chart.reference:
        # "helm inspect" outputs a two-document yaml block, where the second is the parsed default values.yaml
        output = inspect(release.chart.reference, release.chart.version)
        chart_default_values = ruamel.yaml.safe_load(output)
        values = merge_values(from_dict=chart_default_values, onto_dict=values)
    if for_env is not None and len(for_env.defaultValues) > 0:
        default_values = values_list_to_dict(for_env.defaultValues, for_env, skip_value_from=skip_value_from)
        values = merge_values(from_dict=default_values, onto_dict=values)
    if release.valuesFile:
        values_file = resolve_file_path(release.valuesFile, relative_to_file=release._from_file)
        values = merge_values(from_dict=load_values_file(values_file), onto_dict=values)
    if for_env is not None and release.values:
        values = merge_values(from_dict=values_list_to_dict(release.values, for_env, skip_value_from=skip_value_from),
                              onto_dict=values)
    return values
